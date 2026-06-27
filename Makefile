.PHONY: all build build-web build-dashboard build-agent run test clean docker

# Variables
BINARY_DASHBOARD = cloudprobe-dashboard
BINARY_AGENT = cloudprobe-agent
DOCKER_IMAGE = cloudprobe/dashboard:latest
WEB_DIR = ./web

# Build targets
all: build

build: build-web build-dashboard build-agent

build-web:
	@echo "Building web frontend..."
	cd $(WEB_DIR) && npm install && npm run build

build-dashboard:
	@echo "Building dashboard..."
	go build -o $(BINARY_DASHBOARD) ./cmd/dashboard

build-agent:
	@echo "Building agent..."
	go build -o $(BINARY_AGENT) ./cmd/agent

run:
	@echo "Starting dashboard..."
	go run ./cmd/dashboard

run-agent:
	@echo "Starting agent..."
	go run ./cmd/agent

dev-web:
	@echo "Starting web dev server..."
	cd $(WEB_DIR) && npm run dev

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_DASHBOARD) $(BINARY_AGENT)
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIR)/node_modules

docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go-grpc_out=. proto/agent.proto

.DEFAULT_GOAL := build
