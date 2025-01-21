package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/internal/processor"
	"github.com/seastar-consulting/checkers/types"
)

// Executor handles the execution of checks
type Executor struct {
	timeout   time.Duration
	processor *processor.Processor
}

// NewExecutor creates a new Executor instance
func NewExecutor(timeout time.Duration) *Executor {
	return &Executor{
		timeout:   timeout,
		processor: processor.NewProcessor(),
	}
}

// ExecuteCheck executes a single check and returns the result
func (e *Executor) ExecuteCheck(ctx context.Context, check types.CheckItem) (types.CheckResult, error) {
	// Create a new context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Check if this is a native check
	if checkFunc, ok := checks.Registry[check.Type]; ok {
		// Run internal check with timeout
		resultChan := make(chan types.CheckResult, 1)
		errChan := make(chan error, 1)

		go func() {
			result, err := checkFunc.Func(check)
			resultChan <- result
			errChan <- err
		}()

		// Wait for either completion or timeout
		select {
		case <-ctxWithTimeout.Done():
			if ctxWithTimeout.Err() == context.DeadlineExceeded {
				return types.CheckResult{
					Name:   check.Name,
					Type:   check.Type,
					Status: types.Error,
					Output: "command execution timed out",
				}, context.DeadlineExceeded
			}
			return types.CheckResult{}, ctxWithTimeout.Err()
		case err := <-errChan:
			result := <-resultChan
			if err != nil {
				return types.CheckResult{
					Name:   check.Name,
					Type:   check.Type,
					Status: types.Error,
					Error:  fmt.Sprintf("failed to execute check: %v", err),
				}, nil
			}

			// Add name and type if not set
			if result.Name == "" {
				result.Name = check.Name
			}
			if result.Type == "" {
				result.Type = check.Type
			}

			return result, nil
		}
	}

	// Handle command-based check
	if check.Type != "command" {
		return types.CheckResult{
			Name:   check.Name,
			Type:   check.Type,
			Status: types.Error,
			Output: fmt.Sprintf("unsupported check type: %s", check.Type),
		}, nil
	}

	if check.Command == "" {
		return types.CheckResult{
			Name:   check.Name,
			Type:   check.Type,
			Status: types.Error,
			Output: "no command specified",
		}, nil
	}

	// Prepare command
	cmd := exec.CommandContext(ctxWithTimeout, "bash", "-c", "set -eo pipefail; "+check.Command)
	if check.Parameters != nil {
		for key, value := range check.Parameters {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start command
	if err := cmd.Start(); err != nil {
		return types.CheckResult{
			Name:   check.Name,
			Type:   check.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("failed to start command: %v", err),
		}, nil
	}

	// Wait for command with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for either command completion or timeout
	select {
	case <-ctxWithTimeout.Done():
		// Kill the process if it's still running
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		if ctxWithTimeout.Err() == context.DeadlineExceeded {
			return types.CheckResult{
				Name:   check.Name,
				Type:   check.Type,
				Status: types.Error,
				Output: "command execution timed out",
			}, context.DeadlineExceeded
		}
		return types.CheckResult{}, ctxWithTimeout.Err()
	case err := <-done:
		// Get command output
		output := strings.TrimSpace(stdout.String())
		if stderr.Len() > 0 {
			if output != "" {
				output += "\n"
			}
			output += strings.TrimSpace(stderr.String())
		}

		// Handle command execution errors
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Create a direct CheckResult for exit error
				return types.CheckResult{
					Name:   check.Name,
					Type:   check.Type,
					Status: types.Error,
					Output: output,
					Error:  fmt.Sprintf("command failed with exit code %d", exitErr.ExitCode()),
				}, nil
			}
			// Create a direct CheckResult for other errors
			return types.CheckResult{
				Name:   check.Name,
				Type:   check.Type,
				Status: types.Error,
				Error:  err.Error(),
			}, nil
		}

		// Try to parse output as JSON first
		var jsonOutput map[string]interface{}
		if err := json.Unmarshal([]byte(output), &jsonOutput); err == nil {
			// If output is valid JSON, let processor handle it
			return e.processor.ProcessOutput(check.Name, check.Type, jsonOutput), nil
		}

		// If not JSON, create a simple output map
		rawOutput := map[string]interface{}{
			"output": output,
		}

		// Process the raw output into a CheckResult
		return e.processor.ProcessOutput(check.Name, check.Type, rawOutput), nil
	}
}
