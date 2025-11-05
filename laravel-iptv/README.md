# 🚀 Enterprise IPTV Platform (Laravel Edition)

A modern, scalable, and secure IPTV streaming platform built with Laravel 11, designed for enterprise-level performance and reliability.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Deployment](#deployment)
- [Testing](#testing)
- [Migration from Legacy](#migration-from-legacy)
- [Contributing](#contributing)

## ✨ Features

### Core Features
- **Multi-Format Streaming**: HLS, MPEG-TS, RTMP support
- **JWT Authentication**: Secure token-based authentication
- **User Management**: Comprehensive user and subscription management
- **Package System**: Flexible package and channel grouping
- **Load Balancing**: Intelligent server selection
- **DVR/Archive**: Time-shifted viewing capability
- **EPG Integration**: Electronic Program Guide support
- **Series Management**: Full TV series organization

### Security
- ✅ JWT token authentication
- ✅ IP whitelisting
- ✅ Geographic restrictions (GeoIP)
- ✅ Rate limiting
- ✅ User agent validation
- ✅ Connection limits
- ✅ ISP locking
- ✅ Comprehensive logging

### Performance
- ✅ Redis caching
- ✅ Database query optimization
- ✅ CDN integration ready
- ✅ Horizontal scaling support
- ✅ Queue system for heavy tasks
- ✅ Connection pooling

### Monitoring & Observability
- ✅ Structured logging (Monolog)
- ✅ Activity tracking
- ✅ Real-time analytics
- ✅ Health check endpoints
- ✅ Performance metrics

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway                           │
│                    (Nginx Load Balancer)                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
           ┌───────────┴───────────┐
           │                       │
    ┌──────▼──────┐        ┌──────▼──────┐
    │  Laravel    │        │  Laravel    │
    │  App (PHP)  │        │  App (PHP)  │
    └──────┬──────┘        └──────┬──────┘
           │                       │
           └───────────┬───────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
  ┌─────▼─────┐  ┌────▼────┐  ┌─────▼─────┐
  │ PostgreSQL │  │  Redis  │  │  FFmpeg   │
  │  Database  │  │  Cache  │  │ Streaming │
  └────────────┘  └─────────┘  └───────────┘
```

### Directory Structure

```
laravel-iptv/
├── app/
│   ├── Http/
│   │   ├── Controllers/
│   │   │   ├── Api/
│   │   │   │   ├── AuthController.php
│   │   │   │   └── StreamController.php
│   │   │   └── Admin/
│   │   ├── Middleware/
│   │   │   ├── JwtAuthenticate.php
│   │   │   ├── IpWhitelist.php
│   │   │   ├── GeoBlock.php
│   │   │   └── CheckSubscription.php
│   │   └── Requests/
│   ├── Models/
│   │   ├── User.php
│   │   ├── Stream.php
│   │   ├── Package.php
│   │   ├── Server.php
│   │   └── UserSession.php
│   ├── Services/
│   │   ├── AuthenticationService.php
│   │   ├── StreamingService.php
│   │   ├── GeoIPService.php
│   │   ├── LoadBalancerService.php
│   │   ├── FFmpegService.php
│   │   └── LoggingService.php
│   └── Exceptions/
├── config/
│   ├── database.php
│   ├── cache.php
│   └── streaming.php
├── database/
│   ├── migrations/
│   └── seeders/
├── routes/
│   ├── api.php
│   ├── web.php
│   └── console.php
├── docker/
│   └── nginx/
└── tests/
```

## 📦 Requirements

- **PHP**: 8.2 or higher
- **Composer**: 2.x
- **PostgreSQL**: 16+
- **Redis**: 7+
- **FFmpeg**: 4.x+
- **Docker**: 20+ (optional, recommended)
- **Node.js**: 20+ (for frontend development)

## 🚀 Installation

### Option 1: Docker (Recommended)

```bash
# Clone the repository
cd /home/user/pbsupdatingcodex/laravel-iptv

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
nano .env

# Build and start containers
docker-compose up -d --build

# Install dependencies
docker-compose exec app composer install

# Generate application key
docker-compose exec app php artisan key:generate

# Generate JWT secret
docker-compose exec app php artisan jwt:secret

# Run migrations
docker-compose exec app php artisan migrate --seed

# Access the application
# API: http://localhost:8000
# Admin: http://localhost:8000/admin
```

### Option 2: Manual Installation

```bash
# Install PHP dependencies
composer install

# Copy environment file
cp .env.example .env

# Generate keys
php artisan key:generate
php artisan jwt:secret

# Set up database
# Create PostgreSQL database first, then:
php artisan migrate --seed

# Start queue worker
php artisan queue:work &

# Start scheduler (add to crontab)
* * * * * cd /path-to-project && php artisan schedule:run >> /dev/null 2>&1

# Start development server
php artisan serve
```

## ⚙️ Configuration

### Environment Variables

```env
# Application
APP_NAME="IPTV Platform"
APP_ENV=production
APP_DEBUG=false
APP_URL=https://your-domain.com

# Database
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=iptv_platform
DB_USERNAME=iptv_user
DB_PASSWORD=your_secure_password

# Redis
REDIS_HOST=127.0.0.1
REDIS_PORT=6379

# JWT
JWT_SECRET=your_jwt_secret
JWT_TTL=60

# Streaming
FFMPEG_PATH=/usr/bin/ffmpeg
FFPROBE_PATH=/usr/bin/ffprobe
STREAMS_PATH=/var/streams
ARCHIVES_PATH=/var/archives

# GeoIP
GEOIP_DATABASE_PATH=/path/to/GeoLite2.mmdb
```

### Streaming Configuration

Edit `config/streaming.php`:

```php
return [
    'stream' => [
        'segment_duration' => 10,
        'buffer_size' => 8192,
        'prebuffer_segments' => 3,
    ],
    'transcoding' => [
        'enabled' => true,
        'max_concurrent' => 5,
    ],
    'load_balancing' => [
        'enabled' => true,
        'strategy' => 'least_connections', // or 'round_robin', 'weighted'
    ],
];
```

## 📖 API Documentation

### Authentication

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user123",
  "password": "password"
}

Response:
{
  "success": true,
  "data": {
    "access_token": "eyJ0eXAiOiJKV1QiLCJhbGc...",
    "token_type": "bearer",
    "expires_in": 3600,
    "user": {
      "id": 1,
      "username": "user123",
      "max_connections": 3,
      "expires_at": "2025-12-31T23:59:59Z"
    }
  }
}
```

#### Get User Info
```http
GET /api/v1/auth/me
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": {
    "id": 1,
    "username": "user123",
    "max_connections": 3,
    "active_connections": 1,
    "expires_at": "2025-12-31T23:59:59Z"
  }
}
```

### Streaming

#### Get Streams
```http
GET /api/v1/streams?type=live&category_id=5
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": [
    {
      "id": 101,
      "name": "HBO HD",
      "type": "live",
      "icon_url": "https://cdn.example.com/icons/hbo.png",
      "category": {
        "id": 5,
        "name": "Movies"
      }
    }
  ]
}
```

#### Get Stream URL
```http
GET /api/v1/streams/101/url?container=m3u8
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": {
    "type": "stream",
    "url": "https://server1.example.com/stream/101.m3u8",
    "session_id": 12345,
    "container": "m3u8"
  }
}
```

#### Get HLS Playlist
```http
GET /api/v1/stream/101/playlist.m3u8
Authorization: Bearer {token}

Response: (M3U8 playlist content)
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
/api/v1/stream/101/segment/101_0.ts?token=...
#EXTINF:10.0,
/api/v1/stream/101/segment/101_1.ts?token=...
```

### Categories

```http
GET /api/v1/categories?type=live
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Sports",
      "type": "live",
      "streams_count": 25
    }
  ]
}
```

## 🧪 Testing

```bash
# Run all tests
php artisan test

# Run with coverage
php artisan test --coverage

# Run specific test
php artisan test --filter UserAuthenticationTest

# Run feature tests only
php artisan test --testsuite=Feature

# Run unit tests only
php artisan test --testsuite=Unit
```

## 🔄 Migration from Legacy System

### Step 1: Data Migration

```bash
# Run migration command (will be created)
php artisan migrate:legacy-data

# This will:
# - Import users from old database
# - Import streams and categories
# - Import packages and subscriptions
# - Import active sessions
```

### Step 2: Parallel Running

```nginx
# Run both systems in parallel
# Route new users to Laravel
# Keep existing users on old system temporarily

location /api/v1 {
    proxy_pass http://laravel-app:8000;
}

location /api {
    proxy_pass http://legacy-app:8080;
}
```

### Step 3: Gradual Migration

```bash
# Migrate users in batches
php artisan migrate:users --batch=100

# Monitor both systems
tail -f storage/logs/laravel.log
```

## 🚢 Deployment

### Production Checklist

- [ ] Set `APP_ENV=production`
- [ ] Set `APP_DEBUG=false`
- [ ] Configure proper database credentials
- [ ] Set up SSL certificates
- [ ] Configure Redis for cache and sessions
- [ ] Set up queue workers
- [ ] Configure cron jobs
- [ ] Set up monitoring (Sentry, etc.)
- [ ] Configure backups
- [ ] Set up CDN for streams
- [ ] Configure load balancer
- [ ] Test failover scenarios

### Nginx Production Config

```nginx
server {
    listen 80;
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    root /var/www/html/public;
    index index.php;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=100r/s;
    limit_req zone=api burst=20 nodelay;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:/var/run/php/php8.3-fpm.sock;
        fastcgi_index index.php;
        include fastcgi_params;
    }
}
```

### Supervisor Config

```ini
[program:iptv-queue]
process_name=%(program_name)s_%(process_num)02d
command=php /var/www/html/artisan queue:work --sleep=3 --tries=3
autostart=true
autorestart=true
numprocs=4
user=www-data

[program:iptv-scheduler]
command=sh -c "while true; do php /var/www/html/artisan schedule:run; sleep 60; done"
autostart=true
autorestart=true
user=www-data
```

## 📊 Monitoring

### Health Check

```bash
curl http://localhost:8000/api/health

Response:
{
  "status": "ok",
  "timestamp": "2025-01-01T12:00:00Z"
}
```

### Metrics Endpoint

```bash
# To be implemented with Prometheus
GET /metrics
```

## 🔧 Maintenance

### Clear Cache

```bash
php artisan cache:clear
php artisan config:clear
php artisan route:clear
php artisan view:clear
```

### Database Maintenance

```bash
# Backup database
pg_dump iptv_platform > backup.sql

# Optimize database
php artisan db:optimize

# Clean old sessions
php artisan sessions:cleanup
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

This project is proprietary software. All rights reserved.

## 📞 Support

For support, email support@example.com or open an issue on GitHub.

## 🎯 Roadmap

- [ ] Admin panel (React/Vue)
- [ ] Mobile app integration (Flutter)
- [ ] Advanced analytics dashboard
- [ ] Multi-tenancy support
- [ ] Kubernetes deployment manifests
- [ ] AI-powered recommendations
- [ ] Live transcoding
- [ ] WebRTC support

---

**Built with ❤️ using Laravel 11 | Version 2.0.0**
