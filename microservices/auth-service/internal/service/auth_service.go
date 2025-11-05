package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iptv-platform/auth-service/internal/config"
	"github.com/iptv-platform/auth-service/internal/model"
	"github.com/iptv-platform/auth-service/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExpired        = errors.New("user subscription expired")
	ErrUserDisabled       = errors.New("user account disabled")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUserExists         = errors.New("user already exists")
	ErrWeakPassword       = errors.New("password is too weak")
)

type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtConfig   config.JWTConfig
	logger      *zap.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	jwtConfig config.JWTConfig,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtConfig:   jwtConfig,
		logger:      logger,
	}
}

type LoginRequest struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
	DeviceID  string
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
	User         *model.User
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user from database
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("failed to get user", zap.String("username", req.Username), zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("invalid password attempt", zap.String("username", req.Username))
		return nil, ErrInvalidCredentials
	}

	// Check user status
	if !user.IsActive {
		return nil, ErrUserDisabled
	}
	if user.ExpiresAt != nil && user.ExpiresAt.Before(time.Now()) {
		return nil, ErrUserExpired
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	// Create session
	session := &model.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		DeviceID:     req.DeviceID,
		ExpiresAt:    time.Now().Add(s.jwtConfig.RefreshTokenTTL),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session", zap.Error(err))
		return nil, err
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Warn("failed to update last login", zap.Int64("user_id", user.ID), zap.Error(err))
	}

	s.logger.Info("user logged in successfully",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("ip", req.IPAddress),
	)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtConfig.AccessTokenTTL.Seconds()),
		User:         user,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*model.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, int64(userID))
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Get session from Redis
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check user status
	if !user.IsActive {
		return nil, ErrUserDisabled
	}
	if user.ExpiresAt != nil && user.ExpiresAt.Before(time.Now()) {
		return nil, ErrUserExpired
	}

	// Generate new tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Update session
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(s.jwtConfig.RefreshTokenTTL)
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtConfig.AccessTokenTTL.Seconds()),
		User:         user,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.sessionRepo.DeleteByRefreshToken(ctx, refreshToken)
}

func (s *AuthService) generateAccessToken(user *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"iat":      now.Unix(),
		"exp":      now.Add(s.jwtConfig.AccessTokenTTL).Unix(),
		"iss":      s.jwtConfig.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

func (s *AuthService) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Register request and response types
type RegisterRequest struct {
	Username       string
	Password       string
	Email          string
	PackageID      int64
	MaxConnections int
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// Validate password strength
	if len(req.Password) < 8 {
		return nil, ErrWeakPassword
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	// Create user
	user := &model.User{
		Username:       req.Username,
		Password:       string(hashedPassword),
		Email:          req.Email,
		PackageID:      &req.PackageID,
		MaxConnections: req.MaxConnections,
		IsTrial:        false,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("user registered successfully",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// UpdateProfile request type
type UpdateProfileRequest struct {
	UserID int64
	Email  string
}

// UpdateProfile updates user profile information
func (s *AuthService) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Update fields
	if req.Email != "" {
		user.Email = req.Email
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update profile", zap.Error(err))
		return err
	}

	s.logger.Info("profile updated successfully", zap.Int64("user_id", user.ID))
	return nil
}

// ChangePassword request type
type ChangePasswordRequest struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		s.logger.Warn("invalid old password attempt", zap.Int64("user_id", user.ID))
		return ErrInvalidCredentials
	}

	// Validate new password strength
	if len(req.NewPassword) < 8 {
		return ErrWeakPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update password", zap.Error(err))
		return err
	}

	// Invalidate all sessions for this user
	if err := s.sessionRepo.DeleteAllByUserID(ctx, user.ID); err != nil {
		s.logger.Warn("failed to delete sessions after password change", zap.Int64("user_id", user.ID), zap.Error(err))
	}

	s.logger.Info("password changed successfully", zap.Int64("user_id", user.ID))
	return nil
}

// InitiatePasswordReset creates a password reset token
func (s *AuthService) InitiatePasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists
		return "", ErrUserNotFound
	}

	// Generate reset token
	resetToken, err := s.generateRefreshToken()
	if err != nil {
		return "", err
	}

	// Store reset token in Redis with 1 hour expiration
	resetSession := &model.Session{
		UserID:       user.ID,
		RefreshToken: "reset_" + resetToken, // Prefix to identify reset tokens
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, resetSession); err != nil {
		s.logger.Error("failed to create reset token session", zap.Error(err))
		return "", err
	}

	s.logger.Info("password reset initiated", zap.String("email", email))
	return resetToken, nil
}

// ResetPassword request type
type ResetPasswordRequest struct {
	Token       string
	NewPassword string
}

// ResetPassword resets user password with token
func (s *AuthService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	// Validate new password strength
	if len(req.NewPassword) < 8 {
		return ErrWeakPassword
	}

	// Get session with reset token
	session, err := s.sessionRepo.GetByRefreshToken(ctx, "reset_"+req.Token)
	if err != nil {
		return ErrInvalidToken
	}

	if session.ExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update password", zap.Error(err))
		return err
	}

	// Delete reset token
	if err := s.sessionRepo.DeleteByRefreshToken(ctx, "reset_"+req.Token); err != nil {
		s.logger.Warn("failed to delete reset token", zap.Error(err))
	}

	// Invalidate all sessions for this user
	if err := s.sessionRepo.DeleteAllByUserID(ctx, user.ID); err != nil {
		s.logger.Warn("failed to delete sessions after password reset", zap.Int64("user_id", user.ID), zap.Error(err))
	}

	s.logger.Info("password reset successfully", zap.Int64("user_id", user.ID))
	return nil
}
