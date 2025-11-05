package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/iptv-platform/auth-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*model.Session, error)
	Update(ctx context.Context, session *model.Session) error
	DeleteByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteAllByUserID(ctx context.Context, userID int64) error
}

type sessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(client *redis.Client) SessionRepository {
	return &sessionRepository{client: client}
}

func (r *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	key := fmt.Sprintf("session:refresh:%s", session.RefreshToken)

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := time.Until(session.ExpiresAt)
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*model.Session, error) {
	key := fmt.Sprintf("session:refresh:%s", refreshToken)

	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, err
	}

	var session model.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *model.Session) error {
	// Delete old key
	oldKey := fmt.Sprintf("session:refresh:%s", session.RefreshToken)
	r.client.Del(ctx, oldKey)

	// Create new key
	return r.Create(ctx, session)
}

func (r *sessionRepository) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	key := fmt.Sprintf("session:refresh:%s", refreshToken)
	return r.client.Del(ctx, key).Err()
}

func (r *sessionRepository) DeleteAllByUserID(ctx context.Context, userID int64) error {
	// This would require maintaining a separate index in Redis
	// For now, we'll just return nil
	// TODO: Implement proper session tracking by user ID
	return nil
}
