#!/bin/bash
# Database migration script
# Usage: ./migrate.sh [up|down|reset|status]

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-iptv_user}"
DB_PASSWORD="${DB_PASSWORD:-dev_password_123}"
DB_NAME="${DB_NAME:-iptv_platform}"

MIGRATIONS_DIR="$(dirname "$0")/migrations"
PSQL="PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create migrations table if it doesn't exist
init_migrations_table() {
    log_info "Initializing migrations table..."
    $PSQL -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    " 2>/dev/null || log_warn "Migrations table already exists"
}

# Get applied migrations
get_applied_migrations() {
    $PSQL -t -c "SELECT version FROM schema_migrations ORDER BY version;" 2>/dev/null | tr -d ' '
}

# Apply a single migration
apply_migration() {
    local migration_file=$1
    local version=$(basename "$migration_file" .sql)

    log_info "Applying migration: $version"

    if $PSQL -f "$migration_file"; then
        $PSQL -c "INSERT INTO schema_migrations (version) VALUES ('$version');"
        log_info "✅ Migration $version applied successfully"
        return 0
    else
        log_error "❌ Migration $version failed"
        return 1
    fi
}

# Run all pending migrations
migrate_up() {
    log_info "Running migrations..."

    init_migrations_table

    local applied_migrations=$(get_applied_migrations)
    local migration_count=0

    for migration_file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$migration_file" ]; then
            local version=$(basename "$migration_file" .sql)

            # Check if migration is already applied
            if echo "$applied_migrations" | grep -q "^$version$"; then
                log_info "⏭️  Skipping $version (already applied)"
            else
                if apply_migration "$migration_file"; then
                    ((migration_count++))
                else
                    log_error "Migration failed. Stopping."
                    exit 1
                fi
            fi
        fi
    done

    if [ $migration_count -eq 0 ]; then
        log_info "No pending migrations"
    else
        log_info "✅ Applied $migration_count migration(s)"
    fi
}

# Rollback last migration
migrate_down() {
    log_warn "Rolling back last migration..."

    local last_migration=$(get_applied_migrations | tail -1)

    if [ -z "$last_migration" ]; then
        log_warn "No migrations to roll back"
        return
    fi

    log_warn "This will rollback: $last_migration"
    read -p "Are you sure? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        log_info "Rollback cancelled"
        return
    fi

    # Execute rollback (if rollback file exists)
    local rollback_file="$MIGRATIONS_DIR/${last_migration}_down.sql"

    if [ -f "$rollback_file" ]; then
        if $PSQL -f "$rollback_file"; then
            $PSQL -c "DELETE FROM schema_migrations WHERE version = '$last_migration';"
            log_info "✅ Rolled back $last_migration"
        else
            log_error "❌ Rollback failed"
            exit 1
        fi
    else
        log_error "Rollback file not found: $rollback_file"
        log_warn "Manual rollback required"
        exit 1
    fi
}

# Reset database
migrate_reset() {
    log_warn "⚠️  WARNING: This will DROP ALL TABLES and rerun migrations!"
    read -p "Are you absolutely sure? Type 'yes' to confirm: " confirm

    if [ "$confirm" != "yes" ]; then
        log_info "Reset cancelled"
        return
    fi

    log_info "Dropping all tables..."

    # Drop all tables
    $PSQL -c "
        DO \$\$ DECLARE
            r RECORD;
        BEGIN
            FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
                EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
            END LOOP;
        END \$\$;
    "

    log_info "Database reset complete. Running migrations..."
    migrate_up
}

# Show migration status
migrate_status() {
    log_info "Migration Status:"
    echo "-----------------------------------"

    init_migrations_table 2>/dev/null

    local applied_migrations=$(get_applied_migrations)

    for migration_file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$migration_file" ]; then
            local version=$(basename "$migration_file" .sql)

            if echo "$applied_migrations" | grep -q "^$version$"; then
                echo -e "${GREEN}✅${NC} $version (applied)"
            else
                echo -e "${YELLOW}⏳${NC} $version (pending)"
            fi
        fi
    done

    echo "-----------------------------------"
}

# Test database connection
test_connection() {
    log_info "Testing database connection..."

    if $PSQL -c "SELECT version();" >/dev/null 2>&1; then
        log_info "✅ Connected to database successfully"
        $PSQL -c "SELECT version();"
        return 0
    else
        log_error "❌ Failed to connect to database"
        log_error "Check your connection settings:"
        echo "  DB_HOST=$DB_HOST"
        echo "  DB_PORT=$DB_PORT"
        echo "  DB_USER=$DB_USER"
        echo "  DB_NAME=$DB_NAME"
        return 1
    fi
}

# Main script
case "${1:-up}" in
    up)
        migrate_up
        ;;
    down)
        migrate_down
        ;;
    reset)
        migrate_reset
        ;;
    status)
        migrate_status
        ;;
    test)
        test_connection
        ;;
    *)
        echo "Usage: $0 {up|down|reset|status|test}"
        echo ""
        echo "Commands:"
        echo "  up     - Apply all pending migrations"
        echo "  down   - Rollback the last migration"
        echo "  reset  - Drop all tables and rerun migrations"
        echo "  status - Show migration status"
        echo "  test   - Test database connection"
        exit 1
        ;;
esac
