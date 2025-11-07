#!/bin/bash

# ============================================================================
# IPTV Platform - Database Seeding Script
# ============================================================================
# Populates database with sample data for testing and demo purposes
# Usage: ./seed-database.sh [environment]
# ============================================================================

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
ENVIRONMENT=${1:-development}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IPTV Platform - Database Seeding${NC}"
echo -e "${BLUE}Environment: $ENVIRONMENT${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Load environment variables
if [ -f "$PROJECT_DIR/.env" ]; then
    set -a
    source "$PROJECT_DIR/.env"
    set +a
fi

# Function to execute SQL
execute_sql() {
    local sql="$1"
    docker-compose exec -T mysql mysql -u "${MYSQL_USER}" -p"${MYSQL_PASSWORD}" "${MYSQL_DATABASE}" -e "$sql"
}

# 1. Create Admin User
echo -e "${GREEN}>>> Creating admin user...${NC}"
ADMIN_PASSWORD_HASH='$2a$10$CwTycUXWue0Thq9StjUM0uJ8VBLj.RqkrJVLxQ6YxPRKC8Z5qhimy' # Admin@123

execute_sql "
INSERT IGNORE INTO users (id, email, username, password, full_name, role, subscription_type, max_devices, max_streams, is_active, created_at)
VALUES (1, 'admin@iptv.example.com', 'admin', '$ADMIN_PASSWORD_HASH', 'Administrator', 'admin', 'enterprise', 10, 5, 1, NOW());
"
echo "✓ Admin user created (email: admin@iptv.example.com, password: Admin@123)"

# 2. Create Sample Users
echo -e "${GREEN}>>> Creating sample users...${NC}"
execute_sql "
INSERT IGNORE INTO users (email, username, password, full_name, role, subscription_type, max_devices, max_streams, subscription_expiry, is_active, created_at)
VALUES 
('user1@example.com', 'user1', '$ADMIN_PASSWORD_HASH', 'John Doe', 'user', 'premium', 3, 2, DATE_ADD(NOW(), INTERVAL 1 YEAR), 1, NOW()),
('user2@example.com', 'user2', '$ADMIN_PASSWORD_HASH', 'Jane Smith', 'user', 'basic', 2, 1, DATE_ADD(NOW(), INTERVAL 1 YEAR), 1, NOW()),
('user3@example.com', 'user3', '$ADMIN_PASSWORD_HASH', 'Bob Johnson', 'user', 'premium', 3, 2, DATE_ADD(NOW(), INTERVAL 6 MONTH), 1, NOW()),
('reseller1@example.com', 'reseller1', '$ADMIN_PASSWORD_HASH', 'Reseller One', 'reseller', 'enterprise', 5, 3, DATE_ADD(NOW(), INTERVAL 1 YEAR), 1, NOW());
"
echo "✓ Sample users created (4 users)"

# 3. Create Categories
echo -e "${GREEN}>>> Creating categories...${NC}"
execute_sql "
INSERT IGNORE INTO categories (id, name, icon, description, sort_order, is_active, created_at)
VALUES 
(1, 'Sports', '⚽', 'Sports channels and events', 1, 1, NOW()),
(2, 'News', '📰', 'News and current affairs', 2, 1, NOW()),
(3, 'Movies', '🎬', 'Movie channels', 3, 1, NOW()),
(4, 'Entertainment', '🎭', 'Entertainment channels', 4, 1, NOW()),
(5, 'Kids', '🧒', 'Children programming', 5, 1, NOW()),
(6, 'Music', '🎵', 'Music channels', 6, 1, NOW()),
(7, 'Documentary', '🎓', 'Documentary channels', 7, 1, NOW()),
(8, 'Lifestyle', '🏡', 'Lifestyle channels', 8, 1, NOW());
"
echo "✓ Categories created (8 categories)"

# 4. Create Sample Streams
echo -e "${GREEN}>>> Creating sample streams...${NC}"
execute_sql "
INSERT IGNORE INTO streams (name, stream_url, stream_type, logo, category_id, is_live, qualities, epg_enabled, is_active, created_at)
VALUES 
('ESPN HD', 'http://example.com/espn.m3u8', 'hls', 'https://via.placeholder.com/150?text=ESPN', 1, 1, '[\"hd\",\"fhd\"]', 1, 1, NOW()),
('Fox Sports', 'http://example.com/foxsports.m3u8', 'hls', 'https://via.placeholder.com/150?text=Fox+Sports', 1, 1, '[\"hd\"]', 1, 1, NOW()),
('CNN International', 'http://example.com/cnn.m3u8', 'hls', 'https://via.placeholder.com/150?text=CNN', 2, 1, '[\"hd\",\"fhd\"]', 1, 1, NOW()),
('BBC News', 'http://example.com/bbc.m3u8', 'hls', 'https://via.placeholder.com/150?text=BBC', 2, 1, '[\"hd\"]', 1, 1, NOW()),
('HBO', 'http://example.com/hbo.m3u8', 'hls', 'https://via.placeholder.com/150?text=HBO', 3, 1, '[\"hd\",\"fhd\",\"uhd\"]', 1, 1, NOW()),
('Discovery Channel', 'http://example.com/discovery.m3u8', 'hls', 'https://via.placeholder.com/150?text=Discovery', 7, 1, '[\"hd\"]', 1, 1, NOW()),
('MTV', 'http://example.com/mtv.m3u8', 'hls', 'https://via.placeholder.com/150?text=MTV', 6, 1, '[\"hd\"]', 1, 1, NOW()),
('Cartoon Network', 'http://example.com/cartoon.m3u8', 'hls', 'https://via.placeholder.com/150?text=Cartoon', 5, 1, '[\"hd\"]', 1, 1, NOW());
"
echo "✓ Sample streams created (8 streams)"

# 5. Create Sample VOD Movies
echo -e "${GREEN}>>> Creating sample VOD movies...${NC}"
execute_sql "
INSERT IGNORE INTO vod_movies (title, poster, backdrop, year, duration, rating, genres, description, video_url, category_id, is_active, created_at)
VALUES 
('The Matrix', 'https://via.placeholder.com/300x450?text=Matrix', 'https://via.placeholder.com/1920x1080?text=Matrix+BG', 1999, 136, 8.7, '[\"Action\",\"Sci-Fi\"]', 'A computer hacker learns about the true nature of reality and his role in the war against its controllers.', 'http://example.com/matrix.mp4', 3, 1, NOW()),
('Inception', 'https://via.placeholder.com/300x450?text=Inception', 'https://via.placeholder.com/1920x1080?text=Inception+BG', 2010, 148, 8.8, '[\"Action\",\"Sci-Fi\",\"Thriller\"]', 'A thief who steals corporate secrets through dream-sharing technology.', 'http://example.com/inception.mp4', 3, 1, NOW()),
('The Dark Knight', 'https://via.placeholder.com/300x450?text=DarkKnight', 'https://via.placeholder.com/1920x1080?text=DK+BG', 2008, 152, 9.0, '[\"Action\",\"Crime\",\"Drama\"]', 'When the menace known as the Joker wreaks havoc on Gotham.', 'http://example.com/darkknight.mp4', 3, 1, NOW()),
('Interstellar', 'https://via.placeholder.com/300x450?text=Interstellar', 'https://via.placeholder.com/1920x1080?text=Interstellar+BG', 2014, 169, 8.6, '[\"Adventure\",\"Drama\",\"Sci-Fi\"]', 'A team of explorers travel through a wormhole in space.', 'http://example.com/interstellar.mp4', 3, 1, NOW()),
('Pulp Fiction', 'https://via.placeholder.com/300x450?text=PulpFiction', 'https://via.placeholder.com/1920x1080?text=PF+BG', 1994, 154, 8.9, '[\"Crime\",\"Drama\"]', 'The lives of two mob hitmen, a boxer, and more intertwine.', 'http://example.com/pulpfiction.mp4', 3, 1, NOW());
"
echo "✓ Sample movies created (5 movies)"

# 6. Create Sample Series
echo -e "${GREEN}>>> Creating sample series...${NC}"
execute_sql "
INSERT IGNORE INTO vod_series (id, title, poster, backdrop, year, rating, genres, description, season_count, episode_count, is_active, created_at)
VALUES 
(1, 'Breaking Bad', 'https://via.placeholder.com/300x450?text=BreakingBad', 'https://via.placeholder.com/1920x1080?text=BB+BG', 2008, 9.5, '[\"Crime\",\"Drama\",\"Thriller\"]', 'A chemistry teacher turned methamphetamine manufacturer.', 5, 62, 1, NOW()),
(2, 'Game of Thrones', 'https://via.placeholder.com/300x450?text=GoT', 'https://via.placeholder.com/1920x1080?text=GoT+BG', 2011, 9.2, '[\"Action\",\"Adventure\",\"Drama\"]', 'Nine noble families fight for control of the Iron Throne.', 8, 73, 1, NOW()),
(3, 'Stranger Things', 'https://via.placeholder.com/300x450?text=StrangerThings', 'https://via.placeholder.com/1920x1080?text=ST+BG', 2016, 8.7, '[\"Drama\",\"Fantasy\",\"Horror\"]', 'A group of kids encounter supernatural forces.', 4, 34, 1, NOW());
"

# Add episodes for Breaking Bad Season 1
execute_sql "
INSERT IGNORE INTO vod_episodes (series_id, season_num, episode_num, title, description, thumbnail, duration, air_date, video_url, is_active, created_at)
VALUES 
(1, 1, 1, 'Pilot', 'Walter White learns he has terminal cancer.', 'https://via.placeholder.com/300x169?text=BB+S01E01', 58, '2008-01-20', 'http://example.com/bb_s01e01.mp4', 1, NOW()),
(1, 1, 2, 'Cat in the Bag...', 'Walt and Jesse face the aftermath of their first cook.', 'https://via.placeholder.com/300x169?text=BB+S01E02', 48, '2008-01-27', 'http://example.com/bb_s01e02.mp4', 1, NOW()),
(1, 1, 3, '...And the Bag is in the River', 'Walt tries to decide what to do with Krazy-8.', 'https://via.placeholder.com/300x169?text=BB+S01E03', 48, '2008-02-10', 'http://example.com/bb_s01e03.mp4', 1, NOW());
"
echo "✓ Sample series created (3 series + episodes)"

# 7. Create Subscription Packages
echo -e "${GREEN}>>> Creating subscription packages...${NC}"
execute_sql "
INSERT IGNORE INTO packages (name, description, price, billing_cycle, max_devices, max_streams, features, is_active, is_featured, sort_order, created_at)
VALUES 
('Basic', 'Perfect for individuals', 9.99, 'monthly', 1, 1, '[\"HD Streaming\",\"1 Device\",\"Email Support\"]', 1, 0, 1, NOW()),
('Premium', 'Best for families', 19.99, 'monthly', 3, 2, '[\"Full HD Streaming\",\"3 Devices\",\"Priority Support\",\"Offline Downloads\"]', 1, 1, 2, NOW()),
('Enterprise', 'For businesses', 49.99, 'monthly', 10, 5, '[\"4K Streaming\",\"10 Devices\",\"24/7 Support\",\"Offline Downloads\",\"Advanced Analytics\"]', 1, 0, 3, NOW());
"
echo "✓ Subscription packages created (3 packages)"

# 8. Create Sample View History
echo -e "${GREEN}>>> Creating sample view history...${NC}"
execute_sql "
INSERT IGNORE INTO view_history (user_id, content_type, content_id, category_id, quality, watch_duration, completion_percentage, device_type, created_at)
VALUES 
(2, 'movie', 1, 3, 'hd', 3600, 80, 'mobile', NOW() - INTERVAL 1 DAY),
(2, 'stream', 1, 1, 'fhd', 1200, 100, 'tv', NOW() - INTERVAL 2 HOUR),
(3, 'movie', 1, 3, 'hd', 4500, 100, 'web', NOW() - INTERVAL 3 DAY),
(3, 'series', 1, 4, 'hd', 2400, 95, 'mobile', NOW() - INTERVAL 1 DAY),
(2, 'stream', 2, 1, 'uhd', 1800, 100, 'tv', NOW() - INTERVAL 4 HOUR);
"
echo "✓ View history created"

# 9. Create Sample Ratings
echo -e "${GREEN}>>> Creating sample ratings...${NC}"
execute_sql "
INSERT IGNORE INTO ratings (user_id, content_type, content_id, rating, review, created_at)
VALUES 
(2, 'movie', 1, 4.5, 'Great movie! Highly recommended.', NOW() - INTERVAL 1 DAY),
(3, 'movie', 1, 5.0, 'Masterpiece!', NOW() - INTERVAL 2 DAY),
(2, 'stream', 1, 4.0, 'Good quality stream', NOW() - INTERVAL 3 DAY),
(3, 'series', 1, 4.8, 'Best series ever!', NOW() - INTERVAL 4 DAY);
"
echo "✓ Ratings created"

# 10. Create Sample EPG Data
echo -e "${GREEN}>>> Creating sample EPG data...${NC}"
execute_sql "
INSERT IGNORE INTO epg_events (channel_id, title, description, start_time, end_time, category, icon, created_at)
VALUES 
(1, 'Live Football Match', 'Premier League: Team A vs Team B', NOW(), DATE_ADD(NOW(), INTERVAL 2 HOUR), 'Sports', 'https://via.placeholder.com/50?text=⚽', NOW()),
(1, 'Sports Highlights', 'Best moments from today games', DATE_ADD(NOW(), INTERVAL 2 HOUR), DATE_ADD(NOW(), INTERVAL 3 HOUR), 'Sports', 'https://via.placeholder.com/50?text=⚽', NOW()),
(2, 'Breaking News', 'Latest news from around the world', NOW(), DATE_ADD(NOW(), INTERVAL 1 HOUR), 'News', 'https://via.placeholder.com/50?text=📰', NOW()),
(3, 'Movie Marathon', 'Classic action movies', NOW(), DATE_ADD(NOW(), INTERVAL 6 HOUR), 'Movies', 'https://via.placeholder.com/50?text=🎬', NOW());
"
echo "✓ EPG data created"

# Summary
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Database Seeding Completed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Summary:"
echo "- 1 Admin user (admin@iptv.example.com / Admin@123)"
echo "- 4 Sample users (user1-3, reseller1)"
echo "- 8 Categories"
echo "- 8 Live streams"
echo "- 5 VOD movies"
echo "- 3 Series with episodes"
echo "- 3 Subscription packages"
echo "- Sample view history"
echo "- Sample ratings"
echo "- Sample EPG data"
echo ""
echo "You can now login and test the platform with sample data!"
echo ""
