package os

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seastar-consulting/checkers/types"
	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (string, func())
		checkItem types.CheckItem
		want      types.CheckResult
		wantErr   bool
	}{
		{
			name: "existing_file",
			setupFunc: func() (string, func()) {
				f, err := os.CreateTemp("", "test")
				if err != nil {
					t.Fatal(err)
				}
				return f.Name(), func() { os.Remove(f.Name()) }
			},
			checkItem: types.CheckItem{
				Name:       "test-check",
				Type:       "os.file_exists",
				Parameters: map[string]string{}, // path will be added in test
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "os.file_exists",
				Status: types.Success,
				// Output will be checked separately due to dynamic path
			},
		},
		{
			name: "non-existing_file",
			setupFunc: func() (string, func()) {
				return filepath.Join(os.TempDir(), "nonexistent"), func() {}
			},
			checkItem: types.CheckItem{
				Name:       "test-check",
				Type:       "os.file_exists",
				Parameters: map[string]string{}, // path will be added in test
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "os.file_exists",
				Status: types.Failure,
				// Output will be checked separately due to dynamic path
			},
		},
		{
			name: "missing_path_parameter",
			setupFunc: func() (string, func()) {
				return "", func() {}
			},
			checkItem: types.CheckItem{
				Name:       "test-check",
				Type:       "os.file_exists",
				Parameters: map[string]string{},
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "os.file_exists",
				Status: types.Error,
				Error:  "path parameter is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setupFunc()
			defer cleanup()

			if path != "" {
				tt.checkItem.Parameters["path"] = path
			}

			got, err := CheckFileExists(tt.checkItem)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For cases with dynamic paths, check the basic fields first
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.Status, got.Status)
			assert.Equal(t, tt.want.Error, got.Error)

			// For success/failure cases, verify the output contains the path
			if path != "" {
				assert.Contains(t, got.Output, path)
			}
		})
	}
}
