package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iptv-platform/streaming-gateway/internal/config"
	"github.com/iptv-platform/streaming-gateway/internal/model"
	"github.com/iptv-platform/streaming-gateway/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrStreamNotFound        = errors.New("stream not found")
	ErrAccessDenied          = errors.New("access denied")
	ErrMaxConnectionsReached = errors.New("maximum connections reached")
	ErrGeoBlocked            = errors.New("geographic location blocked")
	ErrISPBlocked            = errors.New("ISP blocked")
	ErrServerUnavailable     = errors.New("no server available")
)

type StreamService struct {
	streamRepo   repository.StreamRepository
	serverRepo   repository.ServerRepository
	sessionRepo  repository.SessionRepository
	userRepo     repository.UserRepository
	loadBalancer *LoadBalancerService
	geoIP        *GeoIPService
	config       config.StreamingConfig
	logger       *zap.Logger
}

func NewStreamService(
	streamRepo repository.StreamRepository,
	serverRepo repository.ServerRepository,
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
	loadBalancer *LoadBalancerService,
	geoIP *GeoIPService,
	config config.StreamingConfig,
	logger *zap.Logger,
) *StreamService {
	return &StreamService{
		streamRepo:   streamRepo,
		serverRepo:   serverRepo,
		sessionRepo:  sessionRepo,
		userRepo:     userRepo,
		loadBalancer: loadBalancer,
		geoIP:        geoIP,
		config:       config,
		logger:       logger,
	}
}

type StreamRequest struct {
	UserID    int64
	StreamID  int64
	IPAddress string
	UserAgent string
	Container string
}

type StreamResponse struct {
	Type      string            `json:"type"`
	URL       string            `json:"url"`
	SessionID string            `json:"session_id"`
	Container string            `json:"container"`
	Server    *model.Server     `json:"server,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (s *StreamService) GetStreamURL(ctx context.Context, req *StreamRequest) (*StreamResponse, error) {
	// Get stream
	stream, err := s.streamRepo.GetByID(ctx, req.StreamID)
	if err != nil {
		return nil, ErrStreamNotFound
	}

	if !stream.IsActive {
		return nil, ErrStreamNotFound
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, ErrAccessDenied
	}

	// Check user access to stream
	hasAccess, err := s.streamRepo.UserHasAccess(ctx, user.ID, stream.ID)
	if err != nil || !hasAccess {
		s.logger.Warn("user does not have access to stream",
			zap.Int64("user_id", user.ID),
			zap.Int64("stream_id", stream.ID),
		)
		return nil, ErrAccessDenied
	}

	// Check connection limits
	activeConnections, err := s.sessionRepo.GetActiveConnectionCount(ctx, user.ID)
	if err != nil {
		s.logger.Error("failed to get active connections", zap.Error(err))
	}

	if activeConnections >= user.MaxConnections {
		return nil, ErrMaxConnectionsReached
	}

	// Check geographic restrictions
	if len(stream.AllowedCountries) > 0 {
		country, err := s.geoIP.GetCountry(req.IPAddress)
		if err != nil {
			s.logger.Warn("failed to get country for IP",
				zap.String("ip", req.IPAddress),
				zap.Error(err),
			)
		} else {
			allowed := false
			for _, allowedCountry := range stream.AllowedCountries {
				if country == allowedCountry {
					allowed = true
					break
				}
			}
			if !allowed {
				return nil, ErrGeoBlocked
			}
		}
	}

	// Select best server using load balancer
	server, err := s.loadBalancer.SelectBestServer(ctx, stream.ID, req.IPAddress)
	if err != nil {
		s.logger.Error("failed to select server", zap.Error(err))
		return nil, ErrServerUnavailable
	}

	// Create session
	session := &model.StreamSession{
		UserID:      user.ID,
		StreamID:    stream.ID,
		ServerID:    server.ID,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		Container:   req.Container,
		StartedAt:   time.Now(),
		LastActivity: time.Now(),
	}

	if err := s.sessionRepo.CreateStreamSession(ctx, session); err != nil {
		s.logger.Error("failed to create session", zap.Error(err))
		return nil, err
	}

	// Build stream URL
	streamURL := s.buildStreamURL(stream, server, session, req.Container)

	s.logger.Info("stream URL generated",
		zap.Int64("user_id", user.ID),
		zap.Int64("stream_id", stream.ID),
		zap.Int64("server_id", server.ID),
		zap.String("session_id", session.ID),
	)

	return &StreamResponse{
		Type:      stream.Type,
		URL:       streamURL,
		SessionID: session.ID,
		Container: req.Container,
		Server:    server,
		Metadata: map[string]string{
			"stream_name": stream.Name,
			"bitrate":     fmt.Sprintf("%d", stream.Bitrate),
		},
	}, nil
}

func (s *StreamService) buildStreamURL(
	stream *model.Stream,
	server *model.Server,
	session *model.StreamSession,
	container string,
) string {
	// If CDN is enabled, use CDN URL
	if s.config.CDNEnabled && s.config.CDNBaseURL != "" {
		return fmt.Sprintf("%s/stream/%d/%s.%s?session=%s",
			s.config.CDNBaseURL,
			stream.ID,
			stream.ID,
			container,
			session.ID,
		)
	}

	// Use direct server URL
	protocol := "http"
	if server.HTTPS {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s:%d/stream/%d/%s.%s?session=%s",
		protocol,
		server.Domain,
		server.Port,
		stream.ID,
		stream.ID,
		container,
		session.ID,
	)
}

func (s *StreamService) ListStreams(ctx context.Context, userID int64, filters map[string]interface{}) ([]*model.Stream, error) {
	return s.streamRepo.ListByUser(ctx, userID, filters)
}

func (s *StreamService) GetStream(ctx context.Context, userID, streamID int64) (*model.Stream, error) {
	// Check access
	hasAccess, err := s.streamRepo.UserHasAccess(ctx, userID, streamID)
	if err != nil || !hasAccess {
		return nil, ErrAccessDenied
	}

	return s.streamRepo.GetByID(ctx, streamID)
}

func (s *StreamService) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	return s.sessionRepo.UpdateStreamSessionActivity(ctx, sessionID)
}

func (s *StreamService) CloseSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.CloseStreamSession(ctx, sessionID)
}
