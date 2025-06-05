# P2P Chat Makefile
.PHONY: help build test test-cover clean fmt vet lint run docker-build docker-run demo demo-logs demo-stop demo-clean deps

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the P2P chat binary
	@echo "🔨 Building P2P chat..."
	go build -o p2p ./cmd/p2p

build-linux: ## Build Linux binary for Docker
	@echo "🔨 Building Linux binary..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o p2p-linux ./cmd/p2p

# Test targets
test: ## Run all tests
	@echo "🧪 Running tests..."
	go test ./...

test-cover: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	go test -cover ./...

test-cover-html: ## Generate HTML coverage report
	@echo "🧪 Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# Code quality targets
fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	go vet ./...

lint: ## Run linter (requires golangci-lint)
	@echo "🔍 Running linter..."
	golangci-lint run

# Development targets
run: build ## Build and run locally
	@echo "🚀 Starting P2P chat..."
	./p2p

run-config: build ## Run with example config
	@echo "🚀 Starting P2P chat with config..."
	./p2p -config example-config.yaml

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

clean: ## Clean build artifacts
	@echo "🧹 Cleaning up..."
	rm -f p2p p2p-linux *.out *.html
	docker system prune -f --volumes 2>/dev/null || true

# Docker targets
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t p2p-chat:latest .

docker-run: docker-build ## Run single container
	@echo "🐳 Running Docker container..."
	docker run -it --rm -p 8080:8080 p2p-chat:latest

# Demo targets
demo: ## Start the multi-peer demo
	@echo "🎭 Starting multi-peer demo..."
	@echo "This will create 6 P2P chat peers connected to each other"
	@echo "Use 'make demo-logs' to see logs, 'make demo-stop' to stop"
	docker-compose up -d --build
	@echo ""
	@echo "✅ Demo started! Available peers:"
	@echo "  📡 Bootstrap: http://localhost:8080"
	@echo "  🔗 Peer 1:    http://localhost:8081"
	@echo "  🔗 Peer 2:    http://localhost:8082"
	@echo "  🔗 Peer 3:    http://localhost:8083"
	@echo "  🔗 Peer 4:    http://localhost:8084"
	@echo "  🔗 Peer 5:    http://localhost:8085"
	@echo ""
	@echo "💡 To interact with peers:"
	@echo "  make demo-connect-bootstrap  # Connect to bootstrap peer"
	@echo "  make demo-connect-peer1      # Connect to peer 1"
	@echo "  make demo-logs               # View all logs"
	@echo "  make demo-stop               # Stop demo"

demo-logs: ## Show demo logs
	@echo "📜 Showing demo logs (Ctrl+C to exit)..."
	docker-compose logs -f

demo-status: ## Show demo status
	@echo "📊 Demo status:"
	docker-compose ps

demo-connect-bootstrap: ## Connect to bootstrap peer
	@echo "🔗 Connecting to bootstrap peer..."
	@echo "💬 Type messages and press Enter. Use Ctrl+D to exit."
	docker-compose exec peer-bootstrap /bin/sh -c 'echo "Connected to bootstrap peer" && exec sh'

demo-connect-peer1: ## Connect to peer 1
	@echo "🔗 Connecting to peer 1..."
	@echo "💬 Type messages and press Enter. Use Ctrl+D to exit."
	docker-compose exec peer-1 /bin/sh -c 'echo "Connected to peer 1" && exec sh'

demo-connect-peer2: ## Connect to peer 2
	@echo "🔗 Connecting to peer 2..."
	docker-compose exec peer-2 /bin/sh -c 'echo "Connected to peer 2" && exec sh'

demo-stop: ## Stop the demo
	@echo "🛑 Stopping demo..."
	docker-compose down

demo-clean: ## Stop demo and clean up
	@echo "🧹 Cleaning up demo..."
	docker-compose down -v --rmi all
	docker system prune -f

# Chat simulation
demo-simulate: ## Simulate chat between peers
	@echo "🤖 Simulating chat..."
	@echo "This will send test messages between peers"
	./scripts/simulate-chat.sh

demo-broadcast: ## Run automated broadcast demonstration
	@echo "📡 Starting broadcast demonstration..."
	@echo "This will showcase P2P features with automated messages"
	./scripts/broadcast-demo.sh

demo-broadcast-features: ## Demonstrate specific P2P features
	@echo "🎭 Demonstrating P2P features..."
	./scripts/broadcast-demo.sh features

demo-health: ## Check health of all demo peers
	@echo "🏥 Checking peer health..."
	@echo "Container Status:"
	@docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "Resource Usage:"
	@docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" | head -7
	@echo ""
	@echo "Network Connectivity:"
	@docker network inspect p2p_p2p-network --format '{{range .Containers}}{{.Name}}: {{.IPv4Address}} {{end}}' 2>/dev/null || echo "Network not found"

demo-validate: ## Validate demo configuration and setup
	@echo "🔍 Validating demo setup..."
	@echo "Checking Docker Compose configuration:"
	@docker-compose config --quiet && echo "✅ Configuration valid" || echo "❌ Configuration invalid"
	@echo ""
	@echo "Checking required images:"
	@docker images | grep -E "(p2p|alpine|golang)" | head -5
	@echo ""
	@echo "Checking network setup:"
	@docker network ls | grep p2p || echo "No P2P networks found"

demo-debug: ## Debug demo issues
	@echo "🐛 Demo debugging information..."
	@echo "=== Container Logs (last 10 lines each) ==="
	@for container in p2p-bootstrap p2p-peer-1 p2p-peer-2 p2p-peer-3 p2p-peer-4 p2p-peer-5; do \
		echo "--- $$container ---"; \
		docker logs $$container --tail=10 2>/dev/null || echo "Container not running"; \
		echo ""; \
	done
	@echo "=== Port Usage ==="
	@netstat -tulpn 2>/dev/null | grep -E ":808[0-5]" || echo "No P2P ports in use"
	@echo ""
	@echo "=== Docker System Info ==="
	@docker system df

demo-monitor: ## Monitor demo in real-time
	@echo "📊 Starting real-time monitoring..."
	@echo "Press Ctrl+C to stop monitoring"
	@echo ""
	@while true; do \
		clear; \
		echo "🕐 $(shell date)"; \
		echo ""; \
		echo "📦 Container Status:"; \
		docker-compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null || echo "Demo not running"; \
		echo ""; \
		echo "💾 Resource Usage:"; \
		docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | head -7; \
		echo ""; \
		echo "📡 Recent Activity (last 3 lines):"; \
		docker-compose logs --tail=3 2>/dev/null | tail -3; \
		sleep 5; \
	done

# Development helpers
check: fmt vet test ## Run all code quality checks
	@echo "✅ All checks passed!"

install-tools: ## Install development tools
	@echo "🛠️  Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Version info
version: build ## Show version information
	./p2p -version

# Full pipeline
all: clean deps check build docker-build ## Run complete build pipeline
	@echo "🎉 Build pipeline completed successfully!"