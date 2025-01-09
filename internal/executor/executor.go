package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/seastar-consulting/checkers/internal/types"
)

// Executor handles the execution of checks
type Executor struct {
	timeout   time.Duration
	processor types.Processor
}

// NewExecutor creates a new Executor instance
func NewExecutor(timeout time.Duration, processor types.Processor) *Executor {
	return &Executor{
		timeout:   timeout,
		processor: processor,
	}
}

// ExecuteCheck executes a single check and returns its result
func (e *Executor) ExecuteCheck(check types.CheckItem) types.CheckResult {
	// Handle command-based check
	if check.Type != "command" {
		return types.CheckResult{
			Name:   check.Name,
			Status: types.Error,
			Error:  fmt.Sprintf("unsupported check type: %s", check.Type),
		}
	}

	if check.Command == "" {
		return types.CheckResult{
			Name:   check.Name,
			Status: types.Error,
			Error:  "no command specified",
		}
	}

	// Create command
	cmd := exec.Command("bash", "-c", check.Command)

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range check.Parameters {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Create output buffer
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	// Start command
	if err := cmd.Start(); err != nil {
		return types.CheckResult{
			Name:   check.Name,
			Status: types.Error,
			Error:  fmt.Errorf("failed to start command: %w", err).Error(),
		}
	}

	// Create done channel
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for command to finish or timeout
	select {
	case err := <-done:
		if err != nil {
			return types.CheckResult{
				Name:   check.Name,
				Status: types.Error,
				Error:  fmt.Errorf("command failed: %w", err).Error(),
			}
		}
	case <-time.After(e.timeout):
		if err := cmd.Process.Kill(); err != nil {
			return types.CheckResult{
				Name:   check.Name,
				Status: types.Error,
				Error:  "failed to kill timed out process",
			}
		}
		return types.CheckResult{
			Name:   check.Name,
			Status: types.Error,
			Error:  "command timed out",
		}
	}

	// Process output
	return e.processor.ProcessOutput(check.Name, check.Type, output.Bytes())
}
