#!/bin/bash

# ============================================================================
# IPTV Platform API Test Suite
# ============================================================================
# Description: Comprehensive test suite for all API endpoints
# Usage: ./api-test-suite.sh [base_url]
# Example: ./api-test-suite.sh http://localhost:8080
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${1:-http://localhost:8080}"
API_URL="${BASE_URL}/api/v1"
ADMIN_EMAIL="admin@iptv.example.com"
ADMIN_PASSWORD="Admin@123"
TEST_USER_EMAIL="test@example.com"
TEST_USER_PASSWORD="Test@123"

# Global variables for tokens
ACCESS_TOKEN=""
REFRESH_TOKEN=""
MOBILE_ACCESS_TOKEN=""
DEVICE_ID="test-device-$(date +%s)"

# Helper functions
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

print_test() {
    echo -e "${YELLOW}TEST:${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓ SUCCESS:${NC} $1"
}

print_error() {
    echo -e "${RED}✗ ERROR:${NC} $1"
}

print_response() {
    echo -e "${BLUE}Response:${NC}"
    echo "$1" | jq '.' 2>/dev/null || echo "$1"
    echo ""
}

# Make API call and check response
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local auth_header=$4

    local curl_cmd="curl -s -X $method"
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"

    if [ -n "$auth_header" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $auth_header'"
    fi

    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi

    curl_cmd="$curl_cmd '${API_URL}${endpoint}'"

    eval $curl_cmd
}

# ============================================================================
# 1. AUTHENTICATION TESTS
# ============================================================================
test_authentication() {
    print_header "1. AUTHENTICATION TESTS"

    # Test 1.1: Admin Login
    print_test "1.1 Admin Login"
    response=$(api_call POST "/auth/login" "{
        \"email\": \"$ADMIN_EMAIL\",
        \"password\": \"$ADMIN_PASSWORD\"
    }")

    ACCESS_TOKEN=$(echo "$response" | jq -r '.data.access_token // empty')
    REFRESH_TOKEN=$(echo "$response" | jq -r '.data.refresh_token // empty')

    if [ -n "$ACCESS_TOKEN" ]; then
        print_success "Admin login successful"
        print_response "$response"
    else
        print_error "Admin login failed"
        print_response "$response"
        exit 1
    fi

    # Test 1.2: Get Current User
    print_test "1.2 Get Current User"
    response=$(api_call GET "/auth/me" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 1.3: Refresh Token
    print_test "1.3 Refresh Token"
    response=$(api_call POST "/auth/refresh" "{
        \"refresh_token\": \"$REFRESH_TOKEN\"
    }")
    print_response "$response"

    # Test 1.4: Logout
    print_test "1.4 Logout (Testing, will re-login)"
    response=$(api_call POST "/auth/logout" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Re-login for subsequent tests
    response=$(api_call POST "/auth/login" "{
        \"email\": \"$ADMIN_EMAIL\",
        \"password\": \"$ADMIN_PASSWORD\"
    }")
    ACCESS_TOKEN=$(echo "$response" | jq -r '.data.access_token // empty')
}

# ============================================================================
# 2. USER MANAGEMENT TESTS
# ============================================================================
test_user_management() {
    print_header "2. USER MANAGEMENT TESTS"

    # Test 2.1: List Users
    print_test "2.1 List Users (Paginated)"
    response=$(api_call GET "/users?limit=10&offset=0" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 2.2: Create User
    print_test "2.2 Create New User"
    response=$(api_call POST "/users" "{
        \"email\": \"newuser_$(date +%s)@example.com\",
        \"username\": \"newuser_$(date +%s)\",
        \"password\": \"NewUser@123\",
        \"full_name\": \"Test User\",
        \"subscription_type\": \"premium\",
        \"max_devices\": 3,
        \"max_streams\": 2
    }" "$ACCESS_TOKEN")

    USER_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_response "$response"

    # Test 2.3: Get User by ID
    if [ -n "$USER_ID" ]; then
        print_test "2.3 Get User by ID ($USER_ID)"
        response=$(api_call GET "/users/$USER_ID" "" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 2.4: Update User
        print_test "2.4 Update User"
        response=$(api_call PUT "/users/$USER_ID" "{
            \"full_name\": \"Updated Test User\",
            \"subscription_type\": \"enterprise\"
        }" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 2.5: Suspend User
        print_test "2.5 Suspend User"
        response=$(api_call POST "/users/$USER_ID/suspend" "" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 2.6: Activate User
        print_test "2.6 Activate User"
        response=$(api_call POST "/users/$USER_ID/activate" "" "$ACCESS_TOKEN")
        print_response "$response"
    fi

    # Test 2.7: Search Users
    print_test "2.7 Search Users"
    response=$(api_call GET "/users?search=test&limit=5" "" "$ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 3. STREAM MANAGEMENT TESTS
# ============================================================================
test_stream_management() {
    print_header "3. STREAM MANAGEMENT TESTS"

    # Test 3.1: List Streams
    print_test "3.1 List Streams"
    response=$(api_call GET "/streams?limit=10&offset=0" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 3.2: Create Stream
    print_test "3.2 Create New Stream"
    response=$(api_call POST "/streams" "{
        \"name\": \"Test Stream $(date +%s)\",
        \"stream_url\": \"http://example.com/stream.m3u8\",
        \"stream_type\": \"hls\",
        \"category_id\": 1,
        \"logo\": \"http://example.com/logo.png\",
        \"is_live\": true,
        \"qualities\": [\"hd\", \"fhd\"]
    }" "$ACCESS_TOKEN")

    STREAM_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_response "$response"

    # Test 3.3: Get Stream by ID
    if [ -n "$STREAM_ID" ]; then
        print_test "3.3 Get Stream by ID ($STREAM_ID)"
        response=$(api_call GET "/streams/$STREAM_ID" "" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 3.4: Update Stream
        print_test "3.4 Update Stream"
        response=$(api_call PUT "/streams/$STREAM_ID" "{
            \"name\": \"Updated Test Stream\",
            \"is_live\": false
        }" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 3.5: Toggle Stream Status
        print_test "3.5 Deactivate Stream"
        response=$(api_call POST "/streams/$STREAM_ID/deactivate" "" "$ACCESS_TOKEN")
        print_response "$response"

        print_test "3.6 Activate Stream"
        response=$(api_call POST "/streams/$STREAM_ID/activate" "" "$ACCESS_TOKEN")
        print_response "$response"
    fi

    # Test 3.7: Filter Streams by Category
    print_test "3.7 Filter Streams by Category"
    response=$(api_call GET "/streams?category_id=1&limit=5" "" "$ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 4. CATEGORY MANAGEMENT TESTS
# ============================================================================
test_category_management() {
    print_header "4. CATEGORY MANAGEMENT TESTS"

    # Test 4.1: List Categories
    print_test "4.1 List All Categories"
    response=$(api_call GET "/categories" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 4.2: Create Category
    print_test "4.2 Create New Category"
    response=$(api_call POST "/categories" "{
        \"name\": \"Test Category $(date +%s)\",
        \"icon\": \"📺\",
        \"description\": \"Test category for API testing\"
    }" "$ACCESS_TOKEN")

    CATEGORY_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_response "$response"

    # Test 4.3: Update Category
    if [ -n "$CATEGORY_ID" ]; then
        print_test "4.3 Update Category"
        response=$(api_call PUT "/categories/$CATEGORY_ID" "{
            \"name\": \"Updated Test Category\",
            \"icon\": \"🎬\"
        }" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 4.4: Get Category by ID
        print_test "4.4 Get Category by ID"
        response=$(api_call GET "/categories/$CATEGORY_ID" "" "$ACCESS_TOKEN")
        print_response "$response"
    fi
}

# ============================================================================
# 5. TRANSCODING TESTS
# ============================================================================
test_transcoding() {
    print_header "5. TRANSCODING TESTS"

    # Test 5.1: List Transcoding Jobs
    print_test "5.1 List Transcoding Jobs"
    response=$(api_call GET "/transcoding/jobs?limit=10&offset=0" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 5.2: Create Transcoding Job
    print_test "5.2 Create Transcoding Job"
    response=$(api_call POST "/transcoding/jobs" "{
        \"job_name\": \"Test Transcode $(date +%s)\",
        \"job_type\": \"vod_transcode\",
        \"priority\": 5,
        \"source_url\": \"http://example.com/source.mp4\",
        \"target_url\": \"http://example.com/output.mp4\",
        \"target_quality\": \"hd\",
        \"hardware_acceleration\": false,
        \"encoder\": \"libx264\"
    }" "$ACCESS_TOKEN")

    JOB_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_response "$response"

    # Test 5.3: Get Job by ID
    if [ -n "$JOB_ID" ]; then
        print_test "5.3 Get Transcoding Job by ID"
        response=$(api_call GET "/transcoding/jobs/$JOB_ID" "" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 5.4: Get Job Progress
        print_test "5.4 Get Job Progress"
        response=$(api_call GET "/transcoding/jobs/$JOB_ID/progress" "" "$ACCESS_TOKEN")
        print_response "$response"

        # Test 5.5: Cancel Job
        print_test "5.5 Cancel Transcoding Job"
        response=$(api_call POST "/transcoding/jobs/$JOB_ID/cancel" "" "$ACCESS_TOKEN")
        print_response "$response"
    fi

    # Test 5.6: Get Queue Status
    print_test "5.6 Get Transcoding Queue"
    response=$(api_call GET "/transcoding/queue" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 5.7: Get Workers
    print_test "5.7 Get Transcoding Workers"
    response=$(api_call GET "/transcoding/workers" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 5.8: Get Statistics
    print_test "5.8 Get Transcoding Statistics"
    response=$(api_call GET "/transcoding/stats" "" "$ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 6. MOBILE API TESTS - AUTHENTICATION
# ============================================================================
test_mobile_auth() {
    print_header "6. MOBILE API - AUTHENTICATION"

    # Test 6.1: Mobile Login
    print_test "6.1 Mobile Login"
    response=$(api_call POST "/mobile/auth/login" "{
        \"email\": \"$TEST_USER_EMAIL\",
        \"password\": \"$TEST_USER_PASSWORD\",
        \"device_id\": \"$DEVICE_ID\",
        \"device_name\": \"Test iPhone 15\",
        \"platform\": \"ios\",
        \"app_version\": \"2.0.0\"
    }")

    MOBILE_ACCESS_TOKEN=$(echo "$response" | jq -r '.data.access_token // empty')
    MOBILE_REFRESH_TOKEN=$(echo "$response" | jq -r '.data.refresh_token // empty')

    if [ -n "$MOBILE_ACCESS_TOKEN" ]; then
        print_success "Mobile login successful"
        print_response "$response"
    else
        print_error "Mobile login failed - creating test user first"

        # Create test user for mobile
        api_call POST "/users" "{
            \"email\": \"$TEST_USER_EMAIL\",
            \"username\": \"testuser\",
            \"password\": \"$TEST_USER_PASSWORD\",
            \"full_name\": \"Test Mobile User\",
            \"subscription_type\": \"premium\",
            \"max_devices\": 5,
            \"max_streams\": 3
        }" "$ACCESS_TOKEN"

        # Retry login
        response=$(api_call POST "/mobile/auth/login" "{
            \"email\": \"$TEST_USER_EMAIL\",
            \"password\": \"$TEST_USER_PASSWORD\",
            \"device_id\": \"$DEVICE_ID\",
            \"device_name\": \"Test iPhone 15\",
            \"platform\": \"ios\",
            \"app_version\": \"2.0.0\"
        }")

        MOBILE_ACCESS_TOKEN=$(echo "$response" | jq -r '.data.access_token // empty')
        MOBILE_REFRESH_TOKEN=$(echo "$response" | jq -r '.data.refresh_token // empty')
        print_response "$response"
    fi

    # Test 6.2: Refresh Mobile Token
    if [ -n "$MOBILE_REFRESH_TOKEN" ]; then
        print_test "6.2 Refresh Mobile Token"
        response=$(api_call POST "/mobile/auth/refresh" "{
            \"refresh_token\": \"$MOBILE_REFRESH_TOKEN\"
        }")
        print_response "$response"
    fi
}

# ============================================================================
# 7. MOBILE API TESTS - DEVICE MANAGEMENT
# ============================================================================
test_mobile_devices() {
    print_header "7. MOBILE API - DEVICE MANAGEMENT"

    if [ -z "$MOBILE_ACCESS_TOKEN" ]; then
        print_error "Mobile access token not available. Skipping device tests."
        return
    fi

    # Test 7.1: List Devices
    print_test "7.1 List User Devices"
    response=$(api_call GET "/mobile/devices" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 7.2: Register New Device
    print_test "7.2 Register New Device"
    NEW_DEVICE_ID="test-device-2-$(date +%s)"
    response=$(api_call POST "/mobile/devices" "{
        \"device_id\": \"$NEW_DEVICE_ID\",
        \"device_name\": \"Test Android Phone\",
        \"platform\": \"android\",
        \"os_version\": \"14.0\",
        \"app_version\": \"2.0.0\"
    }" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 7.3: Update Device
    print_test "7.3 Update Device"
    response=$(api_call PUT "/mobile/devices/$DEVICE_ID" "{
        \"device_name\": \"Updated iPhone 15 Pro\",
        \"os_version\": \"17.2\",
        \"app_version\": \"2.1.0\"
    }" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 8. MOBILE API TESTS - CONTENT
# ============================================================================
test_mobile_content() {
    print_header "8. MOBILE API - CONTENT"

    if [ -z "$MOBILE_ACCESS_TOKEN" ]; then
        print_error "Mobile access token not available. Skipping content tests."
        return
    fi

    # Test 8.1: Get Streams
    print_test "8.1 Get Mobile Streams"
    response=$(api_call GET "/mobile/streams?limit=10" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 8.2: Get Stream URL
    print_test "8.2 Get Stream URL"
    response=$(api_call GET "/mobile/streams/1/url?quality=hd" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 8.3: Get VOD Movies
    print_test "8.3 Get VOD Movies"
    response=$(api_call GET "/mobile/vod?limit=10" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 8.4: Get Series
    print_test "8.4 Get Series"
    response=$(api_call GET "/mobile/series" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 8.5: Get Episodes
    print_test "8.5 Get Series Episodes"
    response=$(api_call GET "/mobile/series/1/episodes" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 9. MOBILE API TESTS - FAVORITES & PROGRESS
# ============================================================================
test_mobile_favorites() {
    print_header "9. MOBILE API - FAVORITES & PROGRESS"

    if [ -z "$MOBILE_ACCESS_TOKEN" ]; then
        print_error "Mobile access token not available. Skipping favorites tests."
        return
    fi

    # Test 9.1: Add to Favorites
    print_test "9.1 Add Stream to Favorites"
    response=$(api_call POST "/mobile/favorites" "{
        \"content_type\": \"stream\",
        \"content_id\": 1
    }" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 9.2: Get Favorites
    print_test "9.2 Get Favorites"
    response=$(api_call GET "/mobile/favorites" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 9.3: Update Watch Progress
    print_test "9.3 Update Watch Progress"
    response=$(api_call POST "/mobile/watch-progress" "{
        \"content_type\": \"movie\",
        \"content_id\": 1,
        \"progress\": 45,
        \"current_time\": 2700,
        \"duration\": 6000
    }" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 9.4: Get Continue Watching
    print_test "9.4 Get Continue Watching"
    response=$(api_call GET "/mobile/continue-watching" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 9.5: Remove from Favorites
    print_test "9.5 Remove from Favorites"
    response=$(api_call DELETE "/mobile/favorites?content_type=stream&content_id=1" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 10. MOBILE API TESTS - NOTIFICATIONS
# ============================================================================
test_mobile_notifications() {
    print_header "10. MOBILE API - NOTIFICATIONS"

    if [ -z "$MOBILE_ACCESS_TOKEN" ]; then
        print_error "Mobile access token not available. Skipping notification tests."
        return
    fi

    # Test 10.1: Register Push Token
    print_test "10.1 Register Push Token"
    response=$(api_call POST "/mobile/push/register" "{
        \"token\": \"test_push_token_$(date +%s)\",
        \"platform\": \"ios\"
    }" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 10.2: Update Notification Settings
    print_test "10.2 Update Notification Settings"
    response=$(api_call PUT "/mobile/notifications/settings" "{
        \"enable_push\": true,
        \"new_content\": true,
        \"live_events\": true,
        \"recommendations\": false,
        \"subscription_expiry\": true,
        \"system_updates\": true
    }" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 10.3: Get Notifications
    print_test "10.3 Get Notifications"
    response=$(api_call GET "/mobile/notifications?limit=20" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 11. MOBILE API TESTS - DOWNLOADS
# ============================================================================
test_mobile_downloads() {
    print_header "11. MOBILE API - OFFLINE DOWNLOADS"

    if [ -z "$MOBILE_ACCESS_TOKEN" ]; then
        print_error "Mobile access token not available. Skipping download tests."
        return
    fi

    # Test 11.1: Request Download
    print_test "11.1 Request Download"
    response=$(api_call POST "/mobile/downloads" "{
        \"content_type\": \"movie\",
        \"content_id\": 1,
        \"quality\": \"hd\"
    }" "$MOBILE_ACCESS_TOKEN")

    DOWNLOAD_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_response "$response"

    # Test 11.2: Get Downloads
    print_test "11.2 Get Downloads List"
    response=$(api_call GET "/mobile/downloads" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 11.3: Delete Download
    if [ -n "$DOWNLOAD_ID" ]; then
        print_test "11.3 Delete Download"
        response=$(api_call DELETE "/mobile/downloads/$DOWNLOAD_ID" "" "$MOBILE_ACCESS_TOKEN")
        print_response "$response"
    fi
}

# ============================================================================
# 12. MOBILE API TESTS - PROFILE & CONFIG
# ============================================================================
test_mobile_profile() {
    print_header "12. MOBILE API - PROFILE & CONFIG"

    if [ -z "$MOBILE_ACCESS_TOKEN" ]; then
        print_error "Mobile access token not available. Skipping profile tests."
        return
    fi

    # Test 12.1: Get Profile
    print_test "12.1 Get User Profile"
    response=$(api_call GET "/mobile/profile" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 12.2: Update Profile
    print_test "12.2 Update Profile"
    response=$(api_call PUT "/mobile/profile" "{
        \"full_name\": \"Updated Mobile User\",
        \"phone\": \"+1234567890\",
        \"country\": \"US\"
    }" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 12.3: Get App Config
    print_test "12.3 Get App Configuration"
    response=$(api_call GET "/mobile/config?platform=ios&version=2.0.0" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"

    # Test 12.4: Get EPG
    print_test "12.4 Get EPG Data"
    START_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    END_TIME=$(date -u -d "+6 hours" +"%Y-%m-%dT%H:%M:%SZ")
    response=$(api_call GET "/mobile/epg?channel_ids=1,2,3&start=$START_TIME&end=$END_TIME" "" "$MOBILE_ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 13. ANALYTICS & REPORTS TESTS
# ============================================================================
test_analytics() {
    print_header "13. ANALYTICS & REPORTS"

    # Test 13.1: Get Dashboard Stats
    print_test "13.1 Get Dashboard Statistics"
    response=$(api_call GET "/analytics/dashboard" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 13.2: Get Revenue Reports
    print_test "13.2 Get Revenue Reports"
    response=$(api_call GET "/analytics/revenue?period=30days" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 13.3: Get User Growth
    print_test "13.3 Get User Growth Report"
    response=$(api_call GET "/analytics/users/growth?period=12months" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 13.4: Get Top Streams
    print_test "13.4 Get Top Performing Streams"
    response=$(api_call GET "/analytics/streams/top?limit=10" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 13.5: Get Active Sessions
    print_test "13.5 Get Active Sessions"
    response=$(api_call GET "/sessions/active" "" "$ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 14. PACKAGE & BILLING TESTS
# ============================================================================
test_packages_billing() {
    print_header "14. PACKAGES & BILLING"

    # Test 14.1: List Packages
    print_test "14.1 List Subscription Packages"
    response=$(api_call GET "/packages" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 14.2: Create Package
    print_test "14.2 Create Package"
    response=$(api_call POST "/packages" "{
        \"name\": \"Test Premium $(date +%s)\",
        \"description\": \"Test premium package\",
        \"price\": 29.99,
        \"billing_cycle\": \"monthly\",
        \"max_devices\": 5,
        \"max_streams\": 3,
        \"features\": [\"HD Streaming\", \"Offline Downloads\", \"Priority Support\"]
    }" "$ACCESS_TOKEN")

    PACKAGE_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_response "$response"

    # Test 14.3: Get Billing Transactions
    print_test "14.3 Get Billing Transactions"
    response=$(api_call GET "/billing/transactions?limit=20" "" "$ACCESS_TOKEN")
    print_response "$response"
}

# ============================================================================
# 15. RESELLER TESTS
# ============================================================================
test_resellers() {
    print_header "15. RESELLER MANAGEMENT"

    # Test 15.1: List Resellers
    print_test "15.1 List Resellers"
    response=$(api_call GET "/resellers?limit=10" "" "$ACCESS_TOKEN")
    print_response "$response"

    # Test 15.2: Create Reseller
    print_test "15.2 Create Reseller"
    response=$(api_call POST "/resellers" "{
        \"email\": \"reseller_$(date +%s)@example.com\",
        \"username\": \"reseller_$(date +%s)\",
        \"password\": \"Reseller@123\",
        \"full_name\": \"Test Reseller\",
        \"commission_rate\": 15.0,
        \"max_users\": 100
    }" "$ACCESS_TOKEN")

    RESELLER_ID=$(echo "$response" | jq -r '.data.id // empty')
    print_response "$response"

    # Test 15.3: Get Reseller Statistics
    if [ -n "$RESELLER_ID" ]; then
        print_test "15.3 Get Reseller Statistics"
        response=$(api_call GET "/resellers/$RESELLER_ID/stats" "" "$ACCESS_TOKEN")
        print_response "$response"
    fi
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║     IPTV Platform API Test Suite                      ║"
    echo "║     Testing Base URL: $BASE_URL"
    echo "╚════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_error "jq is not installed. Please install jq for JSON parsing."
        echo "Install: apt-get install jq (Ubuntu/Debian) or brew install jq (macOS)"
        exit 1
    fi

    # Run test suites
    test_authentication
    test_user_management
    test_stream_management
    test_category_management
    test_transcoding
    test_mobile_auth
    test_mobile_devices
    test_mobile_content
    test_mobile_favorites
    test_mobile_notifications
    test_mobile_downloads
    test_mobile_profile
    test_analytics
    test_packages_billing
    test_resellers

    # Summary
    print_header "TEST SUITE COMPLETED"
    echo -e "${GREEN}All API tests have been executed.${NC}"
    echo -e "${YELLOW}Review the output above for any errors or failures.${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review failed tests and fix issues"
    echo "2. Run integration tests with real data"
    echo "3. Perform load testing"
    echo "4. Security testing and penetration testing"
}

# Run main function
main "$@"
