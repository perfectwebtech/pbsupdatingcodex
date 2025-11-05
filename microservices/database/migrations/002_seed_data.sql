-- Migration: 002_seed_data
-- Description: Populate database with initial test data
-- Author: IPTV Platform Team
-- Date: 2025-01-05

BEGIN;

-- Insert packages
INSERT INTO packages (name, description, price, duration_days, max_connections, is_trial, is_active) VALUES
('Free Trial', '7-day free trial with limited content', 0.00, 7, 1, TRUE, TRUE),
('Basic', 'Basic package with 100+ channels', 9.99, 30, 1, FALSE, TRUE),
('Standard', 'Standard package with 300+ channels', 19.99, 30, 2, FALSE, TRUE),
('Premium', 'Premium package with 500+ channels and VOD', 39.99, 30, 5, FALSE, TRUE),
('Enterprise', 'Enterprise package with all content', 99.99, 30, 10, FALSE, TRUE)
ON CONFLICT DO NOTHING;

-- Insert categories
INSERT INTO categories (name, type, parent_id, sort_order, icon_url) VALUES
-- Live TV categories
('Live TV', 'live', NULL, 1, '/icons/live-tv.png'),
('Sports', 'live', 1, 2, '/icons/sports.png'),
('News', 'live', 1, 3, '/icons/news.png'),
('Entertainment', 'live', 1, 4, '/icons/entertainment.png'),
('Kids', 'live', 1, 5, '/icons/kids.png'),

-- VOD categories
('Movies', 'vod', NULL, 10, '/icons/movies.png'),
('Action', 'vod', 6, 11, '/icons/action.png'),
('Comedy', 'vod', 6, 12, '/icons/comedy.png'),
('Drama', 'vod', 6, 13, '/icons/drama.png'),
('Horror', 'vod', 6, 14, '/icons/horror.png'),
('Sci-Fi', 'vod', 6, 15, '/icons/scifi.png'),

-- Series categories
('TV Series', 'series', NULL, 20, '/icons/series.png'),
('Documentaries', 'series', NULL, 21, '/icons/documentaries.png')
ON CONFLICT DO NOTHING;

-- Insert sample streams (Live TV)
INSERT INTO streams (name, type, category_id, source_urls, icon_url, bitrate, is_active, metadata) VALUES
-- Sports channels
('ESPN HD', 'live', 2, ARRAY['http://stream1.cdn.com/espn', 'http://stream2.cdn.com/espn'], '/icons/espn.png', 5000000, TRUE, '{"language": "en", "country": "US"}'::jsonb),
('Fox Sports', 'live', 2, ARRAY['http://stream1.cdn.com/foxsports'], '/icons/fox-sports.png', 5000000, TRUE, '{"language": "en", "country": "US"}'::jsonb),
('Sky Sports', 'live', 2, ARRAY['http://stream1.cdn.com/skysports'], '/icons/sky-sports.png', 5000000, TRUE, '{"language": "en", "country": "UK"}'::jsonb),

-- News channels
('CNN', 'live', 3, ARRAY['http://stream1.cdn.com/cnn'], '/icons/cnn.png', 3000000, TRUE, '{"language": "en", "country": "US"}'::jsonb),
('BBC News', 'live', 3, ARRAY['http://stream1.cdn.com/bbc-news'], '/icons/bbc.png', 3000000, TRUE, '{"language": "en", "country": "UK"}'::jsonb),
('Al Jazeera', 'live', 3, ARRAY['http://stream1.cdn.com/aljazeera'], '/icons/aljazeera.png', 3000000, TRUE, '{"language": "en", "country": "QA"}'::jsonb),

-- Entertainment
('HBO', 'live', 4, ARRAY['http://stream1.cdn.com/hbo'], '/icons/hbo.png', 8000000, TRUE, '{"language": "en", "country": "US"}'::jsonb),
('Discovery Channel', 'live', 4, ARRAY['http://stream1.cdn.com/discovery'], '/icons/discovery.png', 5000000, TRUE, '{"language": "en", "country": "US"}'::jsonb),

-- Kids
('Disney Channel', 'live', 5, ARRAY['http://stream1.cdn.com/disney'], '/icons/disney.png', 5000000, TRUE, '{"language": "en", "country": "US"}'::jsonb),
('Cartoon Network', 'live', 5, ARRAY['http://stream1.cdn.com/cartoon-network'], '/icons/cartoon-network.png', 5000000, TRUE, '{"language": "en", "country": "US"}'::jsonb)
ON CONFLICT DO NOTHING;

-- Insert sample VOD content
INSERT INTO streams (name, type, category_id, direct_source, icon_url, duration, is_active, metadata) VALUES
-- Action movies
('The Dark Knight', 'vod', 7, '/media/movies/dark-knight.mp4', '/posters/dark-knight.jpg', 9120, TRUE, '{"year": 2008, "rating": "PG-13", "genre": "Action"}'::jsonb),
('Mad Max: Fury Road', 'vod', 7, '/media/movies/mad-max.mp4', '/posters/mad-max.jpg', 7200, TRUE, '{"year": 2015, "rating": "R", "genre": "Action"}'::jsonb),

-- Comedy movies
('The Grand Budapest Hotel', 'vod', 8, '/media/movies/budapest.mp4', '/posters/budapest.jpg', 5940, TRUE, '{"year": 2014, "rating": "R", "genre": "Comedy"}'::jsonb),
('Superbad', 'vod', 8, '/media/movies/superbad.mp4', '/posters/superbad.jpg', 6780, TRUE, '{"year": 2007, "rating": "R", "genre": "Comedy"}'::jsonb)
ON CONFLICT DO NOTHING;

-- Insert servers
INSERT INTO servers (name, domain, port, https, is_active, weight, max_connections, region) VALUES
('US-East-01', 'stream1.us-east.cdn.com', 443, TRUE, TRUE, 100, 10000, 'us-east-1'),
('US-West-01', 'stream1.us-west.cdn.com', 443, TRUE, TRUE, 100, 10000, 'us-west-1'),
('EU-West-01', 'stream1.eu-west.cdn.com', 443, TRUE, TRUE, 100, 10000, 'eu-west-1'),
('Asia-East-01', 'stream1.asia-east.cdn.com', 443, TRUE, TRUE, 100, 10000, 'asia-east-1')
ON CONFLICT DO NOTHING;

-- Link streams to servers
INSERT INTO stream_servers (stream_id, server_id, status)
SELECT s.id, srv.id, 'active'
FROM streams s
CROSS JOIN servers srv
WHERE s.type = 'live'
ON CONFLICT DO NOTHING;

-- Link packages to streams
-- Free trial: Only news channels
INSERT INTO package_streams (package_id, stream_id)
SELECT 1, s.id
FROM streams s
WHERE s.category_id = 3 -- News
ON CONFLICT DO NOTHING;

-- Basic: Live TV (limited)
INSERT INTO package_streams (package_id, stream_id)
SELECT 2, s.id
FROM streams s
WHERE s.type = 'live' AND s.id <= 5
ON CONFLICT DO NOTHING;

-- Standard: Most live TV
INSERT INTO package_streams (package_id, stream_id)
SELECT 3, s.id
FROM streams s
WHERE s.type = 'live'
ON CONFLICT DO NOTHING;

-- Premium: All live TV + VOD
INSERT INTO package_streams (package_id, stream_id)
SELECT 4, s.id
FROM streams s
ON CONFLICT DO NOTHING;

-- Enterprise: All content
INSERT INTO package_streams (package_id, stream_id)
SELECT 5, s.id
FROM streams s
ON CONFLICT DO NOTHING;

-- Create admin user
-- Password: admin123 (CHANGE IN PRODUCTION!)
-- Hash generated with: bcrypt.hash("admin123", 12)
INSERT INTO users (username, password, email, package_id, max_connections, is_active, admin_enabled, expires_at)
VALUES (
    'admin',
    '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYAO4yI8KeO',
    'admin@iptv-platform.com',
    5, -- Enterprise package
    10,
    TRUE,
    TRUE,
    NOW() + INTERVAL '1 year'
)
ON CONFLICT (username) DO NOTHING;

-- Create test users
INSERT INTO users (username, password, email, package_id, max_connections, is_trial, is_active, expires_at)
VALUES
    (
        'testuser',
        '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYAO4yI8KeO', -- password: admin123
        'test@iptv-platform.com',
        3, -- Standard package
        2,
        FALSE,
        TRUE,
        NOW() + INTERVAL '30 days'
    ),
    (
        'premium_user',
        '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYAO4yI8KeO', -- password: admin123
        'premium@iptv-platform.com',
        4, -- Premium package
        5,
        FALSE,
        TRUE,
        NOW() + INTERVAL '30 days'
    ),
    (
        'trial_user',
        '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYAO4yI8KeO', -- password: admin123
        'trial@iptv-platform.com',
        1, -- Free trial
        1,
        TRUE,
        TRUE,
        NOW() + INTERVAL '7 days'
    )
ON CONFLICT (username) DO NOTHING;

-- Insert sample user interactions (for ML training)
INSERT INTO user_interactions (user_id, stream_id, interaction_type, rating, watch_duration)
SELECT
    u.id,
    s.id,
    CASE
        WHEN random() < 0.7 THEN 'view'
        WHEN random() < 0.85 THEN 'like'
        ELSE 'rate'
    END,
    CASE
        WHEN random() < 0.5 THEN NULL
        ELSE (random() * 5)::DECIMAL(2,1)
    END,
    (random() * 3600)::INT
FROM users u
CROSS JOIN LATERAL (
    SELECT id FROM streams WHERE type = 'live' ORDER BY random() LIMIT 5
) s
WHERE u.username != 'admin'
ON CONFLICT DO NOTHING;

COMMIT;

-- Display summary
SELECT 'Seed data inserted successfully!' as status,
       (SELECT COUNT(*) FROM users) as users_count,
       (SELECT COUNT(*) FROM packages) as packages_count,
       (SELECT COUNT(*) FROM categories) as categories_count,
       (SELECT COUNT(*) FROM streams) as streams_count,
       (SELECT COUNT(*) FROM servers) as servers_count;
