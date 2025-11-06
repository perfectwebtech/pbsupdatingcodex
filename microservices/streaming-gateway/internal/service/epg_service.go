package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// EPGService provides business logic for EPG management
type EPGService struct {
	db *sql.DB
}

// NewEPGService creates a new EPG service
func NewEPGService(db *sql.DB) *EPGService {
	return &EPGService{
		db: db,
	}
}

// =====================================================
// ERRORS
// =====================================================

var (
	ErrEPGProgramNotFound    = errors.New("EPG program not found")
	ErrEPGSourceNotFound     = errors.New("EPG source not found")
	ErrInvalidTimeRange      = errors.New("invalid time range: end time must be after start time")
	ErrScheduleConflict      = errors.New("schedule conflict: overlapping program exists")
	ErrImportInProgress      = errors.New("import already in progress for this source")
)

// =====================================================
// EPG PROGRAM MANAGEMENT
// =====================================================

// ListPrograms lists EPG programs with filters
func (s *EPGService) ListPrograms(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*EPGProgram, int64, error) {
	query := `
		SELECT
			p.id, p.stream_id, p.title, p.description, p.category,
			p.start_time, p.end_time, p.duration, p.image_url, p.icon_url,
			p.episode_number, p.season_number, p.year, p.rating,
			p.directors, p.actors, p.country, p.language,
			p.is_live, p.is_repeat, p.is_premiere,
			p.created_at, p.updated_at,
			s.name as stream_name
		FROM epg_programs p
		INNER JOIN streams s ON p.stream_id = s.id
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM epg_programs WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if streamID, ok := filters["stream_id"].(int64); ok && streamID > 0 {
		query += fmt.Sprintf(" AND p.stream_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND stream_id = $%d", argCount)
		args = append(args, streamID)
		argCount++
	}

	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		query += fmt.Sprintf(" AND p.start_time >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND start_time >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		query += fmt.Sprintf(" AND p.end_time <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND end_time <= $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	if category, ok := filters["category"].(string); ok && category != "" {
		query += fmt.Sprintf(" AND p.category = $%d", argCount)
		countQuery += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
		argCount++
	}

	// Get total count
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count programs: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY p.start_time ASC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch programs: %w", err)
	}
	defer rows.Close()

	var programs []*EPGProgram
	for rows.Next() {
		program := &EPGProgram{}
		var directorsJSON, actorsJSON []byte

		err := rows.Scan(
			&program.ID, &program.StreamID, &program.Title, &program.Description,
			&program.Category, &program.StartTime, &program.EndTime, &program.Duration,
			&program.ImageURL, &program.IconURL, &program.Episode, &program.Season,
			&program.Year, &program.Rating, &directorsJSON, &actorsJSON,
			&program.Country, &program.Language, &program.IsLive, &program.IsRepeat,
			&program.IsPremiere, &program.CreatedAt, &program.UpdatedAt, &program.StreamName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan program: %w", err)
		}

		// Parse JSON arrays
		if len(directorsJSON) > 0 {
			json.Unmarshal(directorsJSON, &program.Directors)
		}
		if len(actorsJSON) > 0 {
			json.Unmarshal(actorsJSON, &program.Actors)
		}

		// Determine status
		now := time.Now()
		if now.After(program.StartTime) && now.Before(program.EndTime) {
			program.Status = "now_playing"
		} else if now.Before(program.StartTime) {
			program.Status = "upcoming"
		} else {
			program.Status = "past"
		}

		programs = append(programs, program)
	}

	return programs, total, nil
}

// GetProgramByID gets an EPG program by ID
func (s *EPGService) GetProgramByID(ctx context.Context, id int64) (*EPGProgram, error) {
	query := `
		SELECT
			p.id, p.stream_id, p.title, p.description, p.category,
			p.start_time, p.end_time, p.duration, p.image_url, p.icon_url,
			p.episode_number, p.season_number, p.year, p.rating,
			p.directors, p.actors, p.country, p.language,
			p.is_live, p.is_repeat, p.is_premiere,
			p.created_at, p.updated_at,
			s.name as stream_name
		FROM epg_programs p
		INNER JOIN streams s ON p.stream_id = s.id
		WHERE p.id = $1
	`

	program := &EPGProgram{}
	var directorsJSON, actorsJSON []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&program.ID, &program.StreamID, &program.Title, &program.Description,
		&program.Category, &program.StartTime, &program.EndTime, &program.Duration,
		&program.ImageURL, &program.IconURL, &program.Episode, &program.Season,
		&program.Year, &program.Rating, &directorsJSON, &actorsJSON,
		&program.Country, &program.Language, &program.IsLive, &program.IsRepeat,
		&program.IsPremiere, &program.CreatedAt, &program.UpdatedAt, &program.StreamName,
	)
	if err == sql.ErrNoRows {
		return nil, ErrEPGProgramNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch program: %w", err)
	}

	// Parse JSON arrays
	if len(directorsJSON) > 0 {
		json.Unmarshal(directorsJSON, &program.Directors)
	}
	if len(actorsJSON) > 0 {
		json.Unmarshal(actorsJSON, &program.Actors)
	}

	return program, nil
}

// CreateProgram creates a new EPG program
func (s *EPGService) CreateProgram(ctx context.Context, req *CreateEPGProgramRequest) (*EPGProgram, error) {
	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	// Validate time range
	if !endTime.After(startTime) {
		return nil, ErrInvalidTimeRange
	}

	// Check for schedule conflicts
	conflicts, err := s.checkScheduleConflicts(ctx, req.StreamID, startTime, endTime, nil)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, ErrScheduleConflict
	}

	// Convert arrays to JSON
	directorsJSON, _ := json.Marshal(req.Directors)
	actorsJSON, _ := json.Marshal(req.Actors)

	query := `
		INSERT INTO epg_programs (
			stream_id, title, description, category, start_time, end_time,
			image_url, icon_url, episode_number, season_number, year, rating,
			directors, actors, country, language, is_live, is_repeat, is_premiere
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, duration, created_at, updated_at
	`

	program := &EPGProgram{
		StreamID:    req.StreamID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		StartTime:   startTime,
		EndTime:     endTime,
		ImageURL:    req.ImageURL,
		IconURL:     req.IconURL,
		Episode:     req.Episode,
		Season:      req.Season,
		Year:        req.Year,
		Rating:      req.Rating,
		Directors:   req.Directors,
		Actors:      req.Actors,
		Country:     req.Country,
		Language:    req.Language,
		IsLive:      req.IsLive,
		IsRepeat:    req.IsRepeat,
		IsPremiere:  req.IsPremiere,
	}

	err = s.db.QueryRowContext(ctx, query,
		program.StreamID, program.Title, program.Description, program.Category,
		program.StartTime, program.EndTime, program.ImageURL, program.IconURL,
		program.Episode, program.Season, program.Year, program.Rating,
		directorsJSON, actorsJSON, program.Country, program.Language,
		program.IsLive, program.IsRepeat, program.IsPremiere,
	).Scan(&program.ID, &program.Duration, &program.CreatedAt, &program.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create program: %w", err)
	}

	return program, nil
}

// UpdateProgram updates an EPG program
func (s *EPGService) UpdateProgram(ctx context.Context, id int64, req *UpdateEPGProgramRequest) (*EPGProgram, error) {
	// Check if program exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM epg_programs WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check program: %w", err)
	}
	if !exists {
		return nil, ErrEPGProgramNotFound
	}

	// Parse times if provided
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format: %w", err)
		}
		startTime = &t
	}

	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format: %w", err)
		}
		endTime = &t
	}

	// Validate time range if both provided
	if startTime != nil && endTime != nil && !endTime.After(*startTime) {
		return nil, ErrInvalidTimeRange
	}

	// Convert arrays to JSON
	directorsJSON, _ := json.Marshal(req.Directors)
	actorsJSON, _ := json.Marshal(req.Actors)

	query := `
		UPDATE epg_programs SET
			title = $1, description = $2, category = $3,
			image_url = $4, episode_number = $5, season_number = $6,
			year = $7, rating = $8, directors = $9, actors = $10,
			is_live = $11, is_repeat = $12, is_premiere = $13,
			updated_at = NOW()
		WHERE id = $14
	`

	_, err = s.db.ExecContext(ctx, query,
		req.Title, req.Description, req.Category, req.ImageURL,
		req.Episode, req.Season, req.Year, req.Rating,
		directorsJSON, actorsJSON, req.IsLive, req.IsRepeat,
		req.IsPremiere, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update program: %w", err)
	}

	return s.GetProgramByID(ctx, id)
}

// DeleteProgram deletes an EPG program
func (s *EPGService) DeleteProgram(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM epg_programs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete program: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrEPGProgramNotFound
	}

	return nil
}

// GetCurrentPrograms gets currently playing programs
func (s *EPGService) GetCurrentPrograms(ctx context.Context, filters map[string]interface{}) ([]*EPGProgram, error) {
	query := `
		SELECT
			p.id, p.stream_id, p.title, p.description, p.category,
			p.start_time, p.end_time, p.duration, p.image_url, p.icon_url,
			p.episode_number, p.season_number, p.year, p.rating,
			p.directors, p.actors, p.country, p.language,
			p.is_live, p.is_repeat, p.is_premiere,
			p.created_at, p.updated_at,
			s.name as stream_name
		FROM epg_programs p
		INNER JOIN streams s ON p.stream_id = s.id
		WHERE NOW() BETWEEN p.start_time AND p.end_time
	`

	args := []interface{}{}
	argCount := 1

	if streamID, ok := filters["stream_id"].(int64); ok && streamID > 0 {
		query += fmt.Sprintf(" AND p.stream_id = $%d", argCount)
		args = append(args, streamID)
		argCount++
	}

	query += " ORDER BY s.name, p.start_time"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current programs: %w", err)
	}
	defer rows.Close()

	var programs []*EPGProgram
	for rows.Next() {
		program := &EPGProgram{}
		var directorsJSON, actorsJSON []byte

		err := rows.Scan(
			&program.ID, &program.StreamID, &program.Title, &program.Description,
			&program.Category, &program.StartTime, &program.EndTime, &program.Duration,
			&program.ImageURL, &program.IconURL, &program.Episode, &program.Season,
			&program.Year, &program.Rating, &directorsJSON, &actorsJSON,
			&program.Country, &program.Language, &program.IsLive, &program.IsRepeat,
			&program.IsPremiere, &program.CreatedAt, &program.UpdatedAt, &program.StreamName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan program: %w", err)
		}

		if len(directorsJSON) > 0 {
			json.Unmarshal(directorsJSON, &program.Directors)
		}
		if len(actorsJSON) > 0 {
			json.Unmarshal(actorsJSON, &program.Actors)
		}

		program.Status = "now_playing"
		programs = append(programs, program)
	}

	return programs, nil
}

// GetSchedule gets EPG schedule for a date
func (s *EPGService) GetSchedule(ctx context.Context, filters map[string]interface{}) ([]*EPGProgram, error) {
	dateStr, ok := filters["date"].(string)
	if !ok {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT
			p.id, p.stream_id, p.title, p.description, p.category,
			p.start_time, p.end_time, p.duration, p.image_url, p.icon_url,
			p.episode_number, p.season_number, p.year, p.rating,
			p.directors, p.actors, p.country, p.language,
			p.is_live, p.is_repeat, p.is_premiere,
			p.created_at, p.updated_at,
			s.name as stream_name
		FROM epg_programs p
		INNER JOIN streams s ON p.stream_id = s.id
		WHERE p.start_time >= $1 AND p.start_time < $2
	`

	args := []interface{}{startOfDay, endOfDay}
	argCount := 3

	if streamID, ok := filters["stream_id"].(int64); ok && streamID > 0 {
		query += fmt.Sprintf(" AND p.stream_id = $%d", argCount)
		args = append(args, streamID)
		argCount++
	}

	query += " ORDER BY p.stream_id, p.start_time"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}
	defer rows.Close()

	var programs []*EPGProgram
	for rows.Next() {
		program := &EPGProgram{}
		var directorsJSON, actorsJSON []byte

		err := rows.Scan(
			&program.ID, &program.StreamID, &program.Title, &program.Description,
			&program.Category, &program.StartTime, &program.EndTime, &program.Duration,
			&program.ImageURL, &program.IconURL, &program.Episode, &program.Season,
			&program.Year, &program.Rating, &directorsJSON, &actorsJSON,
			&program.Country, &program.Language, &program.IsLive, &program.IsRepeat,
			&program.IsPremiere, &program.CreatedAt, &program.UpdatedAt, &program.StreamName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan program: %w", err)
		}

		if len(directorsJSON) > 0 {
			json.Unmarshal(directorsJSON, &program.Directors)
		}
		if len(actorsJSON) > 0 {
			json.Unmarshal(actorsJSON, &program.Actors)
		}

		now := time.Now()
		if now.After(program.StartTime) && now.Before(program.EndTime) {
			program.Status = "now_playing"
		} else if now.Before(program.StartTime) {
			program.Status = "upcoming"
		} else {
			program.Status = "past"
		}

		programs = append(programs, program)
	}

	return programs, nil
}

// =====================================================
// EPG SOURCE MANAGEMENT
// =====================================================

// ListSources lists all EPG sources
func (s *EPGService) ListSources(ctx context.Context) ([]*EPGSource, error) {
	query := `
		SELECT
			id, name, source_type, url, update_interval, format_config,
			is_active, last_sync, last_sync_status, last_error,
			programs_imported, programs_updated, programs_failed,
			created_at, updated_at
		FROM epg_sources
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sources: %w", err)
	}
	defer rows.Close()

	var sources []*EPGSource
	for rows.Next() {
		source := &EPGSource{}
		var formatConfigJSON []byte

		err := rows.Scan(
			&source.ID, &source.Name, &source.SourceType, &source.URL,
			&source.UpdateInterval, &formatConfigJSON, &source.IsActive,
			&source.LastSync, &source.LastSyncStatus, &source.LastError,
			&source.ProgramsImported, &source.ProgramsUpdated, &source.ProgramsFailed,
			&source.CreatedAt, &source.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}

		if len(formatConfigJSON) > 0 {
			json.Unmarshal(formatConfigJSON, &source.FormatConfig)
		}

		sources = append(sources, source)
	}

	return sources, nil
}

// CreateSource creates a new EPG source
func (s *EPGService) CreateSource(ctx context.Context, req *CreateEPGSourceRequest) (*EPGSource, error) {
	formatConfigJSON, _ := json.Marshal(req.FormatConfig)
	credentialsJSON, _ := json.Marshal(req.Credentials)

	query := `
		INSERT INTO epg_sources (
			name, source_type, url, update_interval, format_config, credentials
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, is_active, programs_imported, programs_updated, programs_failed,
				  created_at, updated_at
	`

	source := &EPGSource{
		Name:           req.Name,
		SourceType:     req.SourceType,
		URL:            req.URL,
		UpdateInterval: req.UpdateInterval,
		FormatConfig:   req.FormatConfig,
	}

	err := s.db.QueryRowContext(ctx, query,
		source.Name, source.SourceType, source.URL, source.UpdateInterval,
		formatConfigJSON, credentialsJSON,
	).Scan(&source.ID, &source.IsActive, &source.ProgramsImported,
		&source.ProgramsUpdated, &source.ProgramsFailed, &source.CreatedAt, &source.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	return source, nil
}

// SyncFromSource triggers EPG import from source
func (s *EPGService) SyncFromSource(ctx context.Context, sourceID int64, importedBy *int64) (*EPGImportLog, error) {
	// Check if source exists
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM epg_sources WHERE id = $1 AND is_active = TRUE)
	`, sourceID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check source: %w", err)
	}
	if !exists {
		return nil, ErrEPGSourceNotFound
	}

	// Check if import is already in progress
	var inProgress bool
	err = s.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM epg_import_log
			WHERE source_id = $1 AND status = 'in_progress'
		)
	`, sourceID).Scan(&inProgress)
	if err != nil {
		return nil, fmt.Errorf("failed to check import status: %w", err)
	}
	if inProgress {
		return nil, ErrImportInProgress
	}

	// Create import log
	importLog := &EPGImportLog{
		SourceID:   &sourceID,
		ImportType: "full",
		Status:     "in_progress",
		StartedAt:  time.Now(),
		ImportedBy: importedBy,
	}

	query := `
		INSERT INTO epg_import_log (
			source_id, import_type, status, started_at, imported_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err = s.db.QueryRowContext(ctx, query,
		importLog.SourceID, importLog.ImportType, importLog.Status,
		importLog.StartedAt, importLog.ImportedBy,
	).Scan(&importLog.ID, &importLog.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create import log: %w", err)
	}

	// TODO: In production, trigger async import job here
	// For now, just return the log entry
	// go s.performImport(ctx, sourceID, importLog.ID)

	return importLog, nil
}

// GetImportHistory gets EPG import history
func (s *EPGService) GetImportHistory(ctx context.Context, limit, offset int) ([]*EPGImportLog, int64, error) {
	countQuery := `SELECT COUNT(*) FROM epg_import_log`
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count import logs: %w", err)
	}

	query := `
		SELECT
			id, source_id, import_type, status, programs_processed,
			programs_created, programs_updated, programs_deleted, programs_failed,
			started_at, completed_at, duration_seconds, error_message,
			error_details, import_file, imported_by, created_at
		FROM epg_import_log
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch import logs: %w", err)
	}
	defer rows.Close()

	var logs []*EPGImportLog
	for rows.Next() {
		log := &EPGImportLog{}
		var errorDetailsJSON []byte

		err := rows.Scan(
			&log.ID, &log.SourceID, &log.ImportType, &log.Status,
			&log.ProgramsProcessed, &log.ProgramsCreated, &log.ProgramsUpdated,
			&log.ProgramsDeleted, &log.ProgramsFailed, &log.StartedAt,
			&log.CompletedAt, &log.DurationSeconds, &log.ErrorMessage,
			&errorDetailsJSON, &log.ImportFile, &log.ImportedBy, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan import log: %w", err)
		}

		if len(errorDetailsJSON) > 0 {
			json.Unmarshal(errorDetailsJSON, &log.ErrorDetails)
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// CleanupOldPrograms removes old EPG programs
func (s *EPGService) CleanupOldPrograms(ctx context.Context, daysToKeep int) (int64, error) {
	query := `
		DELETE FROM epg_programs
		WHERE end_time < NOW() - ($1 || ' days')::INTERVAL
	`

	result, err := s.db.ExecContext(ctx, query, daysToKeep)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup programs: %w", err)
	}

	deletedCount, _ := result.RowsAffected()
	return deletedCount, nil
}

// =====================================================
// HELPER FUNCTIONS
// =====================================================

// checkScheduleConflicts checks for overlapping programs
func (s *EPGService) checkScheduleConflicts(ctx context.Context, streamID int64, startTime, endTime time.Time, excludeID *int64) ([]int64, error) {
	query := `
		SELECT id FROM epg_programs
		WHERE stream_id = $1
		AND (
			(start_time <= $2 AND end_time > $2) OR
			(start_time < $3 AND end_time >= $3) OR
			(start_time >= $2 AND end_time <= $3)
		)
	`

	args := []interface{}{streamID, startTime, endTime}

	if excludeID != nil {
		query += " AND id != $4"
		args = append(args, *excludeID)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		conflicts = append(conflicts, id)
	}

	return conflicts, nil
}
