#!/bin/sh
#
# Run tests against a particular database.
#
# Usage:
#   $ ./tools/test.sh <database-name>
BASE_DIR=`dirname $0`/..
ALLOWED_DB=("postgresql" "mysql" "mssql" "sqlite3")

if [[ ! " ${ALLOWED_DB[@]} " =~ " $1 " ]]; then
  echo "Error: Invalid database type. Allowed types are: ${ALLOWED_DB[@]}"
  exit 1
fi

env $(cat "$BASE_DIR/.env.$1" | xargs) go test -p 1 -tags "$1" -cover -coverprofile cover.out ./...

