# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of an AWS SSO credential switcher tool. It provides fast, interactive switching between AWS accounts and roles using AWS SSO authentication.

## Common Development Commands

### Building and Installation
```bash
# Build the application
make build
# or
go build -o bin/aws-switcher .

# Development build with race detection
make dev

# Install dependencies
make deps
# or
go mod tidy

# Install to GOPATH/bin
make install
```

### Testing
```bash
# Run all tests
make test
# or
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run benchmarks (config package has performance benchmarks)
go test -bench=. ./internal/config
```

### Code Quality
```bash
# Format code
make fmt

# Vet code  
make vet

# Run all quality checks (format, vet, test)
make check
```

### Cross-Platform Building
```bash
# Build for multiple platforms
make build-all
```

## Architecture Overview

The application follows a clean, layered architecture:

### Core Packages

- **main.go**: Application entry point, delegates to cmd package
- **internal/cmd**: CLI interface using Cobra framework (`root.go`)
- **internal/config**: Configuration management for persistent storage (`config.go`)
- **internal/aws**: AWS SSO integration and credential management (`sso.go`)  
- **internal/ui**: Interactive user interface with fuzzy search (`selector.go`)

### Key Data Flow

1. **Configuration Loading**: `config.Load()` reads from `~/.aws/switcher-config.json`
2. **Account Fetching**: `aws.FetchAccounts()` uses AWS SSO OIDC flow to retrieve available accounts/roles
3. **Interactive Selection**: `ui.SelectAccount()` provides fuzzy finder interface for account selection
4. **Credential Management**: `aws.GetCredentials()` fetches temporary credentials and `aws.WriteCredentials()` updates `~/.aws/credentials`

### Authentication Flow

The application uses AWS SSO OIDC device authorization flow:
1. Registers a public client with AWS SSO
2. Initiates device authorization 
3. Displays verification URL and user code
4. Polls for authorization completion
5. Exchanges device code for access token

### Configuration Files

- **Config**: `~/.aws/switcher-config.json` stores SSO URL, region, and cached account list
- **Credentials**: `~/.aws/credentials` updated with temporary AWS credentials in default profile
- **AWS Config**: `~/.aws/config` ensured to exist with default profile

### Dependencies

- AWS SDK Go v2 for AWS API interactions
- Cobra for CLI framework
- go-fuzzyfinder for interactive account selection  
- go-ini for INI file manipulation

## Testing Strategy

Each package has comprehensive unit tests:
- Mock AWS API responses for isolated testing
- Benchmark tests for performance-critical config operations
- Integration tests for end-to-end workflows
- Test coverage across all core functionality