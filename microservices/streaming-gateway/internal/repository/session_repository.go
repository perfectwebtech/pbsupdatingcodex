package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iptv-platform/streaming-gateway/internal/model"
	"github.com/redis/go-redis/v9"
)

type SessionRepository interface {
	CreateStreamSession(ctx context.Context, session *model.StreamSession) error
	GetStreamSession(ctx context.Context, sessionID string) (*model.StreamSession, error)
	UpdateStreamSessionActivity(ctx context.Context, sessionID string) error
	CloseStreamSession(ctx context.Context, sessionID string) error
	GetActiveConnectionCount(ctx context.Context, userID int64) (int, error)
}

type sessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(client *redis.Client) SessionRepository {
	return &sessionRepository{client: client}
}

func (r *sessionRepository) CreateStreamSession(ctx context.Context, session *model.StreamSession) error {
	// Generate session ID
	session.ID = uuid.New().String()

	// Store session in Redis
	key := fmt.Sprintf("stream_session:%s", session.ID)

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// Sessions expire after 2 hours of inactivity
	ttl := 2 * time.Hour

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return err
	}

	// Add to user's active sessions set
	userKey := fmt.Sprintf("user_sessions:%d", session.UserID)
	return r.client.SAdd(ctx, userKey, session.ID).Err()
}

func (r *sessionRepository) GetStreamSession(ctx context.Context, sessionID string) (*model.StreamSession, error) {
	key := fmt.Sprintf("stream_session:%s", sessionID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, err
	}

	var session model.StreamSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) UpdateStreamSessionActivity(ctx context.Context, sessionID string) error {
	session, err := r.GetStreamSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()

	key := fmt.Sprintf("stream_session:%s", sessionID)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, 2*time.Hour).Err()
}

func (r *sessionRepository) CloseStreamSession(ctx context.Context, sessionID string) error {
	session, err := r.GetStreamSession(ctx, sessionID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.EndedAt = &now

	// Remove from user's active sessions
	userKey := fmt.Sprintf("user_sessions:%d", session.UserID)
	r.client.SRem(ctx, userKey, sessionID)

	// Delete session
	key := fmt.Sprintf("stream_session:%s", sessionID)
	return r.client.Del(ctx, key).Err()
}

func (r *sessionRepository) GetActiveConnectionCount(ctx context.Context, userID int64) (int, error) {
	userKey := fmt.Sprintf("user_sessions:%d", userID)

	count, err := r.client.SCard(ctx, userKey).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return int(count), nil
}
