package py

import (
	"testing"

	"github.com/seastar-consulting/checkers/types"
	"github.com/stretchr/testify/assert"
)

type pythonCheckType struct {
	module   string
	function string
}

func TestExecutePythonCheck(t *testing.T) {
	tests := []struct {
		name       string
		checkItem  types.CheckItem
		wantStatus string
		wantOutput string
		wantError  string
	}{
		{
			name: "successful count keys",
			checkItem: types.CheckItem{
				Name: "test count keys",
				Type: "py.pack.box:count_keys",
				Parameters: map[string]string{
					"param1": "value1",
					"param2": "value2",
				},
			},
			wantStatus: "success",
			wantOutput: "Found 2 keys in parameters",
		},
		{
			name: "empty parameters",
			checkItem: types.CheckItem{
				Name:       "test empty params",
				Type:       "py.pack.box:count_keys",
				Parameters: map[string]string{},
			},
			wantStatus: "success",
			wantOutput: "Found 0 keys in parameters",
		},
		{
			name: "invalid python module",
			checkItem: types.CheckItem{
				Name: "test invalid module",
				Type: "py.nonexistent.module:function",
				Parameters: map[string]string{
					"param1": "value1",
				},
			},
			wantStatus: "error",
			wantError:  "failed to import Python module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExecutePythonCheck(tt.checkItem)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, got.Status)

			if tt.wantOutput != "" {
				assert.Equal(t, tt.wantOutput, got.Output)
			}
			if tt.wantError != "" {
				assert.Contains(t, got.Error, tt.wantError)
			}
		})
	}
}

// func TestParsePythonCheckType(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		input   string
// 		want    pythonCheckType
// 		wantErr bool
// 	}{
// 		{
// 			name:  "valid check type",
// 			input: "py.pack.box:count_keys",
// 			want: pythonCheckType{
// 				module:   "py.pack.box",
// 				function: "count_keys",
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "invalid format - no colon",
// 			input:   "py.pack.box.count_keys",
// 			wantErr: true,
// 		},
// 		{
// 			name:    "invalid format - empty module",
// 			input:   ":count_keys",
// 			wantErr: true,
// 		},
// 		{
// 			name:    "invalid format - empty function",
// 			input:   "py.pack.box:",
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := parsePythonCheckType(tt.input)
// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				return
// 			}
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.want, got)
// 		})
// 	}
// }
