.PHONY: all run build dev clean vendor migrate

all: build

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

vendor:
	GONOSUMCHECK=* GONOSUMDB=* GOPROXY=file:///Users/jarlen/go/pkg/mod/cache/download,off go mod vendor

dev:
	@echo "Starting infrastructure..."
	docker compose up -d postgres
	@echo "Waiting for postgres..."
	@sleep 3
	@echo "Starting application..."
	go run ./cmd/server

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

clean:
	rm -rf bin/ web/node_modules web/dist

test:
	GONOSUMCHECK=* GONOSUMDB=* GOPROXY=file:///Users/jarlen/go/pkg/mod/cache/download,off go test ./...

migrate:
	@echo "Running migrations..."
	psql "$${DATABASE_URL:-postgres://tiaozao:tiaozao123@localhost:5432/tiaozao?sslmode=disable}" -f migrations/001_create_orders.sql
	@echo "Migrations complete."

lint:
	go vet ./...
