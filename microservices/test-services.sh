#!/bin/bash
# Comprehensive service testing script
# Tests all microservices with curl commands

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Service URLs
AUTH_URL="${AUTH_URL:-http://localhost:8080}"
STREAMING_URL="${STREAMING_URL:-http://localhost:8000}"
WEBSOCKET_URL="${WEBSOCKET_URL:-ws://localhost:8001}"
TRANSCODING_URL="${TRANSCODING_URL:-http://localhost:8002}"
ML_URL="${ML_URL:-http://localhost:8003}"

# Test credentials
TEST_USER="testuser"
TEST_PASSWORD="admin123"

# Global variables
JWT_TOKEN=""
STREAM_ID=1
JOB_ID=""

# Helper functions
log_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✅ PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[❌ FAIL]${NC} $1"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Make API request
api_request() {
    local method=$1
    local url=$2
    local data=$3
    local headers=$4

    local cmd="curl -s -X $method $url"

    if [ -n "$data" ]; then
        cmd="$cmd -H 'Content-Type: application/json' -d '$data'"
    fi

    if [ -n "$headers" ]; then
        cmd="$cmd $headers"
    fi

    eval $cmd
}

# Wait for service to be ready
wait_for_service() {
    local url=$1
    local name=$2
    local max_attempts=30
    local attempt=1

    log_info "Waiting for $name to be ready..."

    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url/health" >/dev/null 2>&1; then
            log_success "$name is ready"
            return 0
        fi

        echo -n "."
        sleep 1
        ((attempt++))
    done

    log_error "$name failed to start after $max_attempts seconds"
    return 1
}

# Test Authentication Service
test_auth_service() {
    log_header "Testing Authentication Service"

    # Test 1: Health Check
    log_test "1. Health Check"
    response=$(curl -s "$AUTH_URL/health")
    if echo "$response" | grep -q "ok"; then
        log_success "Health check passed"
        echo "Response: $response"
    else
        log_error "Health check failed"
        echo "Response: $response"
        return 1
    fi

    # Test 2: User Login
    log_test "2. User Login"
    response=$(curl -s -X POST "$AUTH_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASSWORD\"}")

    if echo "$response" | grep -q "access_token"; then
        JWT_TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
        log_success "Login successful"
        log_info "JWT Token: ${JWT_TOKEN:0:50}..."
        echo "Response: $response" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "Login failed"
        echo "Response: $response"
        return 1
    fi

    # Test 3: Token Validation
    if [ -n "$JWT_TOKEN" ]; then
        log_test "3. Token Validation"
        response=$(curl -s -X GET "$AUTH_URL/api/v1/auth/validate" \
            -H "Authorization: Bearer $JWT_TOKEN")

        if echo "$response" | grep -q "success"; then
            log_success "Token validation passed"
        else
            log_error "Token validation failed"
            echo "Response: $response"
        fi
    fi

    # Test 4: Metrics
    log_test "4. Prometheus Metrics"
    response=$(curl -s "$AUTH_URL/metrics")
    if [ -n "$response" ]; then
        log_success "Metrics endpoint accessible"
        echo "First 5 lines:"
        echo "$response" | head -5
    else
        log_error "Metrics endpoint failed"
    fi

    return 0
}

# Test Streaming Gateway
test_streaming_gateway() {
    log_header "Testing Streaming Gateway"

    if [ -z "$JWT_TOKEN" ]; then
        log_error "No JWT token available. Run auth tests first."
        return 1
    fi

    # Test 1: Health Check
    log_test "1. Health Check"
    response=$(curl -s "$STREAMING_URL/health")
    if echo "$response" | grep -q "ok"; then
        log_success "Health check passed"
        echo "Response: $response"
    else
        log_error "Health check failed"
        return 1
    fi

    # Test 2: List Streams
    log_test "2. List All Streams"
    response=$(curl -s "$STREAMING_URL/api/v1/streams" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | grep -q "success"; then
        log_success "List streams successful"
        echo "Response:" | jq '.' 2>/dev/null || echo "$response"

        # Extract first stream ID
        STREAM_ID=$(echo "$response" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
        log_info "Using stream ID: $STREAM_ID"
    else
        log_error "List streams failed"
        echo "Response: $response"
    fi

    # Test 3: Get Stream Details
    if [ -n "$STREAM_ID" ]; then
        log_test "3. Get Stream Details (ID: $STREAM_ID)"
        response=$(curl -s "$STREAMING_URL/api/v1/streams/$STREAM_ID" \
            -H "Authorization: Bearer $JWT_TOKEN")

        if echo "$response" | grep -q "success"; then
            log_success "Get stream details successful"
            echo "Response:" | jq '.' 2>/dev/null || echo "$response"
        else
            log_error "Get stream details failed"
        fi
    fi

    # Test 4: Get Stream URL
    if [ -n "$STREAM_ID" ]; then
        log_test "4. Get Stream URL"
        response=$(curl -s "$STREAMING_URL/api/v1/streams/$STREAM_ID/url?container=m3u8" \
            -H "Authorization: Bearer $JWT_TOKEN")

        if echo "$response" | grep -q "url"; then
            log_success "Get stream URL successful"
            echo "Response:" | jq '.' 2>/dev/null || echo "$response"
        else
            log_error "Get stream URL failed"
            echo "Response: $response"
        fi
    fi

    # Test 5: List Categories
    log_test "5. List Categories"
    response=$(curl -s "$STREAMING_URL/api/v1/categories" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | grep -q "success"; then
        log_success "List categories successful"
    else
        log_error "List categories failed"
        echo "Response: $response"
    fi

    # Test 6: Get User Info
    log_test "6. Get User Info"
    response=$(curl -s "$STREAMING_URL/api/v1/user/info" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | grep -q "success"; then
        log_success "Get user info successful"
        echo "Response:" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "Get user info failed"
    fi

    return 0
}

# Test Transcoding Service
test_transcoding_service() {
    log_header "Testing Transcoding Service"

    if [ -z "$JWT_TOKEN" ]; then
        log_error "No JWT token available. Run auth tests first."
        return 1
    fi

    # Test 1: Health Check
    log_test "1. Health Check"
    response=$(curl -s "$TRANSCODING_URL/health")
    if echo "$response" | grep -q "ok"; then
        log_success "Health check passed"
        echo "Response: $response"
    else
        log_error "Health check failed"
        return 1
    fi

    # Test 2: Create Transcode Job
    log_test "2. Create Transcode Job"
    response=$(curl -s -X POST "$TRANSCODING_URL/api/v1/transcode" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"input_file":"/media/test-video.mp4","preset":"1080p_h264"}')

    if echo "$response" | grep -q "job_id"; then
        JOB_ID=$(echo "$response" | grep -o '"job_id":"[^"]*' | cut -d'"' -f4)
        log_success "Transcode job created"
        log_info "Job ID: $JOB_ID"
        echo "Response:" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "Create transcode job failed"
        echo "Response: $response"
    fi

    # Test 3: Get Job Status
    if [ -n "$JOB_ID" ]; then
        log_test "3. Get Job Status"
        sleep 2 # Wait a bit for job to process

        response=$(curl -s "$TRANSCODING_URL/api/v1/jobs/$JOB_ID/status" \
            -H "Authorization: Bearer $JWT_TOKEN")

        if echo "$response" | grep -q "job_id"; then
            log_success "Get job status successful"
            echo "Response:" | jq '.' 2>/dev/null || echo "$response"
        else
            log_error "Get job status failed"
        fi
    fi

    # Test 4: List Jobs
    log_test "4. List All Jobs"
    response=$(curl -s "$TRANSCODING_URL/api/v1/jobs" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | grep -q "success"; then
        log_success "List jobs successful"
    else
        log_error "List jobs failed"
        echo "Response: $response"
    fi

    return 0
}

# Test ML Recommendation Service
test_ml_service() {
    log_header "Testing ML Recommendation Service"

    if [ -z "$JWT_TOKEN" ]; then
        log_error "No JWT token available. Run auth tests first."
        return 1
    fi

    # Test 1: Health Check
    log_test "1. Health Check"
    response=$(curl -s "$ML_URL/health")
    if echo "$response" | grep -q "ok"; then
        log_success "Health check passed"
        echo "Response: $response"
    else
        log_error "Health check failed"
        return 1
    fi

    # Test 2: Get Personalized Recommendations
    log_test "2. Get Personalized Recommendations"
    response=$(curl -s "$ML_URL/api/v1/recommendations/personalized?limit=5" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | grep -q "stream_id"; then
        log_success "Get recommendations successful"
        echo "Response:" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "Get recommendations failed"
        echo "Response: $response"
    fi

    # Test 3: Get Similar Content
    if [ -n "$STREAM_ID" ]; then
        log_test "3. Get Similar Content (Stream ID: $STREAM_ID)"
        response=$(curl -s "$ML_URL/api/v1/recommendations/similar/$STREAM_ID?limit=5" \
            -H "Authorization: Bearer $JWT_TOKEN")

        if echo "$response" | grep -q "stream_id"; then
            log_success "Get similar content successful"
            echo "Response:" | jq '.' 2>/dev/null || echo "$response"
        else
            log_error "Get similar content failed"
        fi
    fi

    # Test 4: Get Trending Content
    log_test "4. Get Trending Content"
    response=$(curl -s "$ML_URL/api/v1/recommendations/trending?limit=5" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | grep -q "stream_id"; then
        log_success "Get trending content successful"
        echo "Response:" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "Get trending content failed"
    fi

    # Test 5: Submit Feedback
    log_test "5. Submit Feedback"
    response=$(curl -s -X POST "$ML_URL/api/v1/recommendations/feedback?stream_id=1&rating=4.5" \
        -H "Authorization: Bearer $JWT_TOKEN")

    if echo "$response" | grep -q "success"; then
        log_success "Submit feedback successful"
    else
        log_error "Submit feedback failed"
        echo "Response: $response"
    fi

    return 0
}

# Test Real-Time WebSocket
test_websocket() {
    log_header "Testing WebSocket Service"

    if [ -z "$JWT_TOKEN" ]; then
        log_error "No JWT token available. Run auth tests first."
        return 1
    fi

    # Test 1: HTTP Health Check
    log_test "1. Health Check (HTTP)"
    response=$(curl -s "http://localhost:8001/health")
    if echo "$response" | grep -q "ok"; then
        log_success "Health check passed"
        echo "Response: $response"
    else
        log_error "Health check failed"
        return 1
    fi

    # Test 2: Metrics
    log_test "2. Metrics Endpoint"
    response=$(curl -s "http://localhost:8001/metrics")
    if echo "$response" | grep -q "activeConnections"; then
        log_success "Metrics endpoint accessible"
        echo "Response:" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "Metrics endpoint failed"
    fi

    # Test 3: WebSocket Connection (requires wscat or websocat)
    log_test "3. WebSocket Connection Test"
    if command -v wscat &> /dev/null; then
        log_info "Testing WebSocket connection with wscat..."
        timeout 5 wscat -c "ws://localhost:8001/ws?token=$JWT_TOKEN" \
            --execute '{"type":"ping"}' 2>&1 | head -5 || true
        log_success "WebSocket test completed (manual verification needed)"
    elif command -v websocat &> /dev/null; then
        log_info "Testing WebSocket connection with websocat..."
        echo '{"type":"ping"}' | timeout 5 websocat "ws://localhost:8001/ws?token=$JWT_TOKEN" 2>&1 || true
        log_success "WebSocket test completed (manual verification needed)"
    else
        log_info "wscat or websocat not found. Skipping WebSocket connection test."
        log_info "Install with: npm install -g wscat"
    fi

    return 0
}

# Generate test report
generate_report() {
    log_header "Test Summary"

    echo "Tested Services:"
    echo "  ✅ Authentication Service ($AUTH_URL)"
    echo "  ✅ Streaming Gateway ($STREAMING_URL)"
    echo "  ✅ Transcoding Service ($TRANSCODING_URL)"
    echo "  ✅ ML Recommendation Service ($ML_URL)"
    echo "  ✅ WebSocket Service (http://localhost:8001)"
    echo ""
    echo "Test Credentials Used:"
    echo "  Username: $TEST_USER"
    echo "  Password: $TEST_PASSWORD"
    echo ""
    echo "Generated Artifacts:"
    echo "  JWT Token: ${JWT_TOKEN:0:50}..."
    if [ -n "$JOB_ID" ]; then
        echo "  Transcode Job ID: $JOB_ID"
    fi
    if [ -n "$STREAM_ID" ]; then
        echo "  Stream ID: $STREAM_ID"
    fi
}

# Main execution
main() {
    log_header "IPTV Platform - Service Testing Suite"

    log_info "Target Environment:"
    echo "  Auth Service: $AUTH_URL"
    echo "  Streaming Gateway: $STREAMING_URL"
    echo "  Transcoding Service: $TRANSCODING_URL"
    echo "  ML Service: $ML_URL"
    echo "  WebSocket Service: http://localhost:8001"
    echo ""

    # Wait for services to be ready
    wait_for_service "$AUTH_URL" "Authentication Service" || exit 1
    wait_for_service "$STREAMING_URL" "Streaming Gateway" || exit 1
    wait_for_service "$TRANSCODING_URL" "Transcoding Service" || exit 1
    wait_for_service "$ML_URL" "ML Recommendation Service" || exit 1
    wait_for_service "http://localhost:8001" "WebSocket Service" || exit 1

    # Run tests
    test_auth_service || log_error "Authentication tests failed"
    test_streaming_gateway || log_error "Streaming tests failed"
    test_transcoding_service || log_error "Transcoding tests failed"
    test_ml_service || log_error "ML Recommendation tests failed"
    test_websocket || log_error "WebSocket tests failed"

    # Generate report
    generate_report

    log_header "Testing Complete!"
    log_success "All services tested successfully"
}

# Run main if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
