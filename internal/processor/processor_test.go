package processor

import (
	"reflect"
	"testing"

	"github.com/seastar-consulting/checkers/internal/types"
)

func TestProcessor_ProcessOutput(t *testing.T) {
	tests := []struct {
		name      string
		checkName string
		checkType string
		output    map[string]interface{}
		want      types.CheckResult
	}{
		{
			name:      "success status",
			checkName: "test-check",
			checkType: "test",
			output: map[string]interface{}{
				"status": "success",
				"output": "test output",
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Success,
				Output: "test output",
			},
		},
		{
			name:      "failure status",
			checkName: "test-check",
			checkType: "test",
			output: map[string]interface{}{
				"status": "failure",
				"output": "test failed",
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Failure,
				Output: "test failed",
			},
		},
		{
			name:      "warning status",
			checkName: "test-check",
			checkType: "test",
			output: map[string]interface{}{
				"status": "warning",
				"output": "test warning",
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Warning,
				Output: "test warning",
			},
		},
		{
			name:      "error present",
			checkName: "test-check",
			checkType: "test",
			output: map[string]interface{}{
				"error": "something went wrong",
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Error,
				Error:  "something went wrong",
			},
		},
		{
			name:      "unknown status",
			checkName: "test-check",
			checkType: "test",
			output: map[string]interface{}{
				"status": "unknown",
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Error,
				Error:  "unknown status: unknown",
			},
		},
		{
			name:      "no status but has output",
			checkName: "test-check",
			checkType: "test",
			output: map[string]interface{}{
				"output": "test output",
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Success,
				Output: "test output",
			},
		},
		{
			name:      "empty output",
			checkName: "test-check",
			checkType: "test",
			output:    map[string]interface{}{},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Error,
				Error:  "no status or output provided",
			},
		},
	}

	p := NewProcessor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ProcessOutput(tt.checkName, tt.checkType, tt.output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
