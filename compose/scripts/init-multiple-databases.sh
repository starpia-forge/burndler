#!/bin/bash
# Initialize multiple PostgreSQL databases
# This script is called by PostgreSQL during container initialization

set -e
set -u

function create_user_and_database() {
    local database=$1
    echo "Creating database '$database' for user '$POSTGRES_USER'"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
        CREATE DATABASE $database;
        GRANT ALL PRIVILEGES ON DATABASE $database TO $POSTGRES_USER;
EOSQL
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
    echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
    for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
        if [ "$db" != "$POSTGRES_DB" ]; then
            create_user_and_database $db
        fi
    done
    echo "Multiple databases created"
fi