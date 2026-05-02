#!/bin/bash

set -e

# Migration script for FMS database
# Usage: ./migrate.sh [up|down|status] [connection-string]
# Example: ./migrate.sh up "postgres://user:password@localhost:5432/fms"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/migrations"

ACTION="${1:-up}"
DB_URL="${2:-postgres://postgres:postgres@localhost:5432/fms?sslmode=disable}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Find all migration files
find_migrations() {
    find "$MIGRATIONS_DIR" -name "*.sql" -type f | sort
}

# Get the last applied migration version from database
get_applied_version() {
    psql "$DB_URL" -t -c "SELECT version FROM _schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null || echo "0"
}

# Create migrations table if it doesn't exist
init_migrations_table() {
    psql "$DB_URL" -c "
        CREATE TABLE IF NOT EXISTS _schema_migrations (
            version BIGINT PRIMARY KEY,
            name TEXT NOT NULL,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    " >/dev/null 2>&1 || true
}

# Apply migrations
migrate_up() {
    echo -e "${YELLOW}Applying migrations...${NC}"
    
    init_migrations_table
    
    local applied_version=$(get_applied_version)
    local migrations=$(find_migrations)
    local count=0
    
    while IFS= read -r migration_file; do
        local filename=$(basename "$migration_file")
        local version="${filename%%_*}"
        
        # Only apply migrations with version greater than applied_version
        if [[ $version -gt $applied_version ]]; then
            echo -e "${YELLOW}Applying: $filename${NC}"
            
            # Run migration
            psql "$DB_URL" -f "$migration_file" >/dev/null 2>&1
            
            # Record migration in database
            psql "$DB_URL" -c "
                INSERT INTO _schema_migrations (version, name) 
                VALUES ($version, '$filename')
                ON CONFLICT DO NOTHING;
            " >/dev/null 2>&1
            
            echo -e "${GREEN}✓ Applied: $filename${NC}"
            ((count++))
        fi
    done <<< "$migrations"
    
    if [[ $count -eq 0 ]]; then
        echo -e "${GREEN}No pending migrations${NC}"
    else
        echo -e "${GREEN}Applied $count migration(s)${NC}"
    fi
}

# Show migration status
migration_status() {
    echo -e "${YELLOW}Migration Status:${NC}"
    
    init_migrations_table
    
    echo "Applied Migrations:"
    psql "$DB_URL" -t -c "
        SELECT version, name, applied_at
        FROM _schema_migrations
        ORDER BY version;
    " | while read line; do
        if [[ ! -z "$line" ]]; then
            echo -e "  ${GREEN}✓${NC} $line"
        fi
    done
    
    echo ""
    echo "Pending Migrations:"
    local applied_version=$(get_applied_version)
    local migrations=$(find_migrations)
    local has_pending=0
    
    while IFS= read -r migration_file; do
        local filename=$(basename "$migration_file")
        local version="${filename%%_*}"
        
        if [[ $version -gt $applied_version ]]; then
            echo -e "  ${YELLOW}◌${NC} $filename"
            has_pending=1
        fi
    done <<< "$migrations"
    
    if [[ $has_pending -eq 0 ]]; then
        echo -e "  ${GREEN}None${NC}"
    fi
}

# Main logic
case "$ACTION" in
    up)
        migrate_up
        ;;
    status)
        migration_status
        ;;
    *)
        echo "Usage: $0 [up|status]"
        echo "  up     - Apply pending migrations (default)"
        echo "  status - Show migration status"
        exit 1
        ;;
esac
