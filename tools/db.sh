#!/bin/bash
#
# Start/Stop test databases.
#
# Subcommands:
#   start <database-name>: Start a particular database.
#   stop <database-name>: Stop a particular database.
#
# Usage:
#   $ ./tools/db.sh start <database-name>
#   $ ./tools/db.sh stop <database-name>
set +e

SCRIPT_DIR=$(dirname "$0")
. "$SCRIPT_DIR/lib.sh"

BASE_DIR=$(dirname "$0")/..
ALLOWED_DBS=("postgresql" "mysql" "mssql")
ALLOWED_COMMANDS=("start" "stop")

function main() {
  local COMMAND="$1"
  if ! elementIn "$COMMAND" "${ALLOWED_COMMANDS[@]}"; then
    echo "Error: Invalid command. Allowed subcommands are: ${ALLOWED_COMMANDS[*]}" >&2
    exit 1
  fi
  local DB_NAME="$2"
  if ! elementIn "$DB_NAME" "${ALLOWED_DBS[@]}"; then
    echo "Error: Invalid database type. Allowed types are: ${ALLOWED_DBS[*]}" >&2
    exit 1
  fi

  case $COMMAND in
  start)
    env $(cat "$BASE_DIR/.env.$DB_NAME" | xargs) docker compose -f "$BASE_DIR/docker-compose.$DB_NAME.yml" up -d
    ;;
  stop)
    env $(cat "$BASE_DIR/.env.$DB_NAME" | xargs) docker compose -f "$BASE_DIR/docker-compose.$DB_NAME.yml" down
    ;;
  *)
    echo "Error: Invalid command. Use 'start' or 'stop'."
    exit 1
    ;;
  esac
}

main "$@"
