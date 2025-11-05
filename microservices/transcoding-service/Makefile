.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	cargo build --release

# Run the application
run:
	cargo run

# Run in development mode with hot reload
dev:
	cargo watch -x run

# Run tests
test:
	cargo test

# Run tests with coverage
test-coverage:
	cargo tarpaulin --out Html

# Build Docker image
docker-build:
	docker build -t iptv-transcoding-service:latest .

# Run Docker container
docker-run:
	docker run -p 8002:8002 --env-file .env iptv-transcoding-service:latest

# Clean build artifacts
clean:
	cargo clean

# Format code
fmt:
	cargo fmt

# Lint code
lint:
	cargo clippy -- -D warnings

# Check for security vulnerabilities
audit:
	cargo audit

# Update dependencies
update:
	cargo update
