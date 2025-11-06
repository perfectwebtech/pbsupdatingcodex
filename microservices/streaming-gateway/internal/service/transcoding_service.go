package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrJobNotFound      = errors.New("transcoding job not found")
	ErrInvalidStatus    = errors.New("invalid job status")
	ErrMaxRetriesReached = errors.New("maximum retries reached")
)

// TranscodingService implements the transcoding business logic
type TranscodingService struct {
	db              *sql.DB
	ffmpegPath      string
	workers         map[string]*Worker
	workerMutex     sync.RWMutex
	maxWorkers      int
	hardwareAccel   bool
	qualityPresets  map[string]QualityPreset
}

// QualityPreset defines encoding parameters for each quality level
type QualityPreset struct {
	Resolution string
	VideoBitrate int // kbps
	AudioBitrate int // kbps
	Encoder string
	Preset string
	ExtraParams string
}

// Worker represents a transcoding worker
type Worker struct {
	ID string
	Hostname string
	Status string
	CurrentJobs int
	MaxJobs int
	CPUUsage float64
	MemoryUsage float64
	LastHeartbeat time.Time
}

// NewTranscodingService creates a new transcoding service
func NewTranscodingService(db *sql.DB, ffmpegPath string) *TranscodingService {
	service := &TranscodingService{
		db:            db,
		ffmpegPath:    ffmpegPath,
		workers:       make(map[string]*Worker),
		maxWorkers:    4,
		hardwareAccel: false,
		qualityPresets: map[string]QualityPreset{
			"sd": {
				Resolution:   "640x360",
				VideoBitrate: 800,
				AudioBitrate: 96,
				Encoder:      "libx264",
				Preset:       "fast",
				ExtraParams:  "-crf 23",
			},
			"hd": {
				Resolution:   "1280x720",
				VideoBitrate: 2500,
				AudioBitrate: 128,
				Encoder:      "libx264",
				Preset:       "medium",
				ExtraParams:  "-crf 22",
			},
			"fhd": {
				Resolution:   "1920x1080",
				VideoBitrate: 5000,
				AudioBitrate: 192,
				Encoder:      "libx264",
				Preset:       "medium",
				ExtraParams:  "-crf 21",
			},
			"uhd": {
				Resolution:   "3840x2160",
				VideoBitrate: 15000,
				AudioBitrate: 256,
				Encoder:      "libx264",
				Preset:       "slow",
				ExtraParams:  "-crf 20",
			},
		},
	}

	// Start background worker management
	go service.manageWorkers()
	go service.processQueue()

	return service
}

// GetAllJobs retrieves all transcoding jobs with filters
func (s *TranscodingService) GetAllJobs(filters JobFilters) ([]TranscodingJob, error) {
	query := `
		SELECT id, job_name, job_type, status, priority, stream_id,
		       source_url, source_format, source_codec, source_bitrate, source_resolution, source_duration,
		       target_url, target_format, target_quality, target_preset, target_resolution, target_bitrate, target_codec,
		       ffmpeg_command, hardware_acceleration, encoder, audio_codec, audio_bitrate, segment_duration, extra_params,
		       progress, current_frame, total_frames, current_time, fps, bitrate, speed,
		       worker_id, worker_hostname, pid,
		       started_at, completed_at, failed_at, estimated_completion,
		       error_message, error_code, retry_count, max_retries,
		       output_size, output_duration, metadata,
		       created_at, updated_at
		FROM transcoding_jobs
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.JobType != "" {
		query += fmt.Sprintf(" AND job_type = $%d", argIndex)
		args = append(args, filters.JobType)
		argIndex++
	}

	if filters.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, *filters.Priority)
		argIndex++
	}

	if filters.StreamID != nil {
		query += fmt.Sprintf(" AND stream_id = $%d", argIndex)
		args = append(args, *filters.StreamID)
		argIndex++
	}

	query += " ORDER BY priority ASC, created_at ASC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []TranscodingJob
	for rows.Next() {
		var job TranscodingJob
		err := rows.Scan(
			&job.ID, &job.JobName, &job.JobType, &job.Status, &job.Priority, &job.StreamID,
			&job.SourceURL, &job.SourceFormat, &job.SourceCodec, &job.SourceBitrate, &job.SourceResolution, &job.SourceDuration,
			&job.TargetURL, &job.TargetFormat, &job.TargetQuality, &job.TargetPreset, &job.TargetResolution, &job.TargetBitrate, &job.TargetCodec,
			&job.FFmpegCommand, &job.HardwareAcceleration, &job.Encoder, &job.AudioCodec, &job.AudioBitrate, &job.SegmentDuration, &job.ExtraParams,
			&job.Progress, &job.CurrentFrame, &job.TotalFrames, &job.CurrentTime, &job.FPS, &job.Bitrate, &job.Speed,
			&job.WorkerID, &job.WorkerHostname, &job.PID,
			&job.StartedAt, &job.CompletedAt, &job.FailedAt, &job.EstimatedCompletion,
			&job.ErrorMessage, &job.ErrorCode, &job.RetryCount, &job.MaxRetries,
			&job.OutputSize, &job.OutputDuration, &job.Metadata,
			&job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobByID retrieves a single job by ID
func (s *TranscodingService) GetJobByID(id int64) (*TranscodingJob, error) {
	query := `
		SELECT id, job_name, job_type, status, priority, stream_id,
		       source_url, source_format, source_codec, source_bitrate, source_resolution, source_duration,
		       target_url, target_format, target_quality, target_preset, target_resolution, target_bitrate, target_codec,
		       ffmpeg_command, hardware_acceleration, encoder, audio_codec, audio_bitrate, segment_duration, extra_params,
		       progress, current_frame, total_frames, current_time, fps, bitrate, speed,
		       worker_id, worker_hostname, pid,
		       started_at, completed_at, failed_at, estimated_completion,
		       error_message, error_code, retry_count, max_retries,
		       output_size, output_duration, metadata,
		       created_at, updated_at
		FROM transcoding_jobs
		WHERE id = $1
	`

	var job TranscodingJob
	err := s.db.QueryRow(query, id).Scan(
		&job.ID, &job.JobName, &job.JobType, &job.Status, &job.Priority, &job.StreamID,
		&job.SourceURL, &job.SourceFormat, &job.SourceCodec, &job.SourceBitrate, &job.SourceResolution, &job.SourceDuration,
		&job.TargetURL, &job.TargetFormat, &job.TargetQuality, &job.TargetPreset, &job.TargetResolution, &job.TargetBitrate, &job.TargetCodec,
		&job.FFmpegCommand, &job.HardwareAcceleration, &job.Encoder, &job.AudioCodec, &job.AudioBitrate, &job.SegmentDuration, &job.ExtraParams,
		&job.Progress, &job.CurrentFrame, &job.TotalFrames, &job.CurrentTime, &job.FPS, &job.Bitrate, &job.Speed,
		&job.WorkerID, &job.WorkerHostname, &job.PID,
		&job.StartedAt, &job.CompletedAt, &job.FailedAt, &job.EstimatedCompletion,
		&job.ErrorMessage, &job.ErrorCode, &job.RetryCount, &job.MaxRetries,
		&job.OutputSize, &job.OutputDuration, &job.Metadata,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query job: %w", err)
	}

	return &job, nil
}

// CreateJob creates a new transcoding job
func (s *TranscodingService) CreateJob(req *CreateJobRequest) (*TranscodingJob, error) {
	// Get quality preset if not custom
	var targetResolution string
	var targetBitrate int

	if req.TargetQuality != "custom" {
		if preset, ok := s.qualityPresets[req.TargetQuality]; ok {
			targetResolution = preset.Resolution
			targetBitrate = preset.VideoBitrate
			if req.Encoder == "" {
				req.Encoder = preset.Encoder
			}
		}
	} else {
		targetResolution = req.TargetResolution
		targetBitrate = req.TargetBitrate
	}

	// Build FFmpeg command
	ffmpegCmd := s.buildFFmpegCommand(req, targetResolution, targetBitrate)

	query := `
		INSERT INTO transcoding_jobs (
			job_name, job_type, status, priority, stream_id,
			source_url, target_url, target_format, target_quality, target_resolution, target_bitrate, target_codec,
			hardware_acceleration, encoder, audio_codec, audio_bitrate, segment_duration, extra_params,
			ffmpeg_command, retry_count, max_retries
		) VALUES ($1, $2, 'pending', $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, 0, 3)
		RETURNING id, created_at, updated_at
	`

	var job TranscodingJob
	job.JobName = req.JobName
	job.JobType = req.JobType
	job.Priority = req.Priority
	job.StreamID = req.StreamID
	job.SourceURL = req.SourceURL
	job.TargetURL = req.TargetURL
	job.TargetFormat = req.TargetFormat
	job.TargetQuality = req.TargetQuality
	job.TargetResolution = targetResolution
	job.TargetBitrate = targetBitrate
	job.TargetCodec = req.Encoder
	job.HardwareAcceleration = req.HardwareAcceleration
	job.Encoder = req.Encoder
	job.AudioCodec = req.AudioCodec
	job.AudioBitrate = req.AudioBitrate
	job.SegmentDuration = req.SegmentDuration
	job.ExtraParams = req.ExtraParams
	job.FFmpegCommand = ffmpegCmd
	job.Status = "pending"

	err := s.db.QueryRow(query,
		req.JobName, req.JobType, req.Priority, req.StreamID,
		req.SourceURL, req.TargetURL, req.TargetFormat, req.TargetQuality, targetResolution, targetBitrate, req.Encoder,
		req.HardwareAcceleration, req.Encoder, req.AudioCodec, req.AudioBitrate, req.SegmentDuration, req.ExtraParams,
		ffmpegCmd,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Queue the job for processing
	go s.queueJob(job.ID)

	return &job, nil
}

// UpdateJob updates an existing job
func (s *TranscodingService) UpdateJob(id int64, updates *UpdateJobRequest) error {
	query := "UPDATE transcoding_jobs SET "
	args := []interface{}{}
	argIndex := 1

	if updates.Status != "" {
		query += fmt.Sprintf("status = $%d, ", argIndex)
		args = append(args, updates.Status)
		argIndex++
	}

	if updates.Priority != nil {
		query += fmt.Sprintf("priority = $%d, ", argIndex)
		args = append(args, *updates.Priority)
		argIndex++
	}

	query += fmt.Sprintf("updated_at = NOW() WHERE id = $%d", argIndex)
	args = append(args, id)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

// DeleteJob deletes a job
func (s *TranscodingService) DeleteJob(id int64) error {
	query := "DELETE FROM transcoding_jobs WHERE id = $1 AND status NOT IN ('processing')"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("cannot delete job that is currently processing")
	}

	return nil
}

// CancelJob cancels a running or queued job
func (s *TranscodingService) CancelJob(id int64) error {
	job, err := s.GetJobByID(id)
	if err != nil {
		return err
	}

	if job.Status == "completed" || job.Status == "failed" || job.Status == "cancelled" {
		return errors.New("job cannot be cancelled")
	}

	// Kill the process if it's running
	if job.PID > 0 {
		cmd := exec.Command("kill", "-9", fmt.Sprintf("%d", job.PID))
		cmd.Run()
	}

	query := "UPDATE transcoding_jobs SET status = 'cancelled', updated_at = NOW() WHERE id = $1"
	_, err = s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	return nil
}

// RetryJob retries a failed job
func (s *TranscodingService) RetryJob(id int64) error {
	job, err := s.GetJobByID(id)
	if err != nil {
		return err
	}

	if job.Status != "failed" {
		return errors.New("only failed jobs can be retried")
	}

	if job.RetryCount >= job.MaxRetries {
		return ErrMaxRetriesReached
	}

	query := `
		UPDATE transcoding_jobs
		SET status = 'pending', retry_count = retry_count + 1,
		    error_message = NULL, error_code = NULL, failed_at = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err = s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to retry job: %w", err)
	}

	// Queue the job again
	go s.queueJob(id)

	return nil
}

// GetQueueStatus returns the current queue status
func (s *TranscodingService) GetQueueStatus() (*QueueStatus, error) {
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'queued' THEN 1 ELSE 0 END) as queued,
			SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END) as processing,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		FROM transcoding_jobs
		WHERE DATE(created_at) = CURDATE()
	`

	var status QueueStatus
	err := s.db.QueryRow(query).Scan(
		&status.TotalJobs,
		&status.PendingJobs,
		&status.QueuedJobs,
		&status.ProcessingJobs,
		&status.CompletedJobs,
		&status.FailedJobs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	s.workerMutex.RLock()
	status.ActiveWorkers = len(s.workers)
	s.workerMutex.RUnlock()

	status.QueueLength = status.PendingJobs + status.QueuedJobs

	return &status, nil
}

// GetJobProgress returns the progress of a specific job
func (s *TranscodingService) GetJobProgress(id int64) (*JobProgress, error) {
	job, err := s.GetJobByID(id)
	if err != nil {
		return nil, err
	}

	progress := &JobProgress{
		JobID:         job.ID,
		Status:        job.Status,
		Progress:      job.Progress,
		CurrentTime:   job.CurrentTime,
		TotalDuration: job.SourceDuration,
		FPS:           job.FPS,
		Speed:         job.Speed,
		Bitrate:       job.Bitrate,
		EstimatedCompletion: job.EstimatedCompletion,
	}

	// Calculate time remaining
	if job.Speed > 0 && job.SourceDuration > 0 {
		remainingSeconds := float64(job.SourceDuration-job.CurrentTime) / job.Speed
		progress.TimeRemaining = int(remainingSeconds)
	}

	return progress, nil
}

// GetStatistics returns transcoding statistics
func (s *TranscodingService) GetStatistics(days int) (*TranscodingStats, error) {
	query := `
		SELECT
			SUM(total_jobs) as total_jobs,
			SUM(completed_jobs) as completed_jobs,
			SUM(failed_jobs) as failed_jobs,
			SUM(cancelled_jobs) as cancelled_jobs,
			AVG(avg_processing_time) as avg_processing_time,
			SUM(total_processing_time) as total_processing_time,
			AVG(avg_speed) as avg_speed,
			SUM(total_input_size) as total_input_size,
			SUM(total_output_size) as total_output_size,
			AVG(compression_ratio) as compression_ratio,
			SUM(sd_jobs) as sd_jobs,
			SUM(hd_jobs) as hd_jobs,
			SUM(fhd_jobs) as fhd_jobs,
			SUM(uhd_jobs) as uhd_jobs
		FROM transcoding_stats
		WHERE date >= DATE_SUB(CURDATE(), INTERVAL $1 DAY)
	`

	var stats TranscodingStats
	stats.Period = fmt.Sprintf("Last %d days", days)

	err := s.db.QueryRow(query, days).Scan(
		&stats.TotalJobs,
		&stats.CompletedJobs,
		&stats.FailedJobs,
		&stats.CancelledJobs,
		&stats.AvgProcessingTime,
		&stats.TotalProcessingTime,
		&stats.AvgSpeed,
		&stats.TotalInputSize,
		&stats.TotalOutputSize,
		&stats.CompressionRatio,
		&stats.SDJobs,
		&stats.HDJobs,
		&stats.FHDJobs,
		&stats.UHDJobs,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &stats, nil
}

// GetWorkers returns all active workers
func (s *TranscodingService) GetWorkers() ([]Worker, error) {
	s.workerMutex.RLock()
	defer s.workerMutex.RUnlock()

	workers := make([]Worker, 0, len(s.workers))
	for _, worker := range s.workers {
		workers = append(workers, *worker)
	}

	return workers, nil
}

// Private helper methods

func (s *TranscodingService) buildFFmpegCommand(req *CreateJobRequest, resolution string, bitrate int) string {
	parts := []string{s.ffmpegPath, "-i", req.SourceURL}

	// Hardware acceleration
	if req.HardwareAcceleration {
		parts = append(parts, "-hwaccel", "auto")
	}

	// Video codec and settings
	parts = append(parts, "-c:v", req.Encoder)
	parts = append(parts, "-b:v", fmt.Sprintf("%dk", bitrate))
	parts = append(parts, "-s", resolution)

	// Audio codec and settings
	parts = append(parts, "-c:a", req.AudioCodec)
	parts = append(parts, "-b:a", fmt.Sprintf("%dk", req.AudioBitrate))

	// Extra parameters
	if req.ExtraParams != "" {
		parts = append(parts, strings.Split(req.ExtraParams, " ")...)
	}

	// Output
	parts = append(parts, req.TargetURL)

	return strings.Join(parts, " ")
}

func (s *TranscodingService) queueJob(jobID int64) {
	query := "UPDATE transcoding_jobs SET status = 'queued', updated_at = NOW() WHERE id = $1"
	s.db.Exec(query, jobID)
}

func (s *TranscodingService) manageWorkers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.workerMutex.Lock()
		// Remove inactive workers
		for id, worker := range s.workers {
			if time.Since(worker.LastHeartbeat) > 2*time.Minute {
				delete(s.workers, id)
			}
		}
		s.workerMutex.Unlock()
	}
}

func (s *TranscodingService) processQueue() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Get next queued job
		query := `
			SELECT id FROM transcoding_jobs
			WHERE status = 'queued'
			ORDER BY priority ASC, created_at ASC
			LIMIT 1
		`

		var jobID int64
		err := s.db.QueryRow(query).Scan(&jobID)
		if err != nil {
			continue
		}

		// Check if we have available workers
		s.workerMutex.RLock()
		hasCapacity := len(s.workers) < s.maxWorkers
		s.workerMutex.RUnlock()

		if hasCapacity {
			go s.processJob(jobID)
		}
	}
}

func (s *TranscodingService) processJob(jobID int64) {
	// Create worker
	workerID := uuid.New().String()
	worker := &Worker{
		ID:            workerID,
		Status:        "active",
		CurrentJobs:   1,
		MaxJobs:       1,
		LastHeartbeat: time.Now(),
	}

	s.workerMutex.Lock()
	s.workers[workerID] = worker
	s.workerMutex.Unlock()

	// Update job status
	query := `
		UPDATE transcoding_jobs
		SET status = 'processing', worker_id = $1, started_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	s.db.Exec(query, workerID, jobID)

	// Simulate job processing (in production, this would run actual FFmpeg)
	// This is a placeholder that simulates progress updates
	for i := 0; i <= 100; i += 10 {
		time.Sleep(2 * time.Second)
		query := "UPDATE transcoding_jobs SET progress = $1, updated_at = NOW() WHERE id = $2"
		s.db.Exec(query, float64(i), jobID)
	}

	// Mark job as completed
	query = `
		UPDATE transcoding_jobs
		SET status = 'completed', progress = 100, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	s.db.Exec(query, jobID)

	// Remove worker
	s.workerMutex.Lock()
	delete(s.workers, workerID)
	s.workerMutex.Unlock()
}

// Import types from handler (these would normally be in a shared package)
type TranscodingJob struct {
	ID                   int64      `json:"id"`
	JobName              string     `json:"job_name"`
	JobType              string     `json:"job_type"`
	Status               string     `json:"status"`
	Priority             int        `json:"priority"`
	StreamID             *int64     `json:"stream_id,omitempty"`
	SourceURL            string     `json:"source_url"`
	SourceFormat         string     `json:"source_format,omitempty"`
	SourceCodec          string     `json:"source_codec,omitempty"`
	SourceBitrate        int        `json:"source_bitrate,omitempty"`
	SourceResolution     string     `json:"source_resolution,omitempty"`
	SourceDuration       int        `json:"source_duration,omitempty"`
	TargetURL            string     `json:"target_url"`
	TargetFormat         string     `json:"target_format"`
	TargetQuality        string     `json:"target_quality"`
	TargetPreset         string     `json:"target_preset,omitempty"`
	TargetResolution     string     `json:"target_resolution,omitempty"`
	TargetBitrate        int        `json:"target_bitrate,omitempty"`
	TargetCodec          string     `json:"target_codec"`
	FFmpegCommand        string     `json:"ffmpeg_command,omitempty"`
	HardwareAcceleration bool       `json:"hardware_acceleration"`
	Encoder              string     `json:"encoder"`
	AudioCodec           string     `json:"audio_codec"`
	AudioBitrate         int        `json:"audio_bitrate"`
	SegmentDuration      int        `json:"segment_duration"`
	ExtraParams          string     `json:"extra_params,omitempty"`
	Progress             float64    `json:"progress"`
	CurrentFrame         int64      `json:"current_frame"`
	TotalFrames          int64      `json:"total_frames"`
	CurrentTime          int        `json:"current_time"`
	FPS                  float64    `json:"fps"`
	Bitrate              string     `json:"bitrate,omitempty"`
	Speed                float64    `json:"speed"`
	WorkerID             string     `json:"worker_id,omitempty"`
	WorkerHostname       string     `json:"worker_hostname,omitempty"`
	PID                  int        `json:"pid,omitempty"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	FailedAt             *time.Time `json:"failed_at,omitempty"`
	EstimatedCompletion  *time.Time `json:"estimated_completion,omitempty"`
	ErrorMessage         string     `json:"error_message,omitempty"`
	ErrorCode            string     `json:"error_code,omitempty"`
	RetryCount           int        `json:"retry_count"`
	MaxRetries           int        `json:"max_retries"`
	OutputSize           int64      `json:"output_size,omitempty"`
	OutputDuration       int        `json:"output_duration,omitempty"`
	Metadata             string     `json:"metadata,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type CreateJobRequest struct {
	JobName              string  `json:"job_name"`
	JobType              string  `json:"job_type"`
	Priority             int     `json:"priority"`
	StreamID             *int64  `json:"stream_id,omitempty"`
	SourceURL            string  `json:"source_url"`
	TargetURL            string  `json:"target_url"`
	TargetFormat         string  `json:"target_format"`
	TargetQuality        string  `json:"target_quality"`
	TargetResolution     string  `json:"target_resolution,omitempty"`
	TargetBitrate        int     `json:"target_bitrate,omitempty"`
	HardwareAcceleration bool    `json:"hardware_acceleration"`
	Encoder              string  `json:"encoder"`
	AudioCodec           string  `json:"audio_codec"`
	AudioBitrate         int     `json:"audio_bitrate"`
	SegmentDuration      int     `json:"segment_duration"`
	ExtraParams          string  `json:"extra_params,omitempty"`
}

type UpdateJobRequest struct {
	Status   string `json:"status,omitempty"`
	Priority *int   `json:"priority,omitempty"`
}

type JobFilters struct {
	Status   string `json:"status,omitempty"`
	JobType  string `json:"job_type,omitempty"`
	Priority *int   `json:"priority,omitempty"`
	StreamID *int64 `json:"stream_id,omitempty"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

type QueueStatus struct {
	TotalJobs      int `json:"total_jobs"`
	PendingJobs    int `json:"pending_jobs"`
	QueuedJobs     int `json:"queued_jobs"`
	ProcessingJobs int `json:"processing_jobs"`
	CompletedJobs  int `json:"completed_jobs"`
	FailedJobs     int `json:"failed_jobs"`
	ActiveWorkers  int `json:"active_workers"`
	QueueLength    int `json:"queue_length"`
	AvgWaitTime    int `json:"avg_wait_time"`
}

type JobProgress struct {
	JobID               int64      `json:"job_id"`
	Status              string     `json:"status"`
	Progress            float64    `json:"progress"`
	CurrentTime         int        `json:"current_time"`
	TotalDuration       int        `json:"total_duration"`
	FPS                 float64    `json:"fps"`
	Speed               float64    `json:"speed"`
	Bitrate             string     `json:"bitrate"`
	EstimatedCompletion *time.Time `json:"estimated_completion,omitempty"`
	TimeRemaining       int        `json:"time_remaining"`
}

type TranscodingStats struct {
	Period              string  `json:"period"`
	TotalJobs           int     `json:"total_jobs"`
	CompletedJobs       int     `json:"completed_jobs"`
	FailedJobs          int     `json:"failed_jobs"`
	CancelledJobs       int     `json:"cancelled_jobs"`
	AvgProcessingTime   int     `json:"avg_processing_time"`
	TotalProcessingTime int64   `json:"total_processing_time"`
	AvgSpeed            float64 `json:"avg_speed"`
	TotalInputSize      int64   `json:"total_input_size"`
	TotalOutputSize     int64   `json:"total_output_size"`
	CompressionRatio    float64 `json:"compression_ratio"`
	SDJobs              int     `json:"sd_jobs"`
	HDJobs              int     `json:"hd_jobs"`
	FHDJobs             int     `json:"fhd_jobs"`
	UHDJobs             int     `json:"uhd_jobs"`
}
