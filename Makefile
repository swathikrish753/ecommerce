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
