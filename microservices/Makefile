.PHONY: help build up down logs clean test deploy-k8s

help:
	@echo "IPTV Platform - Development Commands"
	@echo ""
	@echo "Local Development (Docker Compose):"
	@echo "  make build          - Build all Docker images"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make logs           - View logs from all services"
	@echo "  make logs-auth      - View auth service logs"
	@echo "  make logs-stream    - View streaming gateway logs"
	@echo "  make logs-ws        - View WebSocket server logs"
	@echo "  make clean          - Remove all containers, volumes, and images"
	@echo "  make restart        - Restart all services"
	@echo ""
	@echo "Kubernetes Deployment:"
	@echo "  make deploy-k8s     - Deploy to Kubernetes"
	@echo "  make delete-k8s     - Delete from Kubernetes"
	@echo "  make status-k8s     - Check Kubernetes status"
	@echo "  make logs-k8s       - View Kubernetes logs"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-auth      - Test auth service"
	@echo "  make test-stream    - Test streaming gateway"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate     - Run database migrations"
	@echo "  make db-seed        - Seed database with test data"
	@echo "  make db-shell       - Open PostgreSQL shell"

# Docker Compose commands
build:
	docker-compose build

up:
	docker-compose up -d
	@echo "Services are starting..."
	@echo "Auth Service:      http://localhost:8080"
	@echo "Streaming Gateway: http://localhost:8000"
	@echo "WebSocket Server:  ws://localhost:8001/ws"
	@echo "PostgreSQL:        localhost:5432"
	@echo "Redis:             localhost:6379"

down:
	docker-compose down

logs:
	docker-compose logs -f

logs-auth:
	docker-compose logs -f auth-service

logs-stream:
	docker-compose logs -f streaming-gateway

logs-ws:
	docker-compose logs -f realtime-ws

restart:
	docker-compose restart

clean:
	docker-compose down -v
	docker system prune -af

# Kubernetes commands
deploy-k8s:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/postgres.yaml
	kubectl apply -f k8s/redis.yaml
	kubectl apply -f k8s/auth-service.yaml
	kubectl apply -f k8s/streaming-gateway.yaml
	kubectl apply -f k8s/realtime-ws.yaml
	kubectl apply -f k8s/ingress.yaml
	@echo "Kubernetes deployment completed!"

delete-k8s:
	kubectl delete -f k8s/ --recursive

status-k8s:
	kubectl get all -n iptv-platform

logs-k8s:
	kubectl logs -f -n iptv-platform -l app=streaming-gateway

# Database commands
db-migrate:
	docker-compose exec postgres psql -U iptv_user -d iptv_platform -f /migrations/init.sql

db-seed:
	@echo "Seeding database..."
	# Add seed commands here

db-shell:
	docker-compose exec postgres psql -U iptv_user -d iptv_platform

# Testing
test:
	@echo "Running tests..."
	cd auth-service && make test
	cd streaming-gateway && make test

test-auth:
	cd auth-service && make test

test-stream:
	cd streaming-gateway && make test

# Quick development cycle
dev: down build up logs
