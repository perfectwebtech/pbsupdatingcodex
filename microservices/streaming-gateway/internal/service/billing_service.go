package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BillingService provides business logic for billing and invoicing
type BillingService struct {
	db *sql.DB
}

// NewBillingService creates a new billing service
func NewBillingService(db *sql.DB) *BillingService {
	return &BillingService{
		db: db,
	}
}

// =====================================================
// ERRORS
// =====================================================

var (
	ErrInvoiceNotFound        = errors.New("invoice not found")
	ErrInvoiceAlreadyPaid     = errors.New("invoice is already paid")
	ErrInvoiceCannotBeUpdated = errors.New("invoice cannot be updated in current status")
	ErrInvalidInvoiceStatus   = errors.New("invalid invoice status")
	ErrPaymentMethodNotFound  = errors.New("payment method not found")
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrInsufficientAmount     = errors.New("insufficient payment amount")
	ErrRefundNotAllowed       = errors.New("refund not allowed for this transaction")
	ErrPaymentGatewayFailed   = errors.New("payment gateway error")
)

// =====================================================
// INVOICE MANAGEMENT
// =====================================================

// ListInvoices lists all invoices with pagination and filters
func (s *BillingService) ListInvoices(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*Invoice, int64, error) {
	query := `
		SELECT
			i.id, i.invoice_number, i.user_id, i.reseller_id, i.package_id,
			i.description, i.subtotal, i.tax, i.discount, i.total,
			i.status, i.issue_date, i.due_date, i.paid_at,
			i.payment_method, i.payment_gateway, i.transaction_id,
			i.notes, i.created_at, i.updated_at,
			u.username, u.email,
			p.name as package_name
		FROM invoices i
		INNER JOIN users u ON i.user_id = u.id
		LEFT JOIN packages p ON i.package_id = p.id
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM invoices i WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND i.status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	if userID, ok := filters["user_id"].(int64); ok && userID > 0 {
		query += fmt.Sprintf(" AND i.user_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, userID)
		argCount++
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		query += fmt.Sprintf(" AND (i.invoice_number ILIKE $%d OR u.username ILIKE $%d OR u.email ILIKE $%d)", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM users WHERE id = invoices.user_id AND (username ILIKE $%d OR email ILIKE $%d))", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Get total count
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.UserID, &invoice.ResellerID,
			&invoice.PackageID, &invoice.Description, &invoice.Subtotal, &invoice.Tax,
			&invoice.Discount, &invoice.Total, &invoice.Status, &invoice.IssueDate,
			&invoice.DueDate, &invoice.PaidAt, &invoice.PaymentMethod, &invoice.PaymentGateway,
			&invoice.TransactionID, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.Username, &invoice.Email, &invoice.PackageName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, total, nil
}

// GetInvoiceByID gets an invoice by ID
func (s *BillingService) GetInvoiceByID(ctx context.Context, id int64) (*Invoice, error) {
	query := `
		SELECT
			i.id, i.invoice_number, i.user_id, i.reseller_id, i.package_id,
			i.description, i.subtotal, i.tax, i.discount, i.total,
			i.status, i.issue_date, i.due_date, i.paid_at,
			i.payment_method, i.payment_gateway, i.transaction_id,
			i.notes, i.created_at, i.updated_at,
			u.username, u.email,
			COALESCE(p.name, '') as package_name
		FROM invoices i
		INNER JOIN users u ON i.user_id = u.id
		LEFT JOIN packages p ON i.package_id = p.id
		WHERE i.id = $1
	`

	invoice := &Invoice{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.UserID, &invoice.ResellerID,
		&invoice.PackageID, &invoice.Description, &invoice.Subtotal, &invoice.Tax,
		&invoice.Discount, &invoice.Total, &invoice.Status, &invoice.IssueDate,
		&invoice.DueDate, &invoice.PaidAt, &invoice.PaymentMethod, &invoice.PaymentGateway,
		&invoice.TransactionID, &invoice.Notes, &invoice.CreatedAt, &invoice.UpdatedAt,
		&invoice.Username, &invoice.Email, &invoice.PackageName,
	)
	if err == sql.ErrNoRows {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	return invoice, nil
}

// CreateInvoice creates a new invoice
func (s *BillingService) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*Invoice, error) {
	// Calculate total
	total := req.Subtotal + req.Tax - req.Discount
	if total < 0 {
		total = 0
	}

	// Generate invoice number
	invoiceNumber := s.generateInvoiceNumber()

	// Get reseller ID for this user (if any)
	var resellerID *int64
	err := s.db.QueryRowContext(ctx, `
		SELECT reseller_id FROM reseller_assignments WHERE user_id = $1
	`, req.UserID).Scan(&resellerID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check reseller: %w", err)
	}

	// Insert invoice
	query := `
		INSERT INTO invoices (
			invoice_number, user_id, reseller_id, package_id, description,
			subtotal, tax, discount, total, status, issue_date, due_date, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	invoice := &Invoice{
		InvoiceNumber: invoiceNumber,
		UserID:        req.UserID,
		ResellerID:    resellerID,
		PackageID:     req.PackageID,
		Description:   req.Description,
		Subtotal:      req.Subtotal,
		Tax:           req.Tax,
		Discount:      req.Discount,
		Total:         total,
		Status:        "pending",
		IssueDate:     time.Now(),
		DueDate:       req.DueDate,
		Notes:         req.Notes,
	}

	err = s.db.QueryRowContext(ctx, query,
		invoice.InvoiceNumber, invoice.UserID, invoice.ResellerID, invoice.PackageID,
		invoice.Description, invoice.Subtotal, invoice.Tax, invoice.Discount,
		invoice.Total, invoice.Status, invoice.IssueDate, invoice.DueDate, invoice.Notes,
	).Scan(&invoice.ID, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// UpdateInvoice updates an existing invoice
func (s *BillingService) UpdateInvoice(ctx context.Context, id int64, req *UpdateInvoiceRequest) (*Invoice, error) {
	// Check if invoice exists and can be updated
	var status string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM invoices WHERE id = $1", id).Scan(&status)
	if err == sql.ErrNoRows {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	// Cannot update paid, cancelled, or refunded invoices
	if status == "paid" || status == "cancelled" || status == "refunded" {
		return nil, ErrInvoiceCannotBeUpdated
	}

	// Calculate new total
	total := req.Subtotal + req.Tax - req.Discount
	if total < 0 {
		total = 0
	}

	// Update invoice
	query := `
		UPDATE invoices SET
			description = $1, subtotal = $2, tax = $3, discount = $4,
			total = $5, due_date = $6, notes = $7, updated_at = NOW()
		WHERE id = $8
	`

	_, err = s.db.ExecContext(ctx, query,
		req.Description, req.Subtotal, req.Tax, req.Discount,
		total, req.DueDate, req.Notes, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// DeleteInvoice deletes an invoice (only if not paid)
func (s *BillingService) DeleteInvoice(ctx context.Context, id int64) error {
	// Check if invoice can be deleted
	var status string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM invoices WHERE id = $1", id).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrInvoiceNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to fetch invoice: %w", err)
	}

	if status == "paid" {
		return ErrInvoiceAlreadyPaid
	}

	_, err = s.db.ExecContext(ctx, "DELETE FROM invoices WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	return nil
}

// CancelInvoice cancels an invoice
func (s *BillingService) CancelInvoice(ctx context.Context, id int64) (*Invoice, error) {
	// Check if invoice exists and can be cancelled
	var status string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM invoices WHERE id = $1", id).Scan(&status)
	if err == sql.ErrNoRows {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	if status == "paid" || status == "cancelled" {
		return nil, ErrInvoiceCannotBeUpdated
	}

	_, err = s.db.ExecContext(ctx, "UPDATE invoices SET status = 'cancelled', updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel invoice: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// generateInvoiceNumber generates a unique invoice number
func (s *BillingService) generateInvoiceNumber() string {
	now := time.Now()
	return fmt.Sprintf("INV-%d%02d%02d-%s",
		now.Year(), now.Month(), now.Day(),
		uuid.New().String()[:8],
	)
}

// =====================================================
// PAYMENT PROCESSING
// =====================================================

// ProcessPayment processes a payment for an invoice
func (s *BillingService) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest, ipAddress, userAgent string) (*PaymentTransaction, error) {
	// Get invoice
	invoice, err := s.GetInvoiceByID(ctx, req.InvoiceID)
	if err != nil {
		return nil, err
	}

	// Check if already paid
	if invoice.Status == "paid" {
		return nil, ErrInvoiceAlreadyPaid
	}

	// Validate payment amount
	if req.Amount < invoice.Total {
		return nil, ErrInsufficientAmount
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Process payment through gateway (mock implementation)
	gatewayTransactionID, gatewayResponse, err := s.processPaymentGateway(req.PaymentGateway, req.Amount, req.PaymentDetails)
	if err != nil {
		return nil, fmt.Errorf("payment gateway error: %w", err)
	}

	// Create payment transaction
	transactionQuery := `
		INSERT INTO payment_transactions (
			invoice_id, user_id, payment_method_id, amount, currency,
			status, gateway, gateway_transaction_id, gateway_response,
			ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	transaction := &PaymentTransaction{
		InvoiceID:            &req.InvoiceID,
		UserID:               invoice.UserID,
		PaymentMethodID:      req.PaymentMethodID,
		Amount:               req.Amount,
		Currency:             "USD",
		Status:               "completed",
		Gateway:              req.PaymentGateway,
		GatewayTransactionID: gatewayTransactionID,
		GatewayResponse:      gatewayResponse,
		IPAddress:            ipAddress,
		UserAgent:            userAgent,
	}

	err = tx.QueryRowContext(ctx, transactionQuery,
		transaction.InvoiceID, transaction.UserID, transaction.PaymentMethodID,
		transaction.Amount, transaction.Currency, transaction.Status,
		transaction.Gateway, transaction.GatewayTransactionID, transaction.GatewayResponse,
		transaction.IPAddress, transaction.UserAgent,
	).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update invoice status
	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		UPDATE invoices SET
			status = 'paid',
			paid_at = $1,
			payment_method = $2,
			payment_gateway = $3,
			transaction_id = $4,
			updated_at = NOW()
		WHERE id = $5
	`, now, "gateway", req.PaymentGateway, transaction.GatewayTransactionID, req.InvoiceID)

	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// If user has a reseller, add commission
	if invoice.ResellerID != nil {
		err = s.processResellerCommission(ctx, tx, *invoice.ResellerID, invoice.Total)
		if err != nil {
			// Log error but don't fail payment
			fmt.Printf("Failed to process reseller commission: %v\n", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transaction, nil
}

// processPaymentGateway mocks payment gateway processing
func (s *BillingService) processPaymentGateway(gateway string, amount float64, details map[string]interface{}) (string, map[string]interface{}, error) {
	// In production, integrate with real payment gateways:
	// - Stripe: stripe.PaymentIntents.Create()
	// - PayPal: paypal.ExecutePayment()
	// - Crypto: check blockchain confirmations
	// - Bank Transfer: manual verification

	// Mock successful payment
	transactionID := uuid.New().String()
	response := map[string]interface{}{
		"status":         "success",
		"transaction_id": transactionID,
		"timestamp":      time.Now().Unix(),
		"gateway":        gateway,
		"amount":         amount,
	}

	return transactionID, response, nil
}

// processResellerCommission adds commission to reseller credits
func (s *BillingService) processResellerCommission(ctx context.Context, tx *sql.Tx, resellerID int64, amount float64) error {
	// Get reseller commission rate
	var commissionRate float64
	var credits float64
	err := tx.QueryRowContext(ctx, `
		SELECT commission_rate, credits FROM resellers WHERE id = $1
	`, resellerID).Scan(&commissionRate, &credits)
	if err != nil {
		return err
	}

	// Calculate commission
	commission := amount * (commissionRate / 100.0)
	if commission <= 0 {
		return nil
	}

	newBalance := credits + commission

	// Update reseller credits
	_, err = tx.ExecContext(ctx, `
		UPDATE resellers SET credits = $1, updated_at = NOW() WHERE id = $2
	`, newBalance, resellerID)
	if err != nil {
		return err
	}

	// Log transaction
	_, err = tx.ExecContext(ctx, `
		INSERT INTO reseller_credits_log (
			reseller_id, amount, balance_before, balance_after,
			transaction_type, description
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, resellerID, commission, credits, newBalance, "commission", fmt.Sprintf("Commission from invoice payment: %.2f%%", commissionRate))

	return err
}

// GetPaymentTransactions gets payment transaction history
func (s *BillingService) GetPaymentTransactions(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*PaymentTransaction, int64, error) {
	query := `
		SELECT
			id, invoice_id, user_id, payment_method_id, amount, currency,
			status, gateway, gateway_transaction_id, gateway_response,
			ip_address, user_agent, created_at, updated_at
		FROM payment_transactions
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM payment_transactions WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	if userID, ok := filters["user_id"].(int64); ok && userID > 0 {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, userID)
		argCount++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	// Get total count
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*PaymentTransaction
	for rows.Next() {
		t := &PaymentTransaction{}
		err := rows.Scan(
			&t.ID, &t.InvoiceID, &t.UserID, &t.PaymentMethodID, &t.Amount, &t.Currency,
			&t.Status, &t.Gateway, &t.GatewayTransactionID, &t.GatewayResponse,
			&t.IPAddress, &t.UserAgent, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, total, nil
}

// RefundPayment refunds a payment
func (s *BillingService) RefundPayment(ctx context.Context, transactionID int64, amount float64, reason string) (*PaymentTransaction, error) {
	// Get original transaction
	var originalTransaction PaymentTransaction
	err := s.db.QueryRowContext(ctx, `
		SELECT id, invoice_id, user_id, amount, status, gateway
		FROM payment_transactions WHERE id = $1
	`, transactionID).Scan(
		&originalTransaction.ID, &originalTransaction.InvoiceID,
		&originalTransaction.UserID, &originalTransaction.Amount,
		&originalTransaction.Status, &originalTransaction.Gateway,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTransactionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	// Validate refund
	if originalTransaction.Status != "completed" {
		return nil, ErrRefundNotAllowed
	}
	if amount > originalTransaction.Amount {
		return nil, ErrInsufficientAmount
	}

	// Process refund through gateway (mock)
	gatewayTransactionID := uuid.New().String()
	gatewayResponse := map[string]interface{}{
		"status":                 "refunded",
		"refund_transaction_id":  gatewayTransactionID,
		"original_transaction_id": originalTransaction.ID,
		"amount":                 amount,
		"reason":                 reason,
		"timestamp":              time.Now().Unix(),
	}

	// Create refund transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	refundTransaction := &PaymentTransaction{
		InvoiceID:            originalTransaction.InvoiceID,
		UserID:               originalTransaction.UserID,
		Amount:               -amount, // Negative for refund
		Currency:             "USD",
		Status:               "refunded",
		Gateway:              originalTransaction.Gateway,
		GatewayTransactionID: gatewayTransactionID,
		GatewayResponse:      gatewayResponse,
	}

	query := `
		INSERT INTO payment_transactions (
			invoice_id, user_id, amount, currency, status,
			gateway, gateway_transaction_id, gateway_response
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, query,
		refundTransaction.InvoiceID, refundTransaction.UserID, refundTransaction.Amount,
		refundTransaction.Currency, refundTransaction.Status, refundTransaction.Gateway,
		refundTransaction.GatewayTransactionID, refundTransaction.GatewayResponse,
	).Scan(&refundTransaction.ID, &refundTransaction.CreatedAt, &refundTransaction.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	// Update invoice status
	if originalTransaction.InvoiceID != nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE invoices SET status = 'refunded', updated_at = NOW() WHERE id = $1
		`, *originalTransaction.InvoiceID)
		if err != nil {
			return nil, fmt.Errorf("failed to update invoice: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return refundTransaction, nil
}

// =====================================================
// PAYMENT METHODS
// =====================================================

// ListPaymentMethods lists all payment methods for a user
func (s *BillingService) ListPaymentMethods(ctx context.Context, userID int64) ([]*PaymentMethod, error) {
	query := `
		SELECT
			id, user_id, type, details, gateway_customer_id,
			is_default, is_active, nickname, created_at, updated_at
		FROM payment_methods
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment methods: %w", err)
	}
	defer rows.Close()

	var methods []*PaymentMethod
	for rows.Next() {
		m := &PaymentMethod{}
		err := rows.Scan(
			&m.ID, &m.UserID, &m.Type, &m.Details, &m.GatewayCustomerID,
			&m.IsDefault, &m.IsActive, &m.Nickname, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment method: %w", err)
		}
		methods = append(methods, m)
	}

	return methods, nil
}

// AddPaymentMethod adds a new payment method
func (s *BillingService) AddPaymentMethod(ctx context.Context, req *AddPaymentMethodRequest) (*PaymentMethod, error) {
	// If setting as default, unset other defaults
	if req.IsDefault {
		_, err := s.db.ExecContext(ctx, `
			UPDATE payment_methods SET is_default = FALSE WHERE user_id = $1
		`, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to unset default payment methods: %w", err)
		}
	}

	query := `
		INSERT INTO payment_methods (
			user_id, type, details, gateway_customer_id, is_default, nickname
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, is_active, created_at, updated_at
	`

	method := &PaymentMethod{
		UserID:            req.UserID,
		Type:              req.Type,
		Details:           req.Details,
		GatewayCustomerID: req.GatewayCustomerID,
		IsDefault:         req.IsDefault,
		Nickname:          req.Nickname,
	}

	err := s.db.QueryRowContext(ctx, query,
		method.UserID, method.Type, method.Details, method.GatewayCustomerID,
		method.IsDefault, method.Nickname,
	).Scan(&method.ID, &method.IsActive, &method.CreatedAt, &method.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to add payment method: %w", err)
	}

	return method, nil
}

// DeletePaymentMethod deletes a payment method
func (s *BillingService) DeletePaymentMethod(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE payment_methods SET is_active = FALSE WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrPaymentMethodNotFound
	}

	return nil
}

// SetDefaultPaymentMethod sets a payment method as default
func (s *BillingService) SetDefaultPaymentMethod(ctx context.Context, id int64) (*PaymentMethod, error) {
	// Get the payment method to find user_id
	var userID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id FROM payment_methods WHERE id = $1 AND is_active = TRUE
	`, id).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil, ErrPaymentMethodNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment method: %w", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unset all defaults for this user
	_, err = tx.ExecContext(ctx, `
		UPDATE payment_methods SET is_default = FALSE WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to unset defaults: %w", err)
	}

	// Set new default
	_, err = tx.ExecContext(ctx, `
		UPDATE payment_methods SET is_default = TRUE WHERE id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to set default: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated method
	method := &PaymentMethod{}
	err = s.db.QueryRowContext(ctx, `
		SELECT
			id, user_id, type, details, gateway_customer_id,
			is_default, is_active, nickname, created_at, updated_at
		FROM payment_methods WHERE id = $1
	`, id).Scan(
		&method.ID, &method.UserID, &method.Type, &method.Details,
		&method.GatewayCustomerID, &method.IsDefault, &method.IsActive,
		&method.Nickname, &method.CreatedAt, &method.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated method: %w", err)
	}

	return method, nil
}

// =====================================================
// REVENUE STATISTICS
// =====================================================

// GetRevenueStats gets comprehensive revenue statistics
func (s *BillingService) GetRevenueStats(ctx context.Context) (*RevenueStats, error) {
	stats := &RevenueStats{
		RevenueByGateway: make(map[string]float64),
		RevenueByMonth:   []MonthlyRevenue{},
	}

	// Total revenue
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total), 0) FROM invoices WHERE status = 'paid'
	`).Scan(&stats.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}

	// Monthly revenue
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total), 0) FROM invoices
		WHERE status = 'paid' AND paid_at >= date_trunc('month', CURRENT_DATE)
	`).Scan(&stats.MonthlyRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}

	// Weekly revenue
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total), 0) FROM invoices
		WHERE status = 'paid' AND paid_at >= date_trunc('week', CURRENT_DATE)
	`).Scan(&stats.WeeklyRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly revenue: %w", err)
	}

	// Daily revenue
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total), 0) FROM invoices
		WHERE status = 'paid' AND paid_at >= CURRENT_DATE
	`).Scan(&stats.DailyRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily revenue: %w", err)
	}

	// Invoice counts
	err = s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'paid') as paid,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'overdue') as overdue,
			COALESCE(SUM(total) FILTER (WHERE status = 'pending'), 0) as pending_amount,
			COALESCE(SUM(total) FILTER (WHERE status = 'overdue'), 0) as overdue_amount
		FROM invoices
	`).Scan(&stats.TotalInvoices, &stats.PaidInvoices, &stats.PendingInvoices,
		&stats.OverdueInvoices, &stats.PendingAmount, &stats.OverdueAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice counts: %w", err)
	}

	// Revenue by gateway
	rows, err := s.db.QueryContext(ctx, `
		SELECT payment_gateway, SUM(total) as revenue
		FROM invoices
		WHERE status = 'paid' AND payment_gateway IS NOT NULL
		GROUP BY payment_gateway
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue by gateway: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var gateway string
		var revenue float64
		if err := rows.Scan(&gateway, &revenue); err != nil {
			return nil, fmt.Errorf("failed to scan gateway revenue: %w", err)
		}
		stats.RevenueByGateway[gateway] = revenue
	}

	// Revenue by month (last 12 months)
	rows, err = s.db.QueryContext(ctx, `
		SELECT
			TO_CHAR(paid_at, 'Mon') as month,
			EXTRACT(YEAR FROM paid_at) as year,
			SUM(total) as revenue,
			COUNT(*) as invoices
		FROM invoices
		WHERE status = 'paid' AND paid_at >= CURRENT_DATE - INTERVAL '12 months'
		GROUP BY TO_CHAR(paid_at, 'Mon'), EXTRACT(YEAR FROM paid_at), EXTRACT(MONTH FROM paid_at)
		ORDER BY EXTRACT(YEAR FROM paid_at), EXTRACT(MONTH FROM paid_at)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var mr MonthlyRevenue
		if err := rows.Scan(&mr.Month, &mr.Year, &mr.Revenue, &mr.Invoices); err != nil {
			return nil, fmt.Errorf("failed to scan monthly revenue: %w", err)
		}
		stats.RevenueByMonth = append(stats.RevenueByMonth, mr)
	}

	return stats, nil
}
