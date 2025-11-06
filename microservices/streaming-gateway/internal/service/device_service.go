package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Device types
const (
	DeviceTypeMAG      = "mag"
	DeviceTypeEnigma2  = "enigma2"
	DeviceTypeAndroid  = "android"
	DeviceTypeIOS      = "ios"
	DeviceTypeWeb      = "web"
	DeviceTypeSTB      = "stb"
	DeviceTypeSmartTV  = "smart_tv"
)

// Device errors
var (
	ErrDeviceNotFound      = errors.New("device not found")
	ErrDuplicateMAC        = errors.New("device with this MAC address already exists")
	ErrInvalidMACAddress   = errors.New("invalid MAC address format")
	ErrDeviceLimitReached  = errors.New("user has reached maximum device limit")
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidDeviceType   = errors.New("invalid device type")
)

// DeviceService handles device management business logic
type DeviceService struct {
	db *sql.DB
}

// NewDeviceService creates a new device service
func NewDeviceService(db *sql.DB) *DeviceService {
	return &DeviceService{
		db: db,
	}
}

// Device model
type Device struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	DeviceName    string     `json:"device_name"`
	DeviceType    string     `json:"device_type"`
	MACAddress    string     `json:"mac_address"`
	IPAddress     *string    `json:"ip_address,omitempty"`
	UserAgent     *string    `json:"user_agent,omitempty"`
	Model         *string    `json:"model,omitempty"`
	OSVersion     *string    `json:"os_version,omitempty"`
	AppVersion    *string    `json:"app_version,omitempty"`
	IsActive      bool       `json:"is_active"`
	IsBlocked     bool       `json:"is_blocked"`
	LastSeenAt    *time.Time `json:"last_seen_at,omitempty"`
	ActivatedAt   *time.Time `json:"activated_at,omitempty"`
	BlockedAt     *time.Time `json:"blocked_at,omitempty"`
	BlockedReason *string    `json:"blocked_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// DeviceSession model
type DeviceSession struct {
	ID          int64      `json:"id"`
	DeviceID    int64      `json:"device_id"`
	UserID      int64      `json:"user_id"`
	StreamID    *int64     `json:"stream_id,omitempty"`
	IPAddress   string     `json:"ip_address"`
	UserAgent   string     `json:"user_agent"`
	StartedAt   time.Time  `json:"started_at"`
	LastPingAt  time.Time  `json:"last_ping_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	Duration    *int       `json:"duration,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// DeviceStats model
type DeviceStats struct {
	TotalDevices    int64            `json:"total_devices"`
	ActiveDevices   int64            `json:"active_devices"`
	BlockedDevices  int64            `json:"blocked_devices"`
	DevicesByType   map[string]int64 `json:"devices_by_type"`
	RecentDevices   int64            `json:"recent_devices"`   // Last 7 days
	ActiveSessions  int64            `json:"active_sessions"`
	TotalSessions   int64            `json:"total_sessions"`
}

// RegisterDeviceRequest represents device registration request
type RegisterDeviceRequest struct {
	UserID      int64   `json:"user_id" binding:"required"`
	DeviceName  string  `json:"device_name" binding:"required"`
	DeviceType  string  `json:"device_type" binding:"required"`
	MACAddress  string  `json:"mac_address" binding:"required"`
	IPAddress   *string `json:"ip_address,omitempty"`
	UserAgent   *string `json:"user_agent,omitempty"`
	Model       *string `json:"model,omitempty"`
	OSVersion   *string `json:"os_version,omitempty"`
	AppVersion  *string `json:"app_version,omitempty"`
}

// UpdateDeviceRequest represents device update request
type UpdateDeviceRequest struct {
	DeviceName *string `json:"device_name,omitempty"`
	IPAddress  *string `json:"ip_address,omitempty"`
	UserAgent  *string `json:"user_agent,omitempty"`
	Model      *string `json:"model,omitempty"`
	OSVersion  *string `json:"os_version,omitempty"`
	AppVersion *string `json:"app_version,omitempty"`
}

// BlockDeviceRequest represents device block request
type BlockDeviceRequest struct {
	Reason string `json:"reason"`
}

// CreateSessionRequest represents session creation request
type CreateSessionRequest struct {
	DeviceID  int64   `json:"device_id" binding:"required"`
	UserID    int64   `json:"user_id" binding:"required"`
	StreamID  *int64  `json:"stream_id,omitempty"`
	IPAddress string  `json:"ip_address" binding:"required"`
	UserAgent string  `json:"user_agent"`
}

// RegisterDevice registers a new device
func (s *DeviceService) RegisterDevice(ctx context.Context, req *RegisterDeviceRequest) (*Device, error) {
	// Validate device type
	if !s.isValidDeviceType(req.DeviceType) {
		return nil, ErrInvalidDeviceType
	}

	// Validate MAC address format
	if !s.isValidMACAddress(req.MACAddress) {
		return nil, ErrInvalidMACAddress
	}

	// Normalize MAC address (uppercase, remove separators)
	normalizedMAC := s.normalizeMACAddress(req.MACAddress)

	// Check for duplicate MAC address
	var existingID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM devices WHERE mac_address = $1
	`, normalizedMAC).Scan(&existingID)

	if err == nil {
		return nil, ErrDuplicateMAC
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("error checking duplicate MAC: %w", err)
	}

	// Check device limit per user (configurable, default 5)
	maxDevices := 5
	var currentDeviceCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM devices WHERE user_id = $1 AND is_active = TRUE
	`, req.UserID).Scan(&currentDeviceCount)

	if err != nil {
		return nil, fmt.Errorf("error checking device count: %w", err)
	}

	if currentDeviceCount >= maxDevices {
		return nil, ErrDeviceLimitReached
	}

	// Create device
	now := time.Now()
	device := &Device{
		UserID:      req.UserID,
		DeviceName:  req.DeviceName,
		DeviceType:  req.DeviceType,
		MACAddress:  normalizedMAC,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		Model:       req.Model,
		OSVersion:   req.OSVersion,
		AppVersion:  req.AppVersion,
		IsActive:    true,
		IsBlocked:   false,
		ActivatedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `
		INSERT INTO devices (
			user_id, device_name, device_type, mac_address,
			ip_address, user_agent, model, os_version, app_version,
			is_active, is_blocked, activated_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`

	err = s.db.QueryRowContext(ctx, query,
		device.UserID, device.DeviceName, device.DeviceType, device.MACAddress,
		device.IPAddress, device.UserAgent, device.Model, device.OSVersion, device.AppVersion,
		device.IsActive, device.IsBlocked, device.ActivatedAt, device.CreatedAt, device.UpdatedAt,
	).Scan(&device.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating device: %w", err)
	}

	return device, nil
}

// GetDeviceByID retrieves a device by ID
func (s *DeviceService) GetDeviceByID(ctx context.Context, id int64) (*Device, error) {
	device := &Device{}

	query := `
		SELECT id, user_id, device_name, device_type, mac_address,
			   ip_address, user_agent, model, os_version, app_version,
			   is_active, is_blocked, last_seen_at, activated_at, blocked_at,
			   blocked_reason, created_at, updated_at
		FROM devices
		WHERE id = $1
	`

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&device.ID, &device.UserID, &device.DeviceName, &device.DeviceType, &device.MACAddress,
		&device.IPAddress, &device.UserAgent, &device.Model, &device.OSVersion, &device.AppVersion,
		&device.IsActive, &device.IsBlocked, &device.LastSeenAt, &device.ActivatedAt, &device.BlockedAt,
		&device.BlockedReason, &device.CreatedAt, &device.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrDeviceNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error getting device: %w", err)
	}

	return device, nil
}

// ListDevices retrieves all devices with optional filters
func (s *DeviceService) ListDevices(ctx context.Context, userID *int64, deviceType *string, isActive *bool, isBlocked *bool, limit, offset int) ([]*Device, int64, error) {
	// Build query with filters
	whereClause := []string{}
	args := []interface{}{}
	argPos := 1

	if userID != nil {
		whereClause = append(whereClause, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, *userID)
		argPos++
	}

	if deviceType != nil {
		whereClause = append(whereClause, fmt.Sprintf("device_type = $%d", argPos))
		args = append(args, *deviceType)
		argPos++
	}

	if isActive != nil {
		whereClause = append(whereClause, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *isActive)
		argPos++
	}

	if isBlocked != nil {
		whereClause = append(whereClause, fmt.Sprintf("is_blocked = $%d", argPos))
		args = append(args, *isBlocked)
		argPos++
	}

	whereSQL := ""
	if len(whereClause) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClause, " AND ")
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM devices %s", whereSQL)
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting devices: %w", err)
	}

	// Get devices
	query := fmt.Sprintf(`
		SELECT id, user_id, device_name, device_type, mac_address,
			   ip_address, user_agent, model, os_version, app_version,
			   is_active, is_blocked, last_seen_at, activated_at, blocked_at,
			   blocked_reason, created_at, updated_at
		FROM devices
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing devices: %w", err)
	}
	defer rows.Close()

	devices := []*Device{}
	for rows.Next() {
		device := &Device{}
		err := rows.Scan(
			&device.ID, &device.UserID, &device.DeviceName, &device.DeviceType, &device.MACAddress,
			&device.IPAddress, &device.UserAgent, &device.Model, &device.OSVersion, &device.AppVersion,
			&device.IsActive, &device.IsBlocked, &device.LastSeenAt, &device.ActivatedAt, &device.BlockedAt,
			&device.BlockedReason, &device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning device: %w", err)
		}
		devices = append(devices, device)
	}

	return devices, total, nil
}

// UpdateDevice updates a device
func (s *DeviceService) UpdateDevice(ctx context.Context, id int64, req *UpdateDeviceRequest) (*Device, error) {
	// Check if device exists
	_, err := s.GetDeviceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.DeviceName != nil {
		updates = append(updates, fmt.Sprintf("device_name = $%d", argPos))
		args = append(args, *req.DeviceName)
		argPos++
	}

	if req.IPAddress != nil {
		updates = append(updates, fmt.Sprintf("ip_address = $%d", argPos))
		args = append(args, *req.IPAddress)
		argPos++
	}

	if req.UserAgent != nil {
		updates = append(updates, fmt.Sprintf("user_agent = $%d", argPos))
		args = append(args, *req.UserAgent)
		argPos++
	}

	if req.Model != nil {
		updates = append(updates, fmt.Sprintf("model = $%d", argPos))
		args = append(args, *req.Model)
		argPos++
	}

	if req.OSVersion != nil {
		updates = append(updates, fmt.Sprintf("os_version = $%d", argPos))
		args = append(args, *req.OSVersion)
		argPos++
	}

	if req.AppVersion != nil {
		updates = append(updates, fmt.Sprintf("app_version = $%d", argPos))
		args = append(args, *req.AppVersion)
		argPos++
	}

	if len(updates) == 0 {
		// Nothing to update
		return s.GetDeviceByID(ctx, id)
	}

	// Add updated_at
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, time.Now())
	argPos++

	// Add ID to args
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE devices
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argPos)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error updating device: %w", err)
	}

	return s.GetDeviceByID(ctx, id)
}

// UpdateLastSeen updates device last seen timestamp
func (s *DeviceService) UpdateLastSeen(ctx context.Context, id int64) error {
	query := `
		UPDATE devices
		SET last_seen_at = $1, updated_at = $1
		WHERE id = $2
	`

	_, err := s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error updating last seen: %w", err)
	}

	return nil
}

// DeleteDevice deletes a device (soft delete by setting is_active to false)
func (s *DeviceService) DeleteDevice(ctx context.Context, id int64) error {
	// Check if device exists
	_, err := s.GetDeviceByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete
	query := `
		UPDATE devices
		SET is_active = FALSE, updated_at = $1
		WHERE id = $2
	`

	_, err = s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error deleting device: %w", err)
	}

	// Also terminate any active sessions
	return s.TerminateDeviceSessions(ctx, id)
}

// BlockDevice blocks a device
func (s *DeviceService) BlockDevice(ctx context.Context, id int64, reason string) error {
	// Check if device exists
	_, err := s.GetDeviceByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	query := `
		UPDATE devices
		SET is_blocked = TRUE, blocked_at = $1, blocked_reason = $2, updated_at = $1
		WHERE id = $3
	`

	_, err = s.db.ExecContext(ctx, query, now, reason, id)
	if err != nil {
		return fmt.Errorf("error blocking device: %w", err)
	}

	// Terminate any active sessions
	return s.TerminateDeviceSessions(ctx, id)
}

// UnblockDevice unblocks a device
func (s *DeviceService) UnblockDevice(ctx context.Context, id int64) error {
	// Check if device exists
	_, err := s.GetDeviceByID(ctx, id)
	if err != nil {
		return err
	}

	query := `
		UPDATE devices
		SET is_blocked = FALSE, blocked_at = NULL, blocked_reason = NULL, updated_at = $1
		WHERE id = $2
	`

	_, err = s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error unblocking device: %w", err)
	}

	return nil
}

// CreateSession creates a new device session
func (s *DeviceService) CreateSession(ctx context.Context, req *CreateSessionRequest) (*DeviceSession, error) {
	// Verify device exists and is not blocked
	device, err := s.GetDeviceByID(ctx, req.DeviceID)
	if err != nil {
		return nil, err
	}

	if device.IsBlocked {
		return nil, errors.New("device is blocked")
	}

	// Update device last seen
	err = s.UpdateLastSeen(ctx, req.DeviceID)
	if err != nil {
		return nil, err
	}

	// Create session
	now := time.Now()
	session := &DeviceSession{
		DeviceID:   req.DeviceID,
		UserID:     req.UserID,
		StreamID:   req.StreamID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		StartedAt:  now,
		LastPingAt: now,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	query := `
		INSERT INTO device_sessions (
			device_id, user_id, stream_id, ip_address, user_agent,
			started_at, last_ping_at, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	err = s.db.QueryRowContext(ctx, query,
		session.DeviceID, session.UserID, session.StreamID, session.IPAddress, session.UserAgent,
		session.StartedAt, session.LastPingAt, session.IsActive, session.CreatedAt, session.UpdatedAt,
	).Scan(&session.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	return session, nil
}

// GetDeviceSessions retrieves sessions for a device
func (s *DeviceService) GetDeviceSessions(ctx context.Context, deviceID int64, activeOnly bool, limit, offset int) ([]*DeviceSession, int64, error) {
	// Build query
	whereClause := "device_id = $1"
	args := []interface{}{deviceID}

	if activeOnly {
		whereClause += " AND is_active = TRUE"
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM device_sessions WHERE %s", whereClause)
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting sessions: %w", err)
	}

	// Get sessions
	query := fmt.Sprintf(`
		SELECT id, device_id, user_id, stream_id, ip_address, user_agent,
			   started_at, last_ping_at, ended_at, duration, is_active,
			   created_at, updated_at
		FROM device_sessions
		WHERE %s
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`, whereClause)

	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*DeviceSession{}
	for rows.Next() {
		session := &DeviceSession{}
		err := rows.Scan(
			&session.ID, &session.DeviceID, &session.UserID, &session.StreamID, &session.IPAddress, &session.UserAgent,
			&session.StartedAt, &session.LastPingAt, &session.EndedAt, &session.Duration, &session.IsActive,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

// ListActiveSessions retrieves all active sessions
func (s *DeviceService) ListActiveSessions(ctx context.Context, limit, offset int) ([]*DeviceSession, int64, error) {
	// Get total count
	var total int64
	countQuery := "SELECT COUNT(*) FROM device_sessions WHERE is_active = TRUE"
	err := s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting active sessions: %w", err)
	}

	// Get sessions
	query := `
		SELECT id, device_id, user_id, stream_id, ip_address, user_agent,
			   started_at, last_ping_at, ended_at, duration, is_active,
			   created_at, updated_at
		FROM device_sessions
		WHERE is_active = TRUE
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing active sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*DeviceSession{}
	for rows.Next() {
		session := &DeviceSession{}
		err := rows.Scan(
			&session.ID, &session.DeviceID, &session.UserID, &session.StreamID, &session.IPAddress, &session.UserAgent,
			&session.StartedAt, &session.LastPingAt, &session.EndedAt, &session.Duration, &session.IsActive,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

// UpdateSessionPing updates session last ping timestamp
func (s *DeviceService) UpdateSessionPing(ctx context.Context, sessionID int64) error {
	query := `
		UPDATE device_sessions
		SET last_ping_at = $1, updated_at = $1
		WHERE id = $2 AND is_active = TRUE
	`

	result, err := s.db.ExecContext(ctx, query, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("error updating session ping: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// TerminateSession terminates a session
func (s *DeviceService) TerminateSession(ctx context.Context, sessionID int64) error {
	now := time.Now()
	query := `
		UPDATE device_sessions
		SET is_active = FALSE, ended_at = $1, updated_at = $1
		WHERE id = $2 AND is_active = TRUE
	`

	result, err := s.db.ExecContext(ctx, query, now, sessionID)
	if err != nil {
		return fmt.Errorf("error terminating session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// TerminateDeviceSessions terminates all active sessions for a device
func (s *DeviceService) TerminateDeviceSessions(ctx context.Context, deviceID int64) error {
	now := time.Now()
	query := `
		UPDATE device_sessions
		SET is_active = FALSE, ended_at = $1, updated_at = $1
		WHERE device_id = $2 AND is_active = TRUE
	`

	_, err := s.db.ExecContext(ctx, query, now, deviceID)
	if err != nil {
		return fmt.Errorf("error terminating device sessions: %w", err)
	}

	return nil
}

// GetDeviceStats retrieves device statistics
func (s *DeviceService) GetDeviceStats(ctx context.Context) (*DeviceStats, error) {
	stats := &DeviceStats{
		DevicesByType: make(map[string]int64),
	}

	// Total devices
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices WHERE is_active = TRUE").Scan(&stats.TotalDevices)
	if err != nil {
		return nil, fmt.Errorf("error counting total devices: %w", err)
	}

	// Active devices (seen in last 24 hours)
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM devices
		WHERE is_active = TRUE AND last_seen_at > NOW() - INTERVAL '24 hours'
	`).Scan(&stats.ActiveDevices)
	if err != nil {
		return nil, fmt.Errorf("error counting active devices: %w", err)
	}

	// Blocked devices
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices WHERE is_blocked = TRUE").Scan(&stats.BlockedDevices)
	if err != nil {
		return nil, fmt.Errorf("error counting blocked devices: %w", err)
	}

	// Recent devices (last 7 days)
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM devices
		WHERE created_at > NOW() - INTERVAL '7 days'
	`).Scan(&stats.RecentDevices)
	if err != nil {
		return nil, fmt.Errorf("error counting recent devices: %w", err)
	}

	// Devices by type
	rows, err := s.db.QueryContext(ctx, `
		SELECT device_type, COUNT(*)
		FROM devices
		WHERE is_active = TRUE
		GROUP BY device_type
	`)
	if err != nil {
		return nil, fmt.Errorf("error counting devices by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var deviceType string
		var count int64
		if err := rows.Scan(&deviceType, &count); err != nil {
			return nil, fmt.Errorf("error scanning device type count: %w", err)
		}
		stats.DevicesByType[deviceType] = count
	}

	// Active sessions
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM device_sessions WHERE is_active = TRUE").Scan(&stats.ActiveSessions)
	if err != nil {
		return nil, fmt.Errorf("error counting active sessions: %w", err)
	}

	// Total sessions
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM device_sessions").Scan(&stats.TotalSessions)
	if err != nil {
		return nil, fmt.Errorf("error counting total sessions: %w", err)
	}

	return stats, nil
}

// Helper functions

// isValidDeviceType checks if device type is valid
func (s *DeviceService) isValidDeviceType(deviceType string) bool {
	validTypes := []string{
		DeviceTypeMAG,
		DeviceTypeEnigma2,
		DeviceTypeAndroid,
		DeviceTypeIOS,
		DeviceTypeWeb,
		DeviceTypeSTB,
		DeviceTypeSmartTV,
	}

	for _, vt := range validTypes {
		if deviceType == vt {
			return true
		}
	}

	return false
}

// isValidMACAddress validates MAC address format
func (s *DeviceService) isValidMACAddress(mac string) bool {
	// Supports formats: XX:XX:XX:XX:XX:XX, XX-XX-XX-XX-XX-XX, XXXXXXXXXXXX
	macPattern := `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})|([0-9A-Fa-f]{12})$`
	matched, err := regexp.MatchString(macPattern, mac)
	if err != nil {
		return false
	}
	return matched
}

// normalizeMACAddress normalizes MAC address to uppercase without separators
func (s *DeviceService) normalizeMACAddress(mac string) string {
	// Remove separators
	normalized := strings.ReplaceAll(mac, ":", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	// Convert to uppercase
	return strings.ToUpper(normalized)
}
