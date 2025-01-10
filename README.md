# Checkers

[![Go](https://github.com/seastar-consulting/checkers/actions/workflows/go.yml/badge.svg)](https://github.com/seastar-consulting/checkers/actions/workflows/go.yml)
[![Release](https://github.com/seastar-consulting/checkers/actions/workflows/release.yml/badge.svg)](https://github.com/seastar-consulting/checkers/actions/workflows/release.yml)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/seastar-consulting/checkers)](https://github.com/seastar-consulting/checkers/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/seastar-consulting/checkers)](https://goreportcard.com/report/github.com/seastar-consulting/checkers)

Checkers is a diagnostics framework for developer workstations. It helps ensure
that your development environment is correctly configured and running smoothly.

## Features

- Runs multiple system checks in parallel
- Configurable via YAML files
- Supports both command execution and custom check types
- Pretty and JSON output format

## Goals

- **Automated Diagnostics**: Automatically check the configuration and health of your development environment.
- **Resolve Possible Issues**: Automatically resolve or provide suggestions for resolving detected problems.
- **Customizable Checks**: Allow users to define custom checks specific to their projects.
- **Easy Integration**: Seamlessly integrate with existing development workflows and tools.
- **Detailed Reporting**: Provide detailed reports on the status of the development environment. This is useful for
    sharing the diagnostic results with your team in order to get better support.

## Non-Goals

- **Environment Setup**: Checkers does not set up or install development tools and environments.
- **Performance Monitoring**: It is not intended for monitoring the performance of applications or external systems.

## Getting Started

To get started with Checkers, follow the instructions below:

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

## Documentation

For detailed documentation on how to use Checkers and configure checks, visit our [documentation site](https://seastar-consulting.github.io/checkers/).

## Configuration

Checks are defined in YAML files. By default, the tool looks for `checks-sample.yaml` in the current directory.

### Example Configuration

```yaml
checks:
  - name: Check important file exists
    type: os.file_exists
    parameters:
      path: /path/to/important/file

  - name: Check memory
    type: command
    command: |
      memory_info=$(free -m)
      used_percent=$(echo "$memory_info" | awk 'NR==2{printf "%.2f", $3*100/$2}')
      echo "{\"status\":\"success\",\"output\":\"Memory usage: ${used_percent}%\"}"
```

## Library-Provided Checks

Checkers comes with several built-in check types:

### OS Checks (`os` package)
- `os.file_exists`: Check if a file exists at the specified path
  - Parameters:
    - `path`: Path to the file to check

You can also define your own custom checks as shell commands.

### Command Checks
- Type: `command`
- Description: Execute a shell command and process its output
- Requirements:
  - The command MUST output a JSON object with the following schema:
    ```json
    {
      "status": "success|failure|error",
      "output": "human readable message",
      "error": "error message if status is error"
    }
    ```
  - The output will be parsed as JSON and merged with the check result
  - If the command output is not valid JSON, it will be wrapped in a result object

### Custom Checks
You can implement custom checks by registering them in the `checks` package. See the `checks/os` directory for examples.

## License

Checkers is released under the Apache License 2.0. See the LICENSE file for details.

## Development

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
├── checks/         # Built-in check implementations
└── internal/       # Internal packages
    ├── cli/       # CLI implementation
    ├── config/    # Configuration handling
    ├── executor/  # Check execution
    ├── processor/ # Result processing
    └── ui/        # User interface