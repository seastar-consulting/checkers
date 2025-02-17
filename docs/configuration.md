---
layout: default
title: Configuration
nav_order: 3
---

# Configuration

The checks that Checkers is going to run are defined in the `checks` section of
the configuration file. By default, Checkers looks for a file named
`checks.yaml` in the current directory. This page describes the schema and
options of the configuration.

## Basic Structure

```yaml
# Optional: Maximum execution duration
timeout: 5s

# List of checks to run
checks:
  # Simple command check
  - name: Check Docker daemon is running
    type: command
    command: |
      if command -v docker >/dev/null 2>&1; then
        echo '{"status": "success", "output": "Docker daemon is running"}'
      else
        echo '{"status": "failure", "output": "Docker daemon is not running"}'
      fi

  # Check with parameters
  - name: Check .env file exists
    type: os.file_exists
    parameters:
      path: .env
```

This configuration:

1. Sets an execution duration of 5 seconds
2. Defines two different types of checks

## Global Options

| Option  | Type     | Default | Description                   |
| ------- | -------- | ------- | ----------------------------- |
| timeout | duration | 30s     | Timeout for checks to execute |
| checks  | list     | []      | List of checks to run         |

The timeout value accepts Go duration format (e.g., "30s", "1m", "1h"). All
checks exceeding the timeout will be cancelled and a timeout message will be
shown.

## Check Configuration

Each check in the `checks` list requires the following fields:

| Field      | Type   | Required         | Description                                                              |
| ---------- | ------ | ---------------- | ------------------------------------------------------------------------ |
| name       | string | Yes              | Unique identifier for the check                                          |
| type       | string | Yes              | Type of check to perform (e.g., command, os.file_exists)                 |
| command    | string | No\*             | Shell command to execute                                                 |
| parameters | map    | No\*             | Additional parameters specific to check type                             |
| items      | list   | No\*             | List of parameter sets for running multiple variations of the same check |

\* Note: `command`, `parameters`, and `items` are mutually exclusive. A check must have exactly one of these fields.

### Multiple Items Configuration

The `items` field allows you to run the same check with different parameters.
Each item in the list represents a set of parameters for a separate instance of
the check. This is particularly useful when you want to run the same type of
check against multiple targets.

For example, to check multiple executable installations:

```yaml
- name: Check binary installations
  type: os.executable_exists
  items:
    - name: git
    - name: docker
    - name: my-tool
    - name: checkers
```

This will be expanded into multiple checks, each checking a different
executable. The check names will be automatically generated as `{check-name}:
{index}`.

Similarly, you can check multiple S3 buckets:

```yaml
- name: Check S3 access
  type: cloud.aws_s3_access
  items:
    - bucket: "data-bucket"
    - bucket: "backup-bucket"
```

Each item in the list should contain all the parameters required by the check type.

## Command Line Options

The following command-line flags are available:

```bash
checkers [flags]

Flags:
  -c, --config string    Path to config file (default "checks.yaml")
  -t, --timeout string   Timeout for execution (default "30s")
  -v, --verbose         Enable verbose output
  -h, --help           Show help information
```

### Timeout Configuration

The timeout can be configured in two ways:

1. Command-line flag (`--timeout` or `-t`)
2. Configuration file (`timeout` field)

The command-line flag takes precedence over the configuration file. If neither is specified, a default value of 30s is used.

For example:

```bash
# Uses 1m timeout
checkers --timeout 1m

# Uses timeout from checks.yaml, or 30s if not specified
checkers
```

## Best Practices

1. **Group Related Checks**: Organize your checks logically by grouping related items together
2. **Meaningful Names**: Give your checks descriptive names that clearly indicate their purpose
3. **Timeouts**: Set appropriate timeouts to avoid hanging
4. **Error Messages**: Include helpful error messages to make troubleshooting easier
