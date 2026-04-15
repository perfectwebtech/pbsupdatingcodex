package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ReferralProgramService manages user referral and affiliate programs
// Reduces customer acquisition cost (CAC) by 50-70%
type ReferralProgramService struct {
	db     *sql.DB
	config *ReferralConfig
}

// NewReferralProgramService creates a new referral program service
func NewReferralProgramService(db *sql.DB, config *ReferralConfig) *ReferralProgramService {
	if config == nil {
		config = defaultReferralConfig()
	}
	return &ReferralProgramService{
		db:     db,
		config: config,
	}
}

// =====================================================
// MODELS
// =====================================================

// ReferralConfig defines referral program rules
type ReferralConfig struct {
	Enabled                bool    `json:"enabled"`
	ReferrerRewardType     string  `json:"referrer_reward_type"`     // credits, discount, cash
	ReferrerRewardAmount   float64 `json:"referrer_reward_amount"`   // Amount in USD or %
	ReferredRewardType     string  `json:"referred_reward_type"`     // discount, free_trial
	ReferredRewardAmount   float64 `json:"referred_reward_amount"`   // Amount in USD or %
	MinimumPurchaseAmount  float64 `json:"minimum_purchase_amount"`  // Min purchase for reward
	RewardExpiryDays       int     `json:"reward_expiry_days"`       // Days until reward expires
	MaxReferralsPerUser    int     `json:"max_referrals_per_user"`   // 0 = unlimited
	TierEnabled            bool    `json:"tier_enabled"`             // Enable tiered rewards
	AffiliateCommissionPercent float64 `json:"affiliate_commission_percent"`
}

// ReferralCode represents a unique referral code
type ReferralCode struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Code        string    `json:"code"`
	Type        string    `json:"type"` // user, affiliate, campaign
	IsActive    bool      `json:"is_active"`
	UsageCount  int       `json:"usage_count"`
	MaxUses     int       `json:"max_uses"` // 0 = unlimited
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	Metadata    string    `json:"metadata"` // JSON
}

// Referral tracks a referral relationship
type Referral struct {
	ID                string     `json:"id"`
	ReferrerID        string     `json:"referrer_id"`
	ReferredUserID    string     `json:"referred_user_id"`
	ReferralCode      string     `json:"referral_code"`
	Status            string     `json:"status"` // pending, qualified, rewarded, expired
	ReferrerReward    float64    `json:"referrer_reward"`
	ReferredReward    float64    `json:"referred_reward"`
	PurchaseAmount    float64    `json:"purchase_amount"`
	CommissionEarned  float64    `json:"commission_earned"`
	ReferredAt        time.Time  `json:"referred_at"`
	QualifiedAt       *time.Time `json:"qualified_at"`
	RewardedAt        *time.Time `json:"rewarded_at"`
}

// ReferralReward represents a reward earned from referrals
type ReferralReward struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ReferralID   string    `json:"referral_id"`
	Type         string    `json:"type"` // credits, discount, cash
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	Status       string    `json:"status"` // pending, issued, claimed, expired
	ExpiresAt    *time.Time `json:"expires_at"`
	IssuedAt     time.Time `json:"issued_at"`
	ClaimedAt    *time.Time `json:"claimed_at"`
}

// AffiliateTier defines commission tiers
type AffiliateTier struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	MinReferrals     int     `json:"min_referrals"`
	CommissionPercent float64 `json:"commission_percent"`
	BonusReward      float64 `json:"bonus_reward"`
}

// ReferralStats aggregates referral statistics
type ReferralStats struct {
	UserID             string  `json:"user_id"`
	TotalReferrals     int     `json:"total_referrals"`
	QualifiedReferrals int     `json:"qualified_referrals"`
	PendingReferrals   int     `json:"pending_referrals"`
	TotalEarnings      float64 `json:"total_earnings"`
	AvailableBalance   float64 `json:"available_balance"`
	WithdrawnBalance   float64 `json:"withdrawn_balance"`
	CurrentTier        string  `json:"current_tier"`
	NextTierReferrals  int     `json:"next_tier_referrals"`
	ConversionRate     float64 `json:"conversion_rate"`
}

// =====================================================
// REFERRAL CODE GENERATION
// =====================================================

// GenerateReferralCode creates a new referral code for a user
func (s *ReferralProgramService) GenerateReferralCode(ctx context.Context, userID string, codeType string) (*ReferralCode, error) {
	code := generateUniqueCode(8)

	referralCode := &ReferralCode{
		ID:        uuid.New().String(),
		UserID:    userID,
		Code:      code,
		Type:      codeType,
		IsActive:  true,
		MaxUses:   0,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO referral_codes (id, user_id, code, type, is_active, max_uses, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		referralCode.ID, userID, code, codeType, true, 0,
	)
	if err != nil {
		return nil, err
	}

	return referralCode, nil
}

// GetReferralCode retrieves a user's referral code
func (s *ReferralProgramService) GetReferralCode(ctx context.Context, userID string) (*ReferralCode, error) {
	var code ReferralCode
	query := `
		SELECT id, user_id, code, type, is_active, usage_count, max_uses, created_at
		FROM referral_codes
		WHERE user_id = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&code.ID, &code.UserID, &code.Code, &code.Type,
		&code.IsActive, &code.UsageCount, &code.MaxUses, &code.CreatedAt,
	)
	if err == sql.ErrNoRows {
		// Generate new code if none exists
		return s.GenerateReferralCode(ctx, userID, "user")
	}
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// ValidateReferralCode checks if a referral code is valid
func (s *ReferralProgramService) ValidateReferralCode(ctx context.Context, code string) (bool, string, error) {
	var userID string
	var isActive bool
	var usageCount, maxUses int
	var expiresAt *time.Time

	query := `
		SELECT user_id, is_active, usage_count, max_uses, expires_at
		FROM referral_codes
		WHERE code = ?
	`
	err := s.db.QueryRowContext(ctx, query, code).Scan(
		&userID, &isActive, &usageCount, &maxUses, &expiresAt,
	)
	if err != nil {
		return false, "", err
	}

	if !isActive {
		return false, "", errors.New("referral code is inactive")
	}

	if maxUses > 0 && usageCount >= maxUses {
		return false, "", errors.New("referral code has reached maximum uses")
	}

	if expiresAt != nil && time.Now().After(*expiresAt) {
		return false, "", errors.New("referral code has expired")
	}

	return true, userID, nil
}

// =====================================================
// REFERRAL TRACKING
// =====================================================

// RecordReferral records a new referral when a user signs up with a code
func (s *ReferralProgramService) RecordReferral(ctx context.Context, referredUserID, referralCode string) (*Referral, error) {
	// Validate code
	valid, referrerID, err := s.ValidateReferralCode(ctx, referralCode)
	if !valid || err != nil {
		return nil, fmt.Errorf("invalid referral code: %w", err)
	}

	// Prevent self-referral
	if referrerID == referredUserID {
		return nil, errors.New("cannot refer yourself")
	}

	// Check if user already used a referral code
	var existingCount int
	checkQuery := `SELECT COUNT(*) FROM referrals WHERE referred_user_id = ?`
	s.db.QueryRowContext(ctx, checkQuery, referredUserID).Scan(&existingCount)
	if existingCount > 0 {
		return nil, errors.New("user already used a referral code")
	}

	referral := &Referral{
		ID:             uuid.New().String(),
		ReferrerID:     referrerID,
		ReferredUserID: referredUserID,
		ReferralCode:   referralCode,
		Status:         "pending",
		ReferredAt:     time.Now(),
	}

	query := `
		INSERT INTO referrals (id, referrer_id, referred_user_id, referral_code, status, referred_at)
		VALUES (?, ?, ?, ?, 'pending', NOW())
	`
	_, err = s.db.ExecContext(ctx, query,
		referral.ID, referrerID, referredUserID, referralCode,
	)
	if err != nil {
		return nil, err
	}

	// Update code usage count
	updateCodeQuery := `UPDATE referral_codes SET usage_count = usage_count + 1 WHERE code = ?`
	s.db.ExecContext(ctx, updateCodeQuery, referralCode)

	// Issue instant reward for referred user (welcome discount)
	s.IssueReferredUserReward(ctx, referredUserID, referral.ID)

	return referral, nil
}

// QualifyReferral marks a referral as qualified (when referred user makes purchase)
func (s *ReferralProgramService) QualifyReferral(ctx context.Context, referredUserID string, purchaseAmount float64) error {
	// Check minimum purchase requirement
	if purchaseAmount < s.config.MinimumPurchaseAmount {
		return errors.New("purchase amount below minimum for referral reward")
	}

	// Get referral
	var referralID, referrerID string
	query := `
		SELECT id, referrer_id
		FROM referrals
		WHERE referred_user_id = ? AND status = 'pending'
	`
	err := s.db.QueryRowContext(ctx, query, referredUserID).Scan(&referralID, &referrerID)
	if err != nil {
		return err // No pending referral
	}

	// Calculate rewards
	referrerReward := s.calculateReferrerReward(purchaseAmount)
	commissionEarned := s.calculateCommission(purchaseAmount)

	// Update referral status
	updateQuery := `
		UPDATE referrals
		SET status = 'qualified',
		    qualified_at = NOW(),
		    purchase_amount = ?,
		    referrer_reward = ?,
		    commission_earned = ?
		WHERE id = ?
	`
	_, err = s.db.ExecContext(ctx, updateQuery,
		purchaseAmount, referrerReward, commissionEarned, referralID,
	)
	if err != nil {
		return err
	}

	// Issue reward to referrer
	s.IssueReferrerReward(ctx, referrerID, referralID, referrerReward)

	// Check for tier upgrade
	s.CheckAndUpgradeTier(ctx, referrerID)

	return nil
}

// =====================================================
// REWARDS
// =====================================================

// IssueReferredUserReward gives instant reward to new user
func (s *ReferralProgramService) IssueReferredUserReward(ctx context.Context, userID, referralID string) error {
	reward := &ReferralReward{
		ID:         uuid.New().String(),
		UserID:     userID,
		ReferralID: referralID,
		Type:       s.config.ReferredRewardType,
		Amount:     s.config.ReferredRewardAmount,
		Currency:   "USD",
		Status:     "issued",
		IssuedAt:   time.Now(),
	}

	if s.config.RewardExpiryDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, s.config.RewardExpiryDays)
		reward.ExpiresAt = &expiresAt
	}

	query := `
		INSERT INTO referral_rewards (id, user_id, referral_id, type, amount, currency, status, issued_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, 'issued', NOW(), ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		reward.ID, userID, referralID, reward.Type, reward.Amount, reward.Currency, reward.ExpiresAt,
	)

	// Auto-apply discount if applicable
	if reward.Type == "discount" {
		s.applyDiscountToUser(ctx, userID, reward.Amount)
	}

	return err
}

// IssueReferrerReward gives reward to referrer when referred user qualifies
func (s *ReferralProgramService) IssueReferrerReward(ctx context.Context, userID, referralID string, amount float64) error {
	reward := &ReferralReward{
		ID:         uuid.New().String(),
		UserID:     userID,
		ReferralID: referralID,
		Type:       s.config.ReferrerRewardType,
		Amount:     amount,
		Currency:   "USD",
		Status:     "issued",
		IssuedAt:   time.Now(),
	}

	if s.config.RewardExpiryDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, s.config.RewardExpiryDays)
		reward.ExpiresAt = &expiresAt
	}

	query := `
		INSERT INTO referral_rewards (id, user_id, referral_id, type, amount, currency, status, issued_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, 'issued', NOW(), ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		reward.ID, userID, referralID, reward.Type, reward.Amount, reward.Currency, reward.ExpiresAt,
	)

	// Update referral as rewarded
	updateQuery := `UPDATE referrals SET status = 'rewarded', rewarded_at = NOW() WHERE id = ?`
	s.db.ExecContext(ctx, updateQuery, referralID)

	// Add to user balance if cash reward
	if reward.Type == "cash" || reward.Type == "credits" {
		s.addToUserBalance(ctx, userID, amount)
	}

	return err
}

func (s *ReferralProgramService) calculateReferrerReward(purchaseAmount float64) float64 {
	if s.config.ReferrerRewardType == "credits" || s.config.ReferrerRewardType == "cash" {
		return s.config.ReferrerRewardAmount
	}
	// Percentage-based
	return purchaseAmount * (s.config.ReferrerRewardAmount / 100.0)
}

func (s *ReferralProgramService) calculateCommission(purchaseAmount float64) float64 {
	return purchaseAmount * (s.config.AffiliateCommissionPercent / 100.0)
}

// =====================================================
// STATISTICS & REPORTING
// =====================================================

// GetReferralStats returns detailed referral statistics for a user
func (s *ReferralProgramService) GetReferralStats(ctx context.Context, userID string) (*ReferralStats, error) {
	stats := &ReferralStats{
		UserID: userID,
	}

	// Total referrals
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status IN ('qualified', 'rewarded') THEN 1 ELSE 0 END) as qualified,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending
		FROM referrals
		WHERE referrer_id = ?
	`
	s.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalReferrals, &stats.QualifiedReferrals, &stats.PendingReferrals,
	)

	// Earnings
	earningsQuery := `
		SELECT
			SUM(CASE WHEN status = 'issued' OR status = 'claimed' THEN amount ELSE 0 END) as total,
			SUM(CASE WHEN status = 'issued' THEN amount ELSE 0 END) as available,
			SUM(CASE WHEN status = 'claimed' THEN amount ELSE 0 END) as withdrawn
		FROM referral_rewards
		WHERE user_id = ?
	`
	s.db.QueryRowContext(ctx, earningsQuery, userID).Scan(
		&stats.TotalEarnings, &stats.AvailableBalance, &stats.WithdrawnBalance,
	)

	// Conversion rate
	if stats.TotalReferrals > 0 {
		stats.ConversionRate = (float64(stats.QualifiedReferrals) / float64(stats.TotalReferrals)) * 100
	}

	// Current tier
	tier := s.getCurrentTier(ctx, userID, stats.QualifiedReferrals)
	stats.CurrentTier = tier.Name

	return stats, nil
}

// GetTopReferrers returns the top referring users
func (s *ReferralProgramService) GetTopReferrers(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT
			u.id, u.username, u.email,
			COUNT(r.id) as total_referrals,
			SUM(r.commission_earned) as total_commission
		FROM users u
		JOIN referrals r ON r.referrer_id = u.id
		WHERE r.status IN ('qualified', 'rewarded')
		GROUP BY u.id
		ORDER BY total_referrals DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var referrers []map[string]interface{}
	for rows.Next() {
		var id, username, email string
		var totalReferrals int
		var totalCommission float64

		rows.Scan(&id, &username, &email, &totalReferrals, &totalCommission)
		referrers = append(referrers, map[string]interface{}{
			"user_id":          id,
			"username":         username,
			"email":            email,
			"total_referrals":  totalReferrals,
			"total_commission": totalCommission,
		})
	}

	return referrers, nil
}

// =====================================================
// AFFILIATE TIERS
// =====================================================

func (s *ReferralProgramService) getCurrentTier(ctx context.Context, userID string, qualifiedReferrals int) *AffiliateTier {
	// Predefined tiers
	tiers := []AffiliateTier{
		{ID: "bronze", Name: "Bronze", MinReferrals: 0, CommissionPercent: 10, BonusReward: 0},
		{ID: "silver", Name: "Silver", MinReferrals: 10, CommissionPercent: 15, BonusReward: 50},
		{ID: "gold", Name: "Gold", MinReferrals: 25, CommissionPercent: 20, BonusReward: 150},
		{ID: "platinum", Name: "Platinum", MinReferrals: 50, CommissionPercent: 25, BonusReward: 500},
		{ID: "diamond", Name: "Diamond", MinReferrals: 100, CommissionPercent: 30, BonusReward: 1500},
	}

	currentTier := &tiers[0]
	for i := range tiers {
		if qualifiedReferrals >= tiers[i].MinReferrals {
			currentTier = &tiers[i]
		}
	}

	return currentTier
}

func (s *ReferralProgramService) CheckAndUpgradeTier(ctx context.Context, userID string) {
	var qualifiedCount int
	query := `SELECT COUNT(*) FROM referrals WHERE referrer_id = ? AND status IN ('qualified', 'rewarded')`
	s.db.QueryRowContext(ctx, query, userID).Scan(&qualifiedCount)

	currentTier := s.getCurrentTier(ctx, userID, qualifiedCount)

	// Check if tier changed and issue bonus
	var lastTier string
	tierQuery := `SELECT affiliate_tier FROM users WHERE id = ?`
	s.db.QueryRowContext(ctx, tierQuery, userID).Scan(&lastTier)

	if lastTier != currentTier.ID && currentTier.BonusReward > 0 {
		// Issue tier upgrade bonus
		s.issueTierBonus(ctx, userID, currentTier)

		// Update user tier
		updateQuery := `UPDATE users SET affiliate_tier = ? WHERE id = ?`
		s.db.ExecContext(ctx, updateQuery, currentTier.ID, userID)
	}
}

func (s *ReferralProgramService) issueTierBonus(ctx context.Context, userID string, tier *AffiliateTier) {
	query := `
		INSERT INTO referral_rewards (id, user_id, type, amount, currency, status, issued_at)
		VALUES (?, ?, 'tier_bonus', ?, 'USD', 'issued', NOW())
	`
	s.db.ExecContext(ctx, query, uuid.New().String(), userID, tier.BonusReward)

	s.addToUserBalance(ctx, userID, tier.BonusReward)
}

// =====================================================
// HELPERS
// =====================================================

func (s *ReferralProgramService) applyDiscountToUser(ctx context.Context, userID string, amount float64) {
	query := `UPDATE users SET discount_balance = discount_balance + ? WHERE id = ?`
	s.db.ExecContext(ctx, query, amount, userID)
}

func (s *ReferralProgramService) addToUserBalance(ctx context.Context, userID string, amount float64) {
	query := `UPDATE users SET referral_balance = referral_balance + ? WHERE id = ?`
	s.db.ExecContext(ctx, query, amount, userID)
}

func generateUniqueCode(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	code := base64.URLEncoding.EncodeToString(b)[:length]
	return strings.ToUpper(code)
}

func defaultReferralConfig() *ReferralConfig {
	return &ReferralConfig{
		Enabled:                    true,
		ReferrerRewardType:         "cash",
		ReferrerRewardAmount:       10.00,
		ReferredRewardType:         "discount",
		ReferredRewardAmount:       5.00,
		MinimumPurchaseAmount:      9.99,
		RewardExpiryDays:           90,
		MaxReferralsPerUser:        0, // Unlimited
		TierEnabled:                true,
		AffiliateCommissionPercent: 10.0,
	}
}
