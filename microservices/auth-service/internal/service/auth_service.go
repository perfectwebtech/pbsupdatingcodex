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
