# AWS SSO Switcher (Go)

A command-line tool for managing AWS SSO credentials. This is a Go implementation that provides the same functionality as the Python version with improved performance and better error handling.

## Features

- 🔐 Fetch available AWS accounts and roles from SSO
- 🔍 Interactive account selection with fuzzy search
- ⚡ Automatic credential management
- 💾 Configuration persistence
- 🔄 Support for updating account lists
- 🚀 Fast startup and low memory usage
- 🧪 Comprehensive test coverage

## Installation

### From Source

1. Clone the repository
2. Navigate to the `go_switcher` directory
3. Install dependencies:
   ```bash
   make deps
   # or
   go mod tidy
   ```
4. Build the application:
   ```bash
   make build
   # or
   go build -o bin/aws-switcher .
   ```

### Using Go Install

```bash
go install .
```

## Usage

### First-time setup
Set up your SSO configuration and fetch available accounts:
```bash
./bin/aws-switcher --start-url https://your-sso-portal.awsapps.com/start/ --update
```

### Switch between accounts
Launch the interactive account selector:
```bash
./bin/aws-switcher
```

### Update account list
Refresh the list of available accounts from SSO:
```bash
./bin/aws-switcher --update
```

### Command line options
- `--start-url`: AWS SSO start URL
- `--sso-region`: AWS SSO region (default: us-east-1)
- `--update`: Update account list from SSO
- `--help`: Show help information

## Configuration

The tool stores configuration in `~/.aws/switcher-config.json` and updates credentials in `~/.aws/credentials`.

### Example Configuration File
```json
{
  "start_url": "https://your-sso-portal.awsapps.com/start/",
  "sso_region": "us-east-1",
  "accounts": [
    {
      "id": "123456789012",
      "name": "Production Account",
      "role_name": "AdminRole"
    },
    {
      "id": "123456789013",
      "name": "Development Account", 
      "role_name": "DevRole"
    }
  ]
}
```

## Development

### Running Tests
```bash
# Run all tests
make test
# or
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run benchmarks
go test -bench=. ./internal/config
```

### Building
```bash
# Development build
make dev

# Production build
make build

# Build for multiple platforms
make build-all
```

### Code Quality
```bash
# Format code
make fmt

# Vet code
make vet

# Run all checks
make check
```

## Dependencies

- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS API interaction
- [Cobra CLI](https://github.com/spf13/cobra) - Command line interface
- [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) - Interactive fuzzy search
- [go-ini](https://github.com/go-ini/ini) - INI file parsing

## Project Structure

```
go_switcher/
├── main.go                     # Application entry point
├── integration_test.go         # Integration tests
├── internal/
│   ├── aws/
│   │   ├── sso.go             # AWS SSO functionality
│   │   └── sso_test.go        # AWS tests
│   ├── cmd/
│   │   ├── root.go            # CLI command definitions
│   │   └── root_test.go       # CLI tests
│   ├── config/
│   │   ├── config.go          # Configuration management
│   │   ├── config_test.go     # Config tests
│   │   └── bench_test.go      # Performance benchmarks
│   └── ui/
│       ├── selector.go        # Interactive UI
│       └── selector_test.go   # UI tests
├── Makefile                   # Build automation
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
└── README.md                  # This file
```

## Differences from Python Version

1. **Performance**: Significantly faster startup time and lower memory usage
2. **Dependencies**: No need for Python runtime, single binary deployment
3. **Error Handling**: More robust error handling with Go's error system
4. **Type Safety**: Compile-time type checking prevents runtime errors
5. **Testing**: Comprehensive test suite with benchmarks
6. **Cross-Platform**: Easy compilation for multiple platforms

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to ensure code quality
6. Submit a pull request

## License

See LICENSE file for details.
