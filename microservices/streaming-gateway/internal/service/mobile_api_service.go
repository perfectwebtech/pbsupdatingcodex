package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"streaming-gateway/internal/handler"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrDeviceLimitReached = errors.New("maximum device limit reached")
	ErrDeviceNotFound     = errors.New("device not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrContentNotFound    = errors.New("content not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrDownloadNotFound   = errors.New("download not found")
)

// MobileAPIServiceImpl implements the MobileAPIService interface
type MobileAPIServiceImpl struct {
	db             *sql.DB
	jwtSecret      []byte
	cdnBaseURL     string
	maxDevices     int
	tokenExpiry    time.Duration
	refreshExpiry  time.Duration
	downloadExpiry time.Duration
}

// NewMobileAPIService creates a new mobile API service instance
func NewMobileAPIService(db *sql.DB, jwtSecret string, cdnBaseURL string) *MobileAPIServiceImpl {
	return &MobileAPIServiceImpl{
		db:             db,
		jwtSecret:      []byte(jwtSecret),
		cdnBaseURL:     cdnBaseURL,
		maxDevices:     5,
		tokenExpiry:    24 * time.Hour,
		refreshExpiry:  30 * 24 * time.Hour,
		downloadExpiry: 7 * 24 * time.Hour,
	}
}

// ============================================================================
// Authentication
// ============================================================================

func (s *MobileAPIServiceImpl) MobileLogin(credentials *handler.MobileLoginRequest) (*handler.MobileAuthResponse, error) {
	// Get user from database
	var user handler.MobileUser
	var passwordHash string
	var subscriptionExpiry sql.NullTime

	query := `
		SELECT id, email, username, full_name, avatar, subscription_type,
		       subscription_expiry, max_devices, max_streams, password
		FROM users
		WHERE email = ? AND is_active = 1
	`
	err := s.db.QueryRow(query, credentials.Email).Scan(
		&user.ID, &user.Email, &user.Username, &user.FullName, &user.Avatar,
		&user.SubscriptionType, &subscriptionExpiry, &user.MaxDevices,
		&user.MaxStreams, &passwordHash,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(credentials.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if subscriptionExpiry.Valid {
		user.SubscriptionExpiry = subscriptionExpiry.Time
	}

	// Check subscription expiry
	if user.SubscriptionExpiry.Before(time.Now()) {
		return nil, errors.New("subscription expired")
	}

	// Count active devices
	var activeDevices int
	err = s.db.QueryRow("SELECT COUNT(*) FROM devices WHERE user_id = ? AND is_active = 1", user.ID).Scan(&activeDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to count devices: %w", err)
	}
	user.ActiveDevices = activeDevices

	// Register or update device
	_, err = s.RegisterDevice(&handler.DeviceRegistration{
		DeviceID:   credentials.DeviceID,
		DeviceName: credentials.DeviceName,
		Platform:   credentials.Platform,
		AppVersion: credentials.AppVersion,
		UserID:     user.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	// Generate JWT tokens
	accessToken, err := s.generateAccessToken(user.ID, credentials.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID, credentials.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	_, err = s.db.Exec(`
		INSERT INTO mobile_tokens (user_id, device_id, refresh_token, expires_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE refresh_token = VALUES(refresh_token), expires_at = VALUES(expires_at)
	`, user.ID, credentials.DeviceID, refreshToken, time.Now().Add(s.refreshExpiry))
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Update last login
	_, err = s.db.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", user.ID)
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Warning: failed to update last login: %v\n", err)
	}

	return &handler.MobileAuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.tokenExpiry),
		User:         user,
	}, nil
}

func (s *MobileAPIServiceImpl) RefreshToken(refreshToken string) (*handler.MobileAuthResponse, error) {
	// Validate refresh token
	claims := &MobileTokenClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if refresh token exists in database
	var storedToken string
	var expiresAt time.Time
	err = s.db.QueryRow(`
		SELECT refresh_token, expires_at FROM mobile_tokens
		WHERE user_id = ? AND device_id = ? AND refresh_token = ?
	`, claims.UserID, claims.DeviceID, refreshToken).Scan(&storedToken, &expiresAt)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if time.Now().After(expiresAt) {
		return nil, ErrInvalidToken
	}

	// Get user info
	var user handler.MobileUser
	var subscriptionExpiry sql.NullTime
	query := `
		SELECT id, email, username, full_name, avatar, subscription_type,
		       subscription_expiry, max_devices, max_streams
		FROM users WHERE id = ? AND is_active = 1
	`
	err = s.db.QueryRow(query, claims.UserID).Scan(
		&user.ID, &user.Email, &user.Username, &user.FullName, &user.Avatar,
		&user.SubscriptionType, &subscriptionExpiry, &user.MaxDevices, &user.MaxStreams,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if subscriptionExpiry.Valid {
		user.SubscriptionExpiry = subscriptionExpiry.Time
	}

	// Count active devices
	var activeDevices int
	s.db.QueryRow("SELECT COUNT(*) FROM devices WHERE user_id = ? AND is_active = 1", user.ID).Scan(&activeDevices)
	user.ActiveDevices = activeDevices

	// Generate new tokens
	newAccessToken, err := s.generateAccessToken(user.ID, claims.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user.ID, claims.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update refresh token
	_, err = s.db.Exec(`
		UPDATE mobile_tokens SET refresh_token = ?, expires_at = ?
		WHERE user_id = ? AND device_id = ?
	`, newRefreshToken, time.Now().Add(s.refreshExpiry), user.ID, claims.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to update refresh token: %w", err)
	}

	return &handler.MobileAuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(s.tokenExpiry),
		User:         user,
	}, nil
}

func (s *MobileAPIServiceImpl) MobileLogout(deviceID string, userID int64) error {
	// Delete refresh token
	_, err := s.db.Exec(`
		DELETE FROM mobile_tokens WHERE user_id = ? AND device_id = ?
	`, userID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}

	// Mark device as inactive
	_, err = s.db.Exec(`
		UPDATE devices SET is_active = 0 WHERE device_id = ? AND user_id = ?
	`, deviceID, userID)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	return nil
}

// ============================================================================
// Device Management
// ============================================================================

func (s *MobileAPIServiceImpl) RegisterDevice(device *handler.DeviceRegistration) (*handler.Device, error) {
	// Check device limit
	var activeDevices int
	err := s.db.QueryRow("SELECT COUNT(*) FROM devices WHERE user_id = ? AND is_active = 1", device.UserID).Scan(&activeDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to count devices: %w", err)
	}

	// Get user's max devices
	var maxDevices int
	err = s.db.QueryRow("SELECT max_devices FROM users WHERE id = ?", device.UserID).Scan(&maxDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to get max devices: %w", err)
	}

	// Check if device already exists
	var existingID int64
	err = s.db.QueryRow("SELECT id FROM devices WHERE device_id = ? AND user_id = ?", device.DeviceID, device.UserID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// New device - check limit
		if activeDevices >= maxDevices {
			return nil, ErrDeviceLimitReached
		}

		// Insert new device
		result, err := s.db.Exec(`
			INSERT INTO devices (user_id, device_id, device_name, platform, os_version, app_version, last_active, is_active, created_at)
			VALUES (?, ?, ?, ?, ?, ?, NOW(), 1, NOW())
		`, device.UserID, device.DeviceID, device.DeviceName, device.Platform, device.OSVersion, device.AppVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to insert device: %w", err)
		}

		existingID, _ = result.LastInsertId()
	} else {
		// Update existing device
		_, err = s.db.Exec(`
			UPDATE devices SET device_name = ?, os_version = ?, app_version = ?, last_active = NOW(), is_active = 1
			WHERE id = ?
		`, device.DeviceName, device.OSVersion, device.AppVersion, existingID)
		if err != nil {
			return nil, fmt.Errorf("failed to update device: %w", err)
		}
	}

	// Return device info
	return s.getDeviceByID(existingID)
}

func (s *MobileAPIServiceImpl) UpdateDevice(deviceID string, updates *handler.DeviceUpdate) error {
	var setParts []string
	var args []interface{}

	if updates.DeviceName != nil {
		setParts = append(setParts, "device_name = ?")
		args = append(args, *updates.DeviceName)
	}
	if updates.OSVersion != nil {
		setParts = append(setParts, "os_version = ?")
		args = append(args, *updates.OSVersion)
	}
	if updates.AppVersion != nil {
		setParts = append(setParts, "app_version = ?")
		args = append(args, *updates.AppVersion)
	}

	if len(setParts) == 0 {
		return nil // Nothing to update
	}

	setParts = append(setParts, "last_active = NOW()")
	args = append(args, deviceID)

	query := fmt.Sprintf("UPDATE devices SET %s WHERE device_id = ?", strings.Join(setParts, ", "))
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrDeviceNotFound
	}

	return nil
}

func (s *MobileAPIServiceImpl) GetUserDevices(userID int64) ([]handler.Device, error) {
	query := `
		SELECT id, device_id, device_name, platform, os_version, app_version, last_active, is_active, created_at
		FROM devices WHERE user_id = ? ORDER BY last_active DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []handler.Device
	for rows.Next() {
		var device handler.Device
		err := rows.Scan(
			&device.ID, &device.DeviceID, &device.DeviceName, &device.Platform,
			&device.OSVersion, &device.AppVersion, &device.LastActive,
			&device.IsActive, &device.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (s *MobileAPIServiceImpl) RemoveDevice(deviceID string, userID int64) error {
	// Delete device and associated tokens
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM mobile_tokens WHERE device_id = ? AND user_id = ?", deviceID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}

	result, err := tx.Exec("DELETE FROM devices WHERE device_id = ? AND user_id = ?", deviceID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrDeviceNotFound
	}

	return tx.Commit()
}

// ============================================================================
// Streams
// ============================================================================

func (s *MobileAPIServiceImpl) GetMobileStreams(userID int64, filters handler.MobileStreamFilters) (*handler.MobileStreamsResponse, error) {
	// Build query
	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "s.is_active = 1")

	if filters.CategoryID != nil {
		whereClauses = append(whereClauses, "s.category_id = ?")
		args = append(args, *filters.CategoryID)
	}

	if filters.Search != "" {
		whereClauses = append(whereClauses, "(s.name LIKE ? OR s.description LIKE ?)")
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if filters.Type != "" {
		whereClauses = append(whereClauses, "s.type = ?")
		args = append(args, filters.Type)
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM streams s WHERE %s", whereClause)
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count streams: %w", err)
	}

	// Get streams
	if filters.Limit == 0 {
		filters.Limit = 50
	}

	query := fmt.Sprintf(`
		SELECT s.id, s.name, s.logo, s.category_id, s.type, s.is_live, s.qualities, s.epg_enabled,
		       COALESCE(f.id IS NOT NULL, 0) as is_favorite
		FROM streams s
		LEFT JOIN favorites f ON f.user_id = ? AND f.content_type = 'stream' AND f.content_id = s.id
		WHERE %s
		ORDER BY s.name ASC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append([]interface{}{userID}, args...)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query streams: %w", err)
	}
	defer rows.Close()

	var streams []handler.MobileStream
	for rows.Next() {
		var stream handler.MobileStream
		var qualitiesJSON string

		err := rows.Scan(
			&stream.ID, &stream.Name, &stream.Logo, &stream.CategoryID,
			&stream.Type, &stream.IsLive, &qualitiesJSON, &stream.EPGEnabled,
			&stream.IsFavorite,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stream: %w", err)
		}

		json.Unmarshal([]byte(qualitiesJSON), &stream.Qualities)
		streams = append(streams, stream)
	}

	// Get categories
	categories, err := s.getCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return &handler.MobileStreamsResponse{
		Streams:    streams,
		Categories: categories,
		Total:      total,
	}, nil
}

func (s *MobileAPIServiceImpl) GetStreamURL(streamID int64, deviceID string, quality string) (*handler.StreamURLResponse, error) {
	// Get stream info
	var streamURL, streamType string
	var qualities string
	err := s.db.QueryRow(`
		SELECT stream_url, stream_type, qualities FROM streams WHERE id = ? AND is_active = 1
	`, streamID).Scan(&streamURL, &streamType, &qualities)

	if err == sql.ErrNoRows {
		return nil, ErrContentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	// Generate secure token
	token, err := s.generateStreamToken(streamID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Build URL with token
	expiresAt := time.Now().Add(24 * time.Hour)
	var qualityList []string
	json.Unmarshal([]byte(qualities), &qualityList)

	// Use CDN URL if available
	finalURL := streamURL
	if s.cdnBaseURL != "" {
		finalURL = fmt.Sprintf("%s/stream/%d/%s?token=%s", s.cdnBaseURL, streamID, quality, token)
	}

	return &handler.StreamURLResponse{
		URL:             finalURL,
		Type:            streamType,
		Quality:         quality,
		ExpiresAt:       expiresAt,
		Token:           token,
		CDNEnabled:      s.cdnBaseURL != "",
		AdaptiveBitrate: len(qualityList) > 1,
	}, nil
}

// ============================================================================
// VOD & Series
// ============================================================================

func (s *MobileAPIServiceImpl) GetMobileVOD(userID int64, filters handler.VODFilters) (*handler.MobileVODResponse, error) {
	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "m.is_active = 1")

	if filters.CategoryID != nil {
		whereClauses = append(whereClauses, "m.category_id = ?")
		args = append(args, *filters.CategoryID)
	}

	if filters.Search != "" {
		whereClauses = append(whereClauses, "(m.title LIKE ? OR m.description LIKE ?)")
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if filters.Year != nil {
		whereClauses = append(whereClauses, "m.year = ?")
		args = append(args, *filters.Year)
	}

	if filters.Genre != "" {
		whereClauses = append(whereClauses, "m.genres LIKE ?")
		args = append(args, "%"+filters.Genre+"%")
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vod_movies m WHERE %s", whereClause)
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count movies: %w", err)
	}

	// Get movies
	if filters.Limit == 0 {
		filters.Limit = 50
	}

	query := fmt.Sprintf(`
		SELECT m.id, m.title, m.poster, m.backdrop, m.year, m.duration, m.rating, m.genres, m.description,
		       COALESCE(f.id IS NOT NULL, 0) as is_favorite,
		       COALESCE(wp.progress, 0) as progress
		FROM vod_movies m
		LEFT JOIN favorites f ON f.user_id = ? AND f.content_type = 'movie' AND f.content_id = m.id
		LEFT JOIN watch_progress wp ON wp.user_id = ? AND wp.content_type = 'movie' AND wp.content_id = m.id
		WHERE %s
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append([]interface{}{userID, userID}, args...)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}
	defer rows.Close()

	var movies []handler.MobileMovie
	for rows.Next() {
		var movie handler.MobileMovie
		var genresJSON string

		err := rows.Scan(
			&movie.ID, &movie.Title, &movie.Poster, &movie.Backdrop, &movie.Year,
			&movie.Duration, &movie.Rating, &genresJSON, &movie.Description,
			&movie.IsFavorite, &movie.Progress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie: %w", err)
		}

		json.Unmarshal([]byte(genresJSON), &movie.Genres)
		movies = append(movies, movie)
	}

	return &handler.MobileVODResponse{
		Movies: movies,
		Total:  total,
	}, nil
}

func (s *MobileAPIServiceImpl) GetMobileSeries(userID int64) ([]handler.MobileSeries, error) {
	query := `
		SELECT s.id, s.title, s.poster, s.backdrop, s.year, s.rating, s.genres, s.description,
		       s.season_count, s.episode_count,
		       COALESCE(f.id IS NOT NULL, 0) as is_favorite
		FROM vod_series s
		LEFT JOIN favorites f ON f.user_id = ? AND f.content_type = 'series' AND f.content_id = s.id
		WHERE s.is_active = 1
		ORDER BY s.created_at DESC
		LIMIT 100
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query series: %w", err)
	}
	defer rows.Close()

	var seriesList []handler.MobileSeries
	for rows.Next() {
		var series handler.MobileSeries
		var genresJSON string

		err := rows.Scan(
			&series.ID, &series.Title, &series.Poster, &series.Backdrop, &series.Year,
			&series.Rating, &genresJSON, &series.Description, &series.SeasonCount,
			&series.EpisodeCount, &series.IsFavorite,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan series: %w", err)
		}

		json.Unmarshal([]byte(genresJSON), &series.Genres)
		seriesList = append(seriesList, series)
	}

	return seriesList, nil
}

func (s *MobileAPIServiceImpl) GetEpisodes(seriesID int64) ([]handler.Episode, error) {
	query := `
		SELECT id, series_id, season_num, episode_num, title, description, thumbnail, duration, air_date
		FROM vod_episodes
		WHERE series_id = ?
		ORDER BY season_num ASC, episode_num ASC
	`

	rows, err := s.db.Query(query, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	defer rows.Close()

	var episodes []handler.Episode
	for rows.Next() {
		var episode handler.Episode
		err := rows.Scan(
			&episode.ID, &episode.SeriesID, &episode.SeasonNum, &episode.EpisodeNum,
			&episode.Title, &episode.Description, &episode.Thumbnail, &episode.Duration,
			&episode.AirDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episodes = append(episodes, episode)
	}

	return episodes, nil
}

// ============================================================================
// Favorites & Watchlist
// ============================================================================

func (s *MobileAPIServiceImpl) GetFavorites(userID int64) (*handler.FavoritesResponse, error) {
	response := &handler.FavoritesResponse{
		Streams: []handler.MobileStream{},
		Movies:  []handler.MobileMovie{},
		Series:  []handler.MobileSeries{},
	}

	// Get favorite streams
	streamQuery := `
		SELECT s.id, s.name, s.logo, s.category_id, s.type, s.is_live, s.qualities, s.epg_enabled, 1 as is_favorite
		FROM favorites f
		JOIN streams s ON s.id = f.content_id
		WHERE f.user_id = ? AND f.content_type = 'stream' AND s.is_active = 1
		ORDER BY f.created_at DESC
	`
	streamRows, err := s.db.Query(streamQuery, userID)
	if err == nil {
		defer streamRows.Close()
		for streamRows.Next() {
			var stream handler.MobileStream
			var qualitiesJSON string
			streamRows.Scan(&stream.ID, &stream.Name, &stream.Logo, &stream.CategoryID, &stream.Type, &stream.IsLive, &qualitiesJSON, &stream.EPGEnabled, &stream.IsFavorite)
			json.Unmarshal([]byte(qualitiesJSON), &stream.Qualities)
			response.Streams = append(response.Streams, stream)
		}
	}

	// Get favorite movies
	movieQuery := `
		SELECT m.id, m.title, m.poster, m.backdrop, m.year, m.duration, m.rating, m.genres, m.description, 1 as is_favorite, COALESCE(wp.progress, 0) as progress
		FROM favorites f
		JOIN vod_movies m ON m.id = f.content_id
		LEFT JOIN watch_progress wp ON wp.user_id = ? AND wp.content_type = 'movie' AND wp.content_id = m.id
		WHERE f.user_id = ? AND f.content_type = 'movie' AND m.is_active = 1
		ORDER BY f.created_at DESC
	`
	movieRows, err := s.db.Query(movieQuery, userID, userID)
	if err == nil {
		defer movieRows.Close()
		for movieRows.Next() {
			var movie handler.MobileMovie
			var genresJSON string
			movieRows.Scan(&movie.ID, &movie.Title, &movie.Poster, &movie.Backdrop, &movie.Year, &movie.Duration, &movie.Rating, &genresJSON, &movie.Description, &movie.IsFavorite, &movie.Progress)
			json.Unmarshal([]byte(genresJSON), &movie.Genres)
			response.Movies = append(response.Movies, movie)
		}
	}

	// Get favorite series
	seriesQuery := `
		SELECT s.id, s.title, s.poster, s.backdrop, s.year, s.rating, s.genres, s.description, s.season_count, s.episode_count, 1 as is_favorite
		FROM favorites f
		JOIN vod_series s ON s.id = f.content_id
		WHERE f.user_id = ? AND f.content_type = 'series' AND s.is_active = 1
		ORDER BY f.created_at DESC
	`
	seriesRows, err := s.db.Query(seriesQuery, userID)
	if err == nil {
		defer seriesRows.Close()
		for seriesRows.Next() {
			var series handler.MobileSeries
			var genresJSON string
			seriesRows.Scan(&series.ID, &series.Title, &series.Poster, &series.Backdrop, &series.Year, &series.Rating, &genresJSON, &series.Description, &series.SeasonCount, &series.EpisodeCount, &series.IsFavorite)
			json.Unmarshal([]byte(genresJSON), &series.Genres)
			response.Series = append(response.Series, series)
		}
	}

	return response, nil
}

func (s *MobileAPIServiceImpl) AddFavorite(userID int64, contentType string, contentID int64) error {
	_, err := s.db.Exec(`
		INSERT INTO favorites (user_id, content_type, content_id, created_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE created_at = NOW()
	`, userID, contentType, contentID)

	if err != nil {
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	return nil
}

func (s *MobileAPIServiceImpl) RemoveFavorite(userID int64, contentType string, contentID int64) error {
	result, err := s.db.Exec(`
		DELETE FROM favorites WHERE user_id = ? AND content_type = ? AND content_id = ?
	`, userID, contentType, contentID)

	if err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrContentNotFound
	}

	return nil
}

func (s *MobileAPIServiceImpl) GetWatchlist(userID int64) ([]handler.WatchlistItem, error) {
	query := `
		SELECT content_type, content_id, created_at
		FROM watchlist
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query watchlist: %w", err)
	}
	defer rows.Close()

	var items []handler.WatchlistItem
	for rows.Next() {
		var item handler.WatchlistItem
		rows.Scan(&item.ContentType, &item.ContentID, &item.AddedAt)
		// TODO: Load actual content based on type
		items = append(items, item)
	}

	return items, nil
}

// ============================================================================
// Continue Watching
// ============================================================================

func (s *MobileAPIServiceImpl) GetContinueWatching(userID int64) ([]handler.ContinueWatchingItem, error) {
	query := `
		SELECT content_type, content_id, progress, updated_at
		FROM watch_progress
		WHERE user_id = ? AND progress > 5 AND progress < 95
		ORDER BY updated_at DESC
		LIMIT 20
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query continue watching: %w", err)
	}
	defer rows.Close()

	var items []handler.ContinueWatchingItem
	for rows.Next() {
		var item handler.ContinueWatchingItem
		rows.Scan(&item.ContentType, &item.ContentID, &item.Progress, &item.UpdatedAt)
		// TODO: Load actual content based on type
		items = append(items, item)
	}

	return items, nil
}

func (s *MobileAPIServiceImpl) UpdateWatchProgress(progress *handler.WatchProgress) error {
	_, err := s.db.Exec(`
		INSERT INTO watch_progress (user_id, content_type, content_id, progress, current_time, duration, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			progress = VALUES(progress),
			current_time = VALUES(current_time),
			duration = VALUES(duration),
			updated_at = NOW()
	`, progress.UserID, progress.ContentType, progress.ContentID, progress.Progress, progress.CurrentTime, progress.Duration)

	if err != nil {
		return fmt.Errorf("failed to update watch progress: %w", err)
	}

	return nil
}

// ============================================================================
// Push Notifications
// ============================================================================

func (s *MobileAPIServiceImpl) RegisterPushToken(token *handler.PushToken) error {
	_, err := s.db.Exec(`
		INSERT INTO push_tokens (user_id, device_id, token, platform, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE token = VALUES(token), updated_at = NOW()
	`, token.UserID, token.DeviceID, token.Token, token.Platform)

	if err != nil {
		return fmt.Errorf("failed to register push token: %w", err)
	}

	return nil
}

func (s *MobileAPIServiceImpl) UpdateNotificationSettings(userID int64, settings *handler.NotificationSettings) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO notification_settings (user_id, settings, updated_at)
		VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE settings = VALUES(settings), updated_at = NOW()
	`, userID, settingsJSON)

	if err != nil {
		return fmt.Errorf("failed to update notification settings: %w", err)
	}

	return nil
}

func (s *MobileAPIServiceImpl) GetNotifications(userID int64, limit int) ([]handler.Notification, error) {
	if limit == 0 {
		limit = 50
	}

	query := `
		SELECT id, user_id, type, title, message, image_url, action_url, data, is_read, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []handler.Notification
	for rows.Next() {
		var notif handler.Notification
		var imageURL, actionURL, data sql.NullString

		err := rows.Scan(
			&notif.ID, &notif.UserID, &notif.Type, &notif.Title, &notif.Message,
			&imageURL, &actionURL, &data, &notif.IsRead, &notif.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if imageURL.Valid {
			notif.ImageURL = imageURL.String
		}
		if actionURL.Valid {
			notif.ActionURL = actionURL.String
		}
		if data.Valid {
			notif.Data = data.String
		}

		notifications = append(notifications, notif)
	}

	return notifications, nil
}

func (s *MobileAPIServiceImpl) MarkNotificationRead(notificationID int64) error {
	result, err := s.db.Exec(`
		UPDATE notifications SET is_read = 1 WHERE id = ?
	`, notificationID)

	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrContentNotFound
	}

	return nil
}

// ============================================================================
// Offline Downloads
// ============================================================================

func (s *MobileAPIServiceImpl) GetDownloads(deviceID string) ([]handler.Download, error) {
	query := `
		SELECT id, device_id, content_type, content_id, quality, status, progress, file_size, file_path, expires_at, created_at
		FROM downloads
		WHERE device_id = ? AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query downloads: %w", err)
	}
	defer rows.Close()

	var downloads []handler.Download
	for rows.Next() {
		var download handler.Download
		var expiresAt sql.NullTime

		err := rows.Scan(
			&download.ID, &download.DeviceID, &download.ContentType, &download.ContentID,
			&download.Quality, &download.Status, &download.Progress, &download.FileSize,
			&download.FilePath, &expiresAt, &download.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan download: %w", err)
		}

		if expiresAt.Valid {
			download.ExpiresAt = expiresAt.Time
		}

		// TODO: Load actual content
		downloads = append(downloads, download)
	}

	return downloads, nil
}

func (s *MobileAPIServiceImpl) RequestDownload(request *handler.DownloadRequest) (*handler.Download, error) {
	// Check if content exists
	var contentExists bool
	var contentTable string

	switch request.ContentType {
	case "movie":
		contentTable = "vod_movies"
	case "episode":
		contentTable = "vod_episodes"
	default:
		return nil, errors.New("invalid content type for download")
	}

	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ?)", contentTable)
	err := s.db.QueryRow(query, request.ContentID).Scan(&contentExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check content: %w", err)
	}

	if !contentExists {
		return nil, ErrContentNotFound
	}

	// Create download entry
	expiresAt := time.Now().Add(s.downloadExpiry)
	result, err := s.db.Exec(`
		INSERT INTO downloads (device_id, content_type, content_id, quality, status, progress, file_size, expires_at, created_at)
		VALUES (?, ?, ?, ?, 'pending', 0, 0, ?, NOW())
	`, request.DeviceID, request.ContentType, request.ContentID, request.Quality, expiresAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create download: %w", err)
	}

	downloadID, _ := result.LastInsertId()

	// TODO: Queue download job for background processing

	return &handler.Download{
		ID:          downloadID,
		DeviceID:    request.DeviceID,
		ContentType: request.ContentType,
		ContentID:   request.ContentID,
		Quality:     request.Quality,
		Status:      "pending",
		Progress:    0,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *MobileAPIServiceImpl) DeleteDownload(downloadID int64, deviceID string) error {
	// Get file path first
	var filePath sql.NullString
	err := s.db.QueryRow("SELECT file_path FROM downloads WHERE id = ? AND device_id = ?", downloadID, deviceID).Scan(&filePath)
	if err == sql.ErrNoRows {
		return ErrDownloadNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get download: %w", err)
	}

	// Delete database entry
	result, err := s.db.Exec("DELETE FROM downloads WHERE id = ? AND device_id = ?", downloadID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete download: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrDownloadNotFound
	}

	// TODO: Delete actual file if exists

	return nil
}

// ============================================================================
// EPG
// ============================================================================

func (s *MobileAPIServiceImpl) GetMobileEPG(channelIDs []int64, startTime time.Time, endTime time.Time) ([]handler.EPGEvent, error) {
	if len(channelIDs) == 0 {
		return []handler.EPGEvent{}, nil
	}

	// Build IN clause
	placeholders := make([]string, len(channelIDs))
	args := make([]interface{}, 0, len(channelIDs)+2)
	for i, id := range channelIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, startTime, endTime)

	query := fmt.Sprintf(`
		SELECT channel_id, title, description, start_time, end_time, category, icon
		FROM epg_events
		WHERE channel_id IN (%s) AND start_time >= ? AND end_time <= ?
		ORDER BY channel_id ASC, start_time ASC
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query EPG: %w", err)
	}
	defer rows.Close()

	var events []handler.EPGEvent
	for rows.Next() {
		var event handler.EPGEvent
		var description, category, icon sql.NullString

		err := rows.Scan(
			&event.ChannelID, &event.Title, &description,
			&event.StartTime, &event.EndTime, &category, &icon,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan EPG event: %w", err)
		}

		if description.Valid {
			event.Description = description.String
		}
		if category.Valid {
			event.Category = category.String
		}
		if icon.Valid {
			event.Icon = icon.String
		}

		events = append(events, event)
	}

	return events, nil
}

// ============================================================================
// User Profile
// ============================================================================

func (s *MobileAPIServiceImpl) GetMobileProfile(userID int64) (*handler.MobileProfile, error) {
	var profile handler.MobileProfile
	var subscriptionExpiry sql.NullTime
	var avatar sql.NullString

	query := `
		SELECT id, email, username, full_name, avatar, phone, country,
		       subscription_type, subscription_expiry, max_devices, max_streams,
		       created_at, last_login
		FROM users WHERE id = ?
	`

	err := s.db.QueryRow(query, userID).Scan(
		&profile.ID, &profile.Email, &profile.Username, &profile.FullName,
		&avatar, &profile.Phone, &profile.Country,
		&profile.SubscriptionType, &subscriptionExpiry,
		&profile.MaxDevices, &profile.MaxStreams,
		&profile.CreatedAt, &profile.LastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if avatar.Valid {
		profile.Avatar = avatar.String
	}
	if subscriptionExpiry.Valid {
		profile.SubscriptionExpiry = subscriptionExpiry.Time
	}

	// Get active devices count
	s.db.QueryRow("SELECT COUNT(*) FROM devices WHERE user_id = ? AND is_active = 1", userID).Scan(&profile.ActiveDevices)

	return &profile, nil
}

func (s *MobileAPIServiceImpl) UpdateMobileProfile(userID int64, updates *handler.ProfileUpdate) error {
	var setParts []string
	var args []interface{}

	if updates.FullName != nil {
		setParts = append(setParts, "full_name = ?")
		args = append(args, *updates.FullName)
	}
	if updates.Avatar != nil {
		setParts = append(setParts, "avatar = ?")
		args = append(args, *updates.Avatar)
	}
	if updates.Phone != nil {
		setParts = append(setParts, "phone = ?")
		args = append(args, *updates.Phone)
	}
	if updates.Country != nil {
		setParts = append(setParts, "country = ?")
		args = append(args, *updates.Country)
	}

	if len(setParts) == 0 {
		return nil
	}

	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(setParts, ", "))

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ============================================================================
// App Configuration
// ============================================================================

func (s *MobileAPIServiceImpl) GetAppConfig(platform string, version string) (*handler.AppConfig, error) {
	config := &handler.AppConfig{
		MinVersion:       "1.0.0",
		LatestVersion:    "2.0.0",
		ForceUpdate:      false,
		MaintenanceMode:  false,
		Features:         make(map[string]bool),
		DownloadEnabled:  true,
		MaxDownloads:     10,
		DownloadExpiry:   7,
		StreamQualities:  []string{"sd", "hd", "fhd", "uhd"},
		CDNEnabled:       s.cdnBaseURL != "",
		CDNURL:           s.cdnBaseURL,
		SupportEmail:     "support@iptv.example.com",
		SupportPhone:     "+1-555-0123",
		TermsURL:         "https://iptv.example.com/terms",
		PrivacyURL:       "https://iptv.example.com/privacy",
	}

	// Get config from database if exists
	var configJSON string
	err := s.db.QueryRow(`
		SELECT config FROM app_config WHERE platform = ? ORDER BY created_at DESC LIMIT 1
	`, platform).Scan(&configJSON)

	if err == nil {
		json.Unmarshal([]byte(configJSON), config)
	}

	return config, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

type MobileTokenClaims struct {
	UserID   int64  `json:"user_id"`
	DeviceID string `json:"device_id"`
	jwt.StandardClaims
}

func (s *MobileAPIServiceImpl) generateAccessToken(userID int64, deviceID string) (string, error) {
	claims := &MobileTokenClaims{
		UserID:   userID,
		DeviceID: deviceID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenExpiry).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *MobileAPIServiceImpl) generateRefreshToken(userID int64, deviceID string) (string, error) {
	claims := &MobileTokenClaims{
		UserID:   userID,
		DeviceID: deviceID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.refreshExpiry).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *MobileAPIServiceImpl) generateStreamToken(streamID int64, deviceID string) (string, error) {
	data := fmt.Sprintf("%d:%s:%d", streamID, deviceID, time.Now().Unix())
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store token in cache/database
	_, err := s.db.Exec(`
		INSERT INTO stream_tokens (stream_id, device_id, token, expires_at)
		VALUES (?, ?, ?, ?)
	`, streamID, deviceID, token, time.Now().Add(24*time.Hour))

	return token, err
}

func (s *MobileAPIServiceImpl) getDeviceByID(deviceID int64) (*handler.Device, error) {
	var device handler.Device
	query := `
		SELECT id, device_id, device_name, platform, os_version, app_version, last_active, is_active, created_at
		FROM devices WHERE id = ?
	`
	err := s.db.QueryRow(query, deviceID).Scan(
		&device.ID, &device.DeviceID, &device.DeviceName, &device.Platform,
		&device.OSVersion, &device.AppVersion, &device.LastActive,
		&device.IsActive, &device.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return &device, nil
}

func (s *MobileAPIServiceImpl) getCategories() ([]handler.Category, error) {
	query := `
		SELECT c.id, c.name, c.icon, COUNT(s.id) as stream_count
		FROM categories c
		LEFT JOIN streams s ON s.category_id = c.id AND s.is_active = 1
		GROUP BY c.id, c.name, c.icon
		ORDER BY c.name ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []handler.Category
	for rows.Next() {
		var category handler.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Icon, &category.StreamCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}
