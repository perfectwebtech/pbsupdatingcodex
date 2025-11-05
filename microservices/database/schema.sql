-- Enterprise IPTV Platform Database Schema
-- PostgreSQL 16+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ========================================
-- USERS & AUTHENTICATION
-- ========================================

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    package_id BIGINT,
    max_connections INT DEFAULT 1,
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
    deleted_at TIMESTAMP,
    CONSTRAINT chk_max_connections CHECK (max_connections >= 1 AND max_connections <= 10)
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_active ON users(is_active, expires_at);
CREATE INDEX idx_users_package ON users(package_id);

-- ========================================
-- PACKAGES
-- ========================================

CREATE TABLE packages (
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

CREATE TABLE package_streams (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    stream_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(package_id, stream_id)
);

CREATE INDEX idx_package_streams_package ON package_streams(package_id);
CREATE INDEX idx_package_streams_stream ON package_streams(stream_id);

-- ========================================
-- STREAMS
-- ========================================

CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'live', 'vod', 'series'
    parent_id BIGINT REFERENCES categories(id),
    sort_order INT DEFAULT 0,
    icon_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_categories_type ON categories(type);
CREATE INDEX idx_categories_parent ON categories(parent_id);

CREATE TABLE streams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'live', 'vod', 'series'
    category_id BIGINT REFERENCES categories(id),
    source_urls TEXT[],
    direct_source TEXT,
    icon_url VARCHAR(500),
    bitrate INT,
    duration INT, -- in seconds for VOD
    is_active BOOLEAN DEFAULT TRUE,
    allowed_countries VARCHAR(2)[],
    metadata JSONB, -- Additional metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_streams_type ON streams(type);
CREATE INDEX idx_streams_category ON streams(category_id);
CREATE INDEX idx_streams_active ON streams(is_active);
CREATE INDEX idx_streams_name ON streams USING gin(to_tsvector('english', name));

-- ========================================
-- SERVERS
-- ========================================

CREATE TABLE servers (
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

CREATE INDEX idx_servers_active ON servers(is_active);
CREATE INDEX idx_servers_region ON servers(region);

CREATE TABLE stream_servers (
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

-- ========================================
-- SESSIONS & ACTIVITY
-- ========================================

CREATE TABLE stream_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id BIGINT NOT NULL REFERENCES users(id),
    stream_id BIGINT NOT NULL REFERENCES streams(id),
    server_id BIGINT REFERENCES servers(id),
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

-- ========================================
-- TRANSCODING
-- ========================================

CREATE TABLE transcode_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id BIGINT NOT NULL REFERENCES users(id),
    input_file VARCHAR(500) NOT NULL,
    output_file VARCHAR(500),
    preset VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'pending', 'processing', 'completed', 'failed', 'cancelled'
    progress DECIMAL(5, 2) DEFAULT 0.00,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_transcode_jobs_user ON transcode_jobs(user_id);
CREATE INDEX idx_transcode_jobs_status ON transcode_jobs(status);
CREATE INDEX idx_transcode_jobs_created ON transcode_jobs(created_at DESC);

-- ========================================
-- RECOMMENDATIONS & ML
-- ========================================

CREATE TABLE user_interactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    stream_id BIGINT NOT NULL REFERENCES streams(id),
    interaction_type VARCHAR(20) NOT NULL, -- 'view', 'like', 'rate', 'share'
    rating DECIMAL(2, 1), -- 0.0 to 5.0
    watch_duration INT, -- in seconds
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_interactions_user ON user_interactions(user_id);
CREATE INDEX idx_user_interactions_stream ON user_interactions(stream_id);
CREATE INDEX idx_user_interactions_type ON user_interactions(interaction_type);
CREATE INDEX idx_user_interactions_created ON user_interactions(created_at DESC);

CREATE TABLE user_recommendations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    stream_id BIGINT NOT NULL REFERENCES streams(id),
    score DECIMAL(5, 4) NOT NULL,
    reason VARCHAR(255),
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, stream_id)
);

CREATE INDEX idx_user_recommendations_user ON user_recommendations(user_id, score DESC);
CREATE INDEX idx_user_recommendations_stream ON user_recommendations(stream_id);

-- ========================================
-- CHAT & REAL-TIME
-- ========================================

CREATE TABLE chat_messages (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    channel VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_channel ON chat_messages(channel, created_at DESC);
CREATE INDEX idx_chat_messages_user ON chat_messages(user_id);

-- ========================================
-- ANALYTICS
-- ========================================

CREATE TABLE stream_analytics (
    id BIGSERIAL PRIMARY KEY,
    stream_id BIGINT NOT NULL REFERENCES streams(id),
    date DATE NOT NULL,
    views INT DEFAULT 0,
    unique_viewers INT DEFAULT 0,
    total_watch_time INT DEFAULT 0, -- in seconds
    avg_watch_time INT DEFAULT 0, -- in seconds
    peak_concurrent_viewers INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(stream_id, date)
);

CREATE INDEX idx_stream_analytics_stream ON stream_analytics(stream_id, date DESC);
CREATE INDEX idx_stream_analytics_date ON stream_analytics(date DESC);

-- ========================================
-- BILLING & PAYMENTS
-- ========================================

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    package_id BIGINT REFERENCES packages(id),
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50),
    transaction_id VARCHAR(255) UNIQUE,
    status VARCHAR(20) NOT NULL, -- 'pending', 'completed', 'failed', 'refunded'
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created ON payments(created_at DESC);

-- ========================================
-- ADMIN LOGS
-- ========================================

CREATE TABLE admin_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_id BIGINT NOT NULL REFERENCES users(id),
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

-- ========================================
-- FUNCTIONS & TRIGGERS
-- ========================================

-- Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_streams_updated_at BEFORE UPDATE ON streams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_packages_updated_at BEFORE UPDATE ON packages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- SEED DATA
-- ========================================

-- Create default packages
INSERT INTO packages (name, description, price, duration_days, max_connections) VALUES
('Free Trial', '7-day free trial', 0.00, 7, 1),
('Basic', 'Basic package with 1 connection', 9.99, 30, 1),
('Standard', 'Standard package with 2 connections', 19.99, 30, 2),
('Premium', 'Premium package with 5 connections', 39.99, 30, 5),
('Enterprise', 'Enterprise package with 10 connections', 99.99, 30, 10);

-- Create default categories
INSERT INTO categories (name, type) VALUES
('Live TV', 'live'),
('Sports', 'live'),
('News', 'live'),
('Movies', 'vod'),
('Series', 'series'),
('Documentaries', 'vod'),
('Kids', 'vod');

-- Create default admin user (password: admin123 - CHANGE IN PRODUCTION!)
-- Password hash generated with bcrypt
INSERT INTO users (username, password, email, is_active, admin_enabled) VALUES
('admin', '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYAO4yI8KeO', 'admin@iptv-platform.com', true, true);

-- ========================================
-- VIEWS
-- ========================================

-- Active streams view
CREATE VIEW active_streams_view AS
SELECT
    s.id,
    s.name,
    s.type,
    c.name as category_name,
    COUNT(DISTINCT ss.id) as active_viewers
FROM streams s
LEFT JOIN categories c ON s.category_id = c.id
LEFT JOIN stream_sessions ss ON s.id = ss.stream_id AND ss.ended_at IS NULL
WHERE s.is_active = true
GROUP BY s.id, s.name, s.type, c.name;

-- User activity view
CREATE VIEW user_activity_view AS
SELECT
    u.id,
    u.username,
    COUNT(DISTINCT ss.id) as total_sessions,
    SUM(EXTRACT(EPOCH FROM (COALESCE(ss.ended_at, NOW()) - ss.started_at)))::INT as total_watch_time,
    MAX(ss.started_at) as last_activity
FROM users u
LEFT JOIN stream_sessions ss ON u.id = ss.user_id
GROUP BY u.id, u.username;

-- ========================================
-- GRANTS
-- ========================================

-- Grant permissions to application user
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO iptv_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO iptv_user;

-- ========================================
-- COMMENTS
-- ========================================

COMMENT ON TABLE users IS 'User accounts and authentication';
COMMENT ON TABLE packages IS 'Subscription packages';
COMMENT ON TABLE streams IS 'Content streams (live TV, VOD, series)';
COMMENT ON TABLE stream_sessions IS 'Active and historical streaming sessions';
COMMENT ON TABLE transcode_jobs IS 'Video transcoding jobs';
COMMENT ON TABLE user_interactions IS 'User interactions for ML training';
COMMENT ON TABLE chat_messages IS 'Real-time chat messages';
COMMENT ON TABLE stream_analytics IS 'Daily stream analytics aggregates';

-- ========================================
-- COMPLETED
-- ========================================

SELECT 'Database schema created successfully!' as status;
