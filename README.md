# Checkers

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
3. Build the project:
    ```sh
    go build -o bin/checkers-cli ./cmd/checkers-cli
    ```
4. Run the diagnostics:
    ```sh
    ./bin/checkers-cli
    ```