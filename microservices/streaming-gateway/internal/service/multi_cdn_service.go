package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"
)

// MultiCDNService manages multiple CDN providers with intelligent routing
type MultiCDNService struct {
	db        *sql.DB
	providers map[string]*CDNProvider
	mu        sync.RWMutex
	healthMu  sync.RWMutex
	health    map[string]*CDNHealth
}

// NewMultiCDNService creates a new multi-CDN service
func NewMultiCDNService(db *sql.DB) *MultiCDNService {
	s := &MultiCDNService{
		db:        db,
		providers: make(map[string]*CDNProvider),
		health:    make(map[string]*CDNHealth),
	}
	s.loadProviders()
	go s.startHealthMonitoring()
	return s
}

// =====================================================
// MODELS
// =====================================================

// CDNProvider represents a CDN provider configuration
type CDNProvider struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"` // cloudflare, cloudfront, bunny, fastly, akamai
	BaseURL       string            `json:"base_url"`
	APIKey        string            `json:"api_key"`
	Priority      int               `json:"priority"`
	Weight        int               `json:"weight"`
	IsActive      bool              `json:"is_active"`
	CostPerGB     float64           `json:"cost_per_gb"`
	Regions       []string          `json:"regions"`
	Features      map[string]bool   `json:"features"`
	BandwidthMonthlyLimit int64     `json:"bandwidth_monthly_limit"`
	BandwidthUsedMonth   int64      `json:"bandwidth_used_month"`
}

// CDNHealth represents real-time CDN health metrics
type CDNHealth struct {
	ProviderID    string    `json:"provider_id"`
	IsHealthy     bool      `json:"is_healthy"`
	ResponseTime  int       `json:"response_time_ms"`
	SuccessRate   float64   `json:"success_rate"`
	ErrorRate     float64   `json:"error_rate"`
	LastChecked   time.Time `json:"last_checked"`
	ConsecutiveFails int    `json:"consecutive_failures"`
	UptimePercent float64   `json:"uptime_percent"`
}

// CDNSelection represents a CDN routing decision
type CDNSelection struct {
	ProviderID   string    `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	URL          string    `json:"url"`
	Region       string    `json:"region"`
	Priority     int       `json:"priority"`
	Reason       string    `json:"reason"`
	FallbackURLs []string  `json:"fallback_urls"`
	SelectedAt   time.Time `json:"selected_at"`
}

// CDNStats represents usage statistics per CDN
type CDNStats struct {
	ProviderID    string    `json:"provider_id"`
	Date          string    `json:"date"`
	BandwidthGB   float64   `json:"bandwidth_gb"`
	Requests      int64     `json:"requests"`
	CacheHitRate  float64   `json:"cache_hit_rate"`
	AvgLatencyMs  int       `json:"avg_latency_ms"`
	Cost          float64   `json:"cost"`
	ErrorCount    int64     `json:"error_count"`
}

// =====================================================
// CDN SELECTION ALGORITHM
// =====================================================

// SelectCDN selects the best CDN for a given user based on multiple factors
func (s *MultiCDNService) SelectCDN(ctx context.Context, userIP, contentPath, country string) (*CDNSelection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.providers) == 0 {
		return nil, errors.New("no CDN providers configured")
	}

	// Get healthy providers
	healthyProviders := s.getHealthyProviders(country)
	if len(healthyProviders) == 0 {
		return nil, errors.New("no healthy CDN providers available")
	}

	// Score each provider
	scored := make([]*scoredProvider, 0, len(healthyProviders))
	for _, p := range healthyProviders {
		score := s.calculateProviderScore(p, country)
		scored = append(scored, &scoredProvider{
			provider: p,
			score:    score,
		})
	}

	// Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Select winner
	winner := scored[0].provider

	// Build fallback URLs
	fallbacks := []string{}
	for i := 1; i < len(scored) && i < 3; i++ {
		fallbacks = append(fallbacks, fmt.Sprintf("%s/%s", scored[i].provider.BaseURL, contentPath))
	}

	selection := &CDNSelection{
		ProviderID:   winner.ID,
		ProviderName: winner.Name,
		URL:          fmt.Sprintf("%s/%s", winner.BaseURL, contentPath),
		Region:       country,
		Priority:     winner.Priority,
		Reason:       s.getSelectionReason(winner, country),
		FallbackURLs: fallbacks,
		SelectedAt:   time.Now(),
	}

	// Track usage
	go s.trackUsage(ctx, winner.ID, contentPath)

	return selection, nil
}

type scoredProvider struct {
	provider *CDNProvider
	score    float64
}

// calculateProviderScore scores a CDN provider based on multiple factors
func (s *MultiCDNService) calculateProviderScore(p *CDNProvider, country string) float64 {
	score := 100.0

	// Health score (40% weight)
	s.healthMu.RLock()
	health, exists := s.health[p.ID]
	s.healthMu.RUnlock()

	if exists {
		// Response time (lower is better)
		if health.ResponseTime > 0 {
			score += (1000.0 - float64(health.ResponseTime)) / 10.0 * 0.4
		}
		// Success rate
		score += health.SuccessRate * 40
		// Uptime
		score += health.UptimePercent * 0.4
	}

	// Geo proximity (30% weight)
	geoScore := s.calculateGeoScore(p, country)
	score += geoScore * 30

	// Cost efficiency (20% weight)
	costScore := s.calculateCostScore(p)
	score += costScore * 20

	// Available bandwidth (10% weight)
	if p.BandwidthMonthlyLimit > 0 {
		usagePercent := float64(p.BandwidthUsedMonth) / float64(p.BandwidthMonthlyLimit)
		bandwidthScore := (1.0 - usagePercent) * 10
		score += bandwidthScore
	}

	// Apply weight modifier
	score *= float64(p.Weight) / 100.0

	return score
}

func (s *MultiCDNService) calculateGeoScore(p *CDNProvider, country string) float64 {
	// Check if provider has presence in user's region
	for _, region := range p.Regions {
		if region == country {
			return 1.0 // Best - has POP in country
		}
	}
	// Check continental proximity (simplified)
	return 0.5
}

func (s *MultiCDNService) calculateCostScore(p *CDNProvider) float64 {
	// Lower cost = higher score (max cost considered: $0.10/GB)
	maxCost := 0.10
	if p.CostPerGB >= maxCost {
		return 0.0
	}
	return (maxCost - p.CostPerGB) / maxCost
}

func (s *MultiCDNService) getSelectionReason(p *CDNProvider, country string) string {
	reasons := []string{}

	for _, region := range p.Regions {
		if region == country {
			reasons = append(reasons, "geo-optimal")
			break
		}
	}

	s.healthMu.RLock()
	if health, ok := s.health[p.ID]; ok {
		if health.ResponseTime < 50 {
			reasons = append(reasons, "low-latency")
		}
		if health.SuccessRate > 0.99 {
			reasons = append(reasons, "high-reliability")
		}
	}
	s.healthMu.RUnlock()

	if p.CostPerGB < 0.02 {
		reasons = append(reasons, "cost-optimized")
	}

	if len(reasons) == 0 {
		return "default-selection"
	}

	result := ""
	for i, r := range reasons {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}

// =====================================================
// HEALTH MONITORING
// =====================================================

// startHealthMonitoring continuously monitors CDN health
func (s *MultiCDNService) startHealthMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.checkAllProviders()
	}
}

// checkAllProviders performs health checks on all CDN providers
func (s *MultiCDNService) checkAllProviders() {
	s.mu.RLock()
	providers := make([]*CDNProvider, 0, len(s.providers))
	for _, p := range s.providers {
		providers = append(providers, p)
	}
	s.mu.RUnlock()

	var wg sync.WaitGroup
	for _, p := range providers {
		wg.Add(1)
		go func(provider *CDNProvider) {
			defer wg.Done()
			s.checkProviderHealth(provider)
		}(p)
	}
	wg.Wait()
}

// checkProviderHealth checks a single CDN's health
func (s *MultiCDNService) checkProviderHealth(p *CDNProvider) {
	start := time.Now()
	healthURL := fmt.Sprintf("%s/health", p.BaseURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(healthURL)

	s.healthMu.Lock()
	defer s.healthMu.Unlock()

	health, exists := s.health[p.ID]
	if !exists {
		health = &CDNHealth{
			ProviderID:    p.ID,
			SuccessRate:   1.0,
			UptimePercent: 100.0,
		}
		s.health[p.ID] = health
	}

	health.LastChecked = time.Now()
	responseTime := time.Since(start).Milliseconds()
	health.ResponseTime = int(responseTime)

	if err != nil || resp.StatusCode != http.StatusOK {
		health.IsHealthy = false
		health.ConsecutiveFails++
		health.SuccessRate = math.Max(0, health.SuccessRate-0.05)
		if health.ConsecutiveFails > 3 {
			health.UptimePercent = math.Max(0, health.UptimePercent-1.0)
		}
	} else {
		health.IsHealthy = true
		health.ConsecutiveFails = 0
		health.SuccessRate = math.Min(1.0, health.SuccessRate+0.01)
		health.UptimePercent = math.Min(100, health.UptimePercent+0.1)
		resp.Body.Close()
	}
}

// =====================================================
// PROVIDER MANAGEMENT
// =====================================================

// AddProvider adds a new CDN provider
func (s *MultiCDNService) AddProvider(ctx context.Context, p *CDNProvider) error {
	query := `
		INSERT INTO cdn_providers (id, name, type, base_url, api_key, priority, weight, is_active, cost_per_gb)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Type, p.BaseURL, p.APIKey,
		p.Priority, p.Weight, p.IsActive, p.CostPerGB,
	)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.providers[p.ID] = p
	s.mu.Unlock()

	return nil
}

func (s *MultiCDNService) loadProviders() {
	query := `
		SELECT id, name, type, base_url, COALESCE(api_key, ''), priority, weight, is_active, cost_per_gb
		FROM cdn_providers WHERE is_active = 1
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		p := &CDNProvider{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.APIKey,
			&p.Priority, &p.Weight, &p.IsActive, &p.CostPerGB,
		)
		if err != nil {
			continue
		}
		s.providers[p.ID] = p
	}
}

func (s *MultiCDNService) getHealthyProviders(country string) []*CDNProvider {
	healthy := []*CDNProvider{}
	for _, p := range s.providers {
		if !p.IsActive {
			continue
		}
		s.healthMu.RLock()
		health, exists := s.health[p.ID]
		s.healthMu.RUnlock()

		if !exists || health.IsHealthy {
			healthy = append(healthy, p)
		}
	}
	return healthy
}

// =====================================================
// USAGE TRACKING & ANALYTICS
// =====================================================

func (s *MultiCDNService) trackUsage(ctx context.Context, providerID, contentPath string) {
	query := `
		INSERT INTO cdn_usage (provider_id, content_path, requests, last_request)
		VALUES (?, ?, 1, NOW())
		ON DUPLICATE KEY UPDATE requests = requests + 1, last_request = NOW()
	`
	s.db.ExecContext(ctx, query, providerID, contentPath)
}

// GetCDNStats returns usage statistics for a CDN provider
func (s *MultiCDNService) GetCDNStats(ctx context.Context, providerID string, days int) ([]*CDNStats, error) {
	query := `
		SELECT
			provider_id,
			DATE(date) as date,
			SUM(bandwidth_bytes) / 1024 / 1024 / 1024 as bandwidth_gb,
			SUM(requests) as requests,
			AVG(cache_hit_rate) as cache_hit_rate,
			AVG(avg_latency_ms) as avg_latency_ms,
			SUM(cost) as cost,
			SUM(error_count) as error_count
		FROM cdn_daily_stats
		WHERE provider_id = ? AND date >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY date
		ORDER BY date DESC
	`
	rows, err := s.db.QueryContext(ctx, query, providerID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*CDNStats
	for rows.Next() {
		stat := &CDNStats{}
		err := rows.Scan(
			&stat.ProviderID, &stat.Date, &stat.BandwidthGB, &stat.Requests,
			&stat.CacheHitRate, &stat.AvgLatencyMs, &stat.Cost, &stat.ErrorCount,
		)
		if err != nil {
			continue
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetCostOptimization returns cost optimization recommendations
func (s *MultiCDNService) GetCostOptimization(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Get current month spending per provider
	query := `
		SELECT provider_id, SUM(cost) as total_cost, SUM(bandwidth_bytes)/1024/1024/1024 as gb
		FROM cdn_daily_stats
		WHERE date >= DATE_FORMAT(NOW() ,'%Y-%m-01')
		GROUP BY provider_id
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type providerCost struct {
		ProviderID string  `json:"provider_id"`
		TotalCost  float64 `json:"total_cost"`
		BandwidthGB float64 `json:"bandwidth_gb"`
		EffectiveRate float64 `json:"effective_rate"`
	}

	var costs []providerCost
	totalCost := 0.0
	for rows.Next() {
		c := providerCost{}
		if err := rows.Scan(&c.ProviderID, &c.TotalCost, &c.BandwidthGB); err != nil {
			continue
		}
		if c.BandwidthGB > 0 {
			c.EffectiveRate = c.TotalCost / c.BandwidthGB
		}
		totalCost += c.TotalCost
		costs = append(costs, c)
	}

	result["current_month_total"] = totalCost
	result["per_provider"] = costs
	result["recommendations"] = s.generateCostRecommendations(costs)

	return result, nil
}

func (s *MultiCDNService) generateCostRecommendations(costs []interface{}) []string {
	// Generate cost optimization recommendations
	recommendations := []string{
		"Route static content through cheapest CDN",
		"Use P2P for popular content to reduce CDN bandwidth",
		"Increase cache TTL for static assets",
		"Enable Brotli compression on all CDNs",
		"Consider regional CDN consolidation",
	}
	return recommendations
}
