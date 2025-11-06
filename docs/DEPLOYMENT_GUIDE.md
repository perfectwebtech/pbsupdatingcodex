# 🚀 IPTV Platform - Production Deployment Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Server Setup](#server-setup)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Deployment](#deployment)
6. [Post-Deployment](#post-deployment)
7. [Monitoring](#monitoring)
8. [Backup & Recovery](#backup--recovery)
9. [Troubleshooting](#troubleshooting)
10. [Scaling](#scaling)

---

## Prerequisites

### System Requirements

**Minimum Requirements:**
- CPU: 4 cores
- RAM: 16GB
- Storage: 500GB SSD
- OS: Ubuntu 20.04 LTS or later

**Recommended for Production:**
- CPU: 8+ cores
- RAM: 32GB+
- Storage: 1TB+ NVMe SSD
- OS: Ubuntu 22.04 LTS

### Software Requirements

- Docker 24.0+
- Docker Compose 2.20+
- Git 2.30+
- MySQL 8.0+ (via Docker)
- Redis 7.0+ (via Docker)
- Nginx (via Docker)

---

## Server Setup

### 1. Prepare Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y curl wget git vim htop

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installations
docker --version
docker-compose --version
```

### 2. Configure Firewall

```bash
# Allow SSH
sudo ufw allow 22/tcp

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable
```

### 3. Setup SSL Certificates

**Using Let's Encrypt:**

```bash
# Install certbot
sudo apt install -y certbot

# Generate certificates
sudo certbot certonly --standalone \
  -d iptv.example.com \
  -d api.iptv.example.com \
  -d admin.iptv.example.com \
  -d cdn.iptv.example.com \
  --agree-tos \
  --email admin@iptv.example.com

# Copy certificates
sudo mkdir -p /opt/iptv-platform/nginx/ssl
sudo cp /etc/letsencrypt/live/iptv.example.com/fullchain.pem /opt/iptv-platform/nginx/ssl/
sudo cp /etc/letsencrypt/live/iptv.example.com/privkey.pem /opt/iptv-platform/nginx/ssl/

# Setup auto-renewal
sudo systemctl enable certbot.timer
```

---

## Installation

### 1. Clone Repository

```bash
# Create application directory
sudo mkdir -p /opt/iptv-platform
sudo chown $USER:$USER /opt/iptv-platform
cd /opt/iptv-platform

# Clone repository
git clone https://github.com/your-org/iptv-platform.git .

# Checkout production branch
git checkout main
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit environment file
vim .env
```

**Required Environment Variables:**

```bash
# Database
MYSQL_ROOT_PASSWORD=your_secure_root_password
MYSQL_DATABASE=iptv_platform
MYSQL_USER=iptv_user
MYSQL_PASSWORD=your_secure_password

# Application
JWT_SECRET=your_jwt_secret_min_32_chars
APP_URL=https://iptv.example.com
API_URL=https://api.iptv.example.com

# CDN
CDN_BASE_URL=https://cdn.iptv.example.com

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password

# Payment
STRIPE_SECRET_KEY=sk_live_your_stripe_key
PAYPAL_CLIENT_ID=your_paypal_client_id

# Monitoring
GRAFANA_PASSWORD=your_grafana_password
```

---

## Configuration

### 1. Database Migration

```bash
# Start MySQL container
docker-compose up -d mysql

# Wait for MySQL to be ready
sleep 30

# Run migrations
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE} < migrations/001_initial_schema.sql
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE} < migrations/002_resellers_billing.sql
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE} < migrations/003_streaming_content.sql
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE} < migrations/004_billing_packages.sql
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE} < migrations/005_create_transcoding_jobs_table.sql
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE} < migrations/006_create_mobile_tables.sql
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE} < migrations/007_create_innovative_features_tables.sql

# Verify migrations
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} -e "SHOW TABLES FROM ${MYSQL_DATABASE};"
```

### 2. Create Admin User

```bash
# Access MySQL
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE}

# Create admin user
INSERT INTO users (email, username, password, full_name, role, is_active, created_at)
VALUES (
  'admin@iptv.example.com',
  'admin',
  '$2a$10$your_bcrypt_hashed_password',
  'Administrator',
  'admin',
  1,
  NOW()
);
```

**Generate password hash:**

```bash
# Using htpasswd
htpasswd -bnBC 10 "" your_password | tr -d ':\n'

# Or using Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'your_password', bcrypt.gensalt()).decode())"
```

---

## Deployment

### Automated Deployment

```bash
# Run deployment script
./scripts/deploy.sh production
```

### Manual Deployment

```bash
# 1. Build images
docker-compose build --pull

# 2. Start services
docker-compose up -d

# 3. Check health
docker-compose ps
docker-compose logs -f

# 4. Run tests
./tests/quick-test.sh http://localhost:8080
```

---

## Post-Deployment

### 1. Verify Services

```bash
# Check all services are running
docker-compose ps

# Expected output:
# NAME                      STATUS              PORTS
# iptv_mysql               Up (healthy)        3306/tcp
# iptv_redis               Up (healthy)        6379/tcp
# iptv_streaming_gateway   Up (healthy)        8080/tcp
# iptv_admin_dashboard     Up (healthy)        80/tcp
# iptv_nginx               Up (healthy)        80/tcp, 443/tcp
# iptv_prometheus          Up                  9090/tcp
# iptv_grafana             Up                  3001/tcp
```

### 2. Access Admin Dashboard

1. Open browser: `https://admin.iptv.example.com`
2. Login with admin credentials
3. Complete initial setup wizard

### 3. Configure CDN (Optional)

**Using Cloudflare:**

1. Add domain to Cloudflare
2. Update DNS records:
   ```
   A     @              YOUR_SERVER_IP
   A     api            YOUR_SERVER_IP
   A     admin          YOUR_SERVER_IP
   A     cdn            YOUR_SERVER_IP
   ```
3. Enable SSL/TLS (Full)
4. Configure Page Rules for caching

---

## Monitoring

### Access Monitoring Dashboards

- **Prometheus**: http://server-ip:9090
- **Grafana**: http://server-ip:3001 (admin/password from .env)

### Setup Grafana Dashboards

1. Login to Grafana
2. Add Prometheus data source
3. Import dashboards from `monitoring/grafana/dashboards/`

### Key Metrics to Monitor

- **System**: CPU, Memory, Disk, Network
- **Application**: Request rate, Response time, Error rate
- **Database**: Connections, Queries/sec, Slow queries
- **Transcoding**: Jobs queued, Processing time, Worker utilization
- **Business**: Active users, Streams, Revenue

### Alerts Configuration

Configure alerts in `monitoring/prometheus-alerts.yml`:

```yaml
groups:
  - name: iptv_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"

      - alert: DatabaseDown
        expr: up{job="mysql"} == 0
        for: 1m
        annotations:
          summary: "Database is down"
```

---

## Backup & Recovery

### Automated Backups

Backups run daily at 2 AM via cron:

```bash
# Add to crontab
0 2 * * * /opt/iptv-platform/scripts/backup.sh >> /var/log/iptv-backup.log 2>&1
```

### Manual Backup

```bash
# Run backup script
./scripts/backup.sh

# Backups are stored in: ./backups/
# Format: database_YYYYMMDD_HHMMSS.sql.gz
```

### Restore from Backup

```bash
# Stop services
docker-compose stop

# Restore database
gunzip < backups/database_20250106_020000.sql.gz | \
  docker-compose exec -T mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} ${MYSQL_DATABASE}

# Restore configuration
tar -xzf backups/config_20250106_020000.tar.gz

# Start services
docker-compose up -d
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs streaming-gateway

# Check configuration
docker-compose config

# Restart service
docker-compose restart streaming-gateway
```

### Database Connection Issues

```bash
# Test connection
docker-compose exec mysql mysql -u ${MYSQL_USER} -p${MYSQL_PASSWORD} -e "SELECT 1;"

# Check MySQL logs
docker-compose logs mysql

# Reset password
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} -e "
  ALTER USER '${MYSQL_USER}'@'%' IDENTIFIED BY '${MYSQL_PASSWORD}';
  FLUSH PRIVILEGES;
"
```

### High Memory Usage

```bash
# Check container stats
docker stats

# Limit container memory
# Add to docker-compose.yml:
services:
  streaming-gateway:
    deploy:
      resources:
        limits:
          memory: 2G
```

### Slow Performance

```bash
# Check system resources
htop

# Check database slow queries
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} -e "
  SELECT * FROM mysql.slow_log ORDER BY query_time DESC LIMIT 10;
"

# Enable query cache
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} -e "
  SET GLOBAL query_cache_size = 268435456;
  SET GLOBAL query_cache_type = 1;
"
```

---

## Scaling

### Horizontal Scaling

**Add More Backend Instances:**

```yaml
# docker-compose.yml
services:
  streaming-gateway:
    deploy:
      replicas: 3
```

**Setup Load Balancer:**

```nginx
upstream streaming_gateway {
    least_conn;
    server gateway-1:8080;
    server gateway-2:8080;
    server gateway-3:8080;
}
```

### Vertical Scaling

**Increase Resources:**

```yaml
# docker-compose.yml
services:
  streaming-gateway:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
```

### Database Scaling

**Setup Read Replicas:**

1. Configure MySQL replication
2. Split read/write operations in application
3. Use connection pools

**Enable Query Cache:**

```sql
SET GLOBAL query_cache_size = 268435456;
SET GLOBAL query_cache_type = 1;
```

---

## Maintenance

### Update Platform

```bash
# Pull latest code
git fetch origin
git checkout main
git pull origin main

# Backup before update
./scripts/backup.sh

# Deploy updates
./scripts/deploy.sh production
```

### Update Dependencies

```bash
# Update Docker images
docker-compose pull

# Rebuild with latest base images
docker-compose build --pull --no-cache
```

### Clean Up

```bash
# Remove unused Docker images
docker image prune -a -f

# Remove unused volumes
docker volume prune -f

# Clean old backups (older than 30 days)
find ./backups -name "*.sql.gz" -mtime +30 -delete
```

---

## Security Checklist

- [ ] SSL/TLS certificates configured
- [ ] Firewall rules configured
- [ ] Strong passwords set for all services
- [ ] SSH key authentication enabled
- [ ] Database exposed only to docker network
- [ ] Regular security updates applied
- [ ] Backup encryption enabled
- [ ] Rate limiting configured
- [ ] DDoS protection enabled (Cloudflare)
- [ ] Security headers configured in Nginx
- [ ] Log monitoring configured
- [ ] Intrusion detection setup

---

## Support

- **Documentation**: https://docs.iptv.example.com
- **Email**: support@iptv.example.com
- **Discord**: https://discord.gg/iptv-platform
- **GitHub Issues**: https://github.com/your-org/iptv-platform/issues

---

*Last Updated: January 2025*
*Version: 1.0*
