.PHONY: test lint install build clean ci tidy coverage integration-test

# Run all tests
test:
	go test -v ./...

# Lint code using golangci-lint
lint:
	golangci-lint run ./...

# Install binary to GOPATH/bin
install:
	go install ./cmd/duhrpc

tidy:
	go mod tidy && git diff --exit-code

ci: tidy lint test
	@echo
	@echo "\033[32mEVERYTHING PASSED!\033[0m"

# Build binary
build:
	go build -o duhrpc ./cmd/duhrpc

# Clean build artifacts
clean:
	rm -f duhrpc coverage.out coverage.html
	go clean

# Coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Integration tests
integration-test: build
	@echo "Testing valid spec..."
	./duhrpc lint testdata/valid-spec.yaml
	@echo "\nTesting invalid specs..."
	! ./duhrpc lint testdata/invalid-specs/bad-path-format.yaml
	! ./duhrpc lint testdata/invalid-specs/wrong-http-method.yaml
	! ./duhrpc lint testdata/invalid-specs/has-query-params.yaml
	! ./duhrpc lint testdata/invalid-specs/missing-request-body.yaml
	! ./duhrpc lint testdata/invalid-specs/invalid-status-code.yaml
	! ./duhrpc lint testdata/invalid-specs/missing-success-response.yaml
	! ./duhrpc lint testdata/invalid-specs/invalid-content-type.yaml
	! ./duhrpc lint testdata/invalid-specs/bad-error-schema.yaml
	! ./duhrpc lint testdata/invalid-specs/multiple-violations.yaml
	@echo "\nIntegration tests passed!"
