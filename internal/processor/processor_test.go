package processor

import (
	"testing"

	"github.com/seastar-consulting/checkers/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_ProcessOutput(t *testing.T) {
	tests := []struct {
		name      string
		checkName string
		checkType string
		output    []byte
		want      types.CheckResult
	}{
		{
			name:      "success result",
			checkName: "test-check",
			checkType: "test",
			output:    []byte(`{"status": "success", "output": "test output"}`),
			want: types.CheckResult{
				Name:   "test-check",
				Status: types.Success,
				Output: "test output",
			},
		},
		{
			name:      "failure result",
			checkName: "test-check",
			checkType: "test",
			output:    []byte(`{"status": "failure", "output": "test failed"}`),
			want: types.CheckResult{
				Name:   "test-check",
				Status: types.Failure,
				Output: "test failed",
			},
		},
		{
			name:      "warning result",
			checkName: "test-check",
			checkType: "test",
			output:    []byte(`{"status": "warning", "output": "test warning"}`),
			want: types.CheckResult{
				Name:   "test-check",
				Status: types.Warning,
				Output: "test warning",
			},
		},
		{
			name:      "error result",
			checkName: "test-check",
			checkType: "test",
			output:    []byte(`{"error": "test error"}`),
			want: types.CheckResult{
				Name:   "test-check",
				Status: types.Error,
				Error:  "test error",
			},
		},
		{
			name:      "missing status",
			checkName: "test-check",
			checkType: "test",
			output:    []byte(`{"output": "test output"}`),
			want: types.CheckResult{
				Name:   "test-check",
				Status: types.Error,
				Error:  "missing status field",
			},
		},
		{
			name:      "invalid status",
			checkName: "test-check",
			checkType: "test",
			output:    []byte(`{"status": "invalid", "output": "test output"}`),
			want: types.CheckResult{
				Name:   "test-check",
				Status: types.Error,
				Output: "test output",
			},
		},
		{
			name:      "invalid json",
			checkName: "test-check",
			checkType: "test",
			output:    []byte(`invalid json`),
			want: types.CheckResult{
				Name:   "test-check",
				Status: types.Error,
				Error:  "invalid character 'i' looking for beginning of value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor()
			got := p.ProcessOutput(tt.checkName, tt.checkType, tt.output)
			assert.Equal(t, tt.want, got)
		})
	}
}
