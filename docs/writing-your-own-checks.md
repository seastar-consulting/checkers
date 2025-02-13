---
layout: default
title: Writing Your Own Checks
nav_order: 5
---

# Writing Your Own Checks

The checkers CLI is designed to be easily extensible. This guide will show you how to write your own checks in Go that fit your organization's needs.

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

## Example Project

1. Create a new directory for your checks project:

   ```bash
   mkdir orgchecks
   cd orgchecks
   ```

2. Initialize a new Go module:

   ```bash
   go mod init orgchecks
   ```

3. Add the checkers library as a dependency:

   ```bash
   go get github.com/seastar-consulting/checkers
   ```

4. Create a directory for your checks:

   ```bash
   mkdir -p checks/access
   ```

5. Create a new file `checks/access/api.go` with the example check:

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

6. Create a `main.go` file in the root directory:

   ```go
   package main

   import (
       "fmt"
       "os"

       _ "orgchecks/checks/access"

       "github.com/seastar-consulting/checkers/cmd"
   )

   func main() {
       if err := cmd.Execute(); err != nil {
           fmt.Fprintf(os.Stderr, "Error: %v\n", err)
           os.Exit(1)
       }
   }
   ```

7. Create a `.env` file with your API credentials:

   ```
   API_USERNAME=your-username
   API_PASSWORD=your-password
   ```

8. Create a `checks.yaml` file to define your checks:

   ```yaml
   checks:
     - name: verify-api-access
       type: access.api_access
       parameters:
         url: "https://api.example.com/auth"
   ```

9. Run your checks:

   ```bash
   go run main.go
   ```

If you need to include checks from the standard library, you can use the
`github.com/seastar-consulting/checkers/checks/all` import to import all the
available checks or pick specific packages like
`github.com/seastar-consulting/checkers/checks/cloud` or
`github.com/seastar-consulting/checkers/checks/k8s`. Icluding only specific
packages helps keep the resulting binary small and focused on your needs.

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
