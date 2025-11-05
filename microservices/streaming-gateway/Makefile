.PHONY: build run test docker-build docker-run clean

# Build the application
build:
	go build -o bin/streaming-gateway ./cmd/server

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
	docker build -t iptv-streaming-gateway:latest .

# Run Docker container
docker-run:
	docker run -p 8000:8000 --env-file .env iptv-streaming-gateway:latest

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
	mockgen -source=internal/repository/stream_repository.go -destination=internal/repository/mocks/stream_repository_mock.go
	mockgen -source=internal/repository/server_repository.go -destination=internal/repository/mocks/server_repository_mock.go
