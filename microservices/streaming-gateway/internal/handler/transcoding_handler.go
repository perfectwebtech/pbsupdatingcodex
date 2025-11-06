package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// TranscodingHandler handles all transcoding-related HTTP requests
type TranscodingHandler struct {
	service TranscodingService
}

// TranscodingService interface defines the business logic for transcoding
type TranscodingService interface {
	GetAllJobs(filters JobFilters) ([]TranscodingJob, error)
	GetJobByID(id int64) (*TranscodingJob, error)
	CreateJob(job *CreateJobRequest) (*TranscodingJob, error)
	UpdateJob(id int64, updates *UpdateJobRequest) error
	DeleteJob(id int64) error
	CancelJob(id int64) error
	RetryJob(id int64) error
	GetQueueStatus() (*QueueStatus, error)
	GetJobProgress(id int64) (*JobProgress, error)
	GetStatistics(days int) (*TranscodingStats, error)
	GetWorkers() ([]Worker, error)
}

// Models
type TranscodingJob struct {
	ID                   int64     `json:"id"`
	JobName              string    `json:"job_name"`
	JobType              string    `json:"job_type"`
	Status               string    `json:"status"`
	Priority             int       `json:"priority"`
	StreamID             *int64    `json:"stream_id,omitempty"`
	SourceURL            string    `json:"source_url"`
	SourceFormat         string    `json:"source_format,omitempty"`
	SourceCodec          string    `json:"source_codec,omitempty"`
	SourceBitrate        int       `json:"source_bitrate,omitempty"`
	SourceResolution     string    `json:"source_resolution,omitempty"`
	SourceDuration       int       `json:"source_duration,omitempty"`
	TargetURL            string    `json:"target_url"`
	TargetFormat         string    `json:"target_format"`
	TargetQuality        string    `json:"target_quality"`
	TargetPreset         string    `json:"target_preset,omitempty"`
	TargetResolution     string    `json:"target_resolution,omitempty"`
	TargetBitrate        int       `json:"target_bitrate,omitempty"`
	TargetCodec          string    `json:"target_codec"`
	FFmpegCommand        string    `json:"ffmpeg_command,omitempty"`
	HardwareAcceleration bool      `json:"hardware_acceleration"`
	Encoder              string    `json:"encoder"`
	AudioCodec           string    `json:"audio_codec"`
	AudioBitrate         int       `json:"audio_bitrate"`
	SegmentDuration      int       `json:"segment_duration"`
	ExtraParams          string    `json:"extra_params,omitempty"`
	Progress             float64   `json:"progress"`
	CurrentFrame         int64     `json:"current_frame"`
	TotalFrames          int64     `json:"total_frames"`
	CurrentTime          int       `json:"current_time"`
	FPS                  float64   `json:"fps"`
	Bitrate              string    `json:"bitrate,omitempty"`
	Speed                float64   `json:"speed"`
	WorkerID             string    `json:"worker_id,omitempty"`
	WorkerHostname       string    `json:"worker_hostname,omitempty"`
	PID                  int       `json:"pid,omitempty"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	FailedAt             *time.Time `json:"failed_at,omitempty"`
	EstimatedCompletion  *time.Time `json:"estimated_completion,omitempty"`
	ErrorMessage         string    `json:"error_message,omitempty"`
	ErrorCode            string    `json:"error_code,omitempty"`
	RetryCount           int       `json:"retry_count"`
	MaxRetries           int       `json:"max_retries"`
	OutputSize           int64     `json:"output_size,omitempty"`
	OutputDuration       int       `json:"output_duration,omitempty"`
	Metadata             string    `json:"metadata,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type CreateJobRequest struct {
	JobName              string  `json:"job_name" validate:"required"`
	JobType              string  `json:"job_type" validate:"required,oneof=live_transcode vod_transcode recording clip_generation"`
	Priority             int     `json:"priority" validate:"min=1,max=10"`
	StreamID             *int64  `json:"stream_id,omitempty"`
	SourceURL            string  `json:"source_url" validate:"required"`
	TargetURL            string  `json:"target_url" validate:"required"`
	TargetFormat         string  `json:"target_format" validate:"required"`
	TargetQuality        string  `json:"target_quality" validate:"required,oneof=sd hd fhd uhd custom"`
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
	AvgWaitTime    int `json:"avg_wait_time"` // seconds
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
	TimeRemaining       int        `json:"time_remaining"` // seconds
}

type TranscodingStats struct {
	Period              string  `json:"period"`
	TotalJobs           int     `json:"total_jobs"`
	CompletedJobs       int     `json:"completed_jobs"`
	FailedJobs          int     `json:"failed_jobs"`
	CancelledJobs       int     `json:"cancelled_jobs"`
	AvgProcessingTime   int     `json:"avg_processing_time"` // seconds
	TotalProcessingTime int64   `json:"total_processing_time"` // seconds
	AvgSpeed            float64 `json:"avg_speed"`
	TotalInputSize      int64   `json:"total_input_size"`
	TotalOutputSize     int64   `json:"total_output_size"`
	CompressionRatio    float64 `json:"compression_ratio"`
	SDJobs              int     `json:"sd_jobs"`
	HDJobs              int     `json:"hd_jobs"`
	FHDJobs             int     `json:"fhd_jobs"`
	UHDJobs             int     `json:"uhd_jobs"`
}

type Worker struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	Status       string    `json:"status"` // active, idle, offline
	CurrentJobs  int       `json:"current_jobs"`
	MaxJobs      int       `json:"max_jobs"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// NewTranscodingHandler creates a new transcoding handler
func NewTranscodingHandler(service TranscodingService) *TranscodingHandler {
	return &TranscodingHandler{
		service: service,
	}
}

// RegisterRoutes registers all transcoding routes
func (h *TranscodingHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/transcoding/jobs", h.GetJobs).Methods("GET")
	r.HandleFunc("/transcoding/jobs", h.CreateJob).Methods("POST")
	r.HandleFunc("/transcoding/jobs/{id}", h.GetJob).Methods("GET")
	r.HandleFunc("/transcoding/jobs/{id}", h.UpdateJob).Methods("PUT")
	r.HandleFunc("/transcoding/jobs/{id}", h.DeleteJob).Methods("DELETE")
	r.HandleFunc("/transcoding/jobs/{id}/cancel", h.CancelJob).Methods("POST")
	r.HandleFunc("/transcoding/jobs/{id}/retry", h.RetryJob).Methods("POST")
	r.HandleFunc("/transcoding/jobs/{id}/progress", h.GetJobProgress).Methods("GET")
	r.HandleFunc("/transcoding/queue", h.GetQueue).Methods("GET")
	r.HandleFunc("/transcoding/stats", h.GetStats).Methods("GET")
	r.HandleFunc("/transcoding/workers", h.GetWorkers).Methods("GET")
}

// GetJobs returns all transcoding jobs with optional filters
func (h *TranscodingHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	filters := JobFilters{
		Status:  r.URL.Query().Get("status"),
		JobType: r.URL.Query().Get("job_type"),
		Limit:   50,
		Offset:  0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	jobs, err := h.service.GetAllJobs(filters)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch jobs", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    jobs,
		"count":   len(jobs),
	})
}

// GetJob returns a single transcoding job by ID
func (h *TranscodingHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	job, err := h.service.GetJobByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Job not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    job,
	})
}

// CreateJob creates a new transcoding job
func (h *TranscodingHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set defaults
	if req.Priority == 0 {
		req.Priority = 5
	}
	if req.Encoder == "" {
		req.Encoder = "libx264"
	}
	if req.AudioCodec == "" {
		req.AudioCodec = "aac"
	}
	if req.AudioBitrate == 0 {
		req.AudioBitrate = 128
	}
	if req.SegmentDuration == 0 {
		req.SegmentDuration = 6
	}

	job, err := h.service.CreateJob(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create job", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Transcoding job created successfully",
		"data":    job,
	})
}

// UpdateJob updates an existing transcoding job
func (h *TranscodingHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var req UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.UpdateJob(id, &req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update job", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Job updated successfully",
	})
}

// DeleteJob deletes a transcoding job
func (h *TranscodingHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	if err := h.service.DeleteJob(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete job", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Job deleted successfully",
	})
}

// CancelJob cancels a running or queued transcoding job
func (h *TranscodingHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	if err := h.service.CancelJob(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to cancel job", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Job cancelled successfully",
	})
}

// RetryJob retries a failed transcoding job
func (h *TranscodingHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	if err := h.service.RetryJob(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retry job", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Job queued for retry",
	})
}

// GetJobProgress returns the progress of a specific job
func (h *TranscodingHandler) GetJobProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	progress, err := h.service.GetJobProgress(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get job progress", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    progress,
	})
}

// GetQueue returns the current queue status
func (h *TranscodingHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.GetQueueStatus()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get queue status", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// GetStats returns transcoding statistics
func (h *TranscodingHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	days := 7
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	stats, err := h.service.GetStatistics(days)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get statistics", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetWorkers returns all active transcoding workers
func (h *TranscodingHandler) GetWorkers(w http.ResponseWriter, r *http.Request) {
	workers, err := h.service.GetWorkers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get workers", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    workers,
		"count":   len(workers),
	})
}

// Helper functions
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "message": "Failed to marshal response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	errorMessage := message
	if err != nil {
		errorMessage = message + ": " + err.Error()
	}

	respondWithJSON(w, statusCode, map[string]interface{}{
		"success": false,
		"message": errorMessage,
	})
}
