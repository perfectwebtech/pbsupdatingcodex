package service

import (
	"context"
	"errors"
	"math/rand"
	"sort"

	"github.com/iptv-platform/streaming-gateway/internal/model"
	"github.com/iptv-platform/streaming-gateway/internal/repository"
	"go.uber.org/zap"
)

type LoadBalancerStrategy string

const (
	RoundRobin       LoadBalancerStrategy = "round_robin"
	LeastConnections LoadBalancerStrategy = "least_connections"
	Weighted         LoadBalancerStrategy = "weighted"
	Geographic       LoadBalancerStrategy = "geographic"
)

type LoadBalancerService struct {
	serverRepo repository.ServerRepository
	logger     *zap.Logger
	strategy   LoadBalancerStrategy
	counter    int
}

func NewLoadBalancerService(serverRepo repository.ServerRepository, logger *zap.Logger) *LoadBalancerService {
	return &LoadBalancerService{
		serverRepo: serverRepo,
		logger:     logger,
		strategy:   LeastConnections,
		counter:    0,
	}
}

func (s *LoadBalancerService) SelectBestServer(ctx context.Context, streamID int64, clientIP string) (*model.Server, error) {
	// Get all active servers for this stream
	servers, err := s.serverRepo.GetActiveServersByStream(ctx, streamID)
	if err != nil {
		return nil, err
	}

	if len(servers) == 0 {
		return nil, errors.New("no servers available")
	}

	// Apply strategy
	switch s.strategy {
	case RoundRobin:
		return s.roundRobin(servers), nil
	case LeastConnections:
		return s.leastConnections(servers), nil
	case Weighted:
		return s.weighted(servers), nil
	case Geographic:
		return s.geographic(servers, clientIP), nil
	default:
		return s.leastConnections(servers), nil
	}
}

func (s *LoadBalancerService) roundRobin(servers []*model.Server) *model.Server {
	s.counter++
	index := s.counter % len(servers)
	return servers[index]
}

func (s *LoadBalancerService) leastConnections(servers []*model.Server) *model.Server {
	// Sort servers by active connections
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].ActiveConnections < servers[j].ActiveConnections
	})

	return servers[0]
}

func (s *LoadBalancerService) weighted(servers []*model.Server) *model.Server {
	// Calculate total weight
	totalWeight := 0
	for _, server := range servers {
		totalWeight += server.Weight
	}

	if totalWeight == 0 {
		// Fallback to round robin if no weights configured
		return s.roundRobin(servers)
	}

	// Select server based on weight
	randomWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, server := range servers {
		currentWeight += server.Weight
		if randomWeight < currentWeight {
			return server
		}
	}

	return servers[0]
}

func (s *LoadBalancerService) geographic(servers []*model.Server, clientIP string) *model.Server {
	// This would use GeoIP to find closest server
	// For now, fallback to least connections
	// TODO: Implement geographic routing based on latency
	return s.leastConnections(servers)
}

func (s *LoadBalancerService) GetServerHealth(ctx context.Context, serverID int64) (*model.ServerHealth, error) {
	// This would check server health metrics
	// TODO: Implement health check service
	return &model.ServerHealth{
		ServerID:  serverID,
		Healthy:   true,
		CPUUsage:  0.5,
		MemUsage:  0.6,
		BandWidth: 0.7,
	}, nil
}
