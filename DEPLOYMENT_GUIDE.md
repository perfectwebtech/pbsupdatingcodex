# 🚀 IPTV PLATFORM - COMPLETE DEPLOYMENT GUIDE

**Version:** 2.0 Enterprise Edition
**Architecture:** Microservices (Go, Rust, Deno, Python)
**Status:** Production Ready ✅

---

## 📋 TABLE OF CONTENTS

1. [System Requirements](#system-requirements)
2. [Pre-Installation Setup](#pre-installation-setup)
3. [Quick Start (5 Minutes)](#quick-start-5-minutes)
4. [Detailed Installation](#detailed-installation)
5. [Service Configuration](#service-configuration)
6. [Database Setup](#database-setup)
7. [Testing & Verification](#testing--verification)
8. [Production Deployment](#production-deployment)
9. [Monitoring & Maintenance](#monitoring--maintenance)
10. [Troubleshooting](#troubleshooting)

---

## 💻 SYSTEM REQUIREMENTS

### **Minimum Requirements (Development)**

```
CPU:     4 cores (2.0 GHz+)
RAM:     8 GB
Disk:    50 GB available
OS:      Ubuntu 20.04+, macOS 12+, Windows 11 (WSL2)
```

### **Recommended Requirements (Production)**

```
CPU:     8+ cores (3.0 GHz+)
RAM:     32 GB
Disk:    500 GB SSD
OS:      Ubuntu 22.04 LTS
```

### **Required Software**

```bash
# Check versions
docker --version          # >= 24.0
docker-compose --version  # >= 2.20
git --version            # >= 2.30

# Optional (for development)
go version               # >= 1.22
python3 --version        # >= 3.11
rustc --version          # >= 1.75
deno --version           # >= 1.40
```

---

## 🔧 PRE-INSTALLATION SETUP

### **1. Install Docker & Docker Compose**

#### **Ubuntu/Debian:**
```bash
# Remove old versions
sudo apt-get remove docker docker-engine docker.io containerd runc

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo apt-get install docker-compose-plugin

# Verify installation
docker --version
docker compose version
```

#### **macOS:**
```bash
# Download and install Docker Desktop
# https://www.docker.com/products/docker-desktop

# Or use Homebrew
brew install --cask docker

# Verify
docker --version
docker compose version
```

#### **Windows (WSL2):**
```powershell
# Install WSL2
wsl --install

# Download and install Docker Desktop for Windows
# https://www.docker.com/products/docker-desktop

# Enable WSL2 integration in Docker Desktop settings
```

### **2. Configure System Resources**

#### **Docker Desktop Settings:**
```yaml
Resources:
  CPUs: 4-8
  Memory: 8-16 GB
  Swap: 2 GB
  Disk: 100 GB
```

#### **Linux Kernel Parameters:**
```bash
# Edit /etc/sysctl.conf
sudo nano /etc/sysctl.conf

# Add these lines:
vm.max_map_count=262144
fs.file-max=65536
net.core.somaxconn=65535
net.ipv4.tcp_max_syn_backlog=8192

# Apply changes
sudo sysctl -p
```

### **3. Configure Firewall**

```bash
# Ubuntu/Debian
sudo ufw allow 80/tcp       # HTTP
sudo ufw allow 443/tcp      # HTTPS
sudo ufw allow 8080/tcp     # Auth Service
sudo ufw allow 8000/tcp     # Streaming Gateway
sudo ufw allow 8001/tcp     # WebSocket Service
sudo ufw allow 8002/tcp     # Transcoding Service
sudo ufw allow 8003/tcp     # ML Service

# Apply rules
sudo ufw reload
```

---

## ⚡ QUICK START (5 MINUTES)

Perfect for local development and testing!

```bash
# 1. Clone repository
git clone https://github.com/perfectwebtech/pbsupdatingcodex.git
cd pbsupdatingcodex/microservices

# 2. Start all services
docker compose up -d

# 3. Wait for services to be ready (30-60 seconds)
sleep 60

# 4. Run database migrations
cd database
chmod +x migrate.sh
./migrate.sh up
cd ..

# 5. Run automated tests
chmod +x test-services.sh
./test-services.sh

# 6. Open landing page
# Navigate to: http://localhost:80
# OR open: xdg-open http://localhost:80  # Linux
#          open http://localhost:80       # macOS

# 7. Login with test credentials
# Username: testuser
# Password: admin123
```

**Expected Output:**
```
✅ All services online
✅ Database migrations applied
✅ All tests passed
✅ Platform ready for use
```

---

## 📦 DETAILED INSTALLATION

### **Step 1: Clone and Prepare**

```bash
# Clone repository
git clone https://github.com/perfectwebtech/pbsupdatingcodex.git
cd pbsupdatingcodex

# Create environment configuration
cp microservices/.env.example microservices/.env

# Edit configuration
nano microservices/.env
```

### **Step 2: Environment Configuration**

Create `microservices/.env`:

```bash
# ====================================
# DATABASE CONFIGURATION
# ====================================
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=iptv_platform
POSTGRES_USER=iptv_user
POSTGRES_PASSWORD=change_this_in_production_123!

# ====================================
# REDIS CONFIGURATION
# ====================================
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=change_this_redis_password

# ====================================
# JWT CONFIGURATION
# ====================================
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production
JWT_ALGORITHM=HS256
JWT_EXPIRATION=3600

# ====================================
# SERVICE PORTS
# ====================================
AUTH_SERVICE_PORT=8080
AUTH_GRPC_PORT=50051
STREAMING_SERVICE_PORT=8000
WEBSOCKET_SERVICE_PORT=8001
TRANSCODING_SERVICE_PORT=8002
ML_SERVICE_PORT=8003

# ====================================
# CDN & STREAMING
# ====================================
CDN_BASE_URL=http://cdn.iptv-platform.com
HLS_SEGMENT_DURATION=10
HLS_PLAYLIST_SIZE=5

# ====================================
# TRANSCODING
# ====================================
FFMPEG_THREADS=4
TRANSCODE_QUEUE_SIZE=100
MAX_CONCURRENT_JOBS=5

# ====================================
# ML RECOMMENDATIONS
# ====================================
ML_MODEL_PATH=/models/recommendation_model.h5
ML_BATCH_SIZE=256
ML_UPDATE_INTERVAL=3600

# ====================================
# MONITORING
# ====================================
PROMETHEUS_ENABLED=true
METRICS_PORT=9090
LOG_LEVEL=info

# ====================================
# SECURITY
# ====================================
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60
CORS_ORIGINS=http://localhost:3000,http://localhost:80
SESSION_TIMEOUT=86400
```

### **Step 3: Build Docker Images**

```bash
cd microservices

# Build all images
docker compose build

# Or build individually
docker compose build auth-service
docker compose build streaming-gateway
docker compose build realtime-ws
docker compose build transcoding-service
docker compose build ml-recommendation

# Verify images
docker images | grep iptv
```

**Expected Output:**
```
iptv-platform/auth-service         latest   ...   2.1GB   ...
iptv-platform/streaming-gateway    latest   ...   1.8GB   ...
iptv-platform/realtime-ws          latest   ...   850MB   ...
iptv-platform/transcoding-service  latest   ...   3.2GB   ...
iptv-platform/ml-recommendation    latest   ...   2.8GB   ...
```

### **Step 4: Start Infrastructure Services**

```bash
# Start PostgreSQL and Redis first
docker compose up -d postgres redis

# Wait for databases to be ready
sleep 10

# Check database health
docker compose exec postgres pg_isready -U iptv_user
docker compose exec redis redis-cli ping
```

**Expected Output:**
```
/var/run/postgresql:5432 - accepting connections
PONG
```

### **Step 5: Run Database Migrations**

```bash
cd database

# Make script executable
chmod +x migrate.sh

# Check database connection
./migrate.sh test

# Run migrations
./migrate.sh up

# Verify migration status
./migrate.sh status
```

**Expected Output:**
```
✅ 001_initial_schema (applied at 2025-01-05 10:30:00)
✅ 002_seed_data (applied at 2025-01-05 10:30:05)

Database Schema:
  Tables:   15
  Users:    4
  Streams:  10
  Packages: 5
```

### **Step 6: Start Application Services**

```bash
cd ..

# Start all application services
docker compose up -d auth-service streaming-gateway realtime-ws transcoding-service ml-recommendation

# Check service status
docker compose ps

# Follow logs
docker compose logs -f
```

**Expected Output:**
```
NAME                          STATUS    PORTS
auth-service                  Up        0.0.0.0:8080->8080/tcp, 50051/tcp
streaming-gateway             Up        0.0.0.0:8000->8000/tcp
realtime-ws                   Up        0.0.0.0:8001->8001/tcp
transcoding-service           Up        0.0.0.0:8002->8002/tcp
ml-recommendation             Up        0.0.0.0:8003->8003/tcp
postgres                      Up        5432/tcp
redis                         Up        6379/tcp
```

### **Step 7: Start Load Balancer**

```bash
# Start Nginx load balancer
docker compose up -d nginx

# Verify Nginx configuration
docker compose exec nginx nginx -t

# Check all services
docker compose ps
```

---

## 🗄️ DATABASE SETUP

### **Manual Database Operations**

#### **Connect to PostgreSQL:**
```bash
docker compose exec postgres psql -U iptv_user -d iptv_platform
```

#### **Useful Queries:**
```sql
-- Check all tables
\dt

-- Count records
SELECT 'users' as table_name, COUNT(*) FROM users
UNION ALL SELECT 'streams', COUNT(*) FROM streams
UNION ALL SELECT 'packages', COUNT(*) FROM packages;

-- View test users
SELECT id, username, email, is_active, max_connections
FROM users
WHERE deleted_at IS NULL;

-- Check active streams
SELECT id, name, type, category_id, is_active
FROM streams
WHERE is_active = TRUE
LIMIT 10;

-- View packages
SELECT id, name, price, duration_days, max_connections, is_trial
FROM packages;

-- Exit
\q
```

#### **Database Backup:**
```bash
# Create backup
docker compose exec postgres pg_dump -U iptv_user iptv_platform > backup_$(date +%Y%m%d).sql

# Restore from backup
docker compose exec -T postgres psql -U iptv_user iptv_platform < backup_20250105.sql
```

#### **Reset Database:**
```bash
cd database

# Drop all tables and rerun migrations
./migrate.sh reset

# Or completely remove and recreate
docker compose down -v
docker compose up -d postgres redis
sleep 10
./migrate.sh up
```

---

## 🧪 TESTING & VERIFICATION

### **Automated Test Suite**

```bash
# Run comprehensive test suite
cd microservices
./test-services.sh

# Test specific service
AUTH_URL=http://localhost:8080 ./test-services.sh
```

### **Manual Service Testing**

#### **1. Authentication Service**

```bash
# Health check
curl http://localhost:8080/health

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"admin123"}' | jq

# Save token
export JWT_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"admin123"}' | jq -r '.data.access_token')

echo "Token: $JWT_TOKEN"

# Validate token
curl http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer $JWT_TOKEN" | jq
```

#### **2. Streaming Gateway**

```bash
# List streams
curl http://localhost:8000/api/v1/streams \
  -H "Authorization: Bearer $JWT_TOKEN" | jq

# Get stream details
curl http://localhost:8000/api/v1/streams/1 \
  -H "Authorization: Bearer $JWT_TOKEN" | jq

# Get stream URL
curl "http://localhost:8000/api/v1/streams/1/url?container=m3u8" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq

# List categories
curl http://localhost:8000/api/v1/categories \
  -H "Authorization: Bearer $JWT_TOKEN" | jq
```

#### **3. Transcoding Service**

```bash
# Create transcode job
curl -X POST http://localhost:8002/api/v1/transcode \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"input_file":"/media/test.mp4","preset":"1080p_h264"}' | jq

# Get job status
export JOB_ID="<job_id_from_previous_response>"
curl http://localhost:8002/api/v1/jobs/$JOB_ID/status \
  -H "Authorization: Bearer $JWT_TOKEN" | jq

# List all jobs
curl http://localhost:8002/api/v1/jobs \
  -H "Authorization: Bearer $JWT_TOKEN" | jq
```

#### **4. ML Recommendation Service**

```bash
# Get personalized recommendations
curl "http://localhost:8003/api/v1/recommendations/personalized?limit=5" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq

# Get similar content
curl "http://localhost:8003/api/v1/recommendations/similar/1?limit=5" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq

# Get trending
curl "http://localhost:8003/api/v1/recommendations/trending?limit=10" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq

# Submit feedback
curl -X POST "http://localhost:8003/api/v1/recommendations/feedback?stream_id=1&rating=4.5" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq
```

#### **5. WebSocket Service**

```bash
# Install wscat
npm install -g wscat

# Connect and test
wscat -c "ws://localhost:8001/ws?token=$JWT_TOKEN"

# Send messages (after connecting):
> {"type":"ping"}
> {"type":"subscribe","payload":{"channel":"stream:1"}}
> {"type":"presence","payload":{"streamId":1}}
```

### **Load Testing**

```bash
# Install k6
brew install k6  # macOS
# OR
sudo snap install k6  # Linux

# Create load test script
cat > load-test.js << 'EOF'
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
};

export default function () {
  let loginRes = http.post('http://localhost:8080/api/v1/auth/login',
    JSON.stringify({
      username: 'testuser',
      password: 'admin123',
    }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'login successful': (r) => r.status === 200,
    'has token': (r) => JSON.parse(r.body).data.access_token !== undefined,
  });
}
EOF

# Run load test
k6 run load-test.js
```

---

## 🌐 PRODUCTION DEPLOYMENT

### **Pre-Production Checklist**

- [ ] Change all default passwords in `.env`
- [ ] Generate new JWT secret (64+ characters)
- [ ] Configure SSL/TLS certificates
- [ ] Set up firewall rules
- [ ] Configure backup strategy
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Test disaster recovery procedures
- [ ] Enable rate limiting
- [ ] Configure CDN for static assets
- [ ] Set up CI/CD pipeline
- [ ] Perform security audit
- [ ] Load test with production-like data
- [ ] Document incident response procedures

### **Security Hardening**

```bash
# 1. Generate secure secrets
JWT_SECRET=$(openssl rand -base64 64)
REDIS_PASSWORD=$(openssl rand -base64 32)
POSTGRES_PASSWORD=$(openssl rand -base64 32)

# 2. Update .env file
# 3. Restart services

# 4. Enable HTTPS
# Edit docker-compose.yml nginx service:
# volumes:
#   - ./nginx/ssl:/etc/nginx/ssl
#   - ./nginx/nginx-ssl.conf:/etc/nginx/nginx.conf

# 5. Configure fail2ban
sudo apt-get install fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### **Kubernetes Deployment**

```bash
# Deploy to Kubernetes cluster
cd kubernetes

# Create namespace
kubectl create namespace iptv-platform

# Create secrets
kubectl create secret generic iptv-secrets \
  --from-env-file=../microservices/.env \
  -n iptv-platform

# Deploy infrastructure
kubectl apply -f postgres-deployment.yaml
kubectl apply -f redis-deployment.yaml

# Wait for databases
kubectl wait --for=condition=ready pod -l app=postgres -n iptv-platform --timeout=120s

# Run migrations
kubectl apply -f migration-job.yaml
kubectl wait --for=condition=complete job/database-migration -n iptv-platform --timeout=300s

# Deploy services
kubectl apply -f auth-service-deployment.yaml
kubectl apply -f streaming-gateway-deployment.yaml
kubectl apply -f realtime-ws-deployment.yaml
kubectl apply -f transcoding-service-deployment.yaml
kubectl apply -f ml-recommendation-deployment.yaml

# Deploy ingress
kubectl apply -f ingress.yaml

# Check status
kubectl get pods -n iptv-platform
kubectl get svc -n iptv-platform
```

---

## 📊 MONITORING & MAINTENANCE

### **Service Monitoring**

```bash
# Check service health
docker compose ps

# View logs
docker compose logs -f auth-service
docker compose logs -f --tail=100 streaming-gateway

# Resource usage
docker stats

# Service metrics
curl http://localhost:8080/metrics   # Auth Service (Prometheus format)
curl http://localhost:8001/metrics   # WebSocket metrics (JSON)
```

### **Database Monitoring**

```bash
# Connect to database
docker compose exec postgres psql -U iptv_user -d iptv_platform

# Check database size
SELECT pg_size_pretty(pg_database_size('iptv_platform'));

# Check table sizes
SELECT
  relname as table_name,
  pg_size_pretty(pg_total_relation_size(relid)) as total_size,
  pg_size_pretty(pg_relation_size(relid)) as data_size,
  pg_size_pretty(pg_total_relation_size(relid) - pg_relation_size(relid)) as external_size
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC
LIMIT 10;

# Check active connections
SELECT count(*) FROM pg_stat_activity;

# Check slow queries
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';
```

### **Backup Strategy**

```bash
# Automated daily backups
cat > /usr/local/bin/backup-iptv.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backups/iptv-platform"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Database backup
docker compose exec -T postgres pg_dump -U iptv_user iptv_platform | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# Keep only last 30 days
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +30 -delete

echo "Backup completed: db_$DATE.sql.gz"
EOF

chmod +x /usr/local/bin/backup-iptv.sh

# Add to crontab (daily at 2 AM)
(crontab -l 2>/dev/null; echo "0 2 * * * /usr/local/bin/backup-iptv.sh") | crontab -
```

### **Log Management**

```bash
# Configure log rotation
cat > /etc/logrotate.d/iptv-platform << 'EOF'
/var/lib/docker/containers/*/*.log {
  daily
  rotate 7
  compress
  delaycompress
  missingok
  notifempty
  copytruncate
}
EOF

# Test configuration
logrotate -d /etc/logrotate.d/iptv-platform
```

---

## 🔧 TROUBLESHOOTING

### **Services Not Starting**

```bash
# Check Docker daemon
sudo systemctl status docker

# Check logs
docker compose logs

# Rebuild specific service
docker compose build --no-cache auth-service
docker compose up -d auth-service

# Check resource usage
docker stats
free -h
df -h
```

### **Database Connection Issues**

```bash
# Check PostgreSQL is running
docker compose ps postgres

# Check PostgreSQL logs
docker compose logs postgres

# Test connection
docker compose exec postgres psql -U iptv_user -d iptv_platform -c "SELECT 1;"

# Restart PostgreSQL
docker compose restart postgres
```

### **Authentication Failures**

```bash
# Check auth service logs
docker compose logs auth-service | tail -50

# Verify test user exists
docker compose exec postgres psql -U iptv_user -d iptv_platform \
  -c "SELECT id, username, is_active FROM users WHERE username='testuser';"

# Check JWT_SECRET is set
docker compose exec auth-service env | grep JWT_SECRET
```

### **High Resource Usage**

```bash
# Check resource usage
docker stats --no-stream

# Limit service resources (edit docker-compose.yml)
services:
  auth-service:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G

# Restart with new limits
docker compose up -d
```

### **Port Conflicts**

```bash
# Check what's using a port
sudo lsof -i :8080
sudo netstat -tulpn | grep 8080

# Kill process
sudo kill -9 <PID>

# Or change port in docker-compose.yml
```

---

## 🎯 QUICK REFERENCE

### **Common Commands**

```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down

# Restart specific service
docker compose restart auth-service

# View logs
docker compose logs -f

# Check status
docker compose ps

# Run migrations
cd database && ./migrate.sh up

# Run tests
./test-services.sh

# Access landing page
open http://localhost:80
```

### **Test Credentials**

| Username     | Password  | Package    | Connections |
|--------------|-----------|------------|-------------|
| admin        | admin123  | Enterprise | 10          |
| testuser     | admin123  | Standard   | 2           |
| premium_user | admin123  | Premium    | 5           |
| trial_user   | admin123  | Free Trial | 1           |

### **Service Endpoints**

| Service          | Port | HTTP Endpoint                     | Health Check                  |
|------------------|------|-----------------------------------|-------------------------------|
| Auth Service     | 8080 | http://localhost:8080/api/v1/auth | http://localhost:8080/health  |
| Streaming        | 8000 | http://localhost:8000/api/v1      | http://localhost:8000/health  |
| WebSocket        | 8001 | ws://localhost:8001/ws            | http://localhost:8001/health  |
| Transcoding      | 8002 | http://localhost:8002/api/v1      | http://localhost:8002/health  |
| ML Recommendation| 8003 | http://localhost:8003/api/v1      | http://localhost:8003/health  |
| Nginx (Frontend) | 80   | http://localhost:80               | http://localhost:80           |

---

## 📞 SUPPORT & RESOURCES

### **Documentation**

- Architecture Overview: `ARCHITECTURE.md`
- Feature Checklist: `FEATURE_CHECKLIST.md`
- Testing Guide: `TESTING_GUIDE.md`
- API Documentation: `API_DOCUMENTATION.md`

### **Useful Links**

- Docker Documentation: https://docs.docker.com
- Kubernetes Documentation: https://kubernetes.io/docs
- Go Documentation: https://golang.org/doc
- Deno Documentation: https://deno.land/manual
- Rust Documentation: https://doc.rust-lang.org
- Python/FastAPI: https://fastapi.tiangolo.com

---

**Deployment Guide Complete!** 🎉

Your IPTV Platform is ready for deployment. Follow the steps above for successful installation and operation.
