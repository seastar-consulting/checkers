---
layout: default
title: Writing Your Own Checks
nav_order: 4
---

# Writing Your Own Checks

The checkers CLI is designed to be easily extensible. This guide will show you how to write your own  checks in Go that fit your organization's needs.

## Check Structure

A check is a Go function that:
1. Takes a `types.CheckItem` parameter containing:
   - `Name`: Name of the check
   - `Description`: Optional description
   - `Type`: Type of the check
   - `Command`: Optional command for command-based checks
   - `Parameters`: Map of string parameters passed to the check
2. Returns a `types.CheckResult` and an error, where `CheckResult` contains:
   - `Name`: Name of the check
   - `Type`: Type of the check
   - `Status`: One of `Success`, `Failure`, `Warning`, or `Error`
   - `Output`: Human-readable output message
   - `Error`: Optional error message when Status is Error
3. Is registered with the checks registry using `checks.Register`

## Basic Example

```go
package access

import (
    "encoding/base64"
    "fmt"
    "net/http"
    "os"
    "github.com/joho/godotenv"
    "github.com/seastar-consulting/checkers/checks"
    "github.com/seastar-consulting/checkers/types"
)

func init() {
    // Register your check with a unique name and description
    checks.Register("access.api_access", "Verify API access is authorized", CheckAPIAccess)
}

// CheckAPIAccess verifies that access to an API endpoint is authorized
func CheckAPIAccess(item types.CheckItem) (types.CheckResult, error) {
    // Get parameters from the config
    url, ok := item.Parameters["url"]
    if !ok || url == "" {
        return types.CheckResult{
            Name:   item.Name,
            Type:   item.Type,
            Status: types.Error,
            Error:  "url parameter is required",
        }, nil
    }

    // Load credentials from .env file
    if err := godotenv.Load(); err != nil {
        return types.CheckResult{
            Name:   item.Name,
            Type:   item.Type,
            Status: types.Error,
            Error:  "failed to load .env file",
        }, nil
    }

    username := os.Getenv("API_USERNAME")
    password := os.Getenv("API_PASSWORD")
    if username == "" || password == "" {
        return types.CheckResult{
            Name:   item.Name,
            Type:   item.Type,
            Status: types.Error,
            Error:  "API_USERNAME and API_PASSWORD must be set in .env file",
        }, nil
    }

    // Create request with Basic Auth
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return types.CheckResult{
            Name:   item.Name,
            Type:   item.Type,
            Status: types.Error,
            Error:  fmt.Sprintf("failed to create request: %v", err),
        }, nil
    }

    // Add Basic Auth header
    auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
    req.Header.Add("Authorization", "Basic " + auth)

    // Perform the check
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return types.CheckResult{
            Name:   item.Name,
            Type:   item.Type,
            Status: types.Error,
            Error: fmt.Sprintf("Failed to connect to %s: %v", url, err),
        }, nil
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusUnauthorized {
        return types.CheckResult{
            Name:   item.Name,
            Type:   item.Type,
            Status: types.Failure,
            Output: "API access denied: invalid credentials",
        }, nil
    }

    if resp.StatusCode != http.StatusOK {
        return types.CheckResult{
            Name:   item.Name,
            Type:   item.Type,
            Status: types.Failure,
            Output: fmt.Sprintf("API returned unexpected status code: %d", resp.StatusCode),
        }, nil
    }

    return types.CheckResult{
        Name:   item.Name,
        Type:   item.Type,
        Status: types.Success,
        Output: fmt.Sprintf("Successfully authenticated to API at %s", url),
    }, nil
}
```

## Using Custom Checks

1. Create your check in a new package under the `checks` directory
2. Register it in the package's `init()` function
3. Import the package in `main.go`

```go
package main

import (
	"fmt"
	"os"

	_ "orgchecks/access"

	"github.com/seastar-consulting/checkers/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

4. Define it in your `checks.yaml`:

```yaml
checks:
  - name: verify-api-access
    type: access.api_access
    parameters:
      url: "https://api.example.com/auth"
```

Make sure to create a `.env` file with your API credentials:
```
API_USERNAME=your-username
API_PASSWORD=your-password
```

5. Run checkers with `go run main.go`

## Check Guidelines

1. **Naming Convention**:
   - Use a namespace prefix for your checks (e.g., `tools.`, `access.`)
   - Name your checks in CamelCase starting with the prefix Check
   - Register the checks in snake_case
    - Example: `custom.http_endpoint`, `security.ssl_cert_expiry`

2. **Parameter Handling**:
   - Always validate required parameters
   - Provide sensible defaults for optional parameters
   - Document all parameters in comments
   - Remember that all parameters are strings in the `Parameters` map

3. **Error Handling**:
   - Return errors for configuration/setup issues
   - Use appropriate status in CheckResult:
     - `types.Success`: Check passed
     - `types.Failure`: Check failed but executed correctly
     - `types.Warning`: Check passed with warnings
     - `types.Error`: Check couldn't be executed properly
   - Provide clear, actionable error messages

4. **Result Format**:
   - Always set all required fields in CheckResult:
     - `Name`: From the input CheckItem
     - `Type`: From the input CheckItem
     - `Status`: Success, Failure, Warning, or Error
     - `Output`: A descriptive message
   - Set `Error` field when Status is Error

## Best Practices

1. **Modularity**:
   - Keep checks focused on a single responsibility
   - Break complex checks into smaller, reusable functions
   - Use interfaces for external dependencies to enable testing

2. **Documentation**:
   - Add clear comments explaining the check's purpose
   - Document all parameters and their types
   - Include example usage in YAML format
   - Document any assumptions or limitations

3. **Performance**:
   - Set appropriate timeouts
   - Clean up resources (close connections, files)
