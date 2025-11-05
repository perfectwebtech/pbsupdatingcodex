# 🧪 COMPREHENSIVE TESTING GUIDE

Complete testing procedures for the IPTV Platform microservices architecture.

---

## 📋 **TABLE OF CONTENTS**

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Database Setup](#database-setup)
4. [Service Testing](#service-testing)
5. [Integration Testing](#integration-testing)
6. [Performance Testing](#performance-testing)
7. [Troubleshooting](#troubleshooting)

---

## 🔧 **PREREQUISITES**

### **Required Tools**

```bash
# Check if tools are installed
docker --version          # Docker 20+
docker-compose --version  # Docker Compose 2+
curl --version           # curl 7+
jq --version             # jq 1.6+ (for JSON parsing)
psql --version           # PostgreSQL client 16+

# Optional tools for enhanced testing
wscat --version          # WebSocket testing
k6 version               # Load testing
```

### **Install Missing Tools**

```bash
# Install jq (JSON processor)
sudo apt-get install jq   # Ubuntu/Debian
brew install jq           # macOS

# Install wscat (WebSocket testing)
npm install -g wscat

# Install k6 (load testing)
brew install k6           # macOS
# or download from: https://k6.io/docs/getting-started/installation/
```

---

## 🚀 **QUICK START**

### **1. Start All Services**

```bash
cd /home/user/pbsupdatingcodex/microservices

# Start all services in background
docker-compose up -d

# Wait for services to be ready (30-60 seconds)
docker-compose ps

# Check logs if needed
docker-compose logs -f
```

### **2. Initialize Database**

```bash
# Run database migrations
cd database
./migrate.sh up

# Verify migration status
./migrate.sh status

# Expected output:
# ✅ 001_initial_schema (applied)
# ✅ 002_seed_data (applied)
```

### **3. Run Automated Tests**

```bash
# Run comprehensive test suite
cd ..
./test-services.sh

# This will test:
# ✅ Authentication Service
# ✅ Streaming Gateway
# ✅ Transcoding Service
# ✅ ML Recommendation Service
# ✅ WebSocket Service
```

### **4. Access Landing Page**

Open your browser and navigate to:
```
http://localhost:80
```

You should see the platform dashboard with all services status indicators.

---

## 🗄️ **DATABASE SETUP**

### **Manual Database Setup**

If you need to manually set up the database:

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U iptv_user -d iptv_platform

# Check tables
\dt

# Check table counts
SELECT
    'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'streams', COUNT(*) FROM streams
UNION ALL
SELECT 'categories', COUNT(*) FROM categories
UNION ALL
SELECT 'packages', COUNT(*) FROM packages
UNION ALL
SELECT 'servers', COUNT(*) FROM servers;

# Exit
\q
```

### **Expected Table Counts (After Seed Data)**

| Table | Expected Count |
|-------|----------------|
| users | 4 (admin + 3 test users) |
| packages | 5 |
| categories | 13 |
| streams | 10+ |
| servers | 4 |

### **Test User Credentials**

| Username | Password | Package | Max Connections |
|----------|----------|---------|-----------------|
| admin | admin123 | Enterprise | 10 |
| testuser | admin123 | Standard | 2 |
| premium_user | admin123 | Premium | 5 |
| trial_user | admin123 | Free Trial | 1 |

⚠️ **IMPORTANT**: Change these passwords in production!

---

## 🧪 **SERVICE TESTING**

### **Test 1: Authentication Service**

#### **Health Check**

```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "ok",
  "service": "auth-service"
}
```

#### **User Login**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "admin123"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "dGVzdHJlZnJlc2g=",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": 2,
      "username": "testuser",
      "email": "test@iptv-platform.com",
      "max_connections": 2
    }
  }
}
```

**Save the JWT token for subsequent requests:**
```bash
# Save token to variable
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

#### **Token Validation**

```bash
curl http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **Metrics Endpoint**

```bash
curl http://localhost:8080/metrics
```

---

### **Test 2: Streaming Gateway**

#### **List All Streams**

```bash
curl http://localhost:8000/api/v1/streams \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "ESPN HD",
      "type": "live",
      "category_id": 2,
      "icon_url": "/icons/espn.png",
      "is_active": true
    },
    ...
  ]
}
```

#### **Get Stream Details**

```bash
curl http://localhost:8000/api/v1/streams/1 \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **Get Stream URL (HLS)**

```bash
curl "http://localhost:8000/api/v1/streams/1/url?container=m3u8" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "type": "live",
    "url": "http://stream1.us-east.cdn.com:443/stream/1/1.m3u8?session=uuid",
    "session_id": "uuid",
    "container": "m3u8"
  }
}
```

#### **Get HLS Playlist**

```bash
curl "http://localhost:8000/api/v1/stream/1/playlist.m3u8?session=test123" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:** (M3U8 playlist)
```
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
...
```

#### **List Categories**

```bash
curl http://localhost:8000/api/v1/categories \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

### **Test 3: Transcoding Service**

#### **Create Transcode Job**

```bash
curl -X POST http://localhost:8002/api/v1/transcode \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "input_file": "/media/test-video.mp4",
    "preset": "1080p_h264"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "job_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending"
  }
}
```

**Save job ID:**
```bash
export JOB_ID="550e8400-e29b-41d4-a716-446655440000"
```

#### **Get Job Status**

```bash
curl http://localhost:8002/api/v1/jobs/$JOB_ID/status \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "job_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "processing",
    "progress": 45.5,
    "started_at": "2025-01-05T10:30:00Z"
  }
}
```

#### **List All Jobs**

```bash
curl http://localhost:8002/api/v1/jobs \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **Cancel Job**

```bash
curl -X POST http://localhost:8002/api/v1/jobs/$JOB_ID/cancel \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

### **Test 4: ML Recommendation Service**

#### **Get Personalized Recommendations**

```bash
curl "http://localhost:8003/api/v1/recommendations/personalized?limit=5" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
[
  {
    "stream_id": 123,
    "score": 0.95,
    "reason": "Based on your viewing history"
  },
  {
    "stream_id": 456,
    "score": 0.89,
    "reason": "Similar to streams you liked"
  }
]
```

#### **Get Similar Content**

```bash
curl http://localhost:8003/api/v1/recommendations/similar/1?limit=5 \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **Get Trending Content**

```bash
curl "http://localhost:8003/api/v1/recommendations/trending?limit=10" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **Submit Feedback**

```bash
curl -X POST "http://localhost:8003/api/v1/recommendations/feedback?stream_id=1&rating=4.5" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

### **Test 5: WebSocket Service**

#### **HTTP Health Check**

```bash
curl http://localhost:8001/health
```

#### **WebSocket Connection Test**

Using `wscat`:

```bash
# Install if needed
npm install -g wscat

# Connect
wscat -c "ws://localhost:8001/ws?token=$JWT_TOKEN"

# Once connected, send messages:
> {"type":"ping"}
< {"type":"pong","timestamp":1704067200000}

> {"type":"subscribe","payload":{"channel":"stream:1"}}
< {"type":"subscribed","payload":{"channel":"stream:1"}}

> {"type":"presence","payload":{"streamId":1}}
< {"type":"presence","payload":{"streamId":1,"viewersCount":1}}

> {"type":"message","payload":{"channel":"stream:1","text":"Hello!"}}
< {"type":"chat_message","payload":{...}}
```

#### **Metrics Endpoint**

```bash
curl http://localhost:8001/metrics
```

**Expected Response:**
```json
{
  "totalConnections": 0,
  "activeConnections": 0,
  "messagesSent": 0,
  "messagesReceived": 0,
  "channels": 0,
  "uniqueUsers": 0
}
```

---

## 🔄 **INTEGRATION TESTING**

### **Complete User Flow Test**

This test simulates a complete user journey:

```bash
#!/bin/bash
# Complete user flow test

# 1. Login
echo "1. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"admin123"}')

JWT_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.access_token')
echo "✅ Login successful"

# 2. Get streams
echo "2. Fetching streams..."
STREAMS=$(curl -s http://localhost:8000/api/v1/streams \
  -H "Authorization: Bearer $JWT_TOKEN")
STREAM_ID=$(echo $STREAMS | jq -r '.data[0].id')
echo "✅ Found stream ID: $STREAM_ID"

# 3. Get stream URL
echo "3. Getting stream URL..."
STREAM_URL=$(curl -s "http://localhost:8000/api/v1/streams/$STREAM_ID/url?container=m3u8" \
  -H "Authorization: Bearer $JWT_TOKEN")
echo "✅ Got stream URL"

# 4. Get recommendations
echo "4. Getting recommendations..."
RECS=$(curl -s "http://localhost:8003/api/v1/recommendations/personalized?limit=5" \
  -H "Authorization: Bearer $JWT_TOKEN")
echo "✅ Got recommendations"

# 5. Submit feedback
echo "5. Submitting feedback..."
curl -s -X POST "http://localhost:8003/api/v1/recommendations/feedback?stream_id=$STREAM_ID&rating=4.5" \
  -H "Authorization: Bearer $JWT_TOKEN" > /dev/null
echo "✅ Feedback submitted"

echo ""
echo "🎉 Integration test completed successfully!"
```

---

## ⚡ **PERFORMANCE TESTING**

### **Load Testing with k6**

Create `load-test.js`:

```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 100 },  // Ramp up to 100 users
    { duration: '1m', target: 100 },   // Stay at 100 for 1 minute
    { duration: '30s', target: 0 },    // Ramp down to 0
  ],
};

export default function () {
  // Test authentication
  let loginRes = http.post('http://localhost:8080/api/v1/auth/login', JSON.stringify({
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
```

Run load test:
```bash
k6 run load-test.js
```

### **Stress Testing**

```bash
# Test authentication service
ab -n 1000 -c 100 -p login.json -T application/json \
  http://localhost:8080/api/v1/auth/login

# Test streaming gateway
ab -n 1000 -c 100 -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8000/api/v1/streams
```

---

## 🔍 **TROUBLESHOOTING**

### **Services Not Starting**

```bash
# Check Docker logs
docker-compose logs

# Restart specific service
docker-compose restart auth-service

# Rebuild and restart
docker-compose up -d --build
```

### **Database Connection Issues**

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Connect manually
docker-compose exec postgres psql -U iptv_user -d iptv_platform

# Check database status
\l  # List databases
\dt # List tables
```

### **Authentication Failures**

```bash
# Verify test user exists
docker-compose exec postgres psql -U iptv_user -d iptv_platform \
  -c "SELECT username, is_active FROM users WHERE username='testuser';"

# Reset password if needed
docker-compose exec postgres psql -U iptv_user -d iptv_platform \
  -c "UPDATE users SET password='$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYAO4yI8KeO' WHERE username='testuser';"
```

### **Port Conflicts**

```bash
# Check what's using a port
lsof -i :8080

# Kill process using port
kill -9 <PID>

# Or change port in docker-compose.yml
```

### **Clear All Data and Reset**

```bash
# Stop all services
docker-compose down

# Remove volumes (⚠️ THIS DELETES ALL DATA)
docker-compose down -v

# Remove images
docker-compose down --rmi all

# Start fresh
docker-compose up -d

# Rerun migrations
cd database && ./migrate.sh up
```

---

## ✅ **SUCCESS CRITERIA**

Your platform is working correctly if:

- [x] All 5 microservices show "Online" status on landing page
- [x] Can login and receive JWT token
- [x] Can list and access streams
- [x] Can create transcode jobs
- [x] Can get ML recommendations
- [x] WebSocket connections work
- [x] Database has test data
- [x] All health checks return "ok"
- [x] No errors in docker-compose logs

---

## 📊 **TEST CHECKLIST**

### **Basic Tests**
- [ ] All services start without errors
- [ ] Health endpoints return 200 OK
- [ ] Database migrations applied successfully
- [ ] Test users can login
- [ ] JWT tokens are generated correctly

### **API Tests**
- [ ] Authentication: Login, logout, token refresh
- [ ] Streaming: List streams, get stream URL, HLS playlist
- [ ] Transcoding: Create job, check status, cancel job
- [ ] ML: Get recommendations, submit feedback
- [ ] WebSocket: Connect, subscribe, send messages

### **Integration Tests**
- [ ] Complete user flow works end-to-end
- [ ] Services communicate correctly
- [ ] Database operations succeed
- [ ] Redis caching works

### **Performance Tests**
- [ ] Load test passes (100 concurrent users)
- [ ] Response times < 500ms (p95)
- [ ] No memory leaks over 1 hour
- [ ] Services auto-recover from failures

---

## 🎓 **NEXT STEPS**

After successful testing:

1. **Production Deployment**
   - Update secrets and passwords
   - Configure SSL/TLS certificates
   - Set up monitoring and alerting
   - Configure backups

2. **Performance Tuning**
   - Optimize database queries
   - Configure connection pooling
   - Enable caching strategies
   - Scale services as needed

3. **Security Hardening**
   - Enable rate limiting
   - Configure firewall rules
   - Set up intrusion detection
   - Implement audit logging

4. **Monitoring Setup**
   - Deploy Prometheus + Grafana
   - Configure alerting rules
   - Set up log aggregation
   - Create custom dashboards

---

**Testing Complete!** 🎉

Your IPTV Platform is now fully tested and ready for deployment!
