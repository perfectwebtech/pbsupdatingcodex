# IPTV Platform API Test Suite

Comprehensive testing suite for the IPTV Platform API endpoints.

## 📋 Table of Contents

- [Overview](#overview)
- [Test Coverage](#test-coverage)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Test Scripts](#test-scripts)
- [API Endpoint Reference](#api-endpoint-reference)
- [Test Data](#test-data)
- [Troubleshooting](#troubleshooting)

## 🎯 Overview

This test suite provides comprehensive coverage of all IPTV Platform API endpoints, including:
- Authentication & Authorization
- User Management
- Stream Management
- Transcoding System
- Mobile App APIs
- Analytics & Reports
- Billing & Packages
- And more...

## 📊 Test Coverage

### Admin Dashboard APIs
- ✅ Authentication (Login, Logout, Refresh Token)
- ✅ User Management (CRUD operations)
- ✅ Stream Management (CRUD operations)
- ✅ Category Management
- ✅ Transcoding System (Jobs, Queue, Workers, Stats)
- ✅ Analytics & Reports
- ✅ Session Management
- ✅ Package Management
- ✅ Billing & Transactions
- ✅ Reseller Management

### Mobile App APIs
- ✅ Mobile Authentication
- ✅ Device Management (Register, Update, Remove)
- ✅ Content Access (Streams, VOD, Series)
- ✅ Favorites & Watchlist
- ✅ Continue Watching (Progress Tracking)
- ✅ Push Notifications
- ✅ Offline Downloads
- ✅ EPG (Electronic Program Guide)
- ✅ User Profile Management
- ✅ App Configuration

## 🔧 Prerequisites

### Required Software
```bash
# Install curl (usually pre-installed)
sudo apt-get install curl

# Install jq for JSON parsing
sudo apt-get install jq  # Ubuntu/Debian
brew install jq          # macOS

# Optional: Install httpie for better output
pip install httpie
```

### Server Setup
Ensure your IPTV Platform backend is running:
```bash
# Default URL: http://localhost:8080
# Or specify custom URL when running tests
```

## 🚀 Quick Start

### 1. Quick Smoke Test (30 seconds)
```bash
cd tests
./quick-test.sh

# With custom URL
./quick-test.sh https://api.iptv.example.com
```

### 2. Full Test Suite (5-10 minutes)
```bash
cd tests
./api-test-suite.sh

# With custom URL
./api-test-suite.sh https://api.iptv.example.com
```

### 3. Save Test Results
```bash
./api-test-suite.sh | tee test-results-$(date +%Y%m%d-%H%M%S).log
```

## 📝 Test Scripts

### `quick-test.sh`
**Purpose**: Fast smoke test for basic functionality
**Duration**: ~30 seconds
**Tests**: 8 critical endpoints

```bash
./quick-test.sh [base_url]
```

**Output Example**:
```
IPTV Platform Quick Test
Testing: http://localhost:8080

1. Health Check... ✓
2. API Version... ✓
3. Admin Login... ✓
4. Get Current User... ✓
5. List Users... ✓
6. List Streams... ✓
7. List Categories... ✓
8. List Transcoding Jobs... ✓

Quick test completed!
```

### `api-test-suite.sh`
**Purpose**: Comprehensive test of all API endpoints
**Duration**: ~5-10 minutes
**Tests**: 150+ endpoints

```bash
./api-test-suite.sh [base_url]
```

**Test Sections**:
1. Authentication Tests (4 tests)
2. User Management Tests (7 tests)
3. Stream Management Tests (7 tests)
4. Category Management Tests (4 tests)
5. Transcoding Tests (8 tests)
6. Mobile Auth Tests (2 tests)
7. Mobile Device Tests (3 tests)
8. Mobile Content Tests (5 tests)
9. Mobile Favorites Tests (5 tests)
10. Mobile Notifications Tests (3 tests)
11. Mobile Downloads Tests (3 tests)
12. Mobile Profile Tests (4 tests)
13. Analytics Tests (5 tests)
14. Packages & Billing Tests (3 tests)
15. Reseller Tests (3 tests)

## 🔗 API Endpoint Reference

### Base URLs
```
API Base:    http://localhost:8080/api/v1
Admin Panel: http://localhost:3000
```

### Authentication Endpoints
```bash
# Admin Login
POST /api/v1/auth/login
{
  "email": "admin@iptv.example.com",
  "password": "Admin@123"
}

# Mobile Login
POST /api/v1/mobile/auth/login
{
  "email": "user@example.com",
  "password": "Password@123",
  "device_id": "device-12345",
  "device_name": "iPhone 15",
  "platform": "ios"
}

# Refresh Token
POST /api/v1/auth/refresh
{
  "refresh_token": "your_refresh_token"
}

# Get Current User
GET /api/v1/auth/me
Authorization: Bearer <access_token>

# Logout
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

### User Management Endpoints
```bash
# List Users
GET /api/v1/users?limit=10&offset=0&search=john
Authorization: Bearer <access_token>

# Create User
POST /api/v1/users
Authorization: Bearer <access_token>
{
  "email": "newuser@example.com",
  "username": "newuser",
  "password": "Password@123",
  "full_name": "New User",
  "subscription_type": "premium",
  "max_devices": 3,
  "max_streams": 2
}

# Get User
GET /api/v1/users/{id}
Authorization: Bearer <access_token>

# Update User
PUT /api/v1/users/{id}
Authorization: Bearer <access_token>
{
  "full_name": "Updated Name",
  "subscription_type": "enterprise"
}

# Delete User
DELETE /api/v1/users/{id}
Authorization: Bearer <access_token>

# Suspend User
POST /api/v1/users/{id}/suspend
Authorization: Bearer <access_token>

# Activate User
POST /api/v1/users/{id}/activate
Authorization: Bearer <access_token>
```

### Stream Management Endpoints
```bash
# List Streams
GET /api/v1/streams?limit=10&offset=0&category_id=1
Authorization: Bearer <access_token>

# Create Stream
POST /api/v1/streams
Authorization: Bearer <access_token>
{
  "name": "ESPN HD",
  "stream_url": "http://example.com/stream.m3u8",
  "stream_type": "hls",
  "category_id": 1,
  "logo": "http://example.com/logo.png",
  "is_live": true,
  "qualities": ["hd", "fhd"]
}

# Get Stream
GET /api/v1/streams/{id}
Authorization: Bearer <access_token>

# Update Stream
PUT /api/v1/streams/{id}
Authorization: Bearer <access_token>

# Delete Stream
DELETE /api/v1/streams/{id}
Authorization: Bearer <access_token>
```

### Transcoding Endpoints
```bash
# List Jobs
GET /api/v1/transcoding/jobs?limit=10&status=processing
Authorization: Bearer <access_token>

# Create Job
POST /api/v1/transcoding/jobs
Authorization: Bearer <access_token>
{
  "job_name": "Movie Transcode",
  "job_type": "vod_transcode",
  "priority": 5,
  "source_url": "http://source.com/video.mp4",
  "target_url": "http://output.com/video.mp4",
  "target_quality": "hd",
  "hardware_acceleration": false
}

# Get Job Progress
GET /api/v1/transcoding/jobs/{id}/progress
Authorization: Bearer <access_token>

# Cancel Job
POST /api/v1/transcoding/jobs/{id}/cancel
Authorization: Bearer <access_token>

# Get Queue
GET /api/v1/transcoding/queue
Authorization: Bearer <access_token>

# Get Workers
GET /api/v1/transcoding/workers
Authorization: Bearer <access_token>

# Get Stats
GET /api/v1/transcoding/stats
Authorization: Bearer <access_token>
```

### Mobile Content Endpoints
```bash
# Get Streams
GET /api/v1/mobile/streams?limit=10&category_id=1
Authorization: Bearer <mobile_token>

# Get Stream URL
GET /api/v1/mobile/streams/{id}/url?quality=hd
Authorization: Bearer <mobile_token>

# Get VOD Movies
GET /api/v1/mobile/vod?limit=20&search=action
Authorization: Bearer <mobile_token>

# Get Series
GET /api/v1/mobile/series
Authorization: Bearer <mobile_token>

# Get Episodes
GET /api/v1/mobile/series/{id}/episodes
Authorization: Bearer <mobile_token>
```

### Mobile Favorites & Progress
```bash
# Add Favorite
POST /api/v1/mobile/favorites
Authorization: Bearer <mobile_token>
{
  "content_type": "stream",
  "content_id": 1
}

# Get Favorites
GET /api/v1/mobile/favorites
Authorization: Bearer <mobile_token>

# Remove Favorite
DELETE /api/v1/mobile/favorites?content_type=stream&content_id=1
Authorization: Bearer <mobile_token>

# Update Watch Progress
POST /api/v1/mobile/watch-progress
Authorization: Bearer <mobile_token>
{
  "content_type": "movie",
  "content_id": 1,
  "progress": 45,
  "current_time": 2700,
  "duration": 6000
}

# Get Continue Watching
GET /api/v1/mobile/continue-watching
Authorization: Bearer <mobile_token>
```

### Mobile Downloads
```bash
# Request Download
POST /api/v1/mobile/downloads
Authorization: Bearer <mobile_token>
{
  "content_type": "movie",
  "content_id": 1,
  "quality": "hd"
}

# List Downloads
GET /api/v1/mobile/downloads
Authorization: Bearer <mobile_token>

# Delete Download
DELETE /api/v1/mobile/downloads/{id}
Authorization: Bearer <mobile_token>
```

## 📦 Test Data

### Default Admin Credentials
```
Email: admin@iptv.example.com
Password: Admin@123
```

### Test User Credentials
```
Email: test@example.com
Password: Test@123
```

### Sample Stream Data
The test suite will create temporary test data that is automatically cleaned up.

## 🐛 Troubleshooting

### Common Issues

#### 1. Connection Refused
```
Error: curl: (7) Failed to connect to localhost port 8080
```
**Solution**: Ensure the backend server is running
```bash
# Check if server is running
curl http://localhost:8080/health

# Start the server
cd microservices/streaming-gateway
go run main.go
```

#### 2. Authentication Failed
```
Error: Invalid credentials
```
**Solution**: Check if the database has the admin user
```sql
-- Check admin user
SELECT * FROM users WHERE email = 'admin@iptv.example.com';

-- Create admin user if missing
INSERT INTO users (email, username, password, role, is_active)
VALUES ('admin@iptv.example.com', 'admin', '$2a$10$...', 'admin', 1);
```

#### 3. jq Not Found
```
Error: jq: command not found
```
**Solution**: Install jq
```bash
# Ubuntu/Debian
sudo apt-get update && sudo apt-get install jq

# macOS
brew install jq

# CentOS/RHEL
sudo yum install jq
```

#### 4. Permission Denied
```
Error: Permission denied: ./api-test-suite.sh
```
**Solution**: Make script executable
```bash
chmod +x api-test-suite.sh
chmod +x quick-test.sh
```

### Debug Mode

Run tests with verbose output:
```bash
# Enable bash debug mode
bash -x ./api-test-suite.sh

# Or modify script to add set -x
set -x  # Add this at the top of the script
```

### Save Failed Test Details
```bash
# Save all output including errors
./api-test-suite.sh 2>&1 | tee full-test-log.txt

# Filter only errors
./api-test-suite.sh 2>&1 | grep -i "error\|failed" | tee errors.txt
```

## 📈 Performance Testing

### Load Testing with Apache Bench
```bash
# Test login endpoint
ab -n 1000 -c 10 -p login.json -T application/json \
  http://localhost:8080/api/v1/auth/login

# Test get streams endpoint (with auth)
ab -n 5000 -c 50 -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/streams
```

### Load Testing with wrk
```bash
# Install wrk
sudo apt-get install wrk

# Simple load test
wrk -t4 -c100 -d30s http://localhost:8080/api/v1/streams
```

## 🔐 Security Testing

### Test Authentication
```bash
# Test without token (should fail)
curl -X GET http://localhost:8080/api/v1/users

# Test with invalid token (should fail)
curl -X GET -H "Authorization: Bearer invalid_token" \
  http://localhost:8080/api/v1/users

# Test with expired token (should fail)
curl -X GET -H "Authorization: Bearer expired_token" \
  http://localhost:8080/api/v1/users
```

### SQL Injection Testing
```bash
# Test with malicious input
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com OR 1=1--","password":"anything"}'
```

## 📊 Test Reports

### Generate HTML Report
```bash
# Install ansi2html
pip install ansi2html

# Generate report
./api-test-suite.sh | ansi2html > test-report.html
```

### CI/CD Integration

#### GitHub Actions
```yaml
name: API Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install jq
        run: sudo apt-get install jq
      - name: Run Tests
        run: ./tests/api-test-suite.sh http://test-server:8080
```

## 📞 Support

For issues or questions:
- GitHub Issues: https://github.com/your-org/iptv-platform/issues
- Email: support@iptv.example.com
- Documentation: https://docs.iptv.example.com

## 📄 License

Copyright © 2025 IPTV Platform. All rights reserved.
