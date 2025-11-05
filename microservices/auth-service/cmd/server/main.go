package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iptv-platform/auth-service/internal/config"
	"github.com/iptv-platform/auth-service/internal/database"
	"github.com/iptv-platform/auth-service/internal/handler"
	"github.com/iptv-platform/auth-service/internal/repository"
	"github.com/iptv-platform/auth-service/internal/service"
	pb "github.com/iptv-platform/auth-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(redisClient)

	// Initialize services
	authService := service.NewAuthService(userRepo, sessionRepo, cfg.JWT, logger)

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	)

	// Register service
	authHandler := handler.NewAuthHandler(authService, logger)
	pb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcAddress := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		logger.Fatal("failed to listen", zap.String("address", grpcAddress), zap.Error(err))
	}

	// Start HTTP gateway for REST API
	go startHTTPGateway(cfg.Server.HTTPPort, cfg, authService, logger)

	// Graceful shutdown
	go func() {
		logger.Info("gRPC server started", zap.String("address", grpcAddress))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	grpcServer.GracefulStop()
	logger.Info("server stopped")
}

func loggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			logger.Error("RPC failed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("RPC completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
			)
		}
		return resp, err
	}
}

func startHTTPGateway(port int, cfg *config.Config, authService *service.AuthService, logger *zap.Logger) {
	httpHandler := handler.NewHTTPHandler(authService, logger)
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"auth-service"}`))
	})

	// Metrics endpoint (Prometheus)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement Prometheus metrics
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Auth service metrics\n"))
	})

	// Authentication API endpoints
	mux.HandleFunc("/api/v1/auth/register", httpHandler.EnableCORS(httpHandler.Register))
	mux.HandleFunc("/api/v1/auth/login", httpHandler.EnableCORS(httpHandler.Login))
	mux.HandleFunc("/api/v1/auth/validate", httpHandler.EnableCORS(httpHandler.ValidateToken))
	mux.HandleFunc("/api/v1/auth/refresh", httpHandler.EnableCORS(httpHandler.RefreshToken))
	mux.HandleFunc("/api/v1/auth/logout", httpHandler.EnableCORS(httpHandler.Logout))
	mux.HandleFunc("/api/v1/auth/me", httpHandler.EnableCORS(httpHandler.GetCurrentUser))
	mux.HandleFunc("/api/v1/auth/profile", httpHandler.EnableCORS(httpHandler.UpdateProfile))
	mux.HandleFunc("/api/v1/auth/password", httpHandler.EnableCORS(httpHandler.ChangePassword))
	mux.HandleFunc("/api/v1/auth/forgot", httpHandler.EnableCORS(httpHandler.ForgotPassword))
	mux.HandleFunc("/api/v1/auth/reset", httpHandler.EnableCORS(httpHandler.ResetPassword))

	address := fmt.Sprintf(":%d", port)
	logger.Info("HTTP gateway started", zap.String("address", address))

	if err := http.ListenAndServe(address, mux); err != nil {
		logger.Fatal("HTTP gateway failed", zap.Error(err))
	}
}
