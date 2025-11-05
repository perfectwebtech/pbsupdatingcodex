# 🚀 Quick Start Guide

Get your IPTV platform running in 5 minutes!

## Prerequisites

- Docker & Docker Compose installed
- 4GB RAM minimum
- 10GB disk space

## Installation Steps

### 1. Navigate to Project

```bash
cd /home/user/pbsupdatingcodex/laravel-iptv
```

### 2. Environment Setup

```bash
# Copy environment file
cp .env.example .env

# Edit critical variables
nano .env
```

**Required settings:**
```env
DB_DATABASE=iptv_platform
DB_USERNAME=iptv_user
DB_PASSWORD=your_secure_password_here

# Copy GeoIP database
GEOIP_DATABASE_PATH=/home/user/pbsupdatingcodex/GeoLite2.mmdb
```

### 3. Start Services

```bash
# Build and start all containers
docker-compose up -d --build

# Wait for services to be ready (30-60 seconds)
docker-compose ps
```

### 4. Initialize Application

```bash
# Install dependencies
docker-compose exec app composer install

# Generate keys
docker-compose exec app php artisan key:generate
docker-compose exec app php artisan jwt:secret

# Run migrations
docker-compose exec app php artisan migrate --seed

# Create admin user
docker-compose exec app php artisan user:create-admin
```

### 5. Verify Installation

```bash
# Check health
curl http://localhost:8000/api/health

# Expected response:
# {"status":"ok","timestamp":"..."}
```

## First API Call

### 1. Login

```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJ0eXAiOiJKV1QiLC...",
    "token_type": "bearer",
    "expires_in": 3600
  }
}
```

### 2. Get Streams

```bash
curl http://localhost:8000/api/v1/streams \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Common Commands

```bash
# View logs
docker-compose logs -f app

# Access container shell
docker-compose exec app sh

# Run artisan commands
docker-compose exec app php artisan cache:clear

# Stop services
docker-compose down

# Restart services
docker-compose restart

# View database
docker-compose exec postgres psql -U iptv_user iptv_platform
```

## Troubleshooting

### Services won't start

```bash
# Check logs
docker-compose logs

# Rebuild
docker-compose down -v
docker-compose up -d --build
```

### Database connection error

```bash
# Check postgres is running
docker-compose ps postgres

# Check credentials in .env match docker-compose.yml
```

### Permission errors

```bash
# Fix storage permissions
docker-compose exec app chown -R www-data:www-data storage bootstrap/cache
```

## Next Steps

1. **Configure Servers**: Add streaming servers via API or database
2. **Create Packages**: Set up channel packages
3. **Add Streams**: Import your stream sources
4. **Create Users**: Add subscriber accounts
5. **Test Streaming**: Verify HLS playback works

## Production Deployment

See [README.md](README.md) for full production deployment guide.

## Getting Help

- Check logs: `docker-compose logs -f`
- Documentation: [README.md](README.md)
- Issues: Open a GitHub issue
