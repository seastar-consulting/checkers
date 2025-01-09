package executor

import (
	"context"
	"testing"
	"time"

	"github.com/seastar-consulting/checkers/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_ExecuteCheck(t *testing.T) {
	tests := []struct {
		name    string
		check   types.CheckItem
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "successful echo command",
			check: types.CheckItem{
				Name:    "echo-test",
				Type:    "command",
				Command: `echo '{"status":"success","output":"test output"}'`,
			},
			want: map[string]interface{}{
				"name":    "echo-test",
				"status":  "success",
				"output":  "test output",
			},
			wantErr: false,
		},
		{
			name: "command timeout",
			check: types.CheckItem{
				Name:    "sleep-test",
				Type:    "command",
				Command: "sleep 2",
			},
			want: map[string]interface{}{
				"name":   "sleep-test",
				"status": "error",
				"error":  "command timed out",
			},
			wantErr: false,
		},
		{
			name: "invalid command",
			check: types.CheckItem{
				Name:    "invalid-command",
				Type:    "command",
				Command: "nonexistentcommand",
			},
			want: map[string]interface{}{
				"name":   "invalid-command",
				"status": "error",
				"error":  "bash: line 1: nonexistentcommand: command not found",
			},
			wantErr: false,
		},
		{
			name: "empty command",
			check: types.CheckItem{
				Name: "empty-command",
				Type: "command",
			},
			want: map[string]interface{}{
				"name":   "empty-command",
				"status": "error",
				"error":  "no command specified",
			},
			wantErr: false,
		},
		{
			name: "command with parameters",
			check: types.CheckItem{
				Name:    "param-test",
				Type:    "command",
				Command: "echo $TEST_PARAM",
				Parameters: map[string]string{
					"TEST_PARAM": "test-value",
				},
			},
			want: map[string]interface{}{
				"name":    "param-test",
				"status":  "success",
				"output":  "test-value",
			},
			wantErr: false,
		},
		{
			name: "failing command with exit 1",
			check: types.CheckItem{
				Name:    "test",
				Type:    "command",
				Command: "exit 1",
			},
			want: map[string]interface{}{
				"name":     "test",
				"status":   "error",
				"error":    "command failed with exit code 1",
				"exitCode": 1,
			},
			wantErr: false,
		},
		{
			name: "failing command with pipe",
			check: types.CheckItem{
				Name:    "test",
				Type:    "command",
				Command: "exit 1 | echo hello",
			},
			want: map[string]interface{}{
				"name":     "test",
				"status":   "error",
				"error":    "command failed with exit code 1",
				"exitCode": 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExecutor(1 * time.Second)
			got, err := e.ExecuteCheck(context.Background(), tt.check)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want["status"], got["status"], "status mismatch")
			if errMsg, ok := tt.want["error"].(string); ok {
				assert.Contains(t, got["error"], errMsg, "error message mismatch")
			}
			if output, ok := tt.want["output"].(string); ok {
				assert.Contains(t, got["output"], output, "output mismatch")
			}
			if exitCode, ok := tt.want["exitCode"].(int); ok {
				assert.Equal(t, exitCode, got["exitCode"], "exit code mismatch")
			}
		})
	}
}

func TestExecutor_ExecuteCheckCancellation(t *testing.T) {
	e := NewExecutor(time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to signal when the command starts
	started := make(chan struct{})

	// Create a channel to receive the result
	done := make(chan error)

	go func() {
		close(started) // Signal that we're about to start the command
		_, err := e.ExecuteCheck(ctx, types.CheckItem{
			Name:    "cancel-test",
			Type:    "command",
			Command: "sleep 5",
		})
		done <- err
	}()

	<-started // Wait for the command to start
	time.Sleep(time.Millisecond * 50)
	cancel() // Cancel the context

	select {
	case err := <-done:
		if err == nil {
			t.Error("ExecuteCheck() error = nil, want context canceled error")
		}
	case <-time.After(time.Second):
		t.Error("ExecuteCheck() did not respond to cancellation")
	}
}
