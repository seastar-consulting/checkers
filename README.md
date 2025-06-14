# Checkers

[![Go](https://github.com/seastar-consulting/checkers/actions/workflows/go.yml/badge.svg)](https://github.com/seastar-consulting/checkers/actions/workflows/go.yml)
[![Release](https://github.com/seastar-consulting/checkers/actions/workflows/release.yml/badge.svg)](https://github.com/seastar-consulting/checkers/actions/workflows/release.yml)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/seastar-consulting/checkers)](https://github.com/seastar-consulting/checkers/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/seastar-consulting/checkers)](https://goreportcard.com/report/github.com/seastar-consulting/checkers)

Checkers is a diagnostics framework. It helps ensure a set of criteria is met
through declarative checklists.

It comes with a simple command-line interface that allows you to run a set of
checks on your system and display the results in a human-readable format. It includes a
variety of built-in checks for common tasks, such as checking for the presence of
required files, verifying access to AWS S3, and more.

## Features

- Runs multiple system checks in parallel
- Configurable via YAML files
- Supports both command execution and custom check types
- Pretty and JSON output format

## Quickstart

### Installation

You can install Checkers in one of two ways:

1. Download the binary from [GitHub](https://github.com/seastar-consulting/checkers/releases/latest)
   and add it to your PATH.

2. Using Go:
```bash
go install github.com/seastar-consulting/checkers@latest
```

### Usage

First, you need to create a configuration file named `checks.yaml` in your
current directory. This file should contain the checks to be run and their
configuration.

Here is an example of a `checks.yaml` file:

```yaml
checks:
  - name: Check if .env file exists in current directory
    type: "os.file_exists"
    parameters:
      path: ".env"

  - name: Check S3 access
    type: cloud.aws_s3_access
    parameters:
      bucket: "my-bucket"

  - name: Check access to production K8s namespace
    type: k8s.namespace_access
    parameters:
      namespace: "production"
      context: "prod-cluster"
```

Then the checks:

```bash
# Run with default config file (checks.yaml)
checkers

# Run with custom config file
checkers -c my-checks.yaml

# Run with verbose output
checkers -v

# Run with JSON output format
checkers --output json
```

Example pretty output:
```bash
$ checkers
CLOUD
└── ✅ Check S3 access (cloud.aws_s3_access)

K8S
└── ✅ Check access to production K8s namespace (k8s.namespace_access)

OS
└── ❌ Check if .env file exists in current directory (os.file_exists)
```

Example JSON output:
```json
{
  "results": [
    {
      "name": "Check S3 access",
      "type": "cloud.aws_s3_access",
      "status": "Success",
      "output": "Successfully verified write access to bucket 'my-bucket'"
    }
  ],
  "metadata": {
    "datetime": "2025-02-13T15:50:36+02:00",
    "version": "v0.5.1",
    "os": "darwin/arm64"
  }
}
```

### Command Line Options

- `-c, --config string`: Config file path (default "checks.yaml")
- `-f, --file string`: Output file path. Format will be determined by file extension
- `-h, --help`: Help for checkers
- `-o, --output string`: Output format. One of: pretty, json, html (default "pretty")
- `-t, --timeout duration`: Timeout for each check (default 30s)
- `-v, --verbose`: Enable verbose logging
- `--version`: Version for checkers

## Documentation

For detailed documentation on how to use Checkers and configure checks, visit
our [documentation site](https://seastar-consulting.github.io/checkers/).

## Goals

- **Automated Diagnostics**: Automatically check the configuration and health of your development environment.
- **Resolve Possible Issues**: Automatically resolve or provide suggestions for resolving detected problems.
- **Customizable Checks**: Allow users to define custom checks specific to their projects.
- **Easy Integration**: Seamlessly integrate with existing development workflows and tools.
- **Detailed Reporting**: Provide detailed reports. This is useful for
  sharing the diagnostic results with your team in order to get better support.

## Development

To start developing checkers, follow the instructions below:

1. Clone the repository:
    ```sh
    git clone https://github.com/seastar-consulting/checkers.git
    ```
2. Navigate to the project directory:
    ```sh
    cd checkers
    ```
3. Build and test using Make:
    ```sh
    make build  # Build the binary
    make test   # Run tests
    make all    # Build and test
    ```
4. Run the diagnostics:
    ```sh
    ./bin/checkers
    ```

### Makefile Targets

The project includes a Makefile with the following targets:

- `make build`: Build the binary
- `make test`: Run all tests
- `make all`: Build and test
- `make clean`: Remove build artifacts
- `make release`: Build binaries for multiple platforms (linux/darwin/windows, amd64/arm64)

### Project Structure

```
.
├── checks/        # Built-in check implementations
├── cmd/           # Command-line interface entry points
├── docs/          # Documentation files
├── internal/      # Internal packages
│   ├── cli/       # CLI implementation
│   ├── config/    # Configuration handling
│   ├── executor/  # Check execution
│   ├── processor/ # Result processing
│   └── ui/        # User interface
└── types/         # Common type definitions
```

## License

Checkers is released under the Apache License 2.0. See the LICENSE file for details.
