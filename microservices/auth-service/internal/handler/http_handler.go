package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/iptv-platform/auth-service/internal/service"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

func NewHTTPHandler(authService *service.AuthService, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		authService: authService,
		logger:      logger,
	}
}

// Response structures
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// Request structures
type RegisterRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	PackageID      int64  `json:"package_id,omitempty"`
	MaxConnections int    `json:"max_connections,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DeviceID string `json:"device_id,omitempty"`
}

type UpdateProfileRequest struct {
	Email string `json:"email,omitempty"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register handles user registration
func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		h.sendError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.MaxConnections == 0 {
		req.MaxConnections = 1
	}

	ctx := r.Context()
	registerReq := &service.RegisterRequest{
		Username:       req.Username,
		Password:       req.Password,
		Email:          req.Email,
		PackageID:      req.PackageID,
		MaxConnections: req.MaxConnections,
	}

	user, err := h.authService.Register(ctx, registerReq)
	if err != nil {
		h.logger.Error("registration failed", zap.Error(err))
		switch err {
		case service.ErrUserExists:
			h.sendError(w, "username already exists", http.StatusConflict)
		default:
			h.sendError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	}, "user created successfully", http.StatusCreated)
}

// Login handles user authentication
func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.sendError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	loginReq := &service.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: h.getClientIP(r),
		UserAgent: r.UserAgent(),
		DeviceID:  req.DeviceID,
	}

	resp, err := h.authService.Login(ctx, loginReq)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))
		switch err {
		case service.ErrInvalidCredentials:
			h.sendError(w, "invalid credentials", http.StatusUnauthorized)
		case service.ErrUserExpired:
			h.sendError(w, "subscription expired", http.StatusForbidden)
		case service.ErrUserDisabled:
			h.sendError(w, "account disabled", http.StatusForbidden)
		default:
			h.sendError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
		"user": map[string]interface{}{
			"id":              resp.User.ID,
			"username":        resp.User.Username,
			"email":           resp.User.Email,
			"max_connections": resp.User.MaxConnections,
			"is_trial":        resp.User.IsTrial,
			"is_active":       resp.User.IsActive,
		},
	}, "", http.StatusOK)
}

// ValidateToken validates a JWT token
func (h *HTTPHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := h.extractToken(r)
	if token == "" {
		h.sendError(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := h.authService.ValidateToken(ctx, token)
	if err != nil {
		h.sendError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"valid": true,
		"user": map[string]interface{}{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"max_connections": user.MaxConnections,
			"is_trial":        user.IsTrial,
			"is_active":       user.IsActive,
		},
	}, "", http.StatusOK)
}

// GetCurrentUser returns the authenticated user's information
func (h *HTTPHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := h.extractToken(r)
	if token == "" {
		h.sendError(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := h.authService.ValidateToken(ctx, token)
	if err != nil {
		h.sendError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Get full user details
	fullUser, err := h.authService.GetUserByID(ctx, user.ID)
	if err != nil {
		h.sendError(w, "failed to get user details", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"id":              fullUser.ID,
		"username":        fullUser.Username,
		"email":           fullUser.Email,
		"package_id":      fullUser.PackageID,
		"max_connections": fullUser.MaxConnections,
		"is_trial":        fullUser.IsTrial,
		"is_active":       fullUser.IsActive,
		"expires_at":      fullUser.ExpiresAt,
		"created_at":      fullUser.CreatedAt,
	}, "", http.StatusOK)
}

// UpdateProfile updates user profile information
func (h *HTTPHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := h.extractToken(r)
	if token == "" {
		h.sendError(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := h.authService.ValidateToken(ctx, token)
	if err != nil {
		h.sendError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updateReq := &service.UpdateProfileRequest{
		UserID: user.ID,
		Email:  req.Email,
	}

	if err := h.authService.UpdateProfile(ctx, updateReq); err != nil {
		h.logger.Error("profile update failed", zap.Error(err))
		h.sendError(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, nil, "profile updated successfully", http.StatusOK)
}

// ChangePassword changes user password
func (h *HTTPHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := h.extractToken(r)
	if token == "" {
		h.sendError(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := h.authService.ValidateToken(ctx, token)
	if err != nil {
		h.sendError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		h.sendError(w, "old_password and new_password are required", http.StatusBadRequest)
		return
	}

	changeReq := &service.ChangePasswordRequest{
		UserID:      user.ID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	if err := h.authService.ChangePassword(ctx, changeReq); err != nil {
		h.logger.Error("password change failed", zap.Error(err))
		switch err {
		case service.ErrInvalidCredentials:
			h.sendError(w, "incorrect old password", http.StatusUnauthorized)
		default:
			h.sendError(w, "failed to change password", http.StatusInternalServerError)
		}
		return
	}

	h.sendSuccess(w, nil, "password changed successfully", http.StatusOK)
}

// ForgotPassword initiates password reset process
func (h *HTTPHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		h.sendError(w, "email is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resetToken, err := h.authService.InitiatePasswordReset(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists for security
		h.logger.Error("password reset initiation failed", zap.Error(err))
	}

	// Always return success to prevent user enumeration
	h.sendSuccess(w, map[string]interface{}{
		"message": "If the email exists, a password reset link has been sent",
		"token":   resetToken, // In production, this should be sent via email only
	}, "", http.StatusOK)
}

// ResetPassword resets user password with token
func (h *HTTPHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		h.sendError(w, "token and new_password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resetReq := &service.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}

	if err := h.authService.ResetPassword(ctx, resetReq); err != nil {
		h.logger.Error("password reset failed", zap.Error(err))
		switch err {
		case service.ErrInvalidToken, service.ErrTokenExpired:
			h.sendError(w, "invalid or expired reset token", http.StatusUnauthorized)
		default:
			h.sendError(w, "failed to reset password", http.StatusInternalServerError)
		}
		return
	}

	h.sendSuccess(w, nil, "password reset successfully", http.StatusOK)
}

// RefreshToken refreshes an access token
func (h *HTTPHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		h.sendError(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Error("token refresh failed", zap.Error(err))
		switch err {
		case service.ErrInvalidToken:
			h.sendError(w, "invalid refresh token", http.StatusUnauthorized)
		case service.ErrTokenExpired:
			h.sendError(w, "refresh token expired", http.StatusUnauthorized)
		default:
			h.sendError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	}, "", http.StatusOK)
}

// Logout logs out a user
func (h *HTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.authService.Logout(ctx, req.RefreshToken); err != nil {
		h.sendError(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, nil, "logged out successfully", http.StatusOK)
}

// Helper methods
func (h *HTTPHandler) sendSuccess(w http.ResponseWriter, data interface{}, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *HTTPHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Error:   message,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *HTTPHandler) extractToken(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check query parameter
	return r.URL.Query().Get("token")
}

func (h *HTTPHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// CORS middleware
func (h *HTTPHandler) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Rate limiting middleware (simple implementation)
type RateLimiter struct {
	requests map[string][]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Allow(ip string, limit int, window time.Duration) bool {
	now := time.Now()
	cutoff := now.Add(-window)

	// Clean old requests
	if timestamps, exists := rl.requests[ip]; exists {
		var recent []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		rl.requests[ip] = recent

		if len(recent) >= limit {
			return false
		}
	}

	// Add new request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}
