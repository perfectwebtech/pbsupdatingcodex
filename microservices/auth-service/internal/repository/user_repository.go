package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/iptv-platform/auth-service/internal/model"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	UpdateLastLogin(ctx context.Context, userID int64) error
	Delete(ctx context.Context, id int64) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, username, password, email, package_id, max_connections,
		       is_trial, is_active, admin_enabled, expires_at,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.PackageID,
		&user.MaxConnections,
		&user.IsTrial,
		&user.IsActive,
		&user.AdminEnabled,
		&user.ExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password, email, package_id, max_connections,
		       is_trial, is_active, admin_enabled, expires_at,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.PackageID,
		&user.MaxConnections,
		&user.IsTrial,
		&user.IsActive,
		&user.AdminEnabled,
		&user.ExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, password, email, package_id, max_connections,
		       is_trial, is_active, admin_enabled, expires_at,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.PackageID,
		&user.MaxConnections,
		&user.IsTrial,
		&user.IsActive,
		&user.AdminEnabled,
		&user.ExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password, email, package_id, max_connections,
		                   is_trial, is_active, admin_enabled, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	now := time.Now()
	return r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
		user.PackageID,
		user.MaxConnections,
		user.IsTrial,
		user.IsActive,
		user.AdminEnabled,
		user.ExpiresAt,
		now,
		now,
	).Scan(&user.ID)
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET username = $1, password = $2, email = $3, package_id = $4, max_connections = $5,
		    is_trial = $6, is_active = $7, admin_enabled = $8, expires_at = $9,
		    updated_at = $10
		WHERE id = $11
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
		user.PackageID,
		user.MaxConnections,
		user.IsTrial,
		user.IsActive,
		user.AdminEnabled,
		user.ExpiresAt,
		time.Now(),
		user.ID,
	)

	return err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

var ErrUserNotFound = sql.ErrNoRows
