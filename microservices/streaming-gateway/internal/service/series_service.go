package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SeriesService provides business logic for series and episode management
type SeriesService struct {
	db *sql.DB
}

// NewSeriesService creates a new series service
func NewSeriesService(db *sql.DB) *SeriesService {
	return &SeriesService{
		db: db,
	}
}

// =====================================================
// ERRORS
// =====================================================

var (
	ErrSeriesNotFound       = errors.New("series not found")
	ErrEpisodeNotFound      = errors.New("episode not found")
	ErrDuplicateEpisode     = errors.New("episode already exists for this season and episode number")
	ErrInvalidSeasonNumber  = errors.New("invalid season number")
	ErrInvalidEpisodeNumber = errors.New("invalid episode number")
)

// =====================================================
// SERIES MANAGEMENT
// =====================================================

// ListSeries lists all series with pagination and filters
func (s *SeriesService) ListSeries(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*Series, int64, error) {
	query := `
		SELECT
			s.id, s.name, s.description, s.category_id, s.cover_url, s.backdrop_url,
			s.trailer_url, s.rating, s.release_year, s.genre, s.cast, s.director,
			s.producer, s.is_active, s.is_featured, s.total_seasons, s.total_episodes,
			s.view_count, s.created_at, s.updated_at,
			COALESCE(c.name, '') as category_name
		FROM series s
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.deleted_at IS NULL
	`

	countQuery := `SELECT COUNT(*) FROM series WHERE deleted_at IS NULL`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if categoryID, ok := filters["category_id"].(int64); ok && categoryID > 0 {
		query += fmt.Sprintf(" AND s.category_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, categoryID)
		argCount++
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		query += fmt.Sprintf(" AND (s.name ILIKE $%d OR s.description ILIKE $%d OR s.genre ILIKE $%d)", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR genre ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	if featured, ok := filters["featured"].(bool); ok && featured {
		query += " AND s.is_featured = TRUE"
		countQuery += " AND is_featured = TRUE"
	}

	// Get total count
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count series: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch series: %w", err)
	}
	defer rows.Close()

	var seriesList []*Series
	for rows.Next() {
		series := &Series{}
		var castJSON []byte

		err := rows.Scan(
			&series.ID, &series.Name, &series.Description, &series.CategoryID,
			&series.CoverURL, &series.BackdropURL, &series.TrailerURL, &series.Rating,
			&series.ReleaseYear, &series.Genre, &castJSON, &series.Director,
			&series.Producer, &series.IsActive, &series.IsFeatured, &series.TotalSeasons,
			&series.TotalEpisodes, &series.ViewCount, &series.CreatedAt, &series.UpdatedAt,
			&series.CategoryName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan series: %w", err)
		}

		// Parse cast JSON
		if len(castJSON) > 0 {
			json.Unmarshal(castJSON, &series.Cast)
		}

		seriesList = append(seriesList, series)
	}

	return seriesList, total, nil
}

// GetSeriesByID gets a series by ID
func (s *SeriesService) GetSeriesByID(ctx context.Context, id int64) (*Series, error) {
	query := `
		SELECT
			s.id, s.name, s.description, s.category_id, s.cover_url, s.backdrop_url,
			s.trailer_url, s.rating, s.release_year, s.genre, s.cast, s.director,
			s.producer, s.is_active, s.is_featured, s.total_seasons, s.total_episodes,
			s.view_count, s.created_at, s.updated_at,
			COALESCE(c.name, '') as category_name
		FROM series s
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.id = $1 AND s.deleted_at IS NULL
	`

	series := &Series{}
	var castJSON []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&series.ID, &series.Name, &series.Description, &series.CategoryID,
		&series.CoverURL, &series.BackdropURL, &series.TrailerURL, &series.Rating,
		&series.ReleaseYear, &series.Genre, &castJSON, &series.Director,
		&series.Producer, &series.IsActive, &series.IsFeatured, &series.TotalSeasons,
		&series.TotalEpisodes, &series.ViewCount, &series.CreatedAt, &series.UpdatedAt,
		&series.CategoryName,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSeriesNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch series: %w", err)
	}

	// Parse cast JSON
	if len(castJSON) > 0 {
		json.Unmarshal(castJSON, &series.Cast)
	}

	return series, nil
}

// CreateSeries creates a new series
func (s *SeriesService) CreateSeries(ctx context.Context, req *CreateSeriesRequest) (*Series, error) {
	// Convert cast to JSON
	castJSON, err := json.Marshal(req.Cast)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cast: %w", err)
	}

	query := `
		INSERT INTO series (
			name, description, category_id, cover_url, backdrop_url, trailer_url,
			rating, release_year, genre, cast, director, producer, is_featured
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, is_active, total_seasons, total_episodes, view_count, created_at, updated_at
	`

	series := &Series{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		CoverURL:    req.CoverURL,
		BackdropURL: req.BackdropURL,
		TrailerURL:  req.TrailerURL,
		Rating:      req.Rating,
		ReleaseYear: req.ReleaseYear,
		Genre:       req.Genre,
		Cast:        req.Cast,
		Director:    req.Director,
		Producer:    req.Producer,
		IsFeatured:  req.IsFeatured,
	}

	err = s.db.QueryRowContext(ctx, query,
		series.Name, series.Description, series.CategoryID, series.CoverURL,
		series.BackdropURL, series.TrailerURL, series.Rating, series.ReleaseYear,
		series.Genre, castJSON, series.Director, series.Producer, series.IsFeatured,
	).Scan(&series.ID, &series.IsActive, &series.TotalSeasons, &series.TotalEpisodes,
		&series.ViewCount, &series.CreatedAt, &series.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create series: %w", err)
	}

	return series, nil
}

// UpdateSeries updates an existing series
func (s *SeriesService) UpdateSeries(ctx context.Context, id int64, req *UpdateSeriesRequest) (*Series, error) {
	// Check if series exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM series WHERE id = $1 AND deleted_at IS NULL)", id).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check series: %w", err)
	}
	if !exists {
		return nil, ErrSeriesNotFound
	}

	// Convert cast to JSON
	castJSON, err := json.Marshal(req.Cast)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cast: %w", err)
	}

	// Build update query dynamically based on provided fields
	query := `
		UPDATE series SET
			name = $1, description = $2, category_id = $3, cover_url = $4,
			backdrop_url = $5, trailer_url = $6, rating = $7, release_year = $8,
			genre = $9, cast = $10, director = $11, producer = $12,
			updated_at = NOW()
	`

	args := []interface{}{
		req.Name, req.Description, req.CategoryID, req.CoverURL,
		req.BackdropURL, req.TrailerURL, req.Rating, req.ReleaseYear,
		req.Genre, castJSON, req.Director, req.Producer,
	}

	argCount := 13

	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argCount)
		args = append(args, *req.IsActive)
		argCount++
	}

	if req.IsFeatured != nil {
		query += fmt.Sprintf(", is_featured = $%d", argCount)
		args = append(args, *req.IsFeatured)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update series: %w", err)
	}

	return s.GetSeriesByID(ctx, id)
}

// DeleteSeries soft deletes a series
func (s *SeriesService) DeleteSeries(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE series SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete series: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrSeriesNotFound
	}

	return nil
}

// GetSeriesSeasons gets season information for a series
func (s *SeriesService) GetSeriesSeasons(ctx context.Context, seriesID int64) ([]*SeasonInfo, error) {
	query := `
		SELECT season, COUNT(*) as episode_count
		FROM episodes
		WHERE series_id = $1 AND is_active = TRUE
		GROUP BY season
		ORDER BY season
	`

	rows, err := s.db.QueryContext(ctx, query, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch seasons: %w", err)
	}
	defer rows.Close()

	var seasons []*SeasonInfo
	for rows.Next() {
		season := &SeasonInfo{}
		if err := rows.Scan(&season.Season, &season.EpisodeCount); err != nil {
			return nil, fmt.Errorf("failed to scan season: %w", err)
		}
		seasons = append(seasons, season)
	}

	return seasons, nil
}

// =====================================================
// EPISODE MANAGEMENT
// =====================================================

// ListEpisodes lists episodes for a series
func (s *SeriesService) ListEpisodes(ctx context.Context, seriesID int64, season *int) ([]*Episode, error) {
	query := `
		SELECT
			e.id, e.series_id, e.season, e.episode, e.title, e.description,
			e.stream_url, e.thumbnail_url, e.duration, e.air_date, e.rating,
			e.is_active, e.view_count, e.created_at, e.updated_at,
			s.name as series_name
		FROM episodes e
		INNER JOIN series s ON e.series_id = s.id
		WHERE e.series_id = $1
	`

	args := []interface{}{seriesID}
	argCount := 2

	if season != nil {
		query += fmt.Sprintf(" AND e.season = $%d", argCount)
		args = append(args, *season)
	}

	query += " ORDER BY e.season ASC, e.episode ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", err)
	}
	defer rows.Close()

	var episodes []*Episode
	for rows.Next() {
		episode := &Episode{}
		var airDate sql.NullString

		err := rows.Scan(
			&episode.ID, &episode.SeriesID, &episode.Season, &episode.Episode,
			&episode.Title, &episode.Description, &episode.StreamURL, &episode.ThumbnailURL,
			&episode.Duration, &airDate, &episode.Rating, &episode.IsActive,
			&episode.ViewCount, &episode.CreatedAt, &episode.UpdatedAt, &episode.SeriesName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}

		if airDate.Valid {
			episode.AirDate = &airDate.String
		}

		episodes = append(episodes, episode)
	}

	return episodes, nil
}

// GetEpisodeByID gets an episode by ID
func (s *SeriesService) GetEpisodeByID(ctx context.Context, id int64) (*Episode, error) {
	query := `
		SELECT
			e.id, e.series_id, e.season, e.episode, e.title, e.description,
			e.stream_url, e.thumbnail_url, e.duration, e.air_date, e.rating,
			e.is_active, e.view_count, e.created_at, e.updated_at,
			s.name as series_name
		FROM episodes e
		INNER JOIN series s ON e.series_id = s.id
		WHERE e.id = $1
	`

	episode := &Episode{}
	var airDate sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&episode.ID, &episode.SeriesID, &episode.Season, &episode.Episode,
		&episode.Title, &episode.Description, &episode.StreamURL, &episode.ThumbnailURL,
		&episode.Duration, &airDate, &episode.Rating, &episode.IsActive,
		&episode.ViewCount, &episode.CreatedAt, &episode.UpdatedAt, &episode.SeriesName,
	)
	if err == sql.ErrNoRows {
		return nil, ErrEpisodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episode: %w", err)
	}

	if airDate.Valid {
		episode.AirDate = &airDate.String
	}

	return episode, nil
}

// CreateEpisode creates a new episode
func (s *SeriesService) CreateEpisode(ctx context.Context, req *CreateEpisodeRequest) (*Episode, error) {
	// Validate season and episode numbers
	if req.Season < 1 {
		return nil, ErrInvalidSeasonNumber
	}
	if req.Episode < 1 {
		return nil, ErrInvalidEpisodeNumber
	}

	// Check if series exists
	var seriesExists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM series WHERE id = $1 AND deleted_at IS NULL)", req.SeriesID).Scan(&seriesExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check series: %w", err)
	}
	if !seriesExists {
		return nil, ErrSeriesNotFound
	}

	// Check for duplicate episode
	var duplicateExists bool
	err = s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM episodes WHERE series_id = $1 AND season = $2 AND episode = $3)
	`, req.SeriesID, req.Season, req.Episode).Scan(&duplicateExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if duplicateExists {
		return nil, ErrDuplicateEpisode
	}

	// Convert air_date
	var airDate *time.Time
	if req.AirDate != nil {
		t, err := time.Parse("2006-01-02", *req.AirDate)
		if err == nil {
			airDate = &t
		}
	}

	query := `
		INSERT INTO episodes (
			series_id, season, episode, title, description, stream_url,
			thumbnail_url, duration, air_date, rating
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, is_active, view_count, created_at, updated_at
	`

	episode := &Episode{
		SeriesID:     req.SeriesID,
		Season:       req.Season,
		Episode:      req.Episode,
		Title:        req.Title,
		Description:  req.Description,
		StreamURL:    req.StreamURL,
		ThumbnailURL: req.ThumbnailURL,
		Duration:     req.Duration,
		AirDate:      req.AirDate,
		Rating:       req.Rating,
	}

	err = s.db.QueryRowContext(ctx, query,
		episode.SeriesID, episode.Season, episode.Episode, episode.Title,
		episode.Description, episode.StreamURL, episode.ThumbnailURL,
		episode.Duration, airDate, episode.Rating,
	).Scan(&episode.ID, &episode.IsActive, &episode.ViewCount, &episode.CreatedAt, &episode.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create episode: %w", err)
	}

	return episode, nil
}

// UpdateEpisode updates an existing episode
func (s *SeriesService) UpdateEpisode(ctx context.Context, id int64, req *UpdateEpisodeRequest) (*Episode, error) {
	// Check if episode exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM episodes WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check episode: %w", err)
	}
	if !exists {
		return nil, ErrEpisodeNotFound
	}

	// Convert air_date
	var airDate *time.Time
	if req.AirDate != nil {
		t, err := time.Parse("2006-01-02", *req.AirDate)
		if err == nil {
			airDate = &t
		}
	}

	// Build update query
	query := `
		UPDATE episodes SET
			title = $1, description = $2, stream_url = $3, thumbnail_url = $4,
			duration = $5, air_date = $6, rating = $7, updated_at = NOW()
	`

	args := []interface{}{
		req.Title, req.Description, req.StreamURL, req.ThumbnailURL,
		req.Duration, airDate, req.Rating,
	}

	argCount := 8

	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argCount)
		args = append(args, *req.IsActive)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update episode: %w", err)
	}

	return s.GetEpisodeByID(ctx, id)
}

// DeleteEpisode deletes an episode
func (s *SeriesService) DeleteEpisode(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM episodes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete episode: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrEpisodeNotFound
	}

	return nil
}

// IncrementEpisodeViewCount increments view count for an episode
func (s *SeriesService) IncrementEpisodeViewCount(ctx context.Context, episodeID int64) error {
	// Increment episode view count
	_, err := s.db.ExecContext(ctx, `
		UPDATE episodes SET view_count = view_count + 1 WHERE id = $1
	`, episodeID)
	if err != nil {
		return fmt.Errorf("failed to increment episode view count: %w", err)
	}

	// Also increment series view count
	_, err = s.db.ExecContext(ctx, `
		UPDATE series SET view_count = view_count + 1
		WHERE id = (SELECT series_id FROM episodes WHERE id = $1)
	`, episodeID)
	if err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to increment series view count: %v\n", err)
	}

	return nil
}
