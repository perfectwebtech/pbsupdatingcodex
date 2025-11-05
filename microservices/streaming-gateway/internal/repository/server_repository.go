package repository

import (
	"context"
	"database/sql"

	"github.com/iptv-platform/streaming-gateway/internal/model"
)

type ServerRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Server, error)
	GetActiveServersByStream(ctx context.Context, streamID int64) ([]*model.Server, error)
	UpdateActiveConnections(ctx context.Context, serverID int64, delta int) error
}

type serverRepository struct {
	db *sql.DB
}

func NewServerRepository(db *sql.DB) ServerRepository {
	return &serverRepository{db: db}
}

func (r *serverRepository) GetByID(ctx context.Context, id int64) (*model.Server, error) {
	query := `
		SELECT id, name, domain, port, https, is_active, weight,
		       max_connections, active_connections, region,
		       created_at, updated_at
		FROM servers
		WHERE id = $1
	`

	server := &model.Server{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&server.ID,
		&server.Name,
		&server.Domain,
		&server.Port,
		&server.HTTPS,
		&server.IsActive,
		&server.Weight,
		&server.MaxConnections,
		&server.ActiveConnections,
		&server.Region,
		&server.CreatedAt,
		&server.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return server, nil
}

func (r *serverRepository) GetActiveServersByStream(ctx context.Context, streamID int64) ([]*model.Server, error) {
	query := `
		SELECT DISTINCT s.id, s.name, s.domain, s.port, s.https, s.is_active,
		       s.weight, s.max_connections, s.active_connections, s.region,
		       s.created_at, s.updated_at
		FROM servers s
		INNER JOIN stream_servers ss ON s.id = ss.server_id
		WHERE ss.stream_id = $1
		  AND s.is_active = true
		  AND s.active_connections < s.max_connections
		ORDER BY s.active_connections ASC
	`

	rows, err := r.db.QueryContext(ctx, query, streamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*model.Server
	for rows.Next() {
		server := &model.Server{}
		err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Domain,
			&server.Port,
			&server.HTTPS,
			&server.IsActive,
			&server.Weight,
			&server.MaxConnections,
			&server.ActiveConnections,
			&server.Region,
			&server.CreatedAt,
			&server.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}

	return servers, nil
}

func (r *serverRepository) UpdateActiveConnections(ctx context.Context, serverID int64, delta int) error {
	query := `
		UPDATE servers
		SET active_connections = active_connections + $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, delta, serverID)
	return err
}
