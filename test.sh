#!/bin/sh

ALLOWED_DB=("postgresql" "mysql" "mssql" "sqlite3")

if [[ ! " ${ALLOWED_DB[@]} " =~ " $1 " ]]; then
  echo "Error: Invalid database type. Allowed types are: ${ALLOWED_DB[@]}"
  exit 1
fi

env $(cat ".env.$1" | xargs) go test -p 1 -tags "$1" -cover ./...

