.PHONY: build clean test fmt vet deps install

# Build the application
build:
	go build -o bin/aws-switcher .

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install the binary to GOPATH/bin
install:
	go install .

# Run all checks
check: fmt vet test

# Development build with race detection
dev:
	go build -race -o bin/aws-switcher-dev .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/aws-switcher-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o bin/aws-switcher-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/aws-switcher-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o bin/aws-switcher-windows-amd64.exe .
