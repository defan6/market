#!/bin/bash
set -euo pipefail

export PGUSER=${POSTGRES_USER:-postgres}
export PGPASSWORD=${POSTGRES_PASSWORD:-postgres}

DB_NAME="users"

# Проверяем, есть ли база, и создаём если нет
if ! psql -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo "Creating database $DB_NAME..."
    createdb "$DB_NAME"
else
    echo "Database $DB_NAME already exists"
fi