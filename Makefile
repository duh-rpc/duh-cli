.PHONY: test lint install build clean

# Run all tests
test:
	go test -v ./...

# Lint code using golangci-lint
lint:
	golangci-lint run ./...

# Install binary to GOPATH/bin
install:
	go install ./cmd/duhrpc-lint

# Build binary
build:
	go build -o duhrpc-lint ./cmd/duhrpc-lint

# Clean build artifacts
clean:
	rm -f duhrpc-lint
	go clean
