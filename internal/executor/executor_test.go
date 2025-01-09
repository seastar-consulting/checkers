package executor

import (
	"testing"
	"time"

	"github.com/seastar-consulting/checkers/internal/processor"
	"github.com/seastar-consulting/checkers/internal/types"
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
			name: "successful check",
			check: types.CheckItem{
				Name:    "test",
				Type:    "command",
				Command: "echo '{\"status\": \"success\", \"output\": \"test output\"}'",
			},
			want: types.CheckResult{
				Name:   "test",
				Status: types.Success,
				Output: "test output",
			},
		},
		{
			name: "failing check",
			check: types.CheckItem{
				Name:    "test",
				Type:    "command",
				Command: "echo '{\"status\": \"failure\", \"error\": \"test error\"}' && exit 1",
			},
			want: types.CheckResult{
				Name:   "test",
				Status: types.Error,
				Error:  "command failed: exit status 1",
			},
		},
		{
			name: "invalid json output",
			check: types.CheckItem{
				Name:    "test",
				Type:    "command",
				Command: "echo 'invalid json'",
			},
			want: types.CheckResult{
				Name:   "test",
				Status: types.Error,
				Error:  "invalid character 'i' looking for beginning of value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create processor
			p := processor.NewProcessor()

			// Create executor with processor
			e := NewExecutor(5*time.Second, p)

			// Execute check
			got := e.ExecuteCheck(tt.check)

			// Compare results
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExecutor_ExecuteCheck_Timeout(t *testing.T) {
	// Create processor
	p := processor.NewProcessor()

	// Create executor with short timeout
	e := NewExecutor(100*time.Millisecond, p)

	// Create a check that will timeout
	check := types.CheckItem{
		Name:    "test",
		Type:    "command",
		Command: "sleep 1",
	}

	// Execute check
	result := e.ExecuteCheck(check)

	// Verify result
	assert.Equal(t, types.Error, result.Status)
	assert.Equal(t, "command timed out", result.Error)
}

func TestExecutor_ExecuteCheck_Cancellation(t *testing.T) {
	// Create processor
	p := processor.NewProcessor()

	// Create executor with short timeout
	e := NewExecutor(100*time.Millisecond, p)

	// Create a check that will timeout
	check := types.CheckItem{
		Name:    "test",
		Type:    "command",
		Command: "sleep 1",
	}

	// Execute check
	result := e.ExecuteCheck(check)

	// Verify result
	assert.Equal(t, types.Error, result.Status)
	assert.Equal(t, "command timed out", result.Error)
}
