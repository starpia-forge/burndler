#!/bin/bash
# Wait for PostgreSQL to be ready for connections

set -e

host="$1"
port="$2"
user="$3"
database="$4"
timeout="${5:-30}"

if [ -z "$host" ] || [ -z "$port" ] || [ -z "$user" ] || [ -z "$database" ]; then
    echo "Usage: $0 <host> <port> <user> <database> [timeout]"
    echo "Example: $0 localhost 5432 burndler burndler_dev 30"
    exit 1
fi

echo "Waiting for PostgreSQL at $host:$port to be ready..."

start_time=$(date +%s)
while ! pg_isready -h "$host" -p "$port" -U "$user" -d "$database" >/dev/null 2>&1; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))

    if [ $elapsed -ge $timeout ]; then
        echo "❌ Timeout: PostgreSQL did not become ready within $timeout seconds"
        exit 1
    fi

    echo "⏳ Waiting for PostgreSQL... (${elapsed}s/${timeout}s)"
    sleep 2
done

echo "✅ PostgreSQL is ready at $host:$port!"