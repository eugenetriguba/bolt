#!/bin/bash
#
# Run tests against a particular database.
#
# Usage:
#   $ ./tools/test.sh <database-name>
set +e

SCRIPT_DIR=$(dirname "$0")
. "$SCRIPT_DIR/lib.sh"

BASE_DIR=$(dirname "$0")/..
ALLOWED_DBS=("postgresql" "mysql" "mssql" "sqlite3")

function main() {
  local DB_NAME="$1"
  if ! elementIn "$DB_NAME" "${ALLOWED_DBS[@]}"; then
    echo "Error: Invalid database type. Allowed types are: ${ALLOWED_DBS[*]}" >&2
    exit 1
  fi

  env $(cat "$BASE_DIR/.env.$DB_NAME" | xargs) go test -p 1 -tags "$DB_NAME" -cover -coverprofile cover.out ./...
}

main "$@"
