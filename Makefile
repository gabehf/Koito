ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: all test clean client

api.debug:
	go run cmd/api/main.go

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
	find client/build/client -type f \( -name "*.js" -o -name "*.css" -o -name "*.html" -o -name "*.svg" \) -exec gzip -k -9 {} \;

test: api.test

build: api.build client.build
