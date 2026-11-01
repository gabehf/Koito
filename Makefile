ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: all test clean client api.vet client.typecheck client.test check multiuser.check

api.debug:
	go run cmd/api/main.go

api.test:
	go test ./... -timeout 60s

api.build:
	CGO_ENABLED=1 go build -ldflags='-s -w' -o koito ./cmd/api/main.go

client.dev:
	cd client && corepack yarn run dev

docs.dev:
	cd docs && yarn dev

client.deps:
	cd client && corepack yarn install

client.build: client.deps
	cd client && corepack yarn run build
	find client/build/client -type f \( -name "*.js" -o -name "*.css" -o -name "*.html" -o -name "*.svg" \) -exec gzip -kf -9 {} \;

test: api.test

build: api.build client.build

api.vet:
	CGO_ENABLED=1 go vet ./...

client.typecheck:
	cd client && corepack yarn run typecheck

client.test:
	cd client && corepack yarn run test

check: api.vet api.test client.typecheck client.test

multiuser.check: api.build client.build
	python3 scripts/validate_multiuser.py --binary ./koito
