package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// BillingHandler handles HTTP requests for billing and invoicing
type BillingHandler struct {
	billingService *BillingService
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(billingService *BillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
	}
}

// =====================================================
// REQUEST/RESPONSE DTOs
// =====================================================

// Invoice represents an invoice
type Invoice struct {
	ID              int64     `json:"id"`
	InvoiceNumber   string    `json:"invoice_number"`
	UserID          int64     `json:"user_id"`
	Username        string    `json:"username,omitempty"`
	Email           string    `json:"email,omitempty"`
	ResellerID      *int64    `json:"reseller_id,omitempty"`
	PackageID       *int64    `json:"package_id,omitempty"`
	PackageName     string    `json:"package_name,omitempty"`
	Description     string    `json:"description,omitempty"`
	Subtotal        float64   `json:"subtotal"`
	Tax             float64   `json:"tax"`
	Discount        float64   `json:"discount"`
	Total           float64   `json:"total"`
	Status          string    `json:"status"`
	IssueDate       time.Time `json:"issue_date"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	PaymentMethod   string    `json:"payment_method,omitempty"`
	PaymentGateway  string    `json:"payment_gateway,omitempty"`
	TransactionID   string    `json:"transaction_id,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateInvoiceRequest represents a request to create an invoice
type CreateInvoiceRequest struct {
	UserID         int64      `json:"user_id" binding:"required"`
	PackageID      *int64     `json:"package_id"`
	Description    string     `json:"description"`
	Subtotal       float64    `json:"subtotal" binding:"required,gt=0"`
	Tax            float64    `json:"tax" binding:"min=0"`
	Discount       float64    `json:"discount" binding:"min=0"`
	DueDate        *time.Time `json:"due_date"`
	Notes          string     `json:"notes"`
}

// UpdateInvoiceRequest represents a request to update an invoice
type UpdateInvoiceRequest struct {
	Description    string     `json:"description"`
	Subtotal       float64    `json:"subtotal" binding:"gt=0"`
	Tax            float64    `json:"tax" binding:"min=0"`
	Discount       float64    `json:"discount" binding:"min=0"`
	DueDate        *time.Time `json:"due_date"`
	Notes          string     `json:"notes"`
}

// ProcessPaymentRequest represents a payment processing request
type ProcessPaymentRequest struct {
	InvoiceID        int64  `json:"invoice_id" binding:"required"`
	PaymentMethodID  *int64 `json:"payment_method_id"`
	PaymentGateway   string `json:"payment_gateway" binding:"required,oneof=stripe paypal crypto bank_transfer"`
	Amount           float64 `json:"amount" binding:"required,gt=0"`
	PaymentDetails   map[string]interface{} `json:"payment_details"`
}

// PaymentMethod represents a stored payment method
type PaymentMethod struct {
	ID                int64                  `json:"id"`
	UserID            int64                  `json:"user_id"`
	Type              string                 `json:"type"`
	Details           map[string]interface{} `json:"details,omitempty"`
	GatewayCustomerID string                 `json:"gateway_customer_id,omitempty"`
	IsDefault         bool                   `json:"is_default"`
	IsActive          bool                   `json:"is_active"`
	Nickname          string                 `json:"nickname,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// AddPaymentMethodRequest represents a request to add a payment method
type AddPaymentMethodRequest struct {
	UserID            int64                  `json:"user_id" binding:"required"`
	Type              string                 `json:"type" binding:"required,oneof=credit_card paypal stripe crypto bank_transfer"`
	Details           map[string]interface{} `json:"details"`
	GatewayCustomerID string                 `json:"gateway_customer_id"`
	IsDefault         bool                   `json:"is_default"`
	Nickname          string                 `json:"nickname"`
}

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	ID                   int64                  `json:"id"`
	InvoiceID            *int64                 `json:"invoice_id,omitempty"`
	UserID               int64                  `json:"user_id"`
	PaymentMethodID      *int64                 `json:"payment_method_id,omitempty"`
	Amount               float64                `json:"amount"`
	Currency             string                 `json:"currency"`
	Status               string                 `json:"status"`
	Gateway              string                 `json:"gateway"`
	GatewayTransactionID string                 `json:"gateway_transaction_id,omitempty"`
	GatewayResponse      map[string]interface{} `json:"gateway_response,omitempty"`
	IPAddress            string                 `json:"ip_address,omitempty"`
	UserAgent            string                 `json:"user_agent,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// RevenueStats represents revenue statistics
type RevenueStats struct {
	TotalRevenue      float64            `json:"total_revenue"`
	MonthlyRevenue    float64            `json:"monthly_revenue"`
	WeeklyRevenue     float64            `json:"weekly_revenue"`
	DailyRevenue      float64            `json:"daily_revenue"`
	PendingAmount     float64            `json:"pending_amount"`
	OverdueAmount     float64            `json:"overdue_amount"`
	TotalInvoices     int64              `json:"total_invoices"`
	PaidInvoices      int64              `json:"paid_invoices"`
	PendingInvoices   int64              `json:"pending_invoices"`
	OverdueInvoices   int64              `json:"overdue_invoices"`
	RevenueByGateway  map[string]float64 `json:"revenue_by_gateway"`
	RevenueByMonth    []MonthlyRevenue   `json:"revenue_by_month"`
}

// MonthlyRevenue represents revenue for a specific month
type MonthlyRevenue struct {
	Month    string  `json:"month"`
	Year     int     `json:"year"`
	Revenue  float64 `json:"revenue"`
	Invoices int64   `json:"invoices"`
}

// =====================================================
// INVOICE ENDPOINTS
// =====================================================

// ListInvoices lists all invoices with pagination and filters
// GET /api/v1/admin/billing/invoices
func (h *BillingHandler) ListInvoices(c *gin.Context) {
	ctx := context.Background()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	userID := c.Query("user_id")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build filters
	filters := make(map[string]interface{})
	if status != "" {
		filters["status"] = status
	}
	if userID != "" {
		if uid, err := strconv.ParseInt(userID, 10, 64); err == nil {
			filters["user_id"] = uid
		}
	}
	if search != "" {
		filters["search"] = search
	}

	// Get invoices
	invoices, total, err := h.billingService.ListInvoices(ctx, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch invoices",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": invoices,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetInvoice gets an invoice by ID
// GET /api/v1/admin/billing/invoices/:id
func (h *BillingHandler) GetInvoice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	invoice, err := h.billingService.GetInvoiceByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

// CreateInvoice creates a new invoice
// POST /api/v1/admin/billing/invoices
func (h *BillingHandler) CreateInvoice(c *gin.Context) {
	ctx := context.Background()

	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.billingService.CreateInvoice(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create invoice",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invoice created successfully",
		"data": invoice,
	})
}

// UpdateInvoice updates an existing invoice
// PUT /api/v1/admin/billing/invoices/:id
func (h *BillingHandler) UpdateInvoice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	var req UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.billingService.UpdateInvoice(ctx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update invoice",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice updated successfully",
		"data": invoice,
	})
}

// DeleteInvoice deletes an invoice
// DELETE /api/v1/admin/billing/invoices/:id
func (h *BillingHandler) DeleteInvoice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	if err := h.billingService.DeleteInvoice(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete invoice",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice deleted successfully"})
}

// CancelInvoice cancels an invoice
// POST /api/v1/admin/billing/invoices/:id/cancel
func (h *BillingHandler) CancelInvoice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	invoice, err := h.billingService.CancelInvoice(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel invoice",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice cancelled successfully",
		"data": invoice,
	})
}

// =====================================================
// PAYMENT PROCESSING ENDPOINTS
// =====================================================

// ProcessPayment processes a payment for an invoice
// POST /api/v1/admin/billing/payments
func (h *BillingHandler) ProcessPayment(c *gin.Context) {
	ctx := context.Background()

	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get client IP and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	transaction, err := h.billingService.ProcessPayment(ctx, &req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Payment processing failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment processed successfully",
		"data": transaction,
	})
}

// GetPaymentTransactions gets payment transaction history
// GET /api/v1/admin/billing/transactions
func (h *BillingHandler) GetPaymentTransactions(c *gin.Context) {
	ctx := context.Background()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	userID := c.Query("user_id")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	filters := make(map[string]interface{})
	if userID != "" {
		if uid, err := strconv.ParseInt(userID, 10, 64); err == nil {
			filters["user_id"] = uid
		}
	}
	if status != "" {
		filters["status"] = status
	}

	transactions, total, err := h.billingService.GetPaymentTransactions(ctx, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch transactions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transactions,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// RefundPayment refunds a payment
// POST /api/v1/admin/billing/transactions/:id/refund
func (h *BillingHandler) RefundPayment(c *gin.Context) {
	ctx := context.Background()

	transactionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Reason string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.billingService.RefundPayment(ctx, transactionID, req.Amount, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process refund",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Refund processed successfully",
		"data": refund,
	})
}

// =====================================================
// PAYMENT METHOD ENDPOINTS
// =====================================================

// ListPaymentMethods lists all payment methods for a user
// GET /api/v1/admin/billing/payment-methods
func (h *BillingHandler) ListPaymentMethods(c *gin.Context) {
	ctx := context.Background()

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	methods, err := h.billingService.ListPaymentMethods(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch payment methods",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": methods})
}

// AddPaymentMethod adds a new payment method
// POST /api/v1/admin/billing/payment-methods
func (h *BillingHandler) AddPaymentMethod(c *gin.Context) {
	ctx := context.Background()

	var req AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method, err := h.billingService.AddPaymentMethod(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add payment method",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Payment method added successfully",
		"data": method,
	})
}

// DeletePaymentMethod deletes a payment method
// DELETE /api/v1/admin/billing/payment-methods/:id
func (h *BillingHandler) DeletePaymentMethod(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	if err := h.billingService.DeletePaymentMethod(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete payment method",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method deleted successfully"})
}

// SetDefaultPaymentMethod sets a payment method as default
// POST /api/v1/admin/billing/payment-methods/:id/set-default
func (h *BillingHandler) SetDefaultPaymentMethod(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	method, err := h.billingService.SetDefaultPaymentMethod(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to set default payment method",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Default payment method updated successfully",
		"data": method,
	})
}

// =====================================================
// REVENUE & STATISTICS ENDPOINTS
// =====================================================

// GetRevenueStats gets revenue statistics
// GET /api/v1/admin/billing/revenue/stats
func (h *BillingHandler) GetRevenueStats(c *gin.Context) {
	ctx := context.Background()

	stats, err := h.billingService.GetRevenueStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch revenue statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetUserInvoices gets all invoices for a specific user
// GET /api/v1/billing/invoices (for regular users)
func (h *BillingHandler) GetUserInvoices(c *gin.Context) {
	ctx := context.Background()

	// Get user ID from JWT context (assumed to be set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	filters := map[string]interface{}{
		"user_id": userID.(int64),
	}
	if status != "" {
		filters["status"] = status
	}

	invoices, total, err := h.billingService.ListInvoices(ctx, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch invoices",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": invoices,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// RegisterRoutes registers all billing routes
func (h *BillingHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Admin routes
	admin := router.Group("/admin/billing")
	{
		// Invoice management
		admin.GET("/invoices", h.ListInvoices)
		admin.GET("/invoices/:id", h.GetInvoice)
		admin.POST("/invoices", h.CreateInvoice)
		admin.PUT("/invoices/:id", h.UpdateInvoice)
		admin.DELETE("/invoices/:id", h.DeleteInvoice)
		admin.POST("/invoices/:id/cancel", h.CancelInvoice)

		// Payment processing
		admin.POST("/payments", h.ProcessPayment)
		admin.GET("/transactions", h.GetPaymentTransactions)
		admin.POST("/transactions/:id/refund", h.RefundPayment)

		// Payment methods
		admin.GET("/payment-methods", h.ListPaymentMethods)
		admin.POST("/payment-methods", h.AddPaymentMethod)
		admin.DELETE("/payment-methods/:id", h.DeletePaymentMethod)
		admin.POST("/payment-methods/:id/set-default", h.SetDefaultPaymentMethod)

		// Revenue statistics
		admin.GET("/revenue/stats", h.GetRevenueStats)
	}

	// User routes
	user := router.Group("/billing")
	{
		user.GET("/invoices", h.GetUserInvoices)
	}
}
