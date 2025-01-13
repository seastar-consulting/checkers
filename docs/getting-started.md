---
layout: default
title: Getting Started
nav_order: 2
---

# Getting Started

## Check Types

Checkers provides two main types of checks:

### Command Checks

Command checks allow you to run shell commands and validate their output. They are defined using the `command` type in your configuration. The command must output a JSON with the following schema:

```json
{
  "status": "success|failure|error",
  "output": "string describing the result"
}
```

Example configuration:
```yaml
checks:
  - name: Check Docker version
    type: command
    command: docker --version | jq -n --arg output "$(/usr/local/bin/docker --version)" '{"status":"success","output":$output}'
```

Another example that checks if a specific version of Go is installed:
```yaml
checks:
  - name: Check Go version
    type: command
    command: 'go version | grep -q "go1.21" && echo "{\"status\":\"success\",\"output\":\"Go 1.21 is installed\"}" || echo "{\"status\":\"failure\",\"output\":\"Wrong Go version\"}"'
```

The command must return a valid JSON output matching the schema above. The status field can be:
- `success`: The check passed successfully
- `failure`: The check failed
- `error`: There was an error running the check

### Built-in Library Checks

Checkers comes with several built-in checks for common development environment validations:

#### File System Checks

- `os.file_exists`: Verify that a file exists

  ```yaml
  - name: Check file exists
    type: os.file_exists
    parameters:
      path: "Makefile"
  ```

## Check Results

Each check will return one of the following statuses:

- `success`: The check passed successfully
- `failure`: The check failed
- `error`: There was an error running the check

## Best Practices

1. **Group Related Checks**: Organize your checks logically by grouping related items together
2. **Meaningful Names**: Give your checks descriptive names that clearly indicate their purpose
3. **Timeouts**: Set appropriate timeouts for command checks to avoid hanging
4. **Error Messages**: Include helpful error messages to make troubleshooting easier

## Example Configuration

Here's a complete example that combines various check types:

```yaml
checks:
  # Development tools
  - name: Check Git installation
    type: command
    command: git --version | jq -n --arg output "$(/usr/local/bin/git --version)" '{"status":"success","output":$output}'

  - name: Check Docker
    type: command
    command: docker info | jq -n --arg output "$(/usr/local/bin/docker info)" '{"status":"success","output":$output}'

  # Project configuration
  - name: Check environment file
    type: os.file_exists
    parameters:
      path: .env
```
