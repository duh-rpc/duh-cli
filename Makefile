.PHONY: test lint install build clean ci tidy

# Run all tests
test:
	go test -v ./...

# Lint code using golangci-lint
lint:
	golangci-lint run ./...

# Install binary to GOPATH/bin
install:
	go install ./cmd/duhrpc-lint

tidy:
	go mod tidy && git diff --exit-code

ci: tidy lint test
	@echo
	@echo "\033[32mEVERYTHING PASSED!\033[0m"

# Build binary
build:
	go build -o duhrpc-lint ./cmd/duhrpc-lint

# Clean build artifacts
clean:
	rm -f duhrpc-lint
	go clean
