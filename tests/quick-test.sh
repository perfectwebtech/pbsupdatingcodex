#!/bin/bash

# ============================================================================
# IPTV Platform Quick API Test
# ============================================================================
# Description: Quick smoke test for basic API functionality
# Usage: ./quick-test.sh [base_url]
# ============================================================================

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
BASE_URL="${1:-http://localhost:8080}"
API_URL="${BASE_URL}/api/v1"

echo -e "${YELLOW}IPTV Platform Quick Test${NC}"
echo "Testing: $BASE_URL"
echo ""

# Test 1: Health Check
echo -n "1. Health Check... "
response=$(curl -s "${BASE_URL}/health" || echo "failed")
if [[ $response == *"ok"* ]] || [[ $response == *"healthy"* ]]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ Failed${NC}"
fi

# Test 2: API Version
echo -n "2. API Version... "
response=$(curl -s "${API_URL}/version" || echo "failed")
if [[ $response == *"version"* ]]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ Failed${NC}"
fi

# Test 3: Login
echo -n "3. Admin Login... "
response=$(curl -s -X POST "${API_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@iptv.example.com","password":"Admin@123"}' || echo "failed")

if [[ $response == *"access_token"* ]]; then
    echo -e "${GREEN}✓${NC}"
    TOKEN=$(echo $response | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
else
    echo -e "${RED}✗ Failed${NC}"
    TOKEN=""
fi

# Test 4: Get Current User
if [ -n "$TOKEN" ]; then
    echo -n "4. Get Current User... "
    response=$(curl -s "${API_URL}/auth/me" \
        -H "Authorization: Bearer $TOKEN" || echo "failed")
    if [[ $response == *"email"* ]]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi
fi

# Test 5: List Users
if [ -n "$TOKEN" ]; then
    echo -n "5. List Users... "
    response=$(curl -s "${API_URL}/users?limit=5" \
        -H "Authorization: Bearer $TOKEN" || echo "failed")
    if [[ $response == *"users"* ]] || [[ $response == *"data"* ]]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi
fi

# Test 6: List Streams
if [ -n "$TOKEN" ]; then
    echo -n "6. List Streams... "
    response=$(curl -s "${API_URL}/streams?limit=5" \
        -H "Authorization: Bearer $TOKEN" || echo "failed")
    if [[ $response == *"streams"* ]] || [[ $response == *"data"* ]]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi
fi

# Test 7: List Categories
if [ -n "$TOKEN" ]; then
    echo -n "7. List Categories... "
    response=$(curl -s "${API_URL}/categories" \
        -H "Authorization: Bearer $TOKEN" || echo "failed")
    if [[ $response == *"categories"* ]] || [[ $response == *"data"* ]]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi
fi

# Test 8: Transcoding Jobs
if [ -n "$TOKEN" ]; then
    echo -n "8. List Transcoding Jobs... "
    response=$(curl -s "${API_URL}/transcoding/jobs?limit=5" \
        -H "Authorization: Bearer $TOKEN" || echo "failed")
    if [[ $response == *"jobs"* ]] || [[ $response == *"data"* ]]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi
fi

echo ""
echo -e "${YELLOW}Quick test completed!${NC}"
echo "Run ./api-test-suite.sh for comprehensive testing"
