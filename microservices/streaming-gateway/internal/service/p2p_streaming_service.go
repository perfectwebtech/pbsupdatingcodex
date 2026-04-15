package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// P2PStreamingService implements peer-to-peer content delivery
// Reduces CDN bandwidth costs by 60-80% for popular content
type P2PStreamingService struct {
	db        *sql.DB
	peers     map[string]*Peer
	swarms    map[string]*Swarm
	mu        sync.RWMutex
	config    *P2PConfig
	analytics *P2PAnalytics
}

// NewP2PStreamingService creates a new P2P streaming service
func NewP2PStreamingService(db *sql.DB, config *P2PConfig) *P2PStreamingService {
	if config == nil {
		config = defaultP2PConfig()
	}

	s := &P2PStreamingService{
		db:        db,
		peers:     make(map[string]*Peer),
		swarms:    make(map[string]*Swarm),
		config:    config,
		analytics: &P2PAnalytics{},
	}

	go s.startSwarmManager()
	go s.startAnalyticsCollector()

	return s
}

// =====================================================
// MODELS
// =====================================================

// P2PConfig configures P2P behavior
type P2PConfig struct {
	Enabled            bool    `json:"enabled"`
	MinPeersForP2P     int     `json:"min_peers_for_p2p"`      // Minimum peers to enable P2P
	MaxPeersPerSwarm   int     `json:"max_peers_per_swarm"`    // Maximum peers in one swarm
	CDNFallbackPercent float64 `json:"cdn_fallback_percent"`   // % of traffic from CDN
	ChunkSize          int     `json:"chunk_size_kb"`          // Chunk size in KB
	TrackerURL         string  `json:"tracker_url"`            // WebRTC signaling server
	STUNServers        []string `json:"stun_servers"`
	TURNServers        []string `json:"turn_servers"`
}

// Peer represents a connected peer
type Peer struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	IPAddress     string    `json:"ip_address"`
	ConnectionID  string    `json:"connection_id"`
	Bandwidth     int64     `json:"bandwidth_mbps"`
	UploadSpeed   int64     `json:"upload_speed_kbps"`
	DownloadSpeed int64     `json:"download_speed_kbps"`
	IsSeeder      bool      `json:"is_seeder"`
	IsActive      bool      `json:"is_active"`
	SharedChunks  []string  `json:"shared_chunks"`
	JoinedAt      time.Time `json:"joined_at"`
	LastSeen      time.Time `json:"last_seen"`
	BytesUploaded int64     `json:"bytes_uploaded"`
	BytesDownloaded int64   `json:"bytes_downloaded"`
}

// Swarm represents a group of peers sharing the same content
type Swarm struct {
	ID          string           `json:"id"`
	ContentID   string           `json:"content_id"`
	ContentType string           `json:"content_type"`
	Peers       map[string]*Peer `json:"peers"`
	TotalPeers  int              `json:"total_peers"`
	Seeders     int              `json:"seeders"`
	Leechers    int              `json:"leechers"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	mu          sync.RWMutex     `json:"-"`
}

// P2PSession represents a user's P2P session
type P2PSession struct {
	ID              string    `json:"id"`
	PeerID          string    `json:"peer_id"`
	ContentID       string    `json:"content_id"`
	SwarmID         string    `json:"swarm_id"`
	CDNBandwidth    int64     `json:"cdn_bandwidth_bytes"`
	P2PBandwidth    int64     `json:"p2p_bandwidth_bytes"`
	P2PPercentage   float64   `json:"p2p_percentage"`
	ChunksFromCDN   int       `json:"chunks_from_cdn"`
	ChunksFromPeers int       `json:"chunks_from_peers"`
	PeersConnected  int       `json:"peers_connected"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
}

// Chunk represents a content chunk
type Chunk struct {
	ID          string   `json:"id"`
	ContentID   string   `json:"content_id"`
	Index       int      `json:"index"`
	Hash        string   `json:"hash"`
	Size        int      `json:"size_bytes"`
	AvailablePeers []string `json:"available_peers"`
}

// P2PAnalytics tracks P2P performance
type P2PAnalytics struct {
	TotalBandwidthSaved int64   `json:"total_bandwidth_saved_bytes"`
	CDNCostSaved        float64 `json:"cdn_cost_saved_usd"`
	P2PEfficiency       float64 `json:"p2p_efficiency_percent"`
	ActiveSwarms        int     `json:"active_swarms"`
	TotalPeers          int     `json:"total_peers"`
	mu                  sync.RWMutex
}

// =====================================================
// PEER MANAGEMENT
// =====================================================

// RegisterPeer registers a new peer
func (s *P2PStreamingService) RegisterPeer(ctx context.Context, userID, ipAddress string) (*Peer, error) {
	peer := &Peer{
		ID:           uuid.New().String(),
		UserID:       userID,
		IPAddress:    ipAddress,
		ConnectionID: generateConnectionID(),
		IsActive:     true,
		SharedChunks: []string{},
		JoinedAt:     time.Now(),
		LastSeen:     time.Now(),
	}

	s.mu.Lock()
	s.peers[peer.ID] = peer
	s.mu.Unlock()

	// Persist to database
	query := `
		INSERT INTO p2p_peers (id, user_id, ip_address, connection_id, is_active, joined_at, last_seen)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	_, err := s.db.ExecContext(ctx, query, peer.ID, userID, ipAddress, peer.ConnectionID, true)
	if err != nil {
		return nil, err
	}

	return peer, nil
}

// JoinSwarm adds a peer to a content swarm
func (s *P2PStreamingService) JoinSwarm(ctx context.Context, peerID, contentID, contentType string) (*Swarm, error) {
	s.mu.Lock()
	peer, exists := s.peers[peerID]
	if !exists {
		s.mu.Unlock()
		return nil, errors.New("peer not found")
	}
	s.mu.Unlock()

	// Get or create swarm
	swarmID := generateSwarmID(contentID)

	s.mu.Lock()
	swarm, exists := s.swarms[swarmID]
	if !exists {
		swarm = &Swarm{
			ID:          swarmID,
			ContentID:   contentID,
			ContentType: contentType,
			Peers:       make(map[string]*Peer),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		s.swarms[swarmID] = swarm
	}
	s.mu.Unlock()

	// Add peer to swarm
	swarm.mu.Lock()
	swarm.Peers[peerID] = peer
	swarm.TotalPeers = len(swarm.Peers)
	if peer.IsSeeder {
		swarm.Seeders++
	} else {
		swarm.Leechers++
	}
	swarm.UpdatedAt = time.Now()
	swarm.mu.Unlock()

	// Record join in database
	query := `
		INSERT INTO p2p_swarm_members (swarm_id, peer_id, content_id, joined_at)
		VALUES (?, ?, ?, NOW())
	`
	s.db.ExecContext(ctx, query, swarmID, peerID, contentID)

	return swarm, nil
}

// LeaveSwarm removes a peer from a swarm
func (s *P2PStreamingService) LeaveSwarm(ctx context.Context, peerID, swarmID string) error {
	s.mu.RLock()
	swarm, exists := s.swarms[swarmID]
	s.mu.RUnlock()

	if !exists {
		return errors.New("swarm not found")
	}

	swarm.mu.Lock()
	peer, exists := swarm.Peers[peerID]
	if exists {
		delete(swarm.Peers, peerID)
		swarm.TotalPeers = len(swarm.Peers)
		if peer.IsSeeder {
			swarm.Seeders--
		} else {
			swarm.Leechers--
		}
	}
	swarm.mu.Unlock()

	// Clean up empty swarms
	if swarm.TotalPeers == 0 {
		s.mu.Lock()
		delete(s.swarms, swarmID)
		s.mu.Unlock()
	}

	query := `UPDATE p2p_swarm_members SET left_at = NOW() WHERE swarm_id = ? AND peer_id = ?`
	s.db.ExecContext(ctx, query, swarmID, peerID)

	return nil
}

// =====================================================
// CONTENT DELIVERY
// =====================================================

// GetPeersForChunk returns available peers for a specific chunk
func (s *P2PStreamingService) GetPeersForChunk(ctx context.Context, contentID string, chunkIndex int) ([]*Peer, error) {
	swarmID := generateSwarmID(contentID)

	s.mu.RLock()
	swarm, exists := s.swarms[swarmID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("swarm not found")
	}

	// Filter peers that have this chunk
	chunkHash := fmt.Sprintf("%s:%d", contentID, chunkIndex)

	swarm.mu.RLock()
	defer swarm.mu.RUnlock()

	var availablePeers []*Peer
	for _, peer := range swarm.Peers {
		if peer.IsActive && peer.IsSeeder {
			availablePeers = append(availablePeers, peer)
		} else {
			// Check if leecher has this specific chunk
			for _, chunk := range peer.SharedChunks {
				if chunk == chunkHash {
					availablePeers = append(availablePeers, peer)
					break
				}
			}
		}
	}

	// Limit and randomize for load balancing
	if len(availablePeers) > 5 {
		rand.Shuffle(len(availablePeers), func(i, j int) {
			availablePeers[i], availablePeers[j] = availablePeers[j], availablePeers[i]
		})
		availablePeers = availablePeers[:5]
	}

	return availablePeers, nil
}

// ReportChunkDownload records chunk transfer statistics
func (s *P2PStreamingService) ReportChunkDownload(ctx context.Context, peerID string, chunkID string, source string, bytes int64) error {
	s.mu.RLock()
	peer, exists := s.peers[peerID]
	s.mu.RUnlock()

	if !exists {
		return errors.New("peer not found")
	}

	// Update peer statistics
	peer.BytesDownloaded += bytes
	peer.LastSeen = time.Now()

	// Track for analytics
	s.analytics.mu.Lock()
	if source == "p2p" {
		s.analytics.TotalBandwidthSaved += bytes
		s.analytics.CDNCostSaved += float64(bytes) / 1024 / 1024 / 1024 * 0.05 // Assume $0.05/GB CDN cost
	}
	s.analytics.mu.Unlock()

	// Add chunk to peer's shared chunks
	if !contains(peer.SharedChunks, chunkID) {
		peer.SharedChunks = append(peer.SharedChunks, chunkID)
	}

	// Record in database
	query := `
		INSERT INTO p2p_chunk_transfers (peer_id, chunk_id, source, bytes, transferred_at)
		VALUES (?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query, peerID, chunkID, source, bytes)
	return err
}

// ReportChunkUpload records when a peer uploads a chunk
func (s *P2PStreamingService) ReportChunkUpload(ctx context.Context, peerID string, chunkID string, bytes int64) error {
	s.mu.RLock()
	peer, exists := s.peers[peerID]
	s.mu.RUnlock()

	if !exists {
		return errors.New("peer not found")
	}

	peer.BytesUploaded += bytes
	peer.LastSeen = time.Now()

	// Promote to seeder if uploaded enough
	if peer.BytesUploaded > peer.BytesDownloaded*2 && !peer.IsSeeder {
		peer.IsSeeder = true
	}

	query := `UPDATE p2p_peers SET bytes_uploaded = ?, is_seeder = ?, last_seen = NOW() WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, peer.BytesUploaded, peer.IsSeeder, peerID)
	return err
}

// =====================================================
// P2P STRATEGY DECISION
// =====================================================

// ShouldUseP2P determines if P2P should be used for content
func (s *P2PStreamingService) ShouldUseP2P(ctx context.Context, contentID string) (bool, *P2PStrategy, error) {
	if !s.config.Enabled {
		return false, nil, errors.New("P2P disabled")
	}

	swarmID := generateSwarmID(contentID)

	s.mu.RLock()
	swarm, exists := s.swarms[swarmID]
	s.mu.RUnlock()

	strategy := &P2PStrategy{
		UseP2P:         false,
		CDNPercentage:  100,
		P2PPercentage:  0,
		RecommendedPeers: []string{},
	}

	if !exists || swarm.TotalPeers < s.config.MinPeersForP2P {
		// Not enough peers, use CDN only
		return false, strategy, nil
	}

	// Calculate P2P viability
	swarm.mu.RLock()
	seeders := swarm.Seeders
	totalPeers := swarm.TotalPeers
	swarm.mu.RUnlock()

	if seeders == 0 {
		// No seeders, can't use P2P
		return false, strategy, nil
	}

	// Calculate optimal P2P/CDN mix
	p2pPercent := calculateP2PPercentage(totalPeers, seeders)
	cdnPercent := 100.0 - p2pPercent

	// Ensure minimum CDN traffic for reliability
	if cdnPercent < s.config.CDNFallbackPercent {
		cdnPercent = s.config.CDNFallbackPercent
		p2pPercent = 100.0 - cdnPercent
	}

	strategy.UseP2P = true
	strategy.P2PPercentage = p2pPercent
	strategy.CDNPercentage = cdnPercent
	strategy.TotalPeers = totalPeers
	strategy.Seeders = seeders

	return true, strategy, nil
}

// P2PStrategy describes the P2P/CDN mix
type P2PStrategy struct {
	UseP2P           bool     `json:"use_p2p"`
	P2PPercentage    float64  `json:"p2p_percentage"`
	CDNPercentage    float64  `json:"cdn_percentage"`
	TotalPeers       int      `json:"total_peers"`
	Seeders          int      `json:"seeders"`
	RecommendedPeers []string `json:"recommended_peers"`
}

func calculateP2PPercentage(totalPeers, seeders int) float64 {
	if totalPeers < 5 {
		return 20.0
	} else if totalPeers < 20 {
		return 40.0
	} else if totalPeers < 50 {
		return 60.0
	} else if totalPeers < 100 {
		return 75.0
	} else {
		return 85.0 // Max 85% from P2P, 15% from CDN for reliability
	}
}

// =====================================================
// ANALYTICS & REPORTING
// =====================================================

// GetP2PAnalytics returns current P2P statistics
func (s *P2PStreamingService) GetP2PAnalytics(ctx context.Context) *P2PAnalytics {
	s.analytics.mu.RLock()
	defer s.analytics.mu.RUnlock()

	s.mu.RLock()
	defer s.mu.RUnlock()

	analytics := &P2PAnalytics{
		TotalBandwidthSaved: s.analytics.TotalBandwidthSaved,
		CDNCostSaved:        s.analytics.CDNCostSaved,
		ActiveSwarms:        len(s.swarms),
		TotalPeers:          len(s.peers),
	}

	// Calculate P2P efficiency
	var totalBandwidth int64
	for _, peer := range s.peers {
		totalBandwidth += peer.BytesDownloaded
	}

	if totalBandwidth > 0 {
		analytics.P2PEfficiency = (float64(s.analytics.TotalBandwidthSaved) / float64(totalBandwidth)) * 100
	}

	return analytics
}

// GetSwarmInfo returns information about a swarm
func (s *P2PStreamingService) GetSwarmInfo(ctx context.Context, contentID string) (*Swarm, error) {
	swarmID := generateSwarmID(contentID)

	s.mu.RLock()
	swarm, exists := s.swarms[swarmID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("swarm not found")
	}

	return swarm, nil
}

// GetP2PReport generates a detailed P2P performance report
func (s *P2PStreamingService) GetP2PReport(ctx context.Context, days int) (map[string]interface{}, error) {
	query := `
		SELECT
			DATE(created_at) as date,
			COUNT(DISTINCT swarm_id) as active_swarms,
			COUNT(DISTINCT peer_id) as unique_peers,
			SUM(CASE WHEN source = 'p2p' THEN bytes ELSE 0 END) as p2p_bytes,
			SUM(CASE WHEN source = 'cdn' THEN bytes ELSE 0 END) as cdn_bytes,
			SUM(bytes) as total_bytes
		FROM p2p_chunk_transfers
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	rows, err := s.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dailyStats []map[string]interface{}
	var totalP2PBytes, totalCDNBytes int64

	for rows.Next() {
		var date string
		var swarms, peers int
		var p2pBytes, cdnBytes, totalBytes int64

		rows.Scan(&date, &swarms, &peers, &p2pBytes, &cdnBytes, &totalBytes)

		totalP2PBytes += p2pBytes
		totalCDNBytes += cdnBytes

		p2pPercent := 0.0
		if totalBytes > 0 {
			p2pPercent = (float64(p2pBytes) / float64(totalBytes)) * 100
		}

		dailyStats = append(dailyStats, map[string]interface{}{
			"date":          date,
			"active_swarms": swarms,
			"unique_peers":  peers,
			"p2p_bytes":     p2pBytes,
			"cdn_bytes":     cdnBytes,
			"p2p_percent":   p2pPercent,
		})
	}

	totalBytes := totalP2PBytes + totalCDNBytes
	p2pEfficiency := 0.0
	if totalBytes > 0 {
		p2pEfficiency = (float64(totalP2PBytes) / float64(totalBytes)) * 100
	}

	bandwidthSavedGB := float64(totalP2PBytes) / 1024 / 1024 / 1024
	costSaved := bandwidthSavedGB * 0.05 // Assume $0.05/GB

	return map[string]interface{}{
		"period_days":        days,
		"total_p2p_bytes":    totalP2PBytes,
		"total_cdn_bytes":    totalCDNBytes,
		"p2p_efficiency":     p2pEfficiency,
		"bandwidth_saved_gb": bandwidthSavedGB,
		"cost_saved_usd":     costSaved,
		"daily_stats":        dailyStats,
		"current_analytics":  s.GetP2PAnalytics(ctx),
	}, nil
}

// =====================================================
// BACKGROUND TASKS
// =====================================================

// startSwarmManager cleans up inactive swarms
func (s *P2PStreamingService) startSwarmManager() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupInactiveSwarms()
	}
}

func (s *P2PStreamingService) cleanupInactiveSwarms() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for swarmID, swarm := range s.swarms {
		swarm.mu.RLock()
		inactive := now.Sub(swarm.UpdatedAt) > 30*time.Minute
		swarm.mu.RUnlock()

		if inactive {
			delete(s.swarms, swarmID)
		}
	}

	// Cleanup inactive peers
	for peerID, peer := range s.peers {
		if now.Sub(peer.LastSeen) > 1*time.Hour {
			delete(s.peers, peerID)
		}
	}
}

func (s *P2PStreamingService) startAnalyticsCollector() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.collectAnalytics()
	}
}

func (s *P2PStreamingService) collectAnalytics() {
	// Persist analytics to database
	s.analytics.mu.RLock()
	defer s.analytics.mu.RUnlock()

	query := `
		INSERT INTO p2p_analytics (
			bandwidth_saved_bytes, cost_saved_usd, p2p_efficiency, active_swarms, total_peers, recorded_at
		) VALUES (?, ?, ?, ?, ?, NOW())
	`
	s.db.Exec(query,
		s.analytics.TotalBandwidthSaved,
		s.analytics.CDNCostSaved,
		s.analytics.P2PEfficiency,
		s.analytics.ActiveSwarms,
		s.analytics.TotalPeers,
	)
}

// =====================================================
// HELPERS
// =====================================================

func generateSwarmID(contentID string) string {
	hash := sha256.Sum256([]byte(contentID))
	return hex.EncodeToString(hash[:])[:16]
}

func generateConnectionID() string {
	return uuid.New().String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func defaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		Enabled:            true,
		MinPeersForP2P:     5,
		MaxPeersPerSwarm:   100,
		CDNFallbackPercent: 15.0,
		ChunkSize:          256, // 256KB chunks
		TrackerURL:         "wss://tracker.iptv.example.com",
		STUNServers: []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
		},
		TURNServers: []string{
			"turn:turn.iptv.example.com:3478",
		},
	}
}

// MarshalJSON custom JSON marshaling for Swarm (exclude mutex)
func (s *Swarm) MarshalJSON() ([]byte, error) {
	type Alias Swarm
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}
