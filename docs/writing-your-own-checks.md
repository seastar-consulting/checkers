---
layout: default
title: Writing Your Own Checks
nav_order: 3
---

# Writing Your Own Checks

The checkers CLI is designed to be easily extensible. This guide will show you how to write your own custom checks.

## Check Structure

A check is a Go function that:
1. Takes a `map[string]interface{}` of parameters
2. Returns a result `map[string]interface{}` and an error
3. Is registered with the checks registry

## Basic Example

Here's a simple example of a custom check that verifies an HTTP endpoint:

```go
package custom

import (
    "fmt"
    "net/http"
    "github.com/seastar-consulting/checkers/checks"
)

func init() {
    // Register your check with a unique name and description
    checks.Register("custom.http_endpoint", "Verify HTTP endpoint availability", CheckHTTPEndpoint)
}

// CheckHTTPEndpoint verifies that an HTTP endpoint is accessible
func CheckHTTPEndpoint(params map[string]interface{}) (map[string]interface{}, error) {
    // Get parameters from the config
    url, ok := params["url"].(string)
    if !ok {
        return nil, fmt.Errorf("url parameter is required")
    }

    expectedStatus := 200
    if status, ok := params["expected_status"].(float64); ok {
        expectedStatus = int(status)
    }

    // Perform the check
    resp, err := http.Get(url)
    if err != nil {
        return map[string]interface{}{
            "status": "Failure",
            "output": fmt.Sprintf("Failed to connect to %s: %v", url, err),
        }, nil
    }
    defer resp.Body.Close()

    if resp.StatusCode != expectedStatus {
        return map[string]interface{}{
            "status": "Failure",
            "output": fmt.Sprintf("Expected status %d but got %d", expectedStatus, resp.StatusCode),
        }, nil
    }

    return map[string]interface{}{
        "status": "Success",
        "output": fmt.Sprintf("Successfully connected to %s (status: %d)", url, resp.StatusCode),
    }, nil
}
```

## Check Guidelines

1. **Naming Convention**:
   - Use a namespace prefix for your checks (e.g., `custom.`, `infra.`, `security.`)
   - Use descriptive names in snake_case
   - Example: `custom.http_endpoint`, `security.ssl_cert_expiry`

2. **Parameter Handling**:
   - Always validate required parameters
   - Provide sensible defaults for optional parameters
   - Document all parameters in comments
   - Use type assertions safely with the ok-idiom

3. **Error Handling**:
   - Return errors for configuration/setup issues
   - Use the result map for check-specific failures
   - Provide clear, actionable error messages

4. **Result Format**:
   - The result map must include:
     - `status`: "Success" or "Failure"
     - `output`: A descriptive message
   - Optional fields:
     - `error`: Detailed error information
     - Any additional context-specific fields

## Advanced Example

Here's a more complex example that checks SSL certificate expiry:

```go
package security

import (
    "crypto/tls"
    "fmt"
    "net/http"
    "time"
    "github.com/seastar-consulting/checkers/checks"
)

func init() {
    checks.Register("security.ssl_cert_expiry", "Check SSL certificate expiration", CheckSSLCertExpiry)
}

// CheckSSLCertExpiry verifies that an SSL certificate is valid and not expiring soon
func CheckSSLCertExpiry(params map[string]interface{}) (map[string]interface{}, error) {
    // Required parameters
    host, ok := params["host"].(string)
    if !ok {
        return nil, fmt.Errorf("host parameter is required")
    }

    // Optional parameters with defaults
    warningDays := 30
    if days, ok := params["warning_days"].(float64); ok {
        warningDays = int(days)
    }

    // Create custom HTTP client that doesn't verify certificates
    // (we'll do our own verification)
    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        },
    }

    // Connect to the host
    conn, err := tls.Dial("tcp", host+":443", &tls.Config{
        InsecureSkipVerify: true,
    })
    if err != nil {
        return map[string]interface{}{
            "status": "Failure",
            "output": fmt.Sprintf("Failed to connect to %s: %v", host, err),
        }, nil
    }
    defer conn.Close()

    // Get the certificate
    cert := conn.ConnectionState().PeerCertificates[0]
    daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

    if daysUntilExpiry <= 0 {
        return map[string]interface{}{
            "status": "Failure",
            "output": fmt.Sprintf("Certificate for %s has expired", host),
            "expiry_date": cert.NotAfter,
            "issuer": cert.Issuer.CommonName,
        }, nil
    }

    if daysUntilExpiry <= warningDays {
        return map[string]interface{}{
            "status": "Failure",
            "output": fmt.Sprintf("Certificate for %s will expire in %d days", host, daysUntilExpiry),
            "expiry_date": cert.NotAfter,
            "issuer": cert.Issuer.CommonName,
        }, nil
    }

    return map[string]interface{}{
        "status": "Success",
        "output": fmt.Sprintf("Certificate for %s is valid for %d more days", host, daysUntilExpiry),
        "expiry_date": cert.NotAfter,
        "issuer": cert.Issuer.CommonName,
    }, nil
}
```

## Using Custom Checks

1. Create your check in a new package under the `checks` directory
2. Register it in the package's `init()` function
3. Use it in your `checks.yaml`:

```yaml
checks:
  - name: verify-api-health
    type: custom.http_endpoint
    params:
      url: "https://api.example.com/health"
      expected_status: 200

  - name: check-ssl-cert
    type: security.ssl_cert_expiry
    params:
      host: "example.com"
      warning_days: 14
```

## Testing Your Checks

Here's an example of how to test your custom checks:

```go
package custom

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestCheckHTTPEndpoint(t *testing.T) {
    // Create a test server
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer ts.Close()

    // Test cases
    tests := []struct {
        name           string
        params         map[string]interface{}
        expectedStatus string
        expectError    bool
    }{
        {
            name: "successful check",
            params: map[string]interface{}{
                "url": ts.URL,
                "expected_status": float64(200),
            },
            expectedStatus: "Success",
            expectError:    false,
        },
        {
            name: "missing url",
            params: map[string]interface{}{
                "expected_status": float64(200),
            },
            expectedStatus: "",
            expectError:    true,
        },
    }

    // Run tests
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := CheckHTTPEndpoint(tt.params)

            if tt.expectError {
                if err == nil {
                    t.Error("expected error but got none")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }

            status, ok := result["status"].(string)
            if !ok {
                t.Error("status not found in result")
                return
            }

            if status != tt.expectedStatus {
                t.Errorf("expected status %s but got %s", tt.expectedStatus, status)
            }
        })
    }
}
```

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

3. **Testing**:
   - Write unit tests for your checks
   - Test both success and failure cases
   - Mock external dependencies
   - Test parameter validation

4. **Performance**:
   - Set appropriate timeouts
   - Clean up resources (close connections, files)
   - Consider concurrent execution
   - Cache results when appropriate

5. **Security**:
   - Validate and sanitize input parameters
   - Handle sensitive information securely
   - Follow the principle of least privilege
   - Document security implications
