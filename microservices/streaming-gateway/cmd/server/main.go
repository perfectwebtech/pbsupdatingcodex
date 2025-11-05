package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iptv-platform/streaming-gateway/internal/config"
	"github.com/iptv-platform/streaming-gateway/internal/database"
	"github.com/iptv-platform/streaming-gateway/internal/handler"
	"github.com/iptv-platform/streaming-gateway/internal/middleware"
	"github.com/iptv-platform/streaming-gateway/internal/repository"
	"github.com/iptv-platform/streaming-gateway/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
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

	// Set Gin mode
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
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

	// Initialize GeoIP
	geoIPService, err := service.NewGeoIPService(cfg.GeoIP.DatabasePath, logger)
	if err != nil {
		logger.Fatal("failed to initialize GeoIP", zap.Error(err))
	}
	defer geoIPService.Close()

	// Initialize repositories
	streamRepo := repository.NewStreamRepository(db)
	serverRepo := repository.NewServerRepository(db)
	sessionRepo := repository.NewSessionRepository(redisClient)
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	loadBalancer := service.NewLoadBalancerService(serverRepo, logger)
	streamService := service.NewStreamService(
		streamRepo,
		serverRepo,
		sessionRepo,
		userRepo,
		loadBalancer,
		geoIPService,
		cfg.Streaming,
		logger,
	)
	hlsService := service.NewHLSService(cfg.Streaming, logger)

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Metrics())
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "streaming-gateway",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication middleware
		authMiddleware := middleware.JWTAuth(cfg.JWT.Secret, logger)

		// Stream routes
		streamHandler := handler.NewStreamHandler(streamService, hlsService, logger)
		v1.GET("/streams", authMiddleware, streamHandler.ListStreams)
		v1.GET("/streams/:id", authMiddleware, streamHandler.GetStream)
		v1.GET("/streams/:id/url", authMiddleware, streamHandler.GetStreamURL)

		// HLS streaming routes
		v1.GET("/stream/:id/playlist.m3u8", authMiddleware, streamHandler.GetHLSPlaylist)
		v1.GET("/stream/:id/segment/:segment", authMiddleware, streamHandler.GetHLSSegment)

		// Categories
		v1.GET("/categories", authMiddleware, streamHandler.ListCategories)

		// User info
		v1.GET("/user/info", authMiddleware, streamHandler.GetUserInfo)
		v1.GET("/user/active-connections", authMiddleware, streamHandler.GetActiveConnections)
	}

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("streaming gateway started",
			zap.Int("port", cfg.Server.HTTPPort),
			zap.String("environment", cfg.Server.Environment),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
