#!/usr/bin/env sh

container_name="$1-db"

# Start Docker Instance (create if it doesn't exist)
if [ -z "$(docker ps -a | grep "$container_name")" ]; then
	docker run -d -p $2:5432 --name "$container_name" -e POSTGRES_USER=$1 -e POSTGRES_PASSWORD=dev -e POSTGRES_DB=$1 postgres:11-alpine
	timeout 90s sh -c "until docker exec $container_name  pg_isready ; do sleep 5 ; done"
	docker exec -it "$container_name" psql -U $1 -c 'CREATE EXTENSION IF NOT EXISTS pg_trgm; CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'
else
	docker start "$container_name"
fi
