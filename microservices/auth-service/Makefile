.PHONY: proto build run test docker-build docker-run clean

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/auth.proto

# Build the application
build: proto
	go build -o bin/auth-service ./cmd/server

# Run the application
run:
	go run ./cmd/server/main.go

# Run tests
test:
	go test -v -cover ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Build Docker image
docker-build:
	docker build -t iptv-auth-service:latest .

# Run Docker container
docker-run:
	docker run -p 50051:50051 -p 8080:8080 --env-file .env iptv-auth-service:latest

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate mocks for testing
mocks:
	mockgen -source=internal/repository/user_repository.go -destination=internal/repository/mocks/user_repository_mock.go
	mockgen -source=internal/repository/session_repository.go -destination=internal/repository/mocks/session_repository_mock.go
