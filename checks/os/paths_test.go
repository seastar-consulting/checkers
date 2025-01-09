package os

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func() (string, func())
		expectedStatus string
		expectError    bool
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
			expectedStatus: "Success",
		},
		{
			name: "non-existing_file",
			setupFunc: func() (string, func()) {
				return filepath.Join(os.TempDir(), "nonexistent"), func() {}
			},
			expectedStatus: "Failure",
		},
		{
			name: "missing_path_parameter",
			setupFunc: func() (string, func()) {
				return "", func() {}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setupFunc()
			defer cleanup()

			params := map[string]interface{}{}
			if path != "" {
				params["path"] = path
			}

			result, err := FileExists(params)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			status, ok := result["status"].(string)
			if !ok {
				t.Error("status not found in result")
				return
			}

			if status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}
