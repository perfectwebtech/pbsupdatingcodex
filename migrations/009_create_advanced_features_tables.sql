-- =============================================================================
-- Migration 009: Advanced Features Tables
-- Includes: P2P Streaming, Referral Program
-- (GraphQL uses existing tables)
-- =============================================================================

-- =============================================================================
-- P2P STREAMING TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS p2p_peers (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    connection_id VARCHAR(36) NOT NULL,
    bandwidth_mbps INT DEFAULT 0,
    upload_speed_kbps BIGINT DEFAULT 0,
    download_speed_kbps BIGINT DEFAULT 0,
    is_seeder TINYINT(1) DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    bytes_uploaded BIGINT DEFAULT 0,
    bytes_downloaded BIGINT DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_active (is_active, last_seen),
    INDEX idx_seeder (is_seeder, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS p2p_swarms (
    id VARCHAR(36) PRIMARY KEY,
    content_id VARCHAR(36) NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    total_peers INT DEFAULT 0,
    seeders INT DEFAULT 0,
    leechers INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_content (content_id),
    INDEX idx_updated (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS p2p_swarm_members (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    swarm_id VARCHAR(36) NOT NULL,
    peer_id VARCHAR(36) NOT NULL,
    content_id VARCHAR(36) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP NULL,
    UNIQUE KEY uk_swarm_peer (swarm_id, peer_id),
    INDEX idx_peer (peer_id),
    INDEX idx_content (content_id),
    INDEX idx_active (swarm_id, left_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS p2p_chunk_transfers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    peer_id VARCHAR(36) NOT NULL,
    chunk_id VARCHAR(64) NOT NULL,
    source ENUM('p2p', 'cdn') NOT NULL,
    bytes BIGINT NOT NULL,
    transferred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_peer_time (peer_id, transferred_at),
    INDEX idx_source (source, transferred_at),
    INDEX idx_transferred (transferred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS p2p_analytics (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    bandwidth_saved_bytes BIGINT DEFAULT 0,
    cost_saved_usd DECIMAL(12,4) DEFAULT 0,
    p2p_efficiency DECIMAL(5,2) DEFAULT 0,
    active_swarms INT DEFAULT 0,
    total_peers INT DEFAULT 0,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_recorded (recorded_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- REFERRAL PROGRAM TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS referral_codes (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    type ENUM('user', 'affiliate', 'campaign') DEFAULT 'user',
    is_active TINYINT(1) DEFAULT 1,
    usage_count INT DEFAULT 0,
    max_uses INT DEFAULT 0,
    expires_at TIMESTAMP NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_code (code),
    INDEX idx_active (is_active, expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS referrals (
    id VARCHAR(36) PRIMARY KEY,
    referrer_id VARCHAR(36) NOT NULL,
    referred_user_id VARCHAR(36) NOT NULL,
    referral_code VARCHAR(20) NOT NULL,
    status ENUM('pending', 'qualified', 'rewarded', 'expired') DEFAULT 'pending',
    referrer_reward DECIMAL(10,2) DEFAULT 0,
    referred_reward DECIMAL(10,2) DEFAULT 0,
    purchase_amount DECIMAL(10,2) DEFAULT 0,
    commission_earned DECIMAL(10,2) DEFAULT 0,
    referred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    qualified_at TIMESTAMP NULL,
    rewarded_at TIMESTAMP NULL,
    INDEX idx_referrer (referrer_id),
    INDEX idx_referred (referred_user_id),
    INDEX idx_code (referral_code),
    INDEX idx_status (status, referred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS referral_rewards (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    referral_id VARCHAR(36),
    type ENUM('credits', 'discount', 'cash', 'tier_bonus') NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status ENUM('pending', 'issued', 'claimed', 'expired') DEFAULT 'issued',
    expires_at TIMESTAMP NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    claimed_at TIMESTAMP NULL,
    INDEX idx_user_status (user_id, status),
    INDEX idx_referral (referral_id),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS referral_payouts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50),
    payment_details VARCHAR(255),
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    INDEX idx_user (user_id),
    INDEX idx_status (status, requested_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS affiliate_tiers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    min_referrals INT NOT NULL,
    commission_percent DECIMAL(5,2) NOT NULL,
    bonus_reward DECIMAL(10,2) DEFAULT 0,
    color VARCHAR(7),
    icon VARCHAR(50),
    description TEXT,
    UNIQUE KEY uk_name (name),
    INDEX idx_min_referrals (min_referrals)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Add columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_balance DECIMAL(10,2) DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS discount_balance DECIMAL(10,2) DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS affiliate_tier VARCHAR(36) DEFAULT 'bronze';

-- =============================================================================
-- SEED DATA FOR AFFILIATE TIERS
-- =============================================================================

INSERT IGNORE INTO affiliate_tiers (id, name, min_referrals, commission_percent, bonus_reward, color, description) VALUES
    ('bronze', 'Bronze', 0, 10.00, 0, '#CD7F32', 'Starting tier for all users'),
    ('silver', 'Silver', 10, 15.00, 50.00, '#C0C0C0', 'Achieve 10 qualified referrals'),
    ('gold', 'Gold', 25, 20.00, 150.00, '#FFD700', 'Achieve 25 qualified referrals'),
    ('platinum', 'Platinum', 50, 25.00, 500.00, '#E5E4E2', 'Achieve 50 qualified referrals'),
    ('diamond', 'Diamond', 100, 30.00, 1500.00, '#B9F2FF', 'Elite tier - 100+ qualified referrals');

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- User table indexes for new columns
CREATE INDEX IF NOT EXISTS idx_referral_balance ON users(referral_balance);
CREATE INDEX IF NOT EXISTS idx_affiliate_tier ON users(affiliate_tier);

-- =============================================================================
-- END OF MIGRATION 009
-- =============================================================================
