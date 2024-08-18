#!/bin/sh
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
BASE_DIR=`dirname $0`/..
ALLOWED_DB=("postgresql" "mysql" "mssql" "sqlite3")

if [[ ! " ${ALLOWED_DB[@]} " =~ " $2 " ]]; then
  echo "Error: Invalid database type. Allowed types are: ${ALLOWED_DB[@]}"
  exit 1
fi

case "$1" in
  start)
    env $(cat "$BASE_DIR/.env.$2" | xargs) docker compose -f "$BASE_DIR/docker-compose.$2.yml" up -d
    ;;
  stop)
    env $(cat "$BASE_DIR/.env.$2" | xargs) docker compose -f "$BASE_DIR/docker-compose.$2.yml" down
    ;;
  *)
    echo "Error: Invalid command. Use 'start' or 'stop'."
    exit 1
    ;;
esac
