-- =============================================================================
-- Migration 008: Powerful Features Tables
-- Includes: Dynamic Pricing, Live Chat, Multi-CDN, Fraud Detection, AVOD
-- =============================================================================

-- =============================================================================
-- DYNAMIC PRICING ENGINE TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS pricing_rules (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    package_id VARCHAR(36) NOT NULL,
    rule_type ENUM('demand', 'geo', 'time', 'promotional', 'competitor', 'churn_prevention') NOT NULL,
    conditions JSON,
    price_multiplier DECIMAL(5,4) DEFAULT 1.0000,
    min_price DECIMAL(10,2),
    max_price DECIMAL(10,2),
    priority INT DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    starts_at TIMESTAMP NULL,
    ends_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_package_active (package_id, is_active),
    INDEX idx_rule_type (rule_type),
    INDEX idx_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS price_calculations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    package_id VARCHAR(36) NOT NULL,
    base_price DECIMAL(10,2) NOT NULL,
    final_price DECIMAL(10,2) NOT NULL,
    discount_percent DECIMAL(5,2),
    country_code VARCHAR(3),
    applied_rules TEXT,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_package (user_id, package_id),
    INDEX idx_calculated_at (calculated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS price_tests (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    package_id VARCHAR(36) NOT NULL,
    variant_a_price DECIMAL(10,2) NOT NULL,
    variant_b_price DECIMAL(10,2) NOT NULL,
    traffic_split DECIMAL(3,2) DEFAULT 0.50,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active TINYINT(1) DEFAULT 1,
    conversions_a INT DEFAULT 0,
    conversions_b INT DEFAULT 0,
    revenue_a DECIMAL(12,2) DEFAULT 0,
    revenue_b DECIMAL(12,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_active (is_active, start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS price_test_assignments (
    user_id VARCHAR(36) NOT NULL,
    test_id VARCHAR(36) NOT NULL,
    variant CHAR(1) NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    converted TINYINT(1) DEFAULT 0,
    revenue DECIMAL(10,2) DEFAULT 0,
    PRIMARY KEY (user_id, test_id),
    INDEX idx_test_variant (test_id, variant)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- LIVE CHAT TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS chat_rooms (
    id VARCHAR(36) PRIMARY KEY,
    stream_id VARCHAR(36) NOT NULL,
    name VARCHAR(255),
    is_active TINYINT(1) DEFAULT 1,
    is_moderated TINYINT(1) DEFAULT 1,
    slow_mode INT DEFAULT 0,
    max_message_length INT DEFAULT 500,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_stream (stream_id),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS chat_messages (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    message TEXT NOT NULL,
    type ENUM('text', 'emoji', 'gif', 'super_chat', 'system') DEFAULT 'text',
    color VARCHAR(7),
    super_chat_amount DECIMAL(10,2),
    mentions JSON,
    is_deleted TINYINT(1) DEFAULT 0,
    deleted_by VARCHAR(36),
    deleted_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_room_created (room_id, created_at),
    INDEX idx_user (user_id),
    INDEX idx_super_chat (super_chat_amount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS stream_reactions (
    id VARCHAR(36) PRIMARY KEY,
    stream_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    emoji VARCHAR(10) NOT NULL,
    timestamp BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_stream_emoji (stream_id, emoji),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS live_polls (
    id VARCHAR(36) PRIMARY KEY,
    stream_id VARCHAR(36) NOT NULL,
    question TEXT NOT NULL,
    options JSON NOT NULL,
    duration INT NOT NULL,
    is_active TINYINT(1) DEFAULT 1,
    started_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,
    INDEX idx_stream_active (stream_id, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS poll_votes (
    id VARCHAR(36) PRIMARY KEY,
    poll_id VARCHAR(36) NOT NULL,
    option_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    voted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_poll_user (poll_id, user_id),
    INDEX idx_option (option_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS chat_moderation (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    moderator_id VARCHAR(36),
    action ENUM('mute', 'ban', 'warn', 'timeout') NOT NULL,
    reason TEXT,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_room_user (room_id, user_id),
    INDEX idx_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_badges (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    badge_type VARCHAR(50) NOT NULL,
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    is_active TINYINT(1) DEFAULT 1,
    UNIQUE KEY uk_user_badge (user_id, badge_type),
    INDEX idx_user_active (user_id, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- MULTI-CDN TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS cdn_providers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type ENUM('cloudflare', 'cloudfront', 'bunny', 'fastly', 'akamai', 'keycdn', 'custom') NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    api_key VARCHAR(255),
    priority INT DEFAULT 0,
    weight INT DEFAULT 100,
    is_active TINYINT(1) DEFAULT 1,
    cost_per_gb DECIMAL(8,4) DEFAULT 0.0500,
    bandwidth_monthly_limit BIGINT,
    bandwidth_used_month BIGINT DEFAULT 0,
    regions JSON,
    features JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_active_priority (is_active, priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS cdn_health (
    provider_id VARCHAR(36) PRIMARY KEY,
    is_healthy TINYINT(1) DEFAULT 1,
    response_time_ms INT,
    success_rate DECIMAL(5,4) DEFAULT 1.0000,
    error_rate DECIMAL(5,4) DEFAULT 0.0000,
    consecutive_failures INT DEFAULT 0,
    uptime_percent DECIMAL(5,2) DEFAULT 100.00,
    last_checked TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_health_status (is_healthy, last_checked)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS cdn_usage (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    provider_id VARCHAR(36) NOT NULL,
    content_path VARCHAR(500) NOT NULL,
    requests BIGINT DEFAULT 0,
    bandwidth_bytes BIGINT DEFAULT 0,
    last_request TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_provider_path (provider_id, content_path(255)),
    INDEX idx_last_request (last_request)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS cdn_daily_stats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    provider_id VARCHAR(36) NOT NULL,
    date DATE NOT NULL,
    bandwidth_bytes BIGINT DEFAULT 0,
    requests BIGINT DEFAULT 0,
    cache_hit_rate DECIMAL(5,4),
    avg_latency_ms INT,
    cost DECIMAL(10,4) DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    UNIQUE KEY uk_provider_date (provider_id, date),
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- FRAUD DETECTION TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS fraud_checks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    device_id VARCHAR(255),
    risk_score DECIMAL(5,2) NOT NULL,
    decision ENUM('allow', 'challenge', 'block') NOT NULL,
    reasons TEXT,
    country VARCHAR(3),
    is_vpn TINYINT(1) DEFAULT 0,
    is_proxy TINYINT(1) DEFAULT 0,
    is_tor TINYINT(1) DEFAULT 0,
    fingerprint VARCHAR(64),
    metadata JSON,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_checked (user_id, checked_at),
    INDEX idx_ip (ip_address),
    INDEX idx_decision (decision),
    INDEX idx_risk_score (risk_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS device_fingerprints (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    fingerprint VARCHAR(64) NOT NULL,
    components TEXT,
    user_agent VARCHAR(500),
    screen_size VARCHAR(20),
    timezone VARCHAR(50),
    language VARCHAR(10),
    plugins TEXT,
    ip_address VARCHAR(45),
    first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    use_count INT DEFAULT 1,
    is_trusted TINYINT(1) DEFAULT 0,
    UNIQUE KEY uk_user_fingerprint (user_id, fingerprint),
    INDEX idx_fingerprint (fingerprint),
    INDEX idx_last_seen (last_seen)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS suspicious_activities (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    ip_address VARCHAR(45),
    activity_type VARCHAR(50),
    description TEXT,
    severity ENUM('low', 'medium', 'high', 'critical') NOT NULL,
    action VARCHAR(50),
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_severity (user_id, severity),
    INDEX idx_detected (detected_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fraud_blocklist (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    ip_address VARCHAR(45),
    fingerprint VARCHAR(64),
    reason TEXT,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_ip (ip_address),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS login_attempts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL,
    user_id VARCHAR(36),
    success TINYINT(1) NOT NULL,
    user_agent VARCHAR(500),
    country VARCHAR(3),
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ip_attempted (ip_address, attempted_at),
    INDEX idx_user_success (user_id, success, attempted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_locations (
    user_id VARCHAR(36) PRIMARY KEY,
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    ip_address VARCHAR(45),
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_last_seen (last_seen)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ip_intelligence (
    ip_address VARCHAR(45) PRIMARY KEY,
    is_vpn TINYINT(1) DEFAULT 0,
    is_proxy TINYINT(1) DEFAULT 0,
    is_tor TINYINT(1) DEFAULT 0,
    is_datacenter TINYINT(1) DEFAULT 0,
    asn INT,
    organization VARCHAR(255),
    risk_score DECIMAL(5,2),
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_risk (risk_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tor_exit_nodes (
    ip_address VARCHAR(45) PRIMARY KEY,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS api_requests (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36),
    ip_address VARCHAR(45),
    endpoint VARCHAR(255),
    method VARCHAR(10),
    status_code INT,
    response_time_ms INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_ip_created (ip_address, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS active_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    device_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_activity (user_id, last_activity),
    INDEX idx_device (device_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS session_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    ip_address VARCHAR(45),
    country VARCHAR(3),
    device_id VARCHAR(255),
    session_type VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- AVOD (AD-SUPPORTED) TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS advertisers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    company VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),
    balance DECIMAL(12,2) DEFAULT 0,
    total_spent DECIMAL(12,2) DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ad_campaigns (
    id VARCHAR(36) PRIMARY KEY,
    advertiser_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    budget DECIMAL(12,2) NOT NULL,
    spent DECIMAL(12,2) DEFAULT 0,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_advertiser (advertiser_id),
    INDEX idx_active_dates (is_active, start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS advertisements (
    id VARCHAR(36) PRIMARY KEY,
    advertiser_id VARCHAR(36) NOT NULL,
    campaign_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    type ENUM('pre-roll', 'mid-roll', 'post-roll', 'banner', 'overlay', 'interactive', 'any') NOT NULL,
    format ENUM('video', 'image', 'html5', 'vast') DEFAULT 'video',
    media_url VARCHAR(500) NOT NULL,
    click_url VARCHAR(500),
    duration INT DEFAULT 15,
    skippable TINYINT(1) DEFAULT 0,
    skip_after INT DEFAULT 5,
    target_groups JSON,
    target_genres JSON,
    target_countries JSON,
    min_age INT DEFAULT 0,
    max_age INT DEFAULT 0,
    bid_amount DECIMAL(8,4) NOT NULL,
    daily_budget DECIMAL(10,2),
    total_budget DECIMAL(12,2),
    spent_today DECIMAL(10,2) DEFAULT 0,
    spent_total DECIMAL(12,2) DEFAULT 0,
    impressions BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_active_dates (is_active, start_date, end_date),
    INDEX idx_advertiser (advertiser_id),
    INDEX idx_campaign (campaign_id),
    INDEX idx_type_bid (type, bid_amount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ad_impressions (
    id VARCHAR(36) PRIMARY KEY,
    ad_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    stream_id VARCHAR(36),
    country VARCHAR(3),
    device_type VARCHAR(20),
    position VARCHAR(20),
    watched_seconds INT DEFAULT 0,
    completed TINYINT(1) DEFAULT 0,
    skipped TINYINT(1) DEFAULT 0,
    clicked TINYINT(1) DEFAULT 0,
    revenue DECIMAL(8,4) DEFAULT 0,
    confirmed TINYINT(1) DEFAULT 0,
    started_at TIMESTAMP NULL,
    clicked_at TIMESTAMP NULL,
    shown_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ad_shown (ad_id, shown_at),
    INDEX idx_user_ad (user_id, ad_id),
    INDEX idx_user_recent (user_id, shown_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Add ad_free column to packages table
ALTER TABLE packages ADD COLUMN IF NOT EXISTS ad_free TINYINT(1) DEFAULT 0;

-- Add suspended/locked columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_suspended TINYINT(1) DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS suspension_reason TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_locked TINYINT(1) DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP NULL;

-- =============================================================================
-- INDEXES & PERFORMANCE
-- =============================================================================

-- Pre-populate sample CDN providers
INSERT IGNORE INTO cdn_providers (id, name, type, base_url, priority, weight, is_active, cost_per_gb)
VALUES
    ('cdn-cloudflare', 'Cloudflare', 'cloudflare', 'https://cdn.cloudflare.iptv.example.com', 10, 100, 1, 0.0400),
    ('cdn-cloudfront', 'AWS CloudFront', 'cloudfront', 'https://cdn.cloudfront.iptv.example.com', 8, 80, 1, 0.0850),
    ('cdn-bunny', 'BunnyCDN', 'bunny', 'https://cdn.bunny.iptv.example.com', 9, 90, 1, 0.0100),
    ('cdn-fastly', 'Fastly', 'fastly', 'https://cdn.fastly.iptv.example.com', 7, 70, 0, 0.1200);

-- =============================================================================
-- END OF MIGRATION 008
-- =============================================================================
