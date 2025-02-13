---
layout: default
title: Getting Started
nav_order: 2
---

# Getting Started

## Check Types

Checkers provides two main types of checks, built-in and custom. Built-in checks
are pre-defined checks that are included in the library, while custom checks
allow you to write your own checks using any programming language.

### Built-in Library Checks

Checkers comes with several built-in checks for common development environment validations:

- `os.file_exists`: Verify that a file exists
- `cloud.aws_s3_access`: Verify read/write access to an AWS S3 bucket
- `cloud.aws_authentication`: Verify AWS authentication and identity
- `k8s.namespace_access`: Verify access to a Kubernetes namespace

For a complete list of all available built-in checks, see the 
[Built-in Checks]({% link built-in-checks.md %}) section of the documentation.

### Command Checks

Command checks allow you to run shell commands and validate their output. They
are defined using the `command` type in your configuration. The command must
output a JSON with the following schema:

```json
{
  "status": "success|warning|failure|error",
  "output": "string describing the result"
}
```

Example configuration:

```yaml
checks:
  - name: Check Go version
    type: command
    command: 'go version | grep -q "go1.21" && echo "{\"status\":\"success\",\"output\":\"Go 1.21 is installed\"}" || echo "{\"status\":\"failure\",\"output\":\"Wrong Go version\"}"'
```

The command must return a valid JSON output matching the schema above. The status field can be:

- `success`: The check passed successfully
- `warning`: The check passed but with warnings
- `failure`: The check failed
- `error`: An error occurred while running the check

### Custom Checks

You can extend the checkers library by writing your own checks in Go. For details read
[Writing Your Own Checks]({% link writing-your-own-checks.md %})

## Output Formats

Checkers supports two output formats: pretty (default) and JSON.

### Pretty Format

The pretty format provides a human-readable, hierarchical view of check results:

```bash
$ checkers
CLOUD
└── ✅ Check S3 access (cloud.aws_s3_access)

K8S
└── ✅ Check access to production K8s namespace (k8s.namespace_access)

OS
└── ❌ Check if .env file exists (os.file_exists)
```

### JSON Format

The JSON format provides a machine-readable output, useful for automation and integration with other tools:

```bash
$ checkers --output json
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

The JSON output includes:
- `results`: Array of check results, each containing:
  - `name`: Check name
  - `type`: Check type
  - `status`: Check status (Success, Warning, Failure, Error)
  - `output`: Check output message
- `metadata`: Additional information about the execution:
  - `datetime`: Timestamp of the execution in RFC3339 format
  - `version`: Version of the Checkers CLI (includes git details for development builds)
  - `os`: Operating system and architecture

To use JSON output, pass the `--output json` flag:
```bash
checkers --output json
```

The JSON output is particularly useful for sharing the results with your team in order to get support.

For detailed information about the configuration options, see the [Configuration]({% link configuration.md %}) section of the documentation.
