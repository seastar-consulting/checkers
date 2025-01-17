package executor

import (
	"context"
	"testing"
	"time"

	"github.com/seastar-consulting/checkers/types"

	"github.com/stretchr/testify/assert"
)

func TestExecutor_ExecuteCheck(t *testing.T) {
	tests := []struct {
		name    string
		check   types.CheckItem
		want    types.CheckResult
		wantErr bool
	}{
		{
			name: "successful echo command",
			check: types.CheckItem{
				Name:    "echo-test",
				Type:    "command",
				Command: `echo '{"status":"success","output":"test output"}'`,
			},
			want: types.CheckResult{
				Name:   "echo-test",
				Type:   "command",
				Status: types.Success,
				Output: "test output",
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
			want: types.CheckResult{
				Name:   "sleep-test",
				Type:   "command",
				Status: types.Error,
				Output: "command execution timed out",
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
			want: types.CheckResult{
				Name:   "invalid-command",
				Type:   "command",
				Status: types.Error,
				Output: "bash: line 1: nonexistentcommand: command not found",
				Error:  "command failed with exit code 127",
			},
			wantErr: false,
		},
		{
			name: "empty command",
			check: types.CheckItem{
				Name: "empty-command",
				Type: "command",
			},
			want: types.CheckResult{
				Name:   "empty-command",
				Type:   "command",
				Status: types.Error,
				Output: "no command specified",
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
			want: types.CheckResult{
				Name:   "param-test",
				Type:   "command",
				Status: types.Success,
				Output: "test-value",
			},
			wantErr: false,
		},
		{
			name: "command exit code 1",
			check: types.CheckItem{
				Name:    "test",
				Type:    "command",
				Command: "exit 1",
			},
			want: types.CheckResult{
				Name:   "test",
				Type:   "command",
				Status: types.Error,
				Output: "",
				Error:  "command failed with exit code 1",
			},
			wantErr: false,
		},
		{
			name: "pipeline failure",
			check: types.CheckItem{
				Name:    "test",
				Type:    "command",
				Command: "exit 1 | echo hello",
			},
			want: types.CheckResult{
				Name:   "test",
				Type:   "command",
				Status: types.Error,
				Output: "hello",
				Error:  "command failed with exit code 1",
			},
			wantErr: false,
		},
		{
			name: "invalid json output",
			check: types.CheckItem{
				Name:    "invalid-json",
				Type:    "command",
				Command: `echo '{"status":"success","output":invalid_json}'`,
			},
			want: types.CheckResult{
				Name:   "invalid-json",
				Type:   "command",
				Status: types.Success,
				Output: `{"status":"success","output":invalid_json}`,
			},
			wantErr: false,
		},
		{
			name: "unsupported check type",
			check: types.CheckItem{
				Name: "unsupported",
				Type: "unsupported",
			},
			want: types.CheckResult{
				Name:   "unsupported",
				Type:   "unsupported",
				Status: types.Error,
				Output: "unsupported check type: unsupported",
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
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExecutor_ExecuteCheckCancellation(t *testing.T) {
	e := NewExecutor(5 * time.Second)
	check := types.CheckItem{
		Name:    "sleep-test",
		Type:    "command",
		Command: "sleep 2",
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		result, err := e.ExecuteCheck(ctx, check)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Equal(t, types.CheckResult{}, result)
		close(done)
	}()

	// Cancel the context after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the goroutine to finish
	select {
	case <-done:
		// Test passed
	case <-time.After(2 * time.Second):
		t.Fatal("test timed out")
	}
}
