#!/bin/sh

set -e
echo "start migration"
migrate -path /app/migration -database "postgresql://username:pass@postgres_container:5432/bank_go?sslmode=disable" -verbose up
echo "start the app"
exec "$@"