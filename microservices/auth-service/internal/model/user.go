package model

import "time"

type User struct {
	ID             int64      `json:"id"`
	Username       string     `json:"username"`
	Password       string     `json:"-"`
	Email          string     `json:"email,omitempty"`
	PackageID      *int64     `json:"package_id,omitempty"`
	MaxConnections int        `json:"max_connections"`
	IsTrial        bool       `json:"is_trial"`
	IsActive       bool       `json:"is_active"`
	AdminEnabled   bool       `json:"admin_enabled"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	AllowedIPs     []string   `json:"allowed_ips,omitempty"`
	AllowedCountries []string `json:"allowed_countries,omitempty"`
	ISPLock        string     `json:"isp_lock,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
}

type Session struct {
	ID           string    `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DeviceID     string    `json:"device_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
