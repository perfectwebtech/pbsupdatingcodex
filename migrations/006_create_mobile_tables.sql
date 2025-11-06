-- Migration: Create mobile app tables
-- Description: Tables for mobile device management, tokens, notifications, downloads, etc.
-- Version: 006
-- Date: 2025-11-06

-- ============================================================================
-- Devices Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS devices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(255) NOT NULL UNIQUE,
    device_name VARCHAR(255) NOT NULL,
    platform ENUM('ios', 'android') NOT NULL,
    os_version VARCHAR(50) NULL,
    app_version VARCHAR(50) NULL,
    last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_is_active (is_active),
    INDEX idx_last_active (last_active),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Mobile Tokens Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS mobile_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_user_device (user_id, device_id),
    INDEX idx_expires_at (expires_at),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Stream Tokens Table (for secure stream access)
-- ============================================================================
CREATE TABLE IF NOT EXISTS stream_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    stream_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_token (token),
    INDEX idx_stream_id (stream_id),
    INDEX idx_expires_at (expires_at),

    FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Favorites Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS favorites (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY unique_favorite (user_id, content_type, content_id),
    INDEX idx_user_id (user_id),
    INDEX idx_content (content_type, content_id),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Watchlist Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS watchlist (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY unique_watchlist (user_id, content_type, content_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Watch Progress Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS watch_progress (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    content_type ENUM('stream', 'movie', 'series', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    progress INT NOT NULL DEFAULT 0 COMMENT 'Percentage 0-100',
    current_time INT NOT NULL DEFAULT 0 COMMENT 'In seconds',
    duration INT NOT NULL DEFAULT 0 COMMENT 'Total duration in seconds',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY unique_progress (user_id, content_type, content_id),
    INDEX idx_user_id (user_id),
    INDEX idx_updated_at (updated_at),
    INDEX idx_progress (progress),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Push Tokens Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS push_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    token TEXT NOT NULL,
    platform ENUM('ios', 'android') NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_device (user_id, device_id),
    INDEX idx_user_id (user_id),
    INDEX idx_is_active (is_active),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Notification Settings Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS notification_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,
    settings JSON NOT NULL COMMENT 'Notification preferences',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Notifications Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    type VARCHAR(50) NOT NULL COMMENT 'new_content, live_event, subscription, system, etc.',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    image_url VARCHAR(500) NULL,
    action_url VARCHAR(500) NULL,
    data JSON NULL COMMENT 'Additional data',
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_is_read (is_read),
    INDEX idx_created_at (created_at),
    INDEX idx_type (type),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Downloads Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS downloads (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    content_type ENUM('movie', 'episode') NOT NULL,
    content_id BIGINT UNSIGNED NOT NULL,
    quality VARCHAR(20) NOT NULL DEFAULT 'hd',
    status ENUM('pending', 'downloading', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending',
    progress INT NOT NULL DEFAULT 0 COMMENT 'Percentage 0-100',
    file_size BIGINT DEFAULT 0 COMMENT 'In bytes',
    file_path VARCHAR(500) NULL,
    error_message TEXT NULL,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_device_id (device_id),
    INDEX idx_status (status),
    INDEX idx_content (content_type, content_id),
    INDEX idx_expires_at (expires_at),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- App Configuration Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS app_config (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    platform ENUM('ios', 'android', 'all') NOT NULL DEFAULT 'all',
    config JSON NOT NULL COMMENT 'App configuration settings',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_platform (platform),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- EPG Events Table (Electronic Program Guide)
-- ============================================================================
CREATE TABLE IF NOT EXISTS epg_events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    channel_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    category VARCHAR(100) NULL,
    icon VARCHAR(500) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_channel_id (channel_id),
    INDEX idx_time_range (start_time, end_time),
    INDEX idx_category (category),

    FOREIGN KEY (channel_id) REFERENCES streams(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- VOD Movies Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS vod_movies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    poster VARCHAR(500) NULL,
    backdrop VARCHAR(500) NULL,
    year INT NULL,
    duration INT NULL COMMENT 'In minutes',
    rating DECIMAL(3,1) DEFAULT 0.0,
    genres JSON NULL,
    description TEXT NULL,
    video_url TEXT NULL,
    category_id BIGINT UNSIGNED NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_title (title),
    INDEX idx_year (year),
    INDEX idx_rating (rating),
    INDEX idx_category_id (category_id),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at),

    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- VOD Series Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS vod_series (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    poster VARCHAR(500) NULL,
    backdrop VARCHAR(500) NULL,
    year INT NULL,
    rating DECIMAL(3,1) DEFAULT 0.0,
    genres JSON NULL,
    description TEXT NULL,
    season_count INT DEFAULT 0,
    episode_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_title (title),
    INDEX idx_year (year),
    INDEX idx_rating (rating),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- VOD Episodes Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS vod_episodes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    series_id BIGINT UNSIGNED NOT NULL,
    season_num INT NOT NULL,
    episode_num INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    thumbnail VARCHAR(500) NULL,
    duration INT NULL COMMENT 'In minutes',
    air_date DATE NULL,
    video_url TEXT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_episode (series_id, season_num, episode_num),
    INDEX idx_series_id (series_id),
    INDEX idx_season (season_num),
    INDEX idx_air_date (air_date),
    INDEX idx_is_active (is_active),

    FOREIGN KEY (series_id) REFERENCES vod_series(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Insert Sample Data
-- ============================================================================

-- Sample VOD Movies
INSERT INTO vod_movies (title, poster, backdrop, year, duration, rating, genres, description, video_url, is_active) VALUES
('The Matrix', '/posters/matrix.jpg', '/backdrops/matrix.jpg', 1999, 136, 8.7, '["Action", "Sci-Fi"]', 'A computer hacker learns about the true nature of reality.', 'https://cdn.example.com/movies/matrix.mp4', 1),
('Inception', '/posters/inception.jpg', '/backdrops/inception.jpg', 2010, 148, 8.8, '["Action", "Sci-Fi", "Thriller"]', 'A thief who steals corporate secrets through dream-sharing technology.', 'https://cdn.example.com/movies/inception.mp4', 1),
('The Dark Knight', '/posters/dark_knight.jpg', '/backdrops/dark_knight.jpg', 2008, 152, 9.0, '["Action", "Crime", "Drama"]', 'Batman faces the Joker in Gotham City.', 'https://cdn.example.com/movies/dark_knight.mp4', 1),
('Interstellar', '/posters/interstellar.jpg', '/backdrops/interstellar.jpg', 2014, 169, 8.6, '["Adventure", "Drama", "Sci-Fi"]', 'A team of explorers travel through a wormhole in space.', 'https://cdn.example.com/movies/interstellar.mp4', 1),
('Pulp Fiction', '/posters/pulp_fiction.jpg', '/backdrops/pulp_fiction.jpg', 1994, 154, 8.9, '["Crime", "Drama"]', 'The lives of two mob hitmen, a boxer, and more intertwine.', 'https://cdn.example.com/movies/pulp_fiction.mp4', 1);

-- Sample VOD Series
INSERT INTO vod_series (title, poster, backdrop, year, rating, genres, description, season_count, episode_count, is_active) VALUES
('Breaking Bad', '/posters/breaking_bad.jpg', '/backdrops/breaking_bad.jpg', 2008, 9.5, '["Crime", "Drama", "Thriller"]', 'A chemistry teacher turned methamphetamine manufacturer.', 5, 62, 1),
('Game of Thrones', '/posters/got.jpg', '/backdrops/got.jpg', 2011, 9.2, '["Action", "Adventure", "Drama"]', 'Nine noble families fight for control of the Iron Throne.', 8, 73, 1),
('Stranger Things', '/posters/stranger_things.jpg', '/backdrops/stranger_things.jpg', 2016, 8.7, '["Drama", "Fantasy", "Horror"]', 'A group of kids encounter supernatural forces in their town.', 4, 34, 1);

-- Sample Episodes for Breaking Bad
INSERT INTO vod_episodes (series_id, season_num, episode_num, title, description, thumbnail, duration, air_date, video_url, is_active) VALUES
(1, 1, 1, 'Pilot', 'Walter White, a chemistry teacher, learns he has terminal cancer.', '/thumbnails/bb_s01e01.jpg', 58, '2008-01-20', 'https://cdn.example.com/series/breaking_bad/s01e01.mp4', 1),
(1, 1, 2, 'Cat''s in the Bag...', 'Walt and Jesse face the aftermath of their first cook.', '/thumbnails/bb_s01e02.jpg', 48, '2008-01-27', 'https://cdn.example.com/series/breaking_bad/s01e02.mp4', 1),
(1, 1, 3, '...And the Bag''s in the River', 'Walt tries to decide what to do with Krazy-8.', '/thumbnails/bb_s01e03.jpg', 48, '2008-02-10', 'https://cdn.example.com/series/breaking_bad/s01e03.mp4', 1);

-- Sample EPG Events
INSERT INTO epg_events (channel_id, title, description, start_time, end_time, category, icon) VALUES
(1, 'Morning News', 'Daily morning news and weather updates', NOW(), DATE_ADD(NOW(), INTERVAL 2 HOUR), 'News', '/icons/news.png'),
(1, 'Sports Highlights', 'Best moments from yesterday''s matches', DATE_ADD(NOW(), INTERVAL 2 HOUR), DATE_ADD(NOW(), INTERVAL 3 HOUR), 'Sports', '/icons/sports.png'),
(2, 'Movie Marathon', 'Classic action movies back-to-back', NOW(), DATE_ADD(NOW(), INTERVAL 6 HOUR), 'Movies', '/icons/movies.png'),
(3, 'Live Concert', 'Rock music festival live broadcast', DATE_ADD(NOW(), INTERVAL 1 HOUR), DATE_ADD(NOW(), INTERVAL 4 HOUR), 'Music', '/icons/music.png'),
(5, 'Football Match', 'Premier League: Team A vs Team B', NOW() + INTERVAL 30 MINUTE, NOW() + INTERVAL 2 HOUR + INTERVAL 30 MINUTE, 'Sports', '/icons/football.png');

-- Sample App Config
INSERT INTO app_config (platform, config, is_active) VALUES
('all', '{
  "min_version": "1.0.0",
  "latest_version": "2.0.0",
  "force_update": false,
  "maintenance_mode": false,
  "features": {
    "offline_downloads": true,
    "push_notifications": true,
    "live_streaming": true,
    "vod": true,
    "series": true,
    "epg": true,
    "favorites": true,
    "continue_watching": true,
    "parental_control": true
  },
  "download_enabled": true,
  "max_downloads": 10,
  "download_expiry_days": 7,
  "stream_qualities": ["sd", "hd", "fhd", "uhd"],
  "support_email": "support@iptv.example.com",
  "support_phone": "+1-555-0123",
  "terms_url": "https://iptv.example.com/terms",
  "privacy_url": "https://iptv.example.com/privacy"
}', 1);

-- ============================================================================
-- Views and Triggers
-- ============================================================================

-- View for user statistics
CREATE OR REPLACE VIEW user_mobile_stats AS
SELECT
    u.id as user_id,
    u.username,
    COUNT(DISTINCT d.id) as device_count,
    COUNT(DISTINCT f.id) as favorite_count,
    COUNT(DISTINCT wp.id) as watching_count,
    COUNT(DISTINCT dl.id) as download_count,
    MAX(d.last_active) as last_active
FROM users u
LEFT JOIN devices d ON d.user_id = u.id AND d.is_active = 1
LEFT JOIN favorites f ON f.user_id = u.id
LEFT JOIN watch_progress wp ON wp.user_id = u.id AND wp.progress > 5 AND wp.progress < 95
LEFT JOIN downloads dl ON dl.device_id IN (SELECT device_id FROM devices WHERE user_id = u.id)
GROUP BY u.id, u.username;

-- Trigger to clean up expired tokens
DELIMITER //

CREATE TRIGGER before_delete_device
BEFORE DELETE ON devices
FOR EACH ROW
BEGIN
    DELETE FROM mobile_tokens WHERE device_id = OLD.device_id;
    DELETE FROM push_tokens WHERE device_id = OLD.device_id;
    DELETE FROM downloads WHERE device_id = OLD.device_id;
END//

-- Trigger to update series counts
CREATE TRIGGER after_episode_insert
AFTER INSERT ON vod_episodes
FOR EACH ROW
BEGIN
    UPDATE vod_series
    SET episode_count = (SELECT COUNT(*) FROM vod_episodes WHERE series_id = NEW.series_id)
    WHERE id = NEW.series_id;
END//

CREATE TRIGGER after_episode_delete
AFTER DELETE ON vod_episodes
FOR EACH ROW
BEGIN
    UPDATE vod_series
    SET episode_count = (SELECT COUNT(*) FROM vod_episodes WHERE series_id = OLD.series_id)
    WHERE id = OLD.series_id;
END//

DELIMITER ;

-- Clean up expired data (run periodically)
-- DELETE FROM stream_tokens WHERE expires_at < NOW();
-- DELETE FROM downloads WHERE status = 'completed' AND expires_at < NOW();
-- DELETE FROM mobile_tokens WHERE expires_at < NOW();
