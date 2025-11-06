#!/bin/bash

# ============================================================================
# IPTV Platform - Backup Script
# ============================================================================
# Usage: ./backup.sh
# ============================================================================

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${BACKUP_DIR:-$PROJECT_DIR/backups}"
RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Load environment variables
if [ -f "$PROJECT_DIR/.env" ]; then
    set -a
    source "$PROJECT_DIR/.env"
    set +a
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IPTV Platform Backup${NC}"
echo -e "${BLUE}Timestamp: $TIMESTAMP${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Database backup
echo -e "${GREEN}>>> Backing up database...${NC}"
docker-compose exec -T mysql mysqldump \
    -u "${MYSQL_USER}" \
    -p"${MYSQL_PASSWORD}" \
    --single-transaction \
    --routines \
    --triggers \
    --events \
    "${MYSQL_DATABASE}" | gzip > "$BACKUP_DIR/database_${TIMESTAMP}.sql.gz"

echo "✓ Database backup completed: database_${TIMESTAMP}.sql.gz"

# Configuration backup
echo -e "${GREEN}>>> Backing up configuration...${NC}"
tar -czf "$BACKUP_DIR/config_${TIMESTAMP}.tar.gz" \
    -C "$PROJECT_DIR" \
    --exclude='.env' \
    --exclude='node_modules' \
    --exclude='vendor' \
    --exclude='.git' \
    docker-compose.yml \
    nginx/ \
    monitoring/ \
    scripts/ \
    2>/dev/null || true

echo "✓ Configuration backup completed: config_${TIMESTAMP}.tar.gz"

# Uploads backup (if exists)
if [ -d "$PROJECT_DIR/uploads" ]; then
    echo -e "${GREEN}>>> Backing up uploads...${NC}"
    tar -czf "$BACKUP_DIR/uploads_${TIMESTAMP}.tar.gz" \
        -C "$PROJECT_DIR" \
        uploads/ \
        2>/dev/null || true
    echo "✓ Uploads backup completed: uploads_${TIMESTAMP}.tar.gz"
fi

# Calculate backup size
BACKUP_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)

# Clean old backups
echo -e "${GREEN}>>> Cleaning old backups (older than $RETENTION_DAYS days)...${NC}"
find "$BACKUP_DIR" -name "*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "*.tar.gz" -type f -mtime +$RETENTION_DAYS -delete

REMAINING_BACKUPS=$(ls -1 "$BACKUP_DIR"/*.sql.gz 2>/dev/null | wc -l)

echo "✓ Cleanup completed"
echo ""

# Upload to S3 (if configured)
if [ "$BACKUP_S3_ENABLED" = "true" ] && [ -n "$AWS_BUCKET" ]; then
    echo -e "${GREEN}>>> Uploading to S3...${NC}"
    
    if command -v aws &> /dev/null; then
        aws s3 cp "$BACKUP_DIR/database_${TIMESTAMP}.sql.gz" \
            "s3://${AWS_BUCKET}/backups/database_${TIMESTAMP}.sql.gz"
        
        aws s3 cp "$BACKUP_DIR/config_${TIMESTAMP}.tar.gz" \
            "s3://${AWS_BUCKET}/backups/config_${TIMESTAMP}.tar.gz"
        
        echo "✓ S3 upload completed"
    else
        echo -e "${YELLOW}Warning: AWS CLI not found, skipping S3 upload${NC}"
    fi
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Backup completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Backup details:"
echo "- Location: $BACKUP_DIR"
echo "- Size: $BACKUP_SIZE"
echo "- Remaining backups: $REMAINING_BACKUPS"
echo "- Retention: $RETENTION_DAYS days"
echo ""

# Create backup manifest
cat > "$BACKUP_DIR/backup_${TIMESTAMP}.manifest" << MANIFEST_EOF
Backup Manifest
===============
Timestamp: $TIMESTAMP
Date: $(date)
Environment: ${APP_ENV:-production}

Files:
- database_${TIMESTAMP}.sql.gz
- config_${TIMESTAMP}.tar.gz
$([ -f "$BACKUP_DIR/uploads_${TIMESTAMP}.tar.gz" ] && echo "- uploads_${TIMESTAMP}.tar.gz")

Database:
- Name: ${MYSQL_DATABASE}
- Size: $(du -sh "$BACKUP_DIR/database_${TIMESTAMP}.sql.gz" | cut -f1)

Total Size: $BACKUP_SIZE
MANIFEST_EOF

echo "Backup manifest created: backup_${TIMESTAMP}.manifest"
