package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FraudDetectionService detects and prevents fraudulent activities
type FraudDetectionService struct {
	db          *sql.DB
	geoService  *GeoIPService
	thresholds  *FraudThresholds
}

// NewFraudDetectionService creates a new fraud detection service
func NewFraudDetectionService(db *sql.DB, geoService *GeoIPService) *FraudDetectionService {
	return &FraudDetectionService{
		db:         db,
		geoService: geoService,
		thresholds: defaultFraudThresholds(),
	}
}

// =====================================================
// MODELS
// =====================================================

// FraudThresholds defines detection sensitivity
type FraudThresholds struct {
	MaxConcurrentDevices  int     // Max devices per user
	MaxConcurrentStreams  int     // Max simultaneous streams
	MaxCountriesPerHour   int     // Max country switches per hour
	GeoVelocityKMH        float64 // Max travel speed (impossible velocity)
	MaxFailedLogins       int     // Max failed logins before lockout
	SuspiciousUserAgents  []string
	BannedASNs           []int
	VPNDetectionEnabled   bool
}

// FraudCheck represents a fraud detection result
type FraudCheck struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	IPAddress     string                 `json:"ip_address"`
	DeviceID      string                 `json:"device_id"`
	UserAgent     string                 `json:"user_agent"`
	RiskScore     float64                `json:"risk_score"` // 0-100
	Decision      string                 `json:"decision"`   // allow, challenge, block
	Reasons       []string               `json:"reasons"`
	Country       string                 `json:"country"`
	IsVPN         bool                   `json:"is_vpn"`
	IsProxy       bool                   `json:"is_proxy"`
	IsTor         bool                   `json:"is_tor"`
	IsDatacenter  bool                   `json:"is_datacenter"`
	Fingerprint   string                 `json:"fingerprint"`
	Metadata      map[string]interface{} `json:"metadata"`
	CheckedAt     time.Time              `json:"checked_at"`
}

// DeviceFingerprint represents a unique device identifier
type DeviceFingerprint struct {
	UserID       string    `json:"user_id"`
	Fingerprint  string    `json:"fingerprint"`
	Components   string    `json:"components"` // JSON of fingerprint components
	UserAgent    string    `json:"user_agent"`
	ScreenSize   string    `json:"screen_size"`
	Timezone     string    `json:"timezone"`
	Language     string    `json:"language"`
	Plugins      string    `json:"plugins"`
	IPAddress    string    `json:"ip_address"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	UseCount     int       `json:"use_count"`
	IsTrusted    bool      `json:"is_trusted"`
}

// SuspiciousActivity represents a detected suspicious activity
type SuspiciousActivity struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ActivityType  string    `json:"activity_type"`
	Description   string    `json:"description"`
	Severity      string    `json:"severity"` // low, medium, high, critical
	IPAddress     string    `json:"ip_address"`
	DeviceID      string    `json:"device_id"`
	Action        string    `json:"action"` // logged, flagged, blocked, suspended
	DetectedAt    time.Time `json:"detected_at"`
}

// LoginAttempt tracks login attempts
type LoginAttempt struct {
	IPAddress    string
	UserID       string
	Success      bool
	UserAgent    string
	Country      string
	AttemptedAt  time.Time
}

// =====================================================
// MAIN FRAUD CHECK
// =====================================================

// CheckFraud performs comprehensive fraud detection
func (s *FraudDetectionService) CheckFraud(ctx context.Context, userID, ipAddress, deviceID, userAgent string) (*FraudCheck, error) {
	check := &FraudCheck{
		ID:        uuid.New().String(),
		UserID:    userID,
		IPAddress: ipAddress,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		RiskScore: 0,
		Reasons:   []string{},
		Metadata:  make(map[string]interface{}),
		CheckedAt: time.Now(),
	}

	// 1. IP reputation checks
	s.checkIPReputation(ctx, check)

	// 2. Geo-velocity check (impossible travel)
	s.checkGeoVelocity(ctx, check)

	// 3. Device fingerprinting
	s.checkDeviceFingerprint(ctx, check)

	// 4. Concurrent stream check
	s.checkConcurrentStreams(ctx, check)

	// 5. Account sharing detection
	s.checkAccountSharing(ctx, check)

	// 6. Bot detection
	s.checkBotBehavior(ctx, check)

	// 7. Rate limiting check
	s.checkRateLimit(ctx, check)

	// 8. Known bad actors
	s.checkBlocklist(ctx, check)

	// Calculate final decision
	check.Decision = s.makeDecision(check.RiskScore)

	// Log the check
	s.logFraudCheck(ctx, check)

	// Take action if needed
	if check.Decision == "block" {
		s.executeBlockAction(ctx, userID, ipAddress, check.Reasons)
	} else if check.Decision == "challenge" {
		s.flagForReview(ctx, check)
	}

	return check, nil
}

// =====================================================
// INDIVIDUAL CHECKS
// =====================================================

// checkIPReputation checks IP against various blacklists
func (s *FraudDetectionService) checkIPReputation(ctx context.Context, check *FraudCheck) {
	// Get geo info
	if s.geoService != nil {
		country, _, err := s.geoService.LookupCountry(check.IPAddress)
		if err == nil {
			check.Country = country
		}
	}

	// Check if IP is from a datacenter (likely VPN/proxy)
	if s.isDatacenterIP(check.IPAddress) {
		check.IsDatacenter = true
		check.RiskScore += 15
		check.Reasons = append(check.Reasons, "datacenter_ip")
	}

	// Check if VPN
	if s.thresholds.VPNDetectionEnabled && s.isVPN(check.IPAddress) {
		check.IsVPN = true
		check.RiskScore += 20
		check.Reasons = append(check.Reasons, "vpn_detected")
	}

	// Check Tor exit nodes
	if s.isTorExit(check.IPAddress) {
		check.IsTor = true
		check.RiskScore += 30
		check.Reasons = append(check.Reasons, "tor_exit_node")
	}

	// Check if IP is in our blocklist
	if s.isIPBlocked(ctx, check.IPAddress) {
		check.RiskScore += 50
		check.Reasons = append(check.Reasons, "ip_blocked")
	}
}

// checkGeoVelocity detects impossible travel patterns
func (s *FraudDetectionService) checkGeoVelocity(ctx context.Context, check *FraudCheck) {
	// Get user's last known location
	var lastLat, lastLng float64
	var lastTime time.Time
	query := `
		SELECT latitude, longitude, last_seen
		FROM user_locations
		WHERE user_id = ?
		ORDER BY last_seen DESC
		LIMIT 1
	`
	err := s.db.QueryRowContext(ctx, query, check.UserID).Scan(&lastLat, &lastLng, &lastTime)
	if err != nil || s.geoService == nil {
		return
	}

	// Get current location
	currentLat, currentLng, err := s.geoService.LookupCoordinates(check.IPAddress)
	if err != nil {
		return
	}

	// Calculate distance and time
	distance := haversineDistance(lastLat, lastLng, currentLat, currentLng)
	timeHours := time.Since(lastTime).Hours()

	if timeHours > 0 {
		velocity := distance / timeHours
		if velocity > s.thresholds.GeoVelocityKMH {
			check.RiskScore += 35
			check.Reasons = append(check.Reasons, fmt.Sprintf("impossible_travel_%.0fkmh", velocity))
			check.Metadata["geo_velocity_kmh"] = velocity
			check.Metadata["distance_km"] = distance
		}
	}

	// Update last location
	s.updateUserLocation(ctx, check.UserID, currentLat, currentLng, check.IPAddress)
}

// checkDeviceFingerprint validates device fingerprint
func (s *FraudDetectionService) checkDeviceFingerprint(ctx context.Context, check *FraudCheck) {
	// Generate fingerprint from user agent and other components
	fingerprint := s.generateFingerprint(check.UserAgent, check.IPAddress, check.DeviceID)
	check.Fingerprint = fingerprint

	// Check if this fingerprint is shared across multiple users
	var sharedUserCount int
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM device_fingerprints
		WHERE fingerprint = ? AND user_id != ?
	`
	s.db.QueryRowContext(ctx, query, fingerprint, check.UserID).Scan(&sharedUserCount)

	if sharedUserCount > 5 {
		check.RiskScore += 25
		check.Reasons = append(check.Reasons, fmt.Sprintf("fingerprint_shared_%d_users", sharedUserCount))
	}

	// Check if fingerprint is new
	var isKnown bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM device_fingerprints WHERE user_id = ? AND fingerprint = ?)`
	s.db.QueryRowContext(ctx, checkQuery, check.UserID, fingerprint).Scan(&isKnown)

	if !isKnown {
		check.RiskScore += 5
		check.Reasons = append(check.Reasons, "new_device")
	}

	// Save/update fingerprint
	s.saveFingerprint(ctx, check.UserID, fingerprint, check.UserAgent, check.IPAddress)
}

// checkConcurrentStreams checks for too many concurrent streams
func (s *FraudDetectionService) checkConcurrentStreams(ctx context.Context, check *FraudCheck) {
	var activeStreams int
	query := `
		SELECT COUNT(DISTINCT device_id)
		FROM active_sessions
		WHERE user_id = ? AND last_activity > DATE_SUB(NOW(), INTERVAL 5 MINUTE)
	`
	s.db.QueryRowContext(ctx, query, check.UserID).Scan(&activeStreams)

	if activeStreams > s.thresholds.MaxConcurrentStreams {
		check.RiskScore += 20
		check.Reasons = append(check.Reasons, fmt.Sprintf("too_many_streams_%d", activeStreams))
		check.Metadata["concurrent_streams"] = activeStreams
	}
}

// checkAccountSharing detects account sharing patterns
func (s *FraudDetectionService) checkAccountSharing(ctx context.Context, check *FraudCheck) {
	// Count unique IPs used in last 24 hours
	var uniqueIPs, uniqueCountries int
	query := `
		SELECT
			COUNT(DISTINCT ip_address) as ips,
			COUNT(DISTINCT country) as countries
		FROM session_logs
		WHERE user_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL 24 HOUR)
	`
	s.db.QueryRowContext(ctx, query, check.UserID).Scan(&uniqueIPs, &uniqueCountries)

	if uniqueIPs > 10 {
		check.RiskScore += 15
		check.Reasons = append(check.Reasons, fmt.Sprintf("multiple_ips_%d", uniqueIPs))
	}

	if uniqueCountries > s.thresholds.MaxCountriesPerHour {
		check.RiskScore += 25
		check.Reasons = append(check.Reasons, fmt.Sprintf("multiple_countries_%d", uniqueCountries))
	}

	check.Metadata["unique_ips_24h"] = uniqueIPs
	check.Metadata["unique_countries_24h"] = uniqueCountries
}

// checkBotBehavior detects bot-like patterns
func (s *FraudDetectionService) checkBotBehavior(ctx context.Context, check *FraudCheck) {
	// Check for suspicious user agents
	for _, suspicious := range s.thresholds.SuspiciousUserAgents {
		if strings.Contains(strings.ToLower(check.UserAgent), suspicious) {
			check.RiskScore += 30
			check.Reasons = append(check.Reasons, "suspicious_user_agent")
			return
		}
	}

	// Empty or generic user agent
	if check.UserAgent == "" || len(check.UserAgent) < 20 {
		check.RiskScore += 15
		check.Reasons = append(check.Reasons, "minimal_user_agent")
	}

	// Check request patterns (too regular = bot)
	var avgInterval float64
	query := `
		SELECT AVG(TIMESTAMPDIFF(SECOND, prev_time, created_at))
		FROM (
			SELECT created_at, LAG(created_at) OVER (ORDER BY created_at) as prev_time
			FROM api_requests
			WHERE user_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)
		) t
		WHERE prev_time IS NOT NULL
	`
	s.db.QueryRowContext(ctx, query, check.UserID).Scan(&avgInterval)

	// If requests are very regular (within 1 second variance), likely a bot
	if avgInterval > 0 && avgInterval < 2.0 {
		check.RiskScore += 20
		check.Reasons = append(check.Reasons, "bot_like_pattern")
	}
}

// checkRateLimit checks for rate limit violations
func (s *FraudDetectionService) checkRateLimit(ctx context.Context, check *FraudCheck) {
	var requestCount int
	query := `
		SELECT COUNT(*)
		FROM api_requests
		WHERE (user_id = ? OR ip_address = ?)
		  AND created_at > DATE_SUB(NOW(), INTERVAL 1 MINUTE)
	`
	s.db.QueryRowContext(ctx, query, check.UserID, check.IPAddress).Scan(&requestCount)

	if requestCount > 100 {
		check.RiskScore += 20
		check.Reasons = append(check.Reasons, fmt.Sprintf("rate_limit_exceeded_%d", requestCount))
	}
}

// checkBlocklist checks user/IP against blocklist
func (s *FraudDetectionService) checkBlocklist(ctx context.Context, check *FraudCheck) {
	var inBlocklist bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM fraud_blocklist
			WHERE (user_id = ? OR ip_address = ? OR fingerprint = ?)
			  AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	s.db.QueryRowContext(ctx, query, check.UserID, check.IPAddress, check.Fingerprint).Scan(&inBlocklist)

	if inBlocklist {
		check.RiskScore += 100
		check.Reasons = append(check.Reasons, "in_blocklist")
	}
}

// =====================================================
// LOGIN ATTEMPT TRACKING
// =====================================================

// RecordLoginAttempt records a login attempt
func (s *FraudDetectionService) RecordLoginAttempt(ctx context.Context, attempt *LoginAttempt) error {
	query := `
		INSERT INTO login_attempts (ip_address, user_id, success, user_agent, country, attempted_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		attempt.IPAddress, attempt.UserID, attempt.Success,
		attempt.UserAgent, attempt.Country,
	)

	// Check failed login count
	if !attempt.Success {
		var failCount int
		countQuery := `
			SELECT COUNT(*) FROM login_attempts
			WHERE (user_id = ? OR ip_address = ?)
			  AND success = 0
			  AND attempted_at > DATE_SUB(NOW(), INTERVAL 15 MINUTE)
		`
		s.db.QueryRowContext(ctx, countQuery, attempt.UserID, attempt.IPAddress).Scan(&failCount)

		if failCount >= s.thresholds.MaxFailedLogins {
			s.lockAccount(ctx, attempt.UserID, attempt.IPAddress)
			return errors.New("account locked due to too many failed attempts")
		}
	}

	return err
}

// =====================================================
// ACTIONS
// =====================================================

func (s *FraudDetectionService) executeBlockAction(ctx context.Context, userID, ipAddress string, reasons []string) {
	reasonStr := strings.Join(reasons, ",")

	// Block the IP
	query := `
		INSERT INTO fraud_blocklist (id, user_id, ip_address, reason, expires_at, created_at)
		VALUES (?, ?, ?, ?, DATE_ADD(NOW(), INTERVAL 24 HOUR), NOW())
	`
	s.db.ExecContext(ctx, query, uuid.New().String(), userID, ipAddress, reasonStr)

	// Suspend user temporarily
	suspendQuery := `UPDATE users SET is_suspended = 1, suspension_reason = ? WHERE id = ?`
	s.db.ExecContext(ctx, suspendQuery, reasonStr, userID)

	// Log activity
	s.logSuspiciousActivity(ctx, userID, ipAddress, "blocked", reasonStr, "critical")
}

func (s *FraudDetectionService) flagForReview(ctx context.Context, check *FraudCheck) {
	reasonStr := strings.Join(check.Reasons, ",")
	s.logSuspiciousActivity(ctx, check.UserID, check.IPAddress, "flagged_for_review", reasonStr, "medium")
}

func (s *FraudDetectionService) lockAccount(ctx context.Context, userID, ipAddress string) {
	query := `
		UPDATE users SET
			is_locked = 1,
			locked_at = NOW(),
			locked_until = DATE_ADD(NOW(), INTERVAL 30 MINUTE)
		WHERE id = ?
	`
	s.db.ExecContext(ctx, query, userID)

	s.logSuspiciousActivity(ctx, userID, ipAddress, "account_locked",
		"Too many failed login attempts", "high")
}

// =====================================================
// HELPERS
// =====================================================

func (s *FraudDetectionService) makeDecision(riskScore float64) string {
	switch {
	case riskScore >= 70:
		return "block"
	case riskScore >= 40:
		return "challenge"
	default:
		return "allow"
	}
}

func (s *FraudDetectionService) generateFingerprint(userAgent, ipAddress, deviceID string) string {
	data := fmt.Sprintf("%s:%s:%s", userAgent, ipAddress, deviceID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *FraudDetectionService) isDatacenterIP(ip string) bool {
	// Check known datacenter IP ranges (simplified)
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	datacenterRanges := []string{
		"3.0.0.0/8",     // AWS
		"35.0.0.0/8",    // Google Cloud
		"104.196.0.0/14", // Google Cloud
		"13.64.0.0/11",  // Microsoft Azure
		"40.64.0.0/10",  // Microsoft Azure
		"167.172.0.0/16", // DigitalOcean
		"178.62.0.0/16",  // DigitalOcean
	}

	for _, cidr := range datacenterRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func (s *FraudDetectionService) isVPN(ip string) bool {
	// In production, integrate with services like:
	// - IPQualityScore, MaxMind, IP2Location
	// For now, check our local cache
	var isVPN bool
	query := `SELECT is_vpn FROM ip_intelligence WHERE ip_address = ?`
	s.db.QueryRow(query, ip).Scan(&isVPN)
	return isVPN
}

func (s *FraudDetectionService) isTorExit(ip string) bool {
	// Check against Tor exit node list (updated periodically)
	var isTor bool
	query := `SELECT EXISTS(SELECT 1 FROM tor_exit_nodes WHERE ip_address = ?)`
	s.db.QueryRow(query, ip).Scan(&isTor)
	return isTor
}

func (s *FraudDetectionService) isIPBlocked(ctx context.Context, ip string) bool {
	var blocked bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM fraud_blocklist
			WHERE ip_address = ? AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	s.db.QueryRowContext(ctx, query, ip).Scan(&blocked)
	return blocked
}

func (s *FraudDetectionService) updateUserLocation(ctx context.Context, userID string, lat, lng float64, ip string) {
	query := `
		INSERT INTO user_locations (user_id, latitude, longitude, ip_address, last_seen)
		VALUES (?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			latitude = VALUES(latitude),
			longitude = VALUES(longitude),
			last_seen = NOW()
	`
	s.db.ExecContext(ctx, query, userID, lat, lng, ip)
}

func (s *FraudDetectionService) saveFingerprint(ctx context.Context, userID, fingerprint, userAgent, ipAddress string) {
	query := `
		INSERT INTO device_fingerprints (user_id, fingerprint, user_agent, ip_address, first_seen, last_seen, use_count)
		VALUES (?, ?, ?, ?, NOW(), NOW(), 1)
		ON DUPLICATE KEY UPDATE
			last_seen = NOW(),
			use_count = use_count + 1
	`
	s.db.ExecContext(ctx, query, userID, fingerprint, userAgent, ipAddress)
}

func (s *FraudDetectionService) logFraudCheck(ctx context.Context, check *FraudCheck) {
	query := `
		INSERT INTO fraud_checks (
			id, user_id, ip_address, device_id, risk_score, decision,
			reasons, country, is_vpn, is_proxy, is_tor, fingerprint, checked_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	reasonStr := strings.Join(check.Reasons, ",")
	s.db.ExecContext(ctx, query,
		check.ID, check.UserID, check.IPAddress, check.DeviceID,
		check.RiskScore, check.Decision, reasonStr, check.Country,
		check.IsVPN, check.IsProxy, check.IsTor, check.Fingerprint,
	)
}

func (s *FraudDetectionService) logSuspiciousActivity(ctx context.Context, userID, ipAddress, action, description, severity string) {
	query := `
		INSERT INTO suspicious_activities (id, user_id, ip_address, action, description, severity, detected_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	s.db.ExecContext(ctx, query, uuid.New().String(), userID, ipAddress, action, description, severity)
}

func defaultFraudThresholds() *FraudThresholds {
	return &FraudThresholds{
		MaxConcurrentDevices: 5,
		MaxConcurrentStreams: 3,
		MaxCountriesPerHour:  3,
		GeoVelocityKMH:       1000.0, // Faster than commercial flight = impossible
		MaxFailedLogins:      5,
		VPNDetectionEnabled:  true,
		SuspiciousUserAgents: []string{
			"curl", "wget", "python", "scrapy", "bot", "crawler",
			"phantomjs", "selenium", "headless",
		},
	}
}

// haversineDistance calculates distance between two GPS coordinates in km
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371.0 // km

	dLat := degreesToRadians(lat2 - lat1)
	dLng := degreesToRadians(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}
