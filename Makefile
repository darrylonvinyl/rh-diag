# Define variables
IMAGE_NAME := rh-diag-test
TEST_CMD := go test -v ./pkg/diag/...
BUILD_CMD := go build -o rh-diag main.go

.PHONY: help test build clean

help:
	@echo "Available commands:"
	@echo "  make test   - Build the test Docker image (if needed) and run the Go test suite"
	@echo "  make build  - Compile the rh-diag binary natively for Linux using the Docker container"
	@echo "  make clean  - Remove the compiled binary and the test Docker image"

# Build the Docker image if it doesn't already exist
.docker-image: build/Dockerfile.test
	sg docker -c "docker build -t $(IMAGE_NAME) -f build/Dockerfile.test ."
	@touch .docker-image

test: .docker-image
	@echo "Running tests inside isolated Docker environment..."
	sg docker -c "docker run --rm -v $$(pwd):/app -w /app $(IMAGE_NAME) bash -c '$(TEST_CMD)'"

build: .docker-image
	@echo "Compiling Linux binary inside Docker..."
	sg docker -c "docker run --rm -v $$(pwd):/app -w /app $(IMAGE_NAME) bash -c '$(BUILD_CMD)'"
	@echo "Build complete. Output binary: ./rh-diag"

clean:
	@echo "Cleaning up..."
	rm -f rh-diag .docker-image
	-sg docker -c "docker rmi $(IMAGE_NAME) 2>/dev/null" || true
