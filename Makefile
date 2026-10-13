.PHONY: build test test-verbose up down logs clean demo

# Build all Go service binaries locally
build:
	go build -v ./...

# Run all unit tests
test:
	go test ./...

# Run unit tests with verbose logs
test-verbose:
	go test -v ./...

# Start the distributed platform with Docker Compose
up:
	docker compose -f deployments/docker-compose.yml up --build -d

# Stop all running containers
down:
	docker compose -f deployments/docker-compose.yml down -v

# View platform logs
logs:
	docker compose -f deployments/docker-compose.yml logs -f

# Run end-to-end interactive simulation demo
demo:
	pwsh -File scripts/demo.ps1
