package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AVODService manages ad-supported video on demand (Advertising Video On Demand)
type AVODService struct {
	db *sql.DB
}

// NewAVODService creates a new AVOD service
func NewAVODService(db *sql.DB) *AVODService {
	return &AVODService{db: db}
}

// =====================================================
// MODELS
// =====================================================

// Advertisement represents an ad creative
type Advertisement struct {
	ID            string    `json:"id"`
	AdvertiserID  string    `json:"advertiser_id"`
	CampaignID    string    `json:"campaign_id"`
	Title         string    `json:"title"`
	Type          string    `json:"type"` // pre-roll, mid-roll, post-roll, banner, overlay, interactive
	Format        string    `json:"format"` // video, image, html5, vast
	MediaURL      string    `json:"media_url"`
	ClickURL      string    `json:"click_url"`
	Duration      int       `json:"duration_seconds"`
	Skippable     bool      `json:"skippable"`
	SkipAfter     int       `json:"skip_after_seconds"`
	TargetGroups  []string  `json:"target_groups"`
	TargetGenres  []string  `json:"target_genres"`
	TargetCountries []string `json:"target_countries"`
	MinAge        int       `json:"min_age"`
	MaxAge        int       `json:"max_age"`
	BidAmount     float64   `json:"bid_amount"` // CPM
	DailyBudget   float64   `json:"daily_budget"`
	TotalBudget   float64   `json:"total_budget"`
	SpentToday    float64   `json:"spent_today"`
	SpentTotal    float64   `json:"spent_total"`
	Impressions   int64     `json:"impressions"`
	Clicks        int64     `json:"clicks"`
	IsActive      bool      `json:"is_active"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CreatedAt     time.Time `json:"created_at"`
}

// AdCampaign represents an advertising campaign
type AdCampaign struct {
	ID            string    `json:"id"`
	AdvertiserID  string    `json:"advertiser_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Budget        float64   `json:"budget"`
	Spent         float64   `json:"spent"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// AdImpression tracks ad impressions
type AdImpression struct {
	ID         string    `json:"id"`
	AdID       string    `json:"ad_id"`
	UserID     string    `json:"user_id"`
	StreamID   string    `json:"stream_id"`
	Country    string    `json:"country"`
	DeviceType string    `json:"device_type"`
	Position   string    `json:"position"` // pre-roll, mid-roll, post-roll
	Watched    int       `json:"watched_seconds"`
	Completed  bool      `json:"completed"`
	Skipped    bool      `json:"skipped"`
	Clicked    bool      `json:"clicked"`
	Revenue    float64   `json:"revenue"`
	ShownAt    time.Time `json:"shown_at"`
}

// AdBreakSchedule represents ad insertion points
type AdBreakSchedule struct {
	ContentID    string  `json:"content_id"`
	PreRoll      bool    `json:"pre_roll"`
	MidRolls     []int   `json:"mid_rolls"` // Array of timestamps in seconds
	PostRoll     bool    `json:"post_roll"`
	MaxAdsPerBreak int   `json:"max_ads_per_break"`
}

// AdRequest represents a request for an ad
type AdRequest struct {
	UserID      string   `json:"user_id"`
	StreamID    string   `json:"stream_id"`
	ContentType string   `json:"content_type"`
	Position    string   `json:"position"`
	Country     string   `json:"country"`
	DeviceType  string   `json:"device_type"`
	UserAge     int      `json:"user_age"`
	UserGender  string   `json:"user_gender"`
	Genres      []string `json:"genres"`
}

// =====================================================
// AD SERVING
// =====================================================

// ServeAd selects and serves the best ad for a user
func (s *AVODService) ServeAd(ctx context.Context, req *AdRequest) (*Advertisement, error) {
	// Check if user has ad-free subscription
	hasAdFree, _ := s.userHasAdFreeSubscription(ctx, req.UserID)
	if hasAdFree {
		return nil, errors.New("user has ad-free subscription")
	}

	// Get eligible ads using real-time bidding
	candidates, err := s.getEligibleAds(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return s.getDefaultAd(ctx, req)
	}

	// Run RTB auction
	winner := s.runAuction(candidates, req)
	if winner == nil {
		return nil, errors.New("no ad selected")
	}

	// Pre-record impression (will be confirmed when actually shown)
	s.preRecordImpression(ctx, winner.ID, req)

	return winner, nil
}

// GetAdBreakSchedule returns the ad insertion schedule for content
func (s *AVODService) GetAdBreakSchedule(ctx context.Context, contentID string, durationSeconds int) (*AdBreakSchedule, error) {
	schedule := &AdBreakSchedule{
		ContentID:      contentID,
		PreRoll:        true,
		PostRoll:       false,
		MaxAdsPerBreak: 2,
		MidRolls:       []int{},
	}

	// Insert mid-roll ads every ~10 minutes for content longer than 15 min
	if durationSeconds > 900 {
		intervalSec := 600 // 10 minutes
		for t := intervalSec; t < durationSeconds-60; t += intervalSec {
			schedule.MidRolls = append(schedule.MidRolls, t)
		}
	}

	// For longer content (>30 min), include post-roll
	if durationSeconds > 1800 {
		schedule.PostRoll = true
	}

	return schedule, nil
}

// =====================================================
// VAST/VPAID GENERATION
// =====================================================

// GenerateVASTResponse creates a VAST 4.0 XML response
func (s *AVODService) GenerateVASTResponse(ctx context.Context, ad *Advertisement, impressionID string) string {
	skippableAttr := ""
	if ad.Skippable {
		skippableAttr = fmt.Sprintf(`skipoffset="00:00:%02d"`, ad.SkipAfter)
	}

	vast := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<VAST version="4.0">
  <Ad id="%s">
    <InLine>
      <AdSystem>IPTV-AVOD</AdSystem>
      <AdTitle><![CDATA[%s]]></AdTitle>
      <Impression><![CDATA[https://api.iptv.example.com/avod/impression/%s]]></Impression>
      <Creatives>
        <Creative id="%s" sequence="1">
          <Linear %s>
            <Duration>00:00:%02d</Duration>
            <TrackingEvents>
              <Tracking event="start"><![CDATA[https://api.iptv.example.com/avod/track/%s/start]]></Tracking>
              <Tracking event="firstQuartile"><![CDATA[https://api.iptv.example.com/avod/track/%s/q1]]></Tracking>
              <Tracking event="midpoint"><![CDATA[https://api.iptv.example.com/avod/track/%s/mid]]></Tracking>
              <Tracking event="thirdQuartile"><![CDATA[https://api.iptv.example.com/avod/track/%s/q3]]></Tracking>
              <Tracking event="complete"><![CDATA[https://api.iptv.example.com/avod/track/%s/complete]]></Tracking>
              <Tracking event="skip"><![CDATA[https://api.iptv.example.com/avod/track/%s/skip]]></Tracking>
            </TrackingEvents>
            <VideoClicks>
              <ClickThrough><![CDATA[%s]]></ClickThrough>
              <ClickTracking><![CDATA[https://api.iptv.example.com/avod/click/%s]]></ClickTracking>
            </VideoClicks>
            <MediaFiles>
              <MediaFile delivery="progressive" type="video/mp4" width="1920" height="1080" bitrate="2500">
                <![CDATA[%s]]>
              </MediaFile>
            </MediaFiles>
          </Linear>
        </Creative>
      </Creatives>
    </InLine>
  </Ad>
</VAST>`,
		ad.ID, ad.Title, impressionID,
		ad.ID, skippableAttr, ad.Duration,
		impressionID, impressionID, impressionID, impressionID, impressionID, impressionID,
		ad.ClickURL, impressionID, ad.MediaURL,
	)

	return vast
}

// =====================================================
// IMPRESSION & TRACKING
// =====================================================

// RecordImpression records when an ad was actually shown
func (s *AVODService) RecordImpression(ctx context.Context, impressionID string) error {
	query := `
		UPDATE ad_impressions
		SET shown_at = NOW(), confirmed = 1
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, impressionID)

	// Update ad statistics
	go s.updateAdStats(ctx, impressionID, "impression")
	return err
}

// TrackAdEvent tracks ad playback events (start, quartile, complete, skip)
func (s *AVODService) TrackAdEvent(ctx context.Context, impressionID, event string) error {
	updates := make(map[string]interface{})

	switch event {
	case "start":
		updates["started_at"] = time.Now()
	case "q1":
		updates["watched_seconds"] = "duration * 0.25"
	case "mid":
		updates["watched_seconds"] = "duration * 0.50"
	case "q3":
		updates["watched_seconds"] = "duration * 0.75"
	case "complete":
		updates["completed"] = true
		updates["watched_seconds"] = "duration"
	case "skip":
		updates["skipped"] = true
	}

	// Build dynamic update query
	if len(updates) == 0 {
		return nil
	}

	query := `UPDATE ad_impressions SET `
	args := []interface{}{}
	first := true
	for k, v := range updates {
		if !first {
			query += ", "
		}
		query += fmt.Sprintf("%s = ?", k)
		args = append(args, v)
		first = false
	}
	query += " WHERE id = ?"
	args = append(args, impressionID)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// RecordClick records when a user clicks on an ad
func (s *AVODService) RecordClick(ctx context.Context, impressionID string) error {
	query := `
		UPDATE ad_impressions
		SET clicked = 1, clicked_at = NOW()
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, impressionID)

	// Update ad click count
	go func() {
		updateQuery := `
			UPDATE advertisements
			SET clicks = clicks + 1
			WHERE id = (SELECT ad_id FROM ad_impressions WHERE id = ?)
		`
		s.db.ExecContext(ctx, updateQuery, impressionID)
	}()

	return err
}

// =====================================================
// REAL-TIME BIDDING (RTB)
// =====================================================

// runAuction runs a real-time bidding auction
func (s *AVODService) runAuction(candidates []*Advertisement, req *AdRequest) *Advertisement {
	if len(candidates) == 0 {
		return nil
	}

	// Score each ad based on multiple factors
	type scoredAd struct {
		ad    *Advertisement
		score float64
	}

	scored := make([]scoredAd, 0, len(candidates))
	for _, ad := range candidates {
		score := s.scoreAd(ad, req)
		scored = append(scored, scoredAd{ad: ad, score: score})
	}

	// Find highest scoring ad
	var winner *Advertisement
	maxScore := -1.0
	for _, sa := range scored {
		if sa.score > maxScore {
			maxScore = sa.score
			winner = sa.ad
		}
	}

	return winner
}

// scoreAd calculates a relevance score for an ad
func (s *AVODService) scoreAd(ad *Advertisement, req *AdRequest) float64 {
	score := 0.0

	// Base bid amount (most important)
	score += ad.BidAmount * 10

	// Targeting match bonuses
	for _, country := range ad.TargetCountries {
		if country == req.Country {
			score += 50
			break
		}
	}

	for _, genre := range ad.TargetGenres {
		for _, userGenre := range req.Genres {
			if strings.EqualFold(genre, userGenre) {
				score += 30
				break
			}
		}
	}

	// Age targeting
	if req.UserAge >= ad.MinAge && (ad.MaxAge == 0 || req.UserAge <= ad.MaxAge) {
		score += 20
	}

	// Frequency capping (less score if shown recently)
	if s.wasRecentlyShown(ad.ID, req.UserID) {
		score *= 0.5
	}

	// Budget pacing (slow down if budget running out)
	if ad.DailyBudget > 0 {
		spentRatio := ad.SpentToday / ad.DailyBudget
		if spentRatio > 0.9 {
			score *= 0.3 // Throttle
		}
	}

	// Add small random factor for variety
	score += rand.Float64() * 5

	return score
}

func (s *AVODService) wasRecentlyShown(adID, userID string) bool {
	var count int
	query := `
		SELECT COUNT(*) FROM ad_impressions
		WHERE ad_id = ? AND user_id = ?
		  AND shown_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)
	`
	s.db.QueryRow(query, adID, userID).Scan(&count)
	return count >= 2 // Cap at 2 impressions per hour per user
}

// =====================================================
// CAMPAIGN MANAGEMENT
// =====================================================

// CreateCampaign creates a new ad campaign
func (s *AVODService) CreateCampaign(ctx context.Context, campaign *AdCampaign) error {
	query := `
		INSERT INTO ad_campaigns (id, advertiser_id, name, description, budget, start_date, end_date, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		campaign.ID, campaign.AdvertiserID, campaign.Name, campaign.Description,
		campaign.Budget, campaign.StartDate, campaign.EndDate, campaign.IsActive,
	)
	return err
}

// CreateAdvertisement creates a new ad
func (s *AVODService) CreateAdvertisement(ctx context.Context, ad *Advertisement) error {
	query := `
		INSERT INTO advertisements (
			id, advertiser_id, campaign_id, title, type, format, media_url, click_url,
			duration, skippable, skip_after, bid_amount, daily_budget, total_budget,
			is_active, start_date, end_date, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		ad.ID, ad.AdvertiserID, ad.CampaignID, ad.Title, ad.Type, ad.Format,
		ad.MediaURL, ad.ClickURL, ad.Duration, ad.Skippable, ad.SkipAfter,
		ad.BidAmount, ad.DailyBudget, ad.TotalBudget, ad.IsActive,
		ad.StartDate, ad.EndDate,
	)
	return err
}

// =====================================================
// ANALYTICS
// =====================================================

// AdvertiserMetrics returns aggregate metrics for an advertiser
type AdvertiserMetrics struct {
	AdvertiserID  string  `json:"advertiser_id"`
	Impressions   int64   `json:"impressions"`
	UniqueViewers int64   `json:"unique_viewers"`
	Clicks        int64   `json:"clicks"`
	CTR           float64 `json:"ctr"`
	CompletionRate float64 `json:"completion_rate"`
	SkipRate      float64 `json:"skip_rate"`
	TotalSpent    float64 `json:"total_spent"`
	AvgCPM        float64 `json:"avg_cpm"`
	AvgCPC        float64 `json:"avg_cpc"`
}

// GetAdvertiserMetrics returns metrics for an advertiser
func (s *AVODService) GetAdvertiserMetrics(ctx context.Context, advertiserID string, days int) (*AdvertiserMetrics, error) {
	metrics := &AdvertiserMetrics{AdvertiserID: advertiserID}

	query := `
		SELECT
			COUNT(*) as impressions,
			COUNT(DISTINCT i.user_id) as unique_viewers,
			SUM(CASE WHEN i.clicked = 1 THEN 1 ELSE 0 END) as clicks,
			SUM(CASE WHEN i.completed = 1 THEN 1 ELSE 0 END) as completions,
			SUM(CASE WHEN i.skipped = 1 THEN 1 ELSE 0 END) as skips,
			SUM(i.revenue) as revenue
		FROM ad_impressions i
		JOIN advertisements a ON a.id = i.ad_id
		WHERE a.advertiser_id = ?
		  AND i.shown_at > DATE_SUB(NOW(), INTERVAL ? DAY)
	`

	var completions, skips int64
	err := s.db.QueryRowContext(ctx, query, advertiserID, days).Scan(
		&metrics.Impressions, &metrics.UniqueViewers, &metrics.Clicks,
		&completions, &skips, &metrics.TotalSpent,
	)
	if err != nil {
		return nil, err
	}

	if metrics.Impressions > 0 {
		metrics.CTR = float64(metrics.Clicks) / float64(metrics.Impressions) * 100
		metrics.CompletionRate = float64(completions) / float64(metrics.Impressions) * 100
		metrics.SkipRate = float64(skips) / float64(metrics.Impressions) * 100
		metrics.AvgCPM = metrics.TotalSpent / float64(metrics.Impressions) * 1000
	}
	if metrics.Clicks > 0 {
		metrics.AvgCPC = metrics.TotalSpent / float64(metrics.Clicks)
	}

	return metrics, nil
}

// =====================================================
// HELPERS
// =====================================================

func (s *AVODService) userHasAdFreeSubscription(ctx context.Context, userID string) (bool, error) {
	var hasAdFree bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users u
			JOIN packages p ON p.id = u.package_id
			WHERE u.id = ? AND p.ad_free = 1 AND u.is_active = 1
		)
	`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&hasAdFree)
	return hasAdFree, err
}

func (s *AVODService) getEligibleAds(ctx context.Context, req *AdRequest) ([]*Advertisement, error) {
	query := `
		SELECT id, advertiser_id, campaign_id, title, type, format, media_url, click_url,
			   duration, skippable, skip_after, bid_amount, daily_budget,
			   COALESCE(spent_today, 0), is_active
		FROM advertisements
		WHERE is_active = 1
		  AND start_date <= NOW()
		  AND end_date >= NOW()
		  AND COALESCE(spent_today, 0) < daily_budget
		  AND (type = ? OR type = 'any')
		ORDER BY bid_amount DESC
		LIMIT 20
	`
	rows, err := s.db.QueryContext(ctx, query, req.Position)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []*Advertisement
	for rows.Next() {
		ad := &Advertisement{}
		err := rows.Scan(
			&ad.ID, &ad.AdvertiserID, &ad.CampaignID, &ad.Title, &ad.Type, &ad.Format,
			&ad.MediaURL, &ad.ClickURL, &ad.Duration, &ad.Skippable, &ad.SkipAfter,
			&ad.BidAmount, &ad.DailyBudget, &ad.SpentToday, &ad.IsActive,
		)
		if err != nil {
			continue
		}
		ads = append(ads, ad)
	}
	return ads, nil
}

func (s *AVODService) getDefaultAd(ctx context.Context, req *AdRequest) (*Advertisement, error) {
	// Return platform's house ad when no commercial ad is available
	return &Advertisement{
		ID:        "default-house-ad",
		Title:     "Upgrade to Premium - Ad-Free Experience",
		Type:      req.Position,
		Format:    "video",
		MediaURL:  "https://cdn.iptv.example.com/ads/upgrade-premium.mp4",
		ClickURL:  "https://iptv.example.com/upgrade",
		Duration:  15,
		Skippable: true,
		SkipAfter: 5,
	}, nil
}

func (s *AVODService) preRecordImpression(ctx context.Context, adID string, req *AdRequest) string {
	id := uuid.New().String()
	query := `
		INSERT INTO ad_impressions (
			id, ad_id, user_id, stream_id, country, device_type, position, shown_at, confirmed
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), 0)
	`
	s.db.ExecContext(ctx, query, id, adID, req.UserID, req.StreamID,
		req.Country, req.DeviceType, req.Position)
	return id
}

func (s *AVODService) updateAdStats(ctx context.Context, impressionID, eventType string) {
	query := `
		UPDATE advertisements a
		JOIN ad_impressions i ON i.ad_id = a.id
		SET a.impressions = a.impressions + 1,
			a.spent_today = a.spent_today + (a.bid_amount / 1000),
			a.spent_total = a.spent_total + (a.bid_amount / 1000)
		WHERE i.id = ?
	`
	s.db.ExecContext(ctx, query, impressionID)
}
