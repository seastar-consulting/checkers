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

func TestCheckExecutableExists(t *testing.T) {
	// Create a temporary executable file
	tmpDir := t.TempDir()
	execPath := filepath.Join(tmpDir, "test-exec")
	err := os.WriteFile(execPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		params     map[string]string
		wantStatus types.CheckStatus
		wantError  bool
	}{
		{
			name: "missing name parameter",
			params: map[string]string{
				"custom_path": tmpDir,
			},
			wantStatus: types.Error,
			wantError:  false,
		},
		{
			name: "executable exists in custom path",
			params: map[string]string{
				"name":        "test-exec",
				"custom_path": tmpDir,
			},
			wantStatus: types.Success,
			wantError:  false,
		},
		{
			name: "executable not found",
			params: map[string]string{
				"name": "non-existent-executable",
			},
			wantStatus: types.Failure,
			wantError:  false,
		},
		{
			name: "executable exists in PATH",
			params: map[string]string{
				"name": "sh", // should exist on any Unix system
			},
			wantStatus: types.Success,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := types.CheckItem{
				Name:       "test",
				Type:      "os.executable_exists",
				Parameters: tt.params,
			}

			got, err := CheckExecutableExists(item)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckExecutableExists() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got.Status != tt.wantStatus {
				t.Errorf("CheckExecutableExists() status = %v, want %v", got.Status, tt.wantStatus)
			}
		})
	}
}
