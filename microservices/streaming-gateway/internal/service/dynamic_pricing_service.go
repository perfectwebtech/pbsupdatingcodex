package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

// DynamicPricingService provides AI-powered dynamic pricing optimization
type DynamicPricingService struct {
	db *sql.DB
}

// NewDynamicPricingService creates a new dynamic pricing service
func NewDynamicPricingService(db *sql.DB) *DynamicPricingService {
	return &DynamicPricingService{db: db}
}

// =====================================================
// ERRORS
// =====================================================

var (
	ErrPricingRuleNotFound = errors.New("pricing rule not found")
	ErrInvalidPriceRange   = errors.New("invalid price range")
	ErrPricingTierNotFound = errors.New("pricing tier not found")
)

// =====================================================
// MODELS
// =====================================================

// PricingRule represents a dynamic pricing rule
type PricingRule struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	PackageID       string    `json:"package_id"`
	RuleType        string    `json:"rule_type"` // demand, geo, time, promotional, competitor
	Conditions      string    `json:"conditions"` // JSON conditions
	PriceMultiplier float64   `json:"price_multiplier"`
	MinPrice        float64   `json:"min_price"`
	MaxPrice        float64   `json:"max_price"`
	Priority        int       `json:"priority"`
	IsActive        bool      `json:"is_active"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PriceCalculation represents a calculated price for a user
type PriceCalculation struct {
	PackageID       string    `json:"package_id"`
	UserID          string    `json:"user_id"`
	BasePrice       float64   `json:"base_price"`
	FinalPrice      float64   `json:"final_price"`
	Discount        float64   `json:"discount"`
	DiscountPercent float64   `json:"discount_percent"`
	Currency        string    `json:"currency"`
	AppliedRules    []string  `json:"applied_rules"`
	Reason          string    `json:"reason"`
	ValidUntil      time.Time `json:"valid_until"`
	GeoCountry      string    `json:"geo_country"`
}

// GeoPricing represents country-specific pricing (PPP)
type GeoPricing struct {
	CountryCode    string  `json:"country_code"`
	CountryName    string  `json:"country_name"`
	PriceMultiplier float64 `json:"price_multiplier"`
	Currency       string  `json:"currency"`
	ExchangeRate   float64 `json:"exchange_rate"`
}

// DemandMetrics represents real-time demand data
type DemandMetrics struct {
	PackageID       string  `json:"package_id"`
	ActiveUsers     int     `json:"active_users"`
	NewSignups24h   int     `json:"new_signups_24h"`
	ChurnRate       float64 `json:"churn_rate"`
	ConversionRate  float64 `json:"conversion_rate"`
	DemandScore     float64 `json:"demand_score"` // 0-100
	CompetitorPrice float64 `json:"competitor_price"`
}

// =====================================================
// DYNAMIC PRICING CALCULATION
// =====================================================

// CalculatePrice calculates the optimal price for a user
func (s *DynamicPricingService) CalculatePrice(ctx context.Context, userID, packageID, countryCode string) (*PriceCalculation, error) {
	// Get base package price
	basePrice, currency, err := s.getBasePrice(ctx, packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get base price: %w", err)
	}

	calc := &PriceCalculation{
		PackageID:    packageID,
		UserID:       userID,
		BasePrice:    basePrice,
		FinalPrice:   basePrice,
		Currency:     currency,
		AppliedRules: []string{},
		ValidUntil:   time.Now().Add(24 * time.Hour),
		GeoCountry:   countryCode,
	}

	// Apply geo-based pricing (Purchasing Power Parity)
	if geoMultiplier, err := s.getGeoMultiplier(ctx, countryCode); err == nil {
		calc.FinalPrice *= geoMultiplier
		if geoMultiplier != 1.0 {
			calc.AppliedRules = append(calc.AppliedRules, "geo_ppp_adjustment")
		}
	}

	// Apply demand-based pricing
	demandMultiplier := s.calculateDemandMultiplier(ctx, packageID)
	calc.FinalPrice *= demandMultiplier
	if demandMultiplier != 1.0 {
		calc.AppliedRules = append(calc.AppliedRules, "demand_pricing")
	}

	// Apply time-based promotions
	if promo := s.getActivePromotion(ctx, packageID); promo != nil {
		calc.FinalPrice *= promo.PriceMultiplier
		calc.AppliedRules = append(calc.AppliedRules, fmt.Sprintf("promo_%s", promo.Name))
		calc.Reason = fmt.Sprintf("Promotional pricing: %s", promo.Name)
	}

	// Apply user-specific discount (loyalty, retention)
	userDiscount := s.calculateUserDiscount(ctx, userID)
	if userDiscount > 0 {
		calc.FinalPrice *= (1 - userDiscount)
		calc.AppliedRules = append(calc.AppliedRules, "user_loyalty_discount")
	}

	// Round to 2 decimal places
	calc.FinalPrice = math.Round(calc.FinalPrice*100) / 100

	// Calculate discount
	calc.Discount = calc.BasePrice - calc.FinalPrice
	if calc.BasePrice > 0 {
		calc.DiscountPercent = math.Round((calc.Discount/calc.BasePrice)*10000) / 100
	}

	// Cache the calculation
	s.cachePriceCalculation(ctx, calc)

	return calc, nil
}

// CalculateChurnPreventionPrice calculates a special price for users at risk of churning
func (s *DynamicPricingService) CalculateChurnPreventionPrice(ctx context.Context, userID, packageID string) (*PriceCalculation, error) {
	// Get user's churn risk score
	churnRisk, err := s.getUserChurnRisk(ctx, userID)
	if err != nil {
		return nil, err
	}

	basePrice, currency, _ := s.getBasePrice(ctx, packageID)

	// Discount based on churn risk (higher risk = bigger discount)
	var discountPercent float64
	switch {
	case churnRisk > 0.8:
		discountPercent = 0.30 // 30% discount for high-risk
	case churnRisk > 0.6:
		discountPercent = 0.20 // 20% discount
	case churnRisk > 0.4:
		discountPercent = 0.10 // 10% discount
	default:
		discountPercent = 0.0
	}

	finalPrice := basePrice * (1 - discountPercent)

	return &PriceCalculation{
		PackageID:       packageID,
		UserID:          userID,
		BasePrice:       basePrice,
		FinalPrice:      math.Round(finalPrice*100) / 100,
		Discount:        basePrice - finalPrice,
		DiscountPercent: discountPercent * 100,
		Currency:        currency,
		AppliedRules:    []string{"churn_prevention"},
		Reason:          fmt.Sprintf("Special offer (churn risk: %.0f%%)", churnRisk*100),
		ValidUntil:      time.Now().Add(72 * time.Hour),
	}, nil
}

// =====================================================
// PRICING RULES MANAGEMENT
// =====================================================

// CreatePricingRule creates a new dynamic pricing rule
func (s *DynamicPricingService) CreatePricingRule(ctx context.Context, rule *PricingRule) error {
	query := `
		INSERT INTO pricing_rules (
			id, name, package_id, rule_type, conditions,
			price_multiplier, min_price, max_price, priority,
			is_active, starts_at, ends_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.PackageID, rule.RuleType, rule.Conditions,
		rule.PriceMultiplier, rule.MinPrice, rule.MaxPrice, rule.Priority,
		rule.IsActive, rule.StartsAt, rule.EndsAt,
	)
	return err
}

// ListPricingRules returns all active pricing rules
func (s *DynamicPricingService) ListPricingRules(ctx context.Context, packageID string) ([]*PricingRule, error) {
	query := `
		SELECT id, name, package_id, rule_type, conditions,
			   price_multiplier, min_price, max_price, priority,
			   is_active, starts_at, ends_at, created_at, updated_at
		FROM pricing_rules
		WHERE package_id = ? AND is_active = 1
		ORDER BY priority DESC
	`
	rows, err := s.db.QueryContext(ctx, query, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*PricingRule
	for rows.Next() {
		r := &PricingRule{}
		err := rows.Scan(
			&r.ID, &r.Name, &r.PackageID, &r.RuleType, &r.Conditions,
			&r.PriceMultiplier, &r.MinPrice, &r.MaxPrice, &r.Priority,
			&r.IsActive, &r.StartsAt, &r.EndsAt, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

// =====================================================
// A/B PRICE TESTING
// =====================================================

// PriceTest represents an A/B price test
type PriceTest struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	PackageID     string    `json:"package_id"`
	VariantA      float64   `json:"variant_a_price"`
	VariantB      float64   `json:"variant_b_price"`
	TrafficSplit  float64   `json:"traffic_split"` // 0-1, percentage going to B
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	IsActive      bool      `json:"is_active"`
	ConversionsA  int       `json:"conversions_a"`
	ConversionsB  int       `json:"conversions_b"`
	RevenueA      float64   `json:"revenue_a"`
	RevenueB      float64   `json:"revenue_b"`
}

// AssignPriceVariant assigns a user to a price test variant
func (s *DynamicPricingService) AssignPriceVariant(ctx context.Context, userID, testID string) (string, float64, error) {
	// Get active test
	test := &PriceTest{}
	query := `
		SELECT id, package_id, variant_a_price, variant_b_price, traffic_split
		FROM price_tests
		WHERE id = ? AND is_active = 1 AND start_date <= NOW() AND end_date >= NOW()
	`
	err := s.db.QueryRowContext(ctx, query, testID).Scan(
		&test.ID, &test.PackageID, &test.VariantA, &test.VariantB, &test.TrafficSplit,
	)
	if err != nil {
		return "", 0, err
	}

	// Deterministic assignment based on user ID hash
	hash := hashUserIDForVariant(userID)
	variant := "A"
	price := test.VariantA
	if hash < test.TrafficSplit {
		variant = "B"
		price = test.VariantB
	}

	// Record assignment
	insertQuery := `
		INSERT INTO price_test_assignments (user_id, test_id, variant, assigned_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE variant = variant
	`
	s.db.ExecContext(ctx, insertQuery, userID, testID, variant)

	return variant, price, nil
}

// =====================================================
// HELPER METHODS
// =====================================================

func (s *DynamicPricingService) getBasePrice(ctx context.Context, packageID string) (float64, string, error) {
	var price float64
	var currency string
	query := `SELECT price, COALESCE(currency, 'USD') FROM packages WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, packageID).Scan(&price, &currency)
	return price, currency, err
}

func (s *DynamicPricingService) getGeoMultiplier(ctx context.Context, countryCode string) (float64, error) {
	// PPP-based pricing multipliers
	// In production, these would come from a database or external API
	pppMultipliers := map[string]float64{
		"US": 1.00, "CA": 0.95, "GB": 1.10, "DE": 1.00, "FR": 1.00,
		"AU": 1.05, "JP": 0.90, "CN": 0.55, "IN": 0.30, "BR": 0.50,
		"MX": 0.55, "RU": 0.45, "TR": 0.40, "ID": 0.35, "PH": 0.40,
		"VN": 0.35, "TH": 0.50, "EG": 0.30, "NG": 0.25, "KE": 0.30,
		"ZA": 0.55, "AR": 0.45, "CO": 0.50, "PE": 0.50, "PK": 0.30,
	}
	if mult, ok := pppMultipliers[countryCode]; ok {
		return mult, nil
	}
	return 1.0, nil // Default to no adjustment
}

func (s *DynamicPricingService) calculateDemandMultiplier(ctx context.Context, packageID string) float64 {
	metrics := s.getDemandMetrics(ctx, packageID)
	if metrics == nil {
		return 1.0
	}

	// Higher demand = higher price (within limits)
	if metrics.DemandScore > 80 {
		return 1.10 // +10% during peak demand
	} else if metrics.DemandScore > 60 {
		return 1.05 // +5%
	} else if metrics.DemandScore < 30 {
		return 0.90 // -10% during low demand to attract users
	} else if metrics.DemandScore < 50 {
		return 0.95 // -5%
	}
	return 1.0
}

func (s *DynamicPricingService) getDemandMetrics(ctx context.Context, packageID string) *DemandMetrics {
	metrics := &DemandMetrics{PackageID: packageID}
	query := `
		SELECT
			(SELECT COUNT(*) FROM users WHERE package_id = ? AND is_active = 1) as active_users,
			(SELECT COUNT(*) FROM users WHERE package_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL 24 HOUR)) as new_signups
	`
	s.db.QueryRowContext(ctx, query, packageID, packageID).Scan(
		&metrics.ActiveUsers, &metrics.NewSignups24h,
	)

	// Calculate demand score (0-100)
	// In production, this would use ML models
	metrics.DemandScore = math.Min(100, float64(metrics.NewSignups24h)*2)
	return metrics
}

func (s *DynamicPricingService) getActivePromotion(ctx context.Context, packageID string) *PricingRule {
	query := `
		SELECT id, name, package_id, rule_type, price_multiplier, priority
		FROM pricing_rules
		WHERE package_id = ?
		  AND rule_type = 'promotional'
		  AND is_active = 1
		  AND starts_at <= NOW()
		  AND ends_at >= NOW()
		ORDER BY priority DESC
		LIMIT 1
	`
	rule := &PricingRule{}
	err := s.db.QueryRowContext(ctx, query, packageID).Scan(
		&rule.ID, &rule.Name, &rule.PackageID, &rule.RuleType, &rule.PriceMultiplier, &rule.Priority,
	)
	if err != nil {
		return nil
	}
	return rule
}

func (s *DynamicPricingService) calculateUserDiscount(ctx context.Context, userID string) float64 {
	// Loyalty discount based on subscription duration
	var months int
	query := `
		SELECT TIMESTAMPDIFF(MONTH, created_at, NOW()) as months
		FROM users
		WHERE id = ?
	`
	s.db.QueryRowContext(ctx, query, userID).Scan(&months)

	switch {
	case months >= 24:
		return 0.15 // 15% loyalty discount for 2+ years
	case months >= 12:
		return 0.10 // 10% for 1+ year
	case months >= 6:
		return 0.05 // 5% for 6+ months
	default:
		return 0.0
	}
}

func (s *DynamicPricingService) getUserChurnRisk(ctx context.Context, userID string) (float64, error) {
	var risk float64
	query := `SELECT COALESCE(churn_risk_score, 0) FROM user_analytics WHERE user_id = ?`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&risk)
	return risk, err
}

func (s *DynamicPricingService) cachePriceCalculation(ctx context.Context, calc *PriceCalculation) {
	// Cache in database for analytics
	query := `
		INSERT INTO price_calculations (
			user_id, package_id, base_price, final_price,
			discount_percent, country_code, applied_rules, calculated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`
	rules := ""
	for i, r := range calc.AppliedRules {
		if i > 0 {
			rules += ","
		}
		rules += r
	}
	s.db.ExecContext(ctx, query,
		calc.UserID, calc.PackageID, calc.BasePrice, calc.FinalPrice,
		calc.DiscountPercent, calc.GeoCountry, rules,
	)
}

func hashUserIDForVariant(userID string) float64 {
	// Simple deterministic hash for A/B variant assignment
	var sum int
	for _, c := range userID {
		sum += int(c)
	}
	return float64(sum%100) / 100.0
}
