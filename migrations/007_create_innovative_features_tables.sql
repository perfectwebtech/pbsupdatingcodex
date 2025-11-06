-- Migration: Create innovative features tables
-- Description: AI recommendations, social features, watch parties, analytics
-- Version: 007
-- Date: 2025-11-06

-- ============================================================================
-- View History Table (for AI recommendations)
-- ============================================================================
CREATE TABLE IF NOT EXISTS view_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NULL,
    quality VARCHAR(20) NULL,
    watch_duration INT DEFAULT 0 COMMENT 'Seconds watched',
    completion_percentage INT DEFAULT 0 COMMENT '0-100',
    device_type VARCHAR(50) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_content (user_id, content_type, content_id),
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_content (content_type, content_id),
    INDEX idx_category (category_id),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Ratings Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS ratings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    rating DECIMAL(2,1) NOT NULL COMMENT '0.0-5.0',
    review TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_rating (user_id, content_type, content_id),
    INDEX idx_content (content_type, content_id),
    INDEX idx_rating (rating),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Watch Parties Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS watch_parties (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    host_user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    party_name VARCHAR(255) NOT NULL,
    party_code VARCHAR(20) NOT NULL UNIQUE COMMENT 'Join code',
    max_participants INT DEFAULT 10,
    is_public BOOLEAN DEFAULT FALSE,
    status ENUM('scheduled', 'active', 'completed', 'cancelled') DEFAULT 'scheduled',
    scheduled_at TIMESTAMP NULL,
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    current_timestamp INT DEFAULT 0 COMMENT 'Sync position in seconds',
    is_playing BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_host (host_user_id),
    INDEX idx_party_code (party_code),
    INDEX idx_status (status),
    INDEX idx_scheduled (scheduled_at),

    FOREIGN KEY (host_user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Watch Party Participants Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS watch_party_participants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    party_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    display_name VARCHAR(100) NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,

    UNIQUE KEY unique_participant (party_id, user_id),
    INDEX idx_party (party_id),
    INDEX idx_user (user_id),

    FOREIGN KEY (party_id) REFERENCES watch_parties(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Watch Party Chat Messages
-- ============================================================================
CREATE TABLE IF NOT EXISTS watch_party_messages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    party_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    message TEXT NOT NULL,
    message_type ENUM('text', 'emoji', 'system') DEFAULT 'text',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_party (party_id),
    INDEX idx_created (created_at),

    FOREIGN KEY (party_id) REFERENCES watch_parties(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Social Shares Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS social_shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode', 'party') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    platform ENUM('facebook', 'twitter', 'whatsapp', 'telegram', 'link') NOT NULL,
    share_message TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user (user_id),
    INDEX idx_content (content_type, content_id),
    INDEX idx_platform (platform),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- User Comments Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_comments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    parent_comment_id BIGINT UNSIGNED NULL COMMENT 'For replies',
    comment TEXT NOT NULL,
    likes_count INT DEFAULT 0,
    is_spoiler BOOLEAN DEFAULT FALSE,
    is_approved BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_content (content_type, content_id),
    INDEX idx_user (user_id),
    INDEX idx_parent (parent_comment_id),
    INDEX idx_approved (is_approved),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES user_comments(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Comment Likes Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS comment_likes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    comment_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY unique_like (comment_id, user_id),
    INDEX idx_comment (comment_id),

    FOREIGN KEY (comment_id) REFERENCES user_comments(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- User Playlists Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_playlists (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    thumbnail VARCHAR(500) NULL,
    items_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user (user_id),
    INDEX idx_public (is_public),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Playlist Items Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS playlist_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    playlist_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    sort_order INT DEFAULT 0,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY unique_item (playlist_id, content_type, content_id),
    INDEX idx_playlist (playlist_id),

    FOREIGN KEY (playlist_id) REFERENCES user_playlists(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Predictive Analytics Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS predictive_analytics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    content_type ENUM('stream', 'movie', 'series') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    prediction_type ENUM('popularity', 'churn_risk', 'revenue', 'quality_issues') NOT NULL,
    prediction_value DECIMAL(10,2) NOT NULL,
    confidence_score DECIMAL(5,2) NOT NULL COMMENT '0-100',
    prediction_date DATE NOT NULL,
    actual_value DECIMAL(10,2) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY unique_prediction (content_type, content_id, prediction_type, prediction_date),
    INDEX idx_content (content_type, content_id),
    INDEX idx_type (prediction_type),
    INDEX idx_date (prediction_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- User Activity Log (for ML training)
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_activity_log (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    activity_type VARCHAR(50) NOT NULL,
    activity_data JSON NULL,
    ip_address VARCHAR(45) NULL,
    user_agent TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user (user_id),
    INDEX idx_type (activity_type),
    INDEX idx_created (created_at),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- A/B Testing Experiments Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS ab_experiments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    experiment_name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    status ENUM('draft', 'active', 'paused', 'completed') DEFAULT 'draft',
    variants JSON NOT NULL COMMENT 'Array of variants with their configs',
    success_metric VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_status (status),
    INDEX idx_dates (start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- A/B Test Assignments Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS ab_test_assignments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    experiment_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    variant_id VARCHAR(50) NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY unique_assignment (experiment_id, user_id),
    INDEX idx_experiment (experiment_id),
    INDEX idx_user (user_id),

    FOREIGN KEY (experiment_id) REFERENCES ab_experiments(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Content Quality Monitoring Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS content_quality_monitoring (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    content_type ENUM('stream', 'movie', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    check_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_available BOOLEAN DEFAULT TRUE,
    response_time_ms INT NULL,
    bitrate_kbps INT NULL,
    fps DECIMAL(5,2) NULL,
    resolution VARCHAR(20) NULL,
    error_rate DECIMAL(5,2) NULL COMMENT 'Percentage',
    buffer_ratio DECIMAL(5,2) NULL COMMENT 'Percentage',
    quality_score INT NULL COMMENT '0-100',

    INDEX idx_content (content_type, content_id),
    INDEX idx_timestamp (check_timestamp),
    INDEX idx_quality (quality_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Sample Data
-- ============================================================================

-- Sample View History
INSERT INTO view_history (user_id, content_type, content_id, category_id, quality, watch_duration, completion_percentage, device_type, created_at) VALUES
(1, 'movie', 1, 1, 'hd', 3600, 80, 'mobile', NOW() - INTERVAL 1 DAY),
(1, 'stream', 1, 2, 'fhd', 1200, 100, 'tv', NOW() - INTERVAL 2 HOUR),
(2, 'movie', 1, 1, 'hd', 4500, 100, 'web', NOW() - INTERVAL 3 DAY),
(2, 'series', 1, 3, 'hd', 2400, 95, 'mobile', NOW() - INTERVAL 1 DAY),
(3, 'stream', 2, 2, 'uhd', 1800, 100, 'tv', NOW() - INTERVAL 4 HOUR);

-- Sample Ratings
INSERT INTO ratings (user_id, content_type, content_id, rating, review) VALUES
(1, 'movie', 1, 4.5, 'Great movie! Highly recommended.'),
(2, 'movie', 1, 5.0, 'Masterpiece!'),
(3, 'stream', 1, 4.0, 'Good quality stream'),
(1, 'series', 1, 4.8, 'Best series ever!');

-- Sample Watch Party
INSERT INTO watch_parties (host_user_id, content_type, content_id, party_name, party_code, max_participants, is_public, status, scheduled_at) VALUES
(1, 'movie', 1, 'Movie Night with Friends', 'PARTY123', 10, true, 'scheduled', NOW() + INTERVAL 2 HOUR);

-- Sample User Playlists
INSERT INTO user_playlists (user_id, name, description, is_public, items_count) VALUES
(1, 'My Favorites', 'Collection of my favorite content', true, 0),
(1, 'Watch Later', 'Content to watch later', false, 0),
(2, 'Action Movies', 'Best action movies', true, 0);

-- ============================================================================
-- Triggers
-- ============================================================================

DELIMITER //

-- Update likes count on comment_likes insert
CREATE TRIGGER after_comment_like_insert
AFTER INSERT ON comment_likes
FOR EACH ROW
BEGIN
    UPDATE user_comments SET likes_count = likes_count + 1 WHERE id = NEW.comment_id;
END//

-- Update likes count on comment_likes delete
CREATE TRIGGER after_comment_like_delete
AFTER DELETE ON comment_likes
FOR EACH ROW
BEGIN
    UPDATE user_comments SET likes_count = likes_count - 1 WHERE id = OLD.comment_id;
END//

-- Update playlist items count
CREATE TRIGGER after_playlist_item_insert
AFTER INSERT ON playlist_items
FOR EACH ROW
BEGIN
    UPDATE user_playlists SET items_count = items_count + 1 WHERE id = NEW.playlist_id;
END//

CREATE TRIGGER after_playlist_item_delete
AFTER DELETE ON playlist_items
FOR EACH ROW
BEGIN
    UPDATE user_playlists SET items_count = items_count - 1 WHERE id = OLD.playlist_id;
END//

DELIMITER ;

-- ============================================================================
-- Views for Analytics
-- ============================================================================

-- Popular content view
CREATE OR REPLACE VIEW popular_content AS
SELECT
    content_type,
    content_id,
    COUNT(DISTINCT user_id) as unique_viewers,
    COUNT(*) as total_views,
    AVG(completion_percentage) as avg_completion,
    AVG(watch_duration) as avg_watch_time
FROM view_history
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY content_type, content_id
ORDER BY unique_viewers DESC, total_views DESC;

-- User engagement view
CREATE OR REPLACE VIEW user_engagement_stats AS
SELECT
    u.id as user_id,
    u.username,
    COUNT(DISTINCT vh.id) as views_count,
    COUNT(DISTINCT r.id) as ratings_count,
    COUNT(DISTINCT c.id) as comments_count,
    COUNT(DISTINCT f.id) as favorites_count,
    MAX(vh.created_at) as last_activity
FROM users u
LEFT JOIN view_history vh ON vh.user_id = u.id
LEFT JOIN ratings r ON r.user_id = u.id
LEFT JOIN user_comments c ON c.user_id = u.id
LEFT JOIN favorites f ON f.user_id = u.id
GROUP BY u.id, u.username;

-- ============================================================================
-- Indexes for Performance
-- ============================================================================

CREATE INDEX idx_view_history_composite ON view_history(user_id, content_type, content_id, created_at);
CREATE INDEX idx_ratings_composite ON ratings(content_type, content_id, rating);
CREATE INDEX idx_watch_parties_active ON watch_parties(status, scheduled_at) WHERE status = 'active' OR status = 'scheduled';
