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
- `error`: There was an error running the check

### Custom Checks

You can extend the checkers library by writing your own checks in Go. For details read
[Writing Your Own Checks]({% link writing-your-own-checks.md %})

## Best Practices

1. **Group Related Checks**: Organize your checks logically by grouping related items together
2. **Meaningful Names**: Give your checks descriptive names that clearly indicate their purpose
3. **Timeouts**: Set appropriate timeouts for command checks to avoid hanging
4. **Error Messages**: Include helpful error messages to make troubleshooting easier
