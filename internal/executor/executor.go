package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/seastar-consulting/checkers/internal/types"
)

// Executor handles the execution of checks
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates a new Executor instance
func NewExecutor(timeout time.Duration) *Executor {
	return &Executor{
		timeout: timeout,
	}
}

// ExecuteCheck executes a single check and returns the raw output
func (e *Executor) ExecuteCheck(ctx context.Context, check types.CheckItem) (map[string]interface{}, error) {
	if check.Command == "" {
		return map[string]interface{}{
			"name":   check.Name,
			"status": "error",
			"error":  "no command specified",
		}, nil
	}

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Wrap command with pipefail and execute through bash
	wrappedCmd := fmt.Sprintf("set -eo pipefail; %s", check.Command)
	cmd := exec.CommandContext(ctx, "bash", "-c", wrappedCmd)

	// Set environment variables for parameters
	if check.Parameters != nil {
		for key, value := range check.Parameters {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	var outputBuf, errBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &errBuf

	result := map[string]interface{}{
		"name": check.Name,
	}

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result["status"] = "error"
			result["error"] = "command timed out"
			return result, nil
		}

		if ctx.Err() == context.Canceled {
			return nil, ctx.Err()
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			errMsg := strings.TrimSpace(errBuf.String())
			if errMsg == "" {
				errMsg = fmt.Sprintf("command failed with exit code %d", exitCode)
			}
			result["status"] = "error"
			result["error"] = errMsg
			result["exitCode"] = exitCode
			return result, nil
		}

		result["status"] = "error"
		result["error"] = err.Error()
		return result, nil
	}

	// Try to parse output as JSON first
	var outputMap map[string]interface{}
	if err := json.Unmarshal(outputBuf.Bytes(), &outputMap); err == nil {
		// Merge command output with result
		for k, v := range outputMap {
			result[k] = v
		}
		if _, ok := result["status"]; !ok {
			result["status"] = "success"
		}
	} else {
		// Raw output
		result["status"] = "success"
		output := strings.TrimSpace(outputBuf.String())
		stderr := strings.TrimSpace(errBuf.String())

		if output != "" {
			result["output"] = output
		}
		if stderr != "" {
			result["error"] = stderr
		}
	}

	return result, nil
}
