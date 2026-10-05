.PHONY: run build test tidy lint docker-auth

run:
	ECOM_SERVICE_NAME=auth ECOM_HTTP_PORT=8080 go run ./services/auth

build:
	go build -o bin/ ./services/...

test:
	go test ./... -race -count=1

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-auth:
	docker build -f services/auth/Dockerfile -t ecommerce/auth:dev .

.PHONY: db-up db-down migrate-up migrate-down

db-up:
	docker compose -f deploy/docker/docker-compose.yml up -d

db-down:
	docker compose -f deploy/docker/docker-compose.yml down

migrate-up:
	migrate -path services/auth/migrations -database "$(ECOM_DB_DSN)" up

migrate-down:
	migrate -path services/auth/migrations -database "$(ECOM_DB_DSN)" down
