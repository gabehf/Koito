.PHONY: all test clean client

postgres.schemadump:
	docker run --rm --network=host --env PGPASSWORD=secret -v "./db:/tmp/dump" \
	postgres pg_dump \
	--schema-only \
	--host=localhost \
	--port=5432 \
	--username=postgres \
	-v --dbname="koitodb" -f "/tmp/dump/schema.sql"

postgres.run:
	docker run --name koito-db -p 5432:5432 -e POSTGRES_PASSWORD=secret -d postgres

postgres.run-scratch:
	docker run --name koito-scratch -p 5433:5432 -e POSTGRES_PASSWORD=secret -d postgres

postgres.start:
	docker start koito-db

postgres.stop:
	docker stop koito-db

postgres.remove:
	docker stop koito-db && docker rm koito-db

postgres.remove-scratch:
	docker stop koito-scratch && docker rm koito-scratch

api.debug:
	KOITO_ALLOWED_HOSTS=* KOITO_LOG_LEVEL=debug KOITO_CONFIG_DIR=test_config_dir KOITO_DATABASE_URL=postgres://postgres:secret@localhost:5432?sslmode=disable go run cmd/api/main.go

api.scratch:
	KOITO_ALLOWED_HOSTS=* KOITO_LOG_LEVEL=debug KOITO_CONFIG_DIR=test_config_dir/scratch KOITO_DATABASE_URL=postgres://postgres:secret@localhost:5433?sslmode=disable go run cmd/api/main.go

api.test:
	go test ./... -timeout 60s

api.build:
	CGO_ENABLED=1 go build -ldflags='-s -w' -o koito ./cmd/api/main.go

client.dev:
	cd client && yarn run dev

docs.dev:
	cd docs && yarn dev

client.deps: 
	cd client && yarn install

client.build: client.deps
	cd client && yarn run build

test: api.test

build: api.build client.build