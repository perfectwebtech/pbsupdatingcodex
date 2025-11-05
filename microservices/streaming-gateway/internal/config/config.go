package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	GeoIP     GeoIPConfig
	Streaming StreamingConfig
}

type ServerConfig struct {
	HTTPPort    int
	Environment string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MinConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type JWTConfig struct {
	Secret string
}

type GeoIPConfig struct {
	DatabasePath string
}

type StreamingConfig struct {
	SegmentDuration    int
	BufferSize         int
	PrebufferSegments  int
	MaxBitrate         int
	TranscodingEnabled bool
	CDNEnabled         bool
	CDNBaseURL         string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			HTTPPort:    getEnvAsInt("HTTP_PORT", 8000),
			Environment: getEnv("ENVIRONMENT", "production"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "iptv_user"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "iptv_streaming"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 100),
			MinConns: getEnvAsInt("DB_MIN_CONNS", 10),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 100),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		GeoIP: GeoIPConfig{
			DatabasePath: getEnv("GEOIP_DATABASE_PATH", "/usr/share/GeoIP/GeoLite2-City.mmdb"),
		},
		Streaming: StreamingConfig{
			SegmentDuration:    getEnvAsInt("SEGMENT_DURATION", 10),
			BufferSize:         getEnvAsInt("BUFFER_SIZE", 8192),
			PrebufferSegments:  getEnvAsInt("PREBUFFER_SEGMENTS", 3),
			MaxBitrate:         getEnvAsInt("MAX_BITRATE", 8000000),
			TranscodingEnabled: getEnvAsBool("TRANSCODING_ENABLED", false),
			CDNEnabled:         getEnvAsBool("CDN_ENABLED", false),
			CDNBaseURL:         getEnv("CDN_BASE_URL", ""),
		},
	}

	// Validate required fields
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
