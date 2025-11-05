package model

import "time"

type Stream struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"` // live, vod, series
	CategoryID       int64     `json:"category_id"`
	SourceURLs       []string  `json:"source_urls"`
	DirectSource     string    `json:"direct_source"`
	IconURL          string    `json:"icon_url"`
	Bitrate          int       `json:"bitrate"`
	Duration         int       `json:"duration"` // For VOD, in seconds
	IsActive         bool      `json:"is_active"`
	AllowedCountries []string  `json:"allowed_countries"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Server struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Domain            string    `json:"domain"`
	Port              int       `json:"port"`
	HTTPS             bool      `json:"https"`
	IsActive          bool      `json:"is_active"`
	Weight            int       `json:"weight"`
	MaxConnections    int       `json:"max_connections"`
	ActiveConnections int       `json:"active_connections"`
	Region            string    `json:"region"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type StreamSession struct {
	ID           string    `json:"id"`
	UserID       int64     `json:"user_id"`
	StreamID     int64     `json:"stream_id"`
	ServerID     int64     `json:"server_id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Container    string    `json:"container"`
	StartedAt    time.Time `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	LastActivity time.Time `json:"last_activity"`
	BytesSent    int64     `json:"bytes_sent"`
}

type Category struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // live, vod, series
	ParentID     *int64    `json:"parent_id,omitempty"`
	SortOrder    int       `json:"sort_order"`
	StreamsCount int       `json:"streams_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID             int64      `json:"id"`
	Username       string     `json:"username"`
	MaxConnections int        `json:"max_connections"`
	IsActive       bool       `json:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type ServerHealth struct {
	ServerID  int64   `json:"server_id"`
	Healthy   bool    `json:"healthy"`
	CPUUsage  float64 `json:"cpu_usage"`
	MemUsage  float64 `json:"mem_usage"`
	BandWidth float64 `json:"bandwidth"`
}
