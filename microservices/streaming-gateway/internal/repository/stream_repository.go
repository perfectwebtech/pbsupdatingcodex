package repository

import (
	"context"
	"database/sql"

	"github.com/iptv-platform/streaming-gateway/internal/model"
	"github.com/lib/pq"
)

type StreamRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Stream, error)
	ListByUser(ctx context.Context, userID int64, filters map[string]interface{}) ([]*model.Stream, error)
	UserHasAccess(ctx context.Context, userID, streamID int64) (bool, error)
}

type streamRepository struct {
	db *sql.DB
}

func NewStreamRepository(db *sql.DB) StreamRepository {
	return &streamRepository{db: db}
}

func (r *streamRepository) GetByID(ctx context.Context, id int64) (*model.Stream, error) {
	query := `
		SELECT id, name, type, category_id, source_urls, direct_source,
		       icon_url, bitrate, duration, is_active, allowed_countries,
		       created_at, updated_at
		FROM streams
		WHERE id = $1 AND is_active = true
	`

	stream := &model.Stream{}
	var allowedCountries pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&stream.ID,
		&stream.Name,
		&stream.Type,
		&stream.CategoryID,
		pq.Array(&stream.SourceURLs),
		&stream.DirectSource,
		&stream.IconURL,
		&stream.Bitrate,
		&stream.Duration,
		&stream.IsActive,
		&allowedCountries,
		&stream.CreatedAt,
		&stream.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	stream.AllowedCountries = allowedCountries

	return stream, nil
}

func (r *streamRepository) ListByUser(ctx context.Context, userID int64, filters map[string]interface{}) ([]*model.Stream, error) {
	query := `
		SELECT DISTINCT s.id, s.name, s.type, s.category_id, s.icon_url,
		       s.bitrate, s.is_active, s.created_at
		FROM streams s
		INNER JOIN package_streams ps ON s.id = ps.stream_id
		INNER JOIN users u ON u.package_id = ps.package_id
		WHERE u.id = $1 AND s.is_active = true
	`

	args := []interface{}{userID}

	if streamType, ok := filters["type"].(string); ok {
		query += " AND s.type = $2"
		args = append(args, streamType)
	}

	if categoryID, ok := filters["category_id"].(int64); ok {
		query += " AND s.category_id = $" + string(len(args)+1)
		args = append(args, categoryID)
	}

	query += " ORDER BY s.name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streams []*model.Stream
	for rows.Next() {
		stream := &model.Stream{}
		err := rows.Scan(
			&stream.ID,
			&stream.Name,
			&stream.Type,
			&stream.CategoryID,
			&stream.IconURL,
			&stream.Bitrate,
			&stream.IsActive,
			&stream.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		streams = append(streams, stream)
	}

	return streams, nil
}

func (r *streamRepository) UserHasAccess(ctx context.Context, userID, streamID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM streams s
			INNER JOIN package_streams ps ON s.id = ps.stream_id
			INNER JOIN users u ON u.package_id = ps.package_id
			WHERE u.id = $1 AND s.id = $2 AND s.is_active = true
		)
	`

	var hasAccess bool
	err := r.db.QueryRowContext(ctx, query, userID, streamID).Scan(&hasAccess)
	if err != nil {
		return false, err
	}

	return hasAccess, nil
}
