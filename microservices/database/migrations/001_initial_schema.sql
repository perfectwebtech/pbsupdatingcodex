-- Migration: 001_initial_schema
-- Description: Create initial database schema with all tables
-- Author: IPTV Platform Team
-- Date: 2025-01-05

BEGIN;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For fuzzy text search

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    package_id BIGINT,
    max_connections INT DEFAULT 1 CHECK (max_connections >= 1 AND max_connections <= 10),
    is_trial BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    admin_enabled BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP,
    allowed_ips JSONB,
    allowed_countries VARCHAR(2)[],
    isp_lock VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active, expires_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_package ON users(package_id);

-- Packages table
CREATE TABLE IF NOT EXISTS packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2),
    duration_days INT,
    max_connections INT DEFAULT 1,
    is_trial BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('live', 'vod', 'series')),
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT DEFAULT 0,
    icon_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_categories_type ON categories(type);
CREATE INDEX idx_categories_parent ON categories(parent_id);

-- Streams table
CREATE TABLE IF NOT EXISTS streams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('live', 'vod', 'series')),
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    source_urls TEXT[],
    direct_source TEXT,
    icon_url VARCHAR(500),
    bitrate INT,
    duration INT,
    is_active BOOLEAN DEFAULT TRUE,
    allowed_countries VARCHAR(2)[],
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_streams_type ON streams(type);
CREATE INDEX idx_streams_category ON streams(category_id);
CREATE INDEX idx_streams_active ON streams(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_streams_name_trgm ON streams USING gin(name gin_trgm_ops);

-- Package-Stream relationship
CREATE TABLE IF NOT EXISTS package_streams (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    stream_id BIGINT NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(package_id, stream_id)
);

CREATE INDEX idx_package_streams_package ON package_streams(package_id);
CREATE INDEX idx_package_streams_stream ON package_streams(stream_id);

-- Servers table
CREATE TABLE IF NOT EXISTS servers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    port INT DEFAULT 80,
    https BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    weight INT DEFAULT 100,
    max_connections INT DEFAULT 1000,
    active_connections INT DEFAULT 0,
    region VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_servers_active ON servers(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_servers_region ON servers(region);

-- Stream-Server relationship
CREATE TABLE IF NOT EXISTS stream_servers (
    id BIGSERIAL PRIMARY KEY,
    stream_id BIGINT NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    server_id BIGINT NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    pid INT,
    status VARCHAR(20),
    bitrate INT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(stream_id, server_id)
);

CREATE INDEX idx_stream_servers_stream ON stream_servers(stream_id);
CREATE INDEX idx_stream_servers_server ON stream_servers(server_id);

-- Stream sessions
CREATE TABLE IF NOT EXISTS stream_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stream_id BIGINT NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    server_id BIGINT REFERENCES servers(id) ON DELETE SET NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    container VARCHAR(20),
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP,
    last_activity TIMESTAMP DEFAULT NOW(),
    bytes_sent BIGINT DEFAULT 0
);

CREATE INDEX idx_stream_sessions_user ON stream_sessions(user_id);
CREATE INDEX idx_stream_sessions_stream ON stream_sessions(stream_id);
CREATE INDEX idx_stream_sessions_active ON stream_sessions(user_id, ended_at) WHERE ended_at IS NULL;
CREATE INDEX idx_stream_sessions_started ON stream_sessions(started_at DESC);

-- Transcode jobs
CREATE TABLE IF NOT EXISTS transcode_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    input_file VARCHAR(500) NOT NULL,
    output_file VARCHAR(500),
    preset VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    progress DECIMAL(5, 2) DEFAULT 0.00 CHECK (progress >= 0 AND progress <= 100),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_transcode_jobs_user ON transcode_jobs(user_id);
CREATE INDEX idx_transcode_jobs_status ON transcode_jobs(status);
CREATE INDEX idx_transcode_jobs_created ON transcode_jobs(created_at DESC);

-- User interactions (for ML)
CREATE TABLE IF NOT EXISTS user_interactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stream_id BIGINT NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    interaction_type VARCHAR(20) NOT NULL CHECK (interaction_type IN ('view', 'like', 'rate', 'share', 'complete')),
    rating DECIMAL(2, 1) CHECK (rating IS NULL OR (rating >= 0 AND rating <= 5)),
    watch_duration INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_interactions_user ON user_interactions(user_id);
CREATE INDEX idx_user_interactions_stream ON user_interactions(stream_id);
CREATE INDEX idx_user_interactions_type ON user_interactions(interaction_type);
CREATE INDEX idx_user_interactions_created ON user_interactions(created_at DESC);

-- User recommendations (cached)
CREATE TABLE IF NOT EXISTS user_recommendations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stream_id BIGINT NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    score DECIMAL(5, 4) NOT NULL,
    reason VARCHAR(255),
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, stream_id)
);

CREATE INDEX idx_user_recommendations_user ON user_recommendations(user_id, score DESC);
CREATE INDEX idx_user_recommendations_stream ON user_recommendations(stream_id);

-- Chat messages
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_channel ON chat_messages(channel, created_at DESC);
CREATE INDEX idx_chat_messages_user ON chat_messages(user_id);

-- Stream analytics
CREATE TABLE IF NOT EXISTS stream_analytics (
    id BIGSERIAL PRIMARY KEY,
    stream_id BIGINT NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views INT DEFAULT 0,
    unique_viewers INT DEFAULT 0,
    total_watch_time INT DEFAULT 0,
    avg_watch_time INT DEFAULT 0,
    peak_concurrent_viewers INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(stream_id, date)
);

CREATE INDEX idx_stream_analytics_stream ON stream_analytics(stream_id, date DESC);
CREATE INDEX idx_stream_analytics_date ON stream_analytics(date DESC);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_id BIGINT REFERENCES packages(id) ON DELETE SET NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50),
    transaction_id VARCHAR(255) UNIQUE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created ON payments(created_at DESC);
CREATE INDEX idx_payments_transaction ON payments(transaction_id);

-- Admin logs
CREATE TABLE IF NOT EXISTS admin_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id BIGINT,
    details JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_admin_logs_admin ON admin_logs(admin_id);
CREATE INDEX idx_admin_logs_action ON admin_logs(action);
CREATE INDEX idx_admin_logs_created ON admin_logs(created_at DESC);

-- Foreign key for users.package_id (added after packages table exists)
ALTER TABLE users ADD CONSTRAINT fk_users_package FOREIGN KEY (package_id) REFERENCES packages(id) ON DELETE SET NULL;

COMMIT;

-- Add comment
COMMENT ON SCHEMA public IS 'Initial schema for IPTV Platform v2.0';
