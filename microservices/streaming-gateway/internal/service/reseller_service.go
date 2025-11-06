package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.uber.org/zap"
)

var (
	ErrResellerNotFound         = errors.New("reseller not found")
	ErrUserAlreadyReseller      = errors.New("user is already a reseller")
	ErrParentResellerNotFound   = errors.New("parent reseller not found")
	ErrInsufficientCredits      = errors.New("insufficient credits")
	ErrResellerMaxUsersReached  = errors.New("reseller maximum users limit reached")
	ErrUserAlreadyAssigned      = errors.New("user already assigned to a reseller")
)

type ResellerService struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewResellerService(db *sql.DB, logger *zap.Logger) *ResellerService {
	return &ResellerService{
		db:     db,
		logger: logger,
	}
}

type Reseller struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"user_id"`
	Username           string     `json:"username"`
	Email              string     `json:"email"`
	ParentID           *int64     `json:"parent_id"`
	ParentUsername     *string    `json:"parent_username,omitempty"`
	Credits            float64    `json:"credits"`
	CommissionRate     float64    `json:"commission_rate"`
	CanCreateResellers bool       `json:"can_create_resellers"`
	MaxUsers           int        `json:"max_users"`
	MaxResellers       int        `json:"max_resellers"`
	CurrentUsers       int        `json:"current_users"`
	CurrentResellers   int        `json:"current_resellers"`
	IsActive           bool       `json:"is_active"`
	Notes              string     `json:"notes,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CreateResellerRequest struct {
	UserID             int64
	ParentID           *int64
	Credits            float64
	CommissionRate     float64
	CanCreateResellers bool
	MaxUsers           int
	MaxResellers       int
	Notes              string
	CreatedBy          int64
}

type UpdateResellerRequest struct {
	CommissionRate     float64
	CanCreateResellers bool
	MaxUsers           int
	MaxResellers       int
	Notes              string
}

type CreditTransaction struct {
	ID              int64     `json:"id"`
	ResellerID      int64     `json:"reseller_id"`
	Amount          float64   `json:"amount"`
	BalanceBefore   float64   `json:"balance_before"`
	BalanceAfter    float64   `json:"balance_after"`
	TransactionType string    `json:"transaction_type"`
	Description     string    `json:"description"`
	CreatedBy       *int64    `json:"created_by"`
	CreatedByName   *string   `json:"created_by_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// ListResellers returns a paginated list of resellers
func (s *ResellerService) ListResellers(ctx context.Context, page, limit int, filters map[string]interface{}) ([]Reseller, int, error) {
	offset := (page - 1) * limit

	query := `
		SELECT
			r.id, r.user_id, u.username, u.email, r.parent_id,
			pu.username as parent_username,
			r.credits, r.commission_rate, r.can_create_resellers,
			r.max_users, r.max_resellers, r.is_active, r.notes,
			r.created_at, r.updated_at,
			COALESCE((SELECT COUNT(*) FROM reseller_assignments WHERE reseller_id = r.id), 0) as current_users,
			COALESCE((SELECT COUNT(*) FROM resellers WHERE parent_id = r.id), 0) as current_resellers
		FROM resellers r
		INNER JOIN users u ON r.user_id = u.id
		LEFT JOIN resellers pr ON r.parent_id = pr.id
		LEFT JOIN users pu ON pr.user_id = pu.id
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM resellers r INNER JOIN users u ON r.user_id = u.id WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	// Add filters
	if search, ok := filters["search"].(string); ok && search != "" {
		query += ` AND (u.username LIKE $` + string(rune(argCount)) + ` OR u.email LIKE $` + string(rune(argCount)) + `)`
		countQuery += ` AND (u.username LIKE $` + string(rune(argCount)) + ` OR u.email LIKE $` + string(rune(argCount)) + `)`
		args = append(args, "%"+search+"%")
		argCount++
	}

	if parentID, ok := filters["parent_id"].(int64); ok {
		query += ` AND r.parent_id = $` + string(rune(argCount))
		countQuery += ` AND r.parent_id = $` + string(rune(argCount))
		args = append(args, parentID)
		argCount++
	}

	query += ` ORDER BY r.created_at DESC LIMIT $` + string(rune(argCount)) + ` OFFSET $` + string(rune(argCount+1))
	args = append(args, limit, offset)

	// Get total count
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get resellers
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	resellers := []Reseller{}
	for rows.Next() {
		var r Reseller
		err := rows.Scan(
			&r.ID, &r.UserID, &r.Username, &r.Email, &r.ParentID,
			&r.ParentUsername,
			&r.Credits, &r.CommissionRate, &r.CanCreateResellers,
			&r.MaxUsers, &r.MaxResellers, &r.IsActive, &r.Notes,
			&r.CreatedAt, &r.UpdatedAt,
			&r.CurrentUsers, &r.CurrentResellers,
		)
		if err != nil {
			return nil, 0, err
		}
		resellers = append(resellers, r)
	}

	return resellers, total, nil
}

// GetResellerByID returns a single reseller
func (s *ResellerService) GetResellerByID(ctx context.Context, resellerID int64) (*Reseller, error) {
	query := `
		SELECT
			r.id, r.user_id, u.username, u.email, r.parent_id,
			pu.username as parent_username,
			r.credits, r.commission_rate, r.can_create_resellers,
			r.max_users, r.max_resellers, r.is_active, r.notes,
			r.created_at, r.updated_at,
			COALESCE((SELECT COUNT(*) FROM reseller_assignments WHERE reseller_id = r.id), 0) as current_users,
			COALESCE((SELECT COUNT(*) FROM resellers WHERE parent_id = r.id), 0) as current_resellers
		FROM resellers r
		INNER JOIN users u ON r.user_id = u.id
		LEFT JOIN resellers pr ON r.parent_id = pr.id
		LEFT JOIN users pu ON pr.user_id = pu.id
		WHERE r.id = $1
	`

	var r Reseller
	err := s.db.QueryRowContext(ctx, query, resellerID).Scan(
		&r.ID, &r.UserID, &r.Username, &r.Email, &r.ParentID,
		&r.ParentUsername,
		&r.Credits, &r.CommissionRate, &r.CanCreateResellers,
		&r.MaxUsers, &r.MaxResellers, &r.IsActive, &r.Notes,
		&r.CreatedAt, &r.UpdatedAt,
		&r.CurrentUsers, &r.CurrentResellers,
	)

	if err == sql.ErrNoRows {
		return nil, ErrResellerNotFound
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// CreateReseller creates a new reseller
func (s *ResellerService) CreateReseller(ctx context.Context, req *CreateResellerRequest) (*Reseller, error) {
	// Check if user is already a reseller
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM resellers WHERE user_id = $1)", req.UserID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyReseller
	}

	// If parent is specified, verify it exists
	if req.ParentID != nil {
		err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM resellers WHERE id = $1)", *req.ParentID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrParentResellerNotFound
		}
	}

	// Create reseller
	query := `
		INSERT INTO resellers (
			user_id, parent_id, credits, commission_rate,
			can_create_resellers, max_users, max_resellers, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	var resellerID int64
	var createdAt, updatedAt time.Time
	err = s.db.QueryRowContext(ctx, query,
		req.UserID, req.ParentID, req.Credits, req.CommissionRate,
		req.CanCreateResellers, req.MaxUsers, req.MaxResellers, req.Notes,
	).Scan(&resellerID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	// Log credit transaction if credits > 0
	if req.Credits > 0 {
		_, err = s.logCreditTransaction(ctx, resellerID, req.Credits, 0, req.Credits, "add", "Initial credits", &req.CreatedBy)
		if err != nil {
			s.logger.Error("failed to log initial credit transaction", zap.Error(err))
		}
	}

	return s.GetResellerByID(ctx, resellerID)
}

// UpdateReseller updates a reseller
func (s *ResellerService) UpdateReseller(ctx context.Context, resellerID int64, req *UpdateResellerRequest) error {
	query := `
		UPDATE resellers
		SET commission_rate = $1,
		    can_create_resellers = $2,
		    max_users = $3,
		    max_resellers = $4,
		    notes = $5,
		    updated_at = NOW()
		WHERE id = $6
	`

	result, err := s.db.ExecContext(ctx, query,
		req.CommissionRate, req.CanCreateResellers,
		req.MaxUsers, req.MaxResellers, req.Notes,
		resellerID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrResellerNotFound
	}

	return nil
}

// DeleteReseller soft deletes a reseller
func (s *ResellerService) DeleteReseller(ctx context.Context, resellerID int64) error {
	query := `UPDATE resellers SET is_active = FALSE, updated_at = NOW() WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, resellerID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrResellerNotFound
	}

	return nil
}

// AddCredits adds credits to a reseller
func (s *ResellerService) AddCredits(ctx context.Context, resellerID int64, amount float64, description string, createdBy int64) (float64, error) {
	// Get current balance
	var currentBalance float64
	err := s.db.QueryRowContext(ctx, "SELECT credits FROM resellers WHERE id = $1", resellerID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return 0, ErrResellerNotFound
	}
	if err != nil {
		return 0, err
	}

	newBalance := currentBalance + amount

	// Update balance
	_, err = s.db.ExecContext(ctx, "UPDATE resellers SET credits = $1, updated_at = NOW() WHERE id = $2", newBalance, resellerID)
	if err != nil {
		return 0, err
	}

	// Log transaction
	_, err = s.logCreditTransaction(ctx, resellerID, amount, currentBalance, newBalance, "add", description, &createdBy)
	if err != nil {
		s.logger.Error("failed to log credit transaction", zap.Error(err))
	}

	return newBalance, nil
}

// DeductCredits deducts credits from a reseller
func (s *ResellerService) DeductCredits(ctx context.Context, resellerID int64, amount float64, description string, createdBy int64) (float64, error) {
	// Get current balance
	var currentBalance float64
	err := s.db.QueryRowContext(ctx, "SELECT credits FROM resellers WHERE id = $1", resellerID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return 0, ErrResellerNotFound
	}
	if err != nil {
		return 0, err
	}

	if currentBalance < amount {
		return 0, ErrInsufficientCredits
	}

	newBalance := currentBalance - amount

	// Update balance
	_, err = s.db.ExecContext(ctx, "UPDATE resellers SET credits = $1, updated_at = NOW() WHERE id = $2", newBalance, resellerID)
	if err != nil {
		return 0, err
	}

	// Log transaction
	_, err = s.logCreditTransaction(ctx, resellerID, -amount, currentBalance, newBalance, "deduct", description, &createdBy)
	if err != nil {
		s.logger.Error("failed to log credit transaction", zap.Error(err))
	}

	return newBalance, nil
}

// logCreditTransaction logs a credit transaction
func (s *ResellerService) logCreditTransaction(ctx context.Context, resellerID int64, amount, balanceBefore, balanceAfter float64, txType, description string, createdBy *int64) (int64, error) {
	query := `
		INSERT INTO reseller_credits_log (
			reseller_id, amount, balance_before, balance_after,
			transaction_type, description, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var logID int64
	err := s.db.QueryRowContext(ctx, query,
		resellerID, amount, balanceBefore, balanceAfter,
		txType, description, createdBy,
	).Scan(&logID)

	return logID, err
}

// GetCreditHistory returns credit transaction history
func (s *ResellerService) GetCreditHistory(ctx context.Context, resellerID int64, page, limit int) ([]CreditTransaction, int, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reseller_credits_log WHERE reseller_id = $1", resellerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get transactions
	query := `
		SELECT
			rcl.id, rcl.reseller_id, rcl.amount, rcl.balance_before, rcl.balance_after,
			rcl.transaction_type, rcl.description, rcl.created_by, u.username as created_by_name,
			rcl.created_at
		FROM reseller_credits_log rcl
		LEFT JOIN users u ON rcl.created_by = u.id
		WHERE rcl.reseller_id = $1
		ORDER BY rcl.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, resellerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	transactions := []CreditTransaction{}
	for rows.Next() {
		var tx CreditTransaction
		err := rows.Scan(
			&tx.ID, &tx.ResellerID, &tx.Amount, &tx.BalanceBefore, &tx.BalanceAfter,
			&tx.TransactionType, &tx.Description, &tx.CreatedBy, &tx.CreatedByName,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// GetResellerCustomers returns customers assigned to a reseller
func (s *ResellerService) GetResellerCustomers(ctx context.Context, resellerID int64, page, limit int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reseller_assignments WHERE reseller_id = $1", resellerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get customers
	query := `
		SELECT
			u.id, u.username, u.email, u.is_active, u.expires_at,
			ra.assigned_at, ra.expires_at as assignment_expires
		FROM reseller_assignments ra
		INNER JOIN users u ON ra.user_id = u.id
		WHERE ra.reseller_id = $1
		ORDER BY ra.assigned_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, resellerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	customers := []map[string]interface{}{}
	for rows.Next() {
		var userID int64
		var username, email string
		var isActive bool
		var userExpires, assignedAt, assignmentExpires sql.NullTime

		err := rows.Scan(&userID, &username, &email, &isActive, &userExpires, &assignedAt, &assignmentExpires)
		if err != nil {
			return nil, 0, err
		}

		customer := map[string]interface{}{
			"id":                 userID,
			"username":           username,
			"email":              email,
			"is_active":          isActive,
			"assigned_at":        assignedAt.Time,
		}

		if userExpires.Valid {
			customer["expires_at"] = userExpires.Time
		}
		if assignmentExpires.Valid {
			customer["assignment_expires"] = assignmentExpires.Time
		}

		customers = append(customers, customer)
	}

	return customers, total, nil
}

// GetResellerStats returns statistics for a reseller
func (s *ResellerService) GetResellerStats(ctx context.Context, resellerID int64) (map[string]interface{}, error) {
	query := `
		SELECT
			r.credits,
			COALESCE((SELECT COUNT(*) FROM reseller_assignments WHERE reseller_id = r.id), 0) as total_customers,
			COALESCE((SELECT COUNT(*) FROM resellers WHERE parent_id = r.id), 0) as total_sub_resellers,
			COALESCE((SELECT SUM(amount) FROM reseller_credits_log WHERE reseller_id = r.id AND transaction_type = 'add'), 0) as total_credits_added,
			COALESCE((SELECT SUM(ABS(amount)) FROM reseller_credits_log WHERE reseller_id = r.id AND transaction_type = 'deduct'), 0) as total_credits_used,
			COALESCE((SELECT SUM(amount) FROM reseller_credits_log WHERE reseller_id = r.id AND transaction_type = 'commission'), 0) as total_commission_earned
		FROM resellers r
		WHERE r.id = $1
	`

	var credits, totalCreditsAdded, totalCreditsUsed, totalCommission float64
	var totalCustomers, totalSubResellers int

	err := s.db.QueryRowContext(ctx, query, resellerID).Scan(
		&credits, &totalCustomers, &totalSubResellers,
		&totalCreditsAdded, &totalCreditsUsed, &totalCommission,
	)

	if err == sql.ErrNoRows {
		return nil, ErrResellerNotFound
	}
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"current_credits":         credits,
		"total_customers":         totalCustomers,
		"total_sub_resellers":     totalSubResellers,
		"total_credits_added":     totalCreditsAdded,
		"total_credits_used":      totalCreditsUsed,
		"total_commission_earned": totalCommission,
	}

	return stats, nil
}

// AssignCustomer assigns a user to a reseller
func (s *ResellerService) AssignCustomer(ctx context.Context, resellerID, userID int64, expiresAt *time.Time) error {
	// Check reseller exists and get current user count
	var currentUsers, maxUsers int
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE((SELECT COUNT(*) FROM reseller_assignments WHERE reseller_id = $1), 0),
			max_users
		FROM resellers WHERE id = $1
	`, resellerID).Scan(&currentUsers, &maxUsers)

	if err == sql.ErrNoRows {
		return ErrResellerNotFound
	}
	if err != nil {
		return err
	}

	if currentUsers >= maxUsers {
		return ErrResellerMaxUsersReached
	}

	// Check if user is already assigned
	var exists bool
	err = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reseller_assignments WHERE user_id = $1)", userID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyAssigned
	}

	// Assign customer
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO reseller_assignments (reseller_id, user_id, expires_at) VALUES ($1, $2, $3)",
		resellerID, userID, expiresAt,
	)

	return err
}

// UnassignCustomer removes a user from a reseller
func (s *ResellerService) UnassignCustomer(ctx context.Context, resellerID, userID int64) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM reseller_assignments WHERE reseller_id = $1 AND user_id = $2",
		resellerID, userID,
	)
	return err
}
