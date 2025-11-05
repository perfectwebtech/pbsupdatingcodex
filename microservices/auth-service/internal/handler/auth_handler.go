package handler

import (
	"context"

	"github.com/iptv-platform/auth-service/internal/service"
	pb "github.com/iptv-platform/auth-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Extract metadata for IP and User-Agent
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	ipAddress := getMetadataValue(md, "x-forwarded-for")
	userAgent := getMetadataValue(md, "user-agent")

	loginReq := &service.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		DeviceID:  req.DeviceId,
	}

	resp, err := h.authService.Login(ctx, loginReq)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		switch err {
		case service.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case service.ErrUserExpired:
			return nil, status.Error(codes.PermissionDenied, "subscription expired")
		case service.ErrUserDisabled:
			return nil, status.Error(codes.PermissionDenied, "account disabled")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &pb.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
		ExpiresIn:    resp.ExpiresIn,
		User: &pb.User{
			Id:             resp.User.ID,
			Username:       resp.User.Username,
			Email:          resp.User.Email,
			MaxConnections: int32(resp.User.MaxConnections),
			IsTrial:        resp.User.IsTrial,
			IsActive:       resp.User.IsActive,
		},
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	user, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		User: &pb.User{
			Id:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			MaxConnections: int32(user.MaxConnections),
			IsTrial:        user.IsTrial,
			IsActive:       user.IsActive,
		},
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	resp, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrInvalidToken:
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		case service.ErrTokenExpired:
			return nil, status.Error(codes.Unauthenticated, "refresh token expired")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &pb.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
		ExpiresIn:    resp.ExpiresIn,
		User: &pb.User{
			Id:             resp.User.ID,
			Username:       resp.User.Username,
			Email:          resp.User.Email,
			MaxConnections: int32(resp.User.MaxConnections),
			IsTrial:        resp.User.IsTrial,
			IsActive:       resp.User.IsActive,
		},
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := h.authService.Logout(ctx, req.RefreshToken); err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &pb.LogoutResponse{
		Success: true,
		Message: "logged out successfully",
	}, nil
}

func getMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
