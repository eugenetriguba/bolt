#!/bin/sh
#
# Start/Stop test databases.
#
# Subcommands:
#   start <database>: Start a particular database.
#   stop <database>: Stop a particular database.
ALLOWED_DB=("postgresql" "mysql" "mssql" "sqlite3")

if [[ ! " ${ALLOWED_DB[@]} " =~ " $2 " ]]; then
  echo "Error: Invalid database type. Allowed types are: ${ALLOWED_DB[@]}"
  exit 1
fi

case "$1" in
  start)
    env $(cat ".env.$2" | xargs) docker compose -f "docker-compose.$2.yml" up -d
    ;;
  stop)
    env $(cat ".env.$2" | xargs) docker compose -f "docker-compose.$2.yml" down
    ;;
  *)
    echo "Error: Invalid command. Use 'start' or 'stop'."
    exit 1
    ;;
esac

