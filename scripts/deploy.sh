#!/bin/bash

# ============================================================================
# IPTV Platform - Deployment Script
# ============================================================================
# Usage: ./deploy.sh [environment]
# Example: ./deploy.sh production
# ============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT=${1:-production}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IPTV Platform Deployment${NC}"
echo -e "${BLUE}Environment: $ENVIRONMENT${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if environment file exists
if [ ! -f "$PROJECT_DIR/.env.$ENVIRONMENT" ]; then
    echo -e "${RED}Error: Environment file .env.$ENVIRONMENT not found${NC}"
    exit 1
fi

# Load environment variables
set -a
source "$PROJECT_DIR/.env.$ENVIRONMENT"
set +a

# Function to print step
print_step() {
    echo ""
    echo -e "${GREEN}>>> $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}Error: $1${NC}"
    exit 1
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}Warning: $1${NC}"
}

# Check prerequisites
print_step "Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
fi

if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed"
fi

if ! command -v git &> /dev/null; then
    print_error "Git is not installed"
fi

echo "✓ All prerequisites met"

# Pull latest code
print_step "Pulling latest code..."
cd "$PROJECT_DIR"
git fetch origin
git checkout "$ENVIRONMENT"
git pull origin "$ENVIRONMENT"

# Create backup
print_step "Creating backup..."
if [ -f "$SCRIPT_DIR/backup.sh" ]; then
    bash "$SCRIPT_DIR/backup.sh"
else
    print_warning "Backup script not found, skipping backup"
fi

# Build Docker images
print_step "Building Docker images..."
docker-compose build --pull

# Stop old containers
print_step "Stopping old containers..."
docker-compose stop

# Start new containers
print_step "Starting new containers..."
docker-compose up -d

# Wait for services to be healthy
print_step "Waiting for services to be healthy..."
sleep 10

max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -f http://localhost:8080/health &> /dev/null; then
        echo "✓ Backend is healthy"
        break
    fi
    attempt=$((attempt + 1))
    echo "Waiting for backend to be healthy... (attempt $attempt/$max_attempts)"
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    print_error "Backend failed to become healthy"
fi

# Run database migrations
print_step "Running database migrations..."
docker-compose exec -T streaming-gateway /app/streaming-gateway migrate up

# Clean up old Docker images
print_step "Cleaning up old Docker images..."
docker image prune -f

# Show running containers
print_step "Deployment complete!"
echo ""
echo "Running containers:"
docker-compose ps

# Run smoke tests
print_step "Running smoke tests..."
if [ -f "$PROJECT_DIR/tests/quick-test.sh" ]; then
    bash "$PROJECT_DIR/tests/quick-test.sh" http://localhost:8080
else
    print_warning "Test script not found, skipping smoke tests"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Access points:"
echo "- API: http://localhost:8080"
echo "- Admin Dashboard: http://localhost:3000"
echo "- Prometheus: http://localhost:9090"
echo "- Grafana: http://localhost:3001"
echo ""
