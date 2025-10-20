#!/usr/bin/env bash
set -euo pipefail

# Setup a test SQLite database and seed random 9-letter codes.
# Usage:
#   scripts/setup_test_db.sh [DB_PATH] [CODES_COUNT]
# Defaults:
#   DB_PATH=a.db
#   CODES_COUNT=20

PROJECT_ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
DB_PATH="${1:-${PROJECT_ROOT_DIR}/a.db}"
CODES_COUNT="${2:-150}"

# Basic logging helpers
log() { printf "[setup-db] %s\n" "$*"; }
err() { printf "[setup-db][ERROR] %s\n" "$*" >&2; }

# Pure bash random 9-letter uppercase code generator (avoids pipelines/SIGPIPE)
rand_code() {
  local alphabet=ABCDEFGHIJKLMNOPQRSTUVWXYZ
  local out=""
  for _ in 1 2 3 4 5 6 7 8 9; do
    local idx=$((RANDOM % 26))
    out+="${alphabet:idx:1}"
  done
  printf '%s\n' "$out"
}

# Trap errors and interrupts for better diagnostics
trap 'err "Aborted (signal or error). Partial DB may exist at: $DB_PATH"' INT TERM
trap 'rc=$?; if [ $rc -ne 0 ]; then err "Script failed with exit code $rc"; fi' EXIT

if ! command -v sqlite3 >/dev/null 2>&1; then
  err "sqlite3 is required but not installed."
  exit 1
fi

log "Using DB: $DB_PATH"
log "Applying schema from: ${PROJECT_ROOT_DIR}/sql/SETUP_DB.sql"
log "sqlite3 version: $(sqlite3 -version)"

# Apply schema (idempotent)
sqlite3 "$DB_PATH" ".read ${PROJECT_ROOT_DIR}/sql/SETUP_DB.sql"

# Helper to get current number of codes
get_codes_count() {
  sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM codes;"
}

# Seed random 6-letter uppercase codes until we reach target count
TARGET=$CODES_COUNT
current=$(get_codes_count)

log "Seeding codes... (current=$current, target=$TARGET)"

# INSERT ADMIN CODE
sqlite3 -batch -noheader "$DB_PATH" "INSERT OR IGNORE INTO codes (code) VALUES ('ADMBSNAKE');" >/dev/null

# INSERT ADM TEAM
sqlite3 -batch -noheader "$DB_PATH" "INSERT OR IGNORE INTO teams (name, code_id, is_admin) VALUES ('Administração', (SELECT id FROM codes WHERE code = 'ADMBSNAKE'), true);" >/dev/null

while [ "$current" -lt "$TARGET" ]; do
  CODE=$(rand_code)
  # Use INSERT OR IGNORE to avoid duplicates; unique constraint on codes(code)
  sqlite3 -batch -noheader "$DB_PATH" "INSERT OR IGNORE INTO codes (code) VALUES ('$CODE');" >/dev/null
  inserted=$(sqlite3 -batch -noheader "$DB_PATH" "SELECT changes();")
#   if [ "${inserted}" = "1" ]; then
#     log "Inserted code: $CODE"
#   else
#     log "Duplicate generated, ignored: $CODE"
#   fi
  current=$(get_codes_count)
  # Optional small sleep to avoid tight loop
  # sleep 0.01
done


log "Done. Codes in DB: $current"

log "Preview of latest codes:";
sqlite3 -header -column "$DB_PATH" "SELECT id, code, created_at FROM codes ORDER BY id ASC LIMIT $CODES_COUNT;"

