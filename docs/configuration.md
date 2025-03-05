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
- name: "Check binary installations"
  type: os.executable_exists
  items:
    - name: git
      path: /usr/local/bin
    - name: docker
```

This will be expanded into multiple checks, each checking a different
executable. By default, check names will be automatically generated as `{check-name}: {index}`.

You can customize the check names using Go template syntax to reference any parameter from your items.
The template has access to all parameters defined in each item. For example:

{% raw %}
```yaml
- name: "Check binary: {{ .name }}"
  type: os.executable_exists
  items:
    - name: git
      path: /usr/local/bin
    - name: docker
```
{% endraw %}

This will create two checks:
1. `Check binary: git` (with parameters `name: git` and `path: /usr/local/bin`)
2. `Check binary: docker` (with parameters `name: docker`)

The template syntax follows Go's [text/template](https://pkg.go.dev/text/template) package rules:
- Use {% raw %}`{{ .key }}`{% endraw %} to reference a parameter value, where `key` is the parameter name
- Parameter names are case-sensitive
- If a referenced parameter is missing, the check will fail validation

Each item in the list must contain all the parameters required by the check
type. The validation will fail if any required parameters are missing.

## Command Line Options

The following command-line flags are available:

```bash
checkers [flags]

Flags:
  -c, --config string     config file path (default "checks.yaml")
  -f, --file string       output file path. Format will be determined by file extension
  -h, --help              help for checkers
  -o, --output string     output format. One of: pretty, json, html (default "pretty")
  -t, --timeout duration  timeout for each check (default 30s)
  -v, --verbose           enable verbose logging
      --version           version for checkers
```

### Output Formats

Checkers supports multiple output formats:

1. **Pretty** (default): Human-readable colored output for terminal viewing
2. **JSON**: Machine-readable JSON format for integration with other tools
3. **HTML**: Rich HTML report with interactive features and styling

You can specify the output format in two ways:

1. Using the `--output` or `-o` flag:
   ```bash
   checkers --output html
   ```

2. Using the `--file` or `-f` flag with an appropriate file extension:
   ```bash
   checkers --file results.html  # Uses HTML format
   checkers --file results.json  # Uses JSON format
   checkers --file results.txt   # Uses Pretty format
   ```

Supported file extensions:
- `.html` - HTML format
- `.json` - JSON format
- `.txt`, `.log`, `.out` - Pretty format

If you specify both `--output` and `--file` flags, the `--output` flag takes precedence.

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
5. **Output Format Selection**: Choose the appropriate output format based on your needs:
   - Use `pretty` for interactive terminal usage
   - Use `json` for integration with other tools or parsing
   - Use `html` for creating shareable reports or documentation
