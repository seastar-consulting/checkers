package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManager_Load(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		configYAML  string
		wantErr     bool
		wantChecks  int
		errContains string
	}{
		{
			name: "valid config",
			configYAML: `
checks:
  - name: test-check
    type: test
    command: echo "test"
`,
			wantErr:    false,
			wantChecks: 1,
		},
		{
			name: "empty checks",
			configYAML: `
checks: []
`,
			wantErr:     true,
			errContains: "no checks defined",
		},
		{
			name: "missing check name",
			configYAML: `
checks:
  - type: test
    command: echo "test"
`,
			wantErr:     true,
			errContains: "check name is required",
		},
		{
			name: "missing check type",
			configYAML: `
checks:
  - name: test-check
    command: echo "test"
`,
			wantErr:     true,
			errContains: "check type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary config file
			configPath := filepath.Join(tmpDir, tt.name+".yaml")
			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			m := NewManager(configPath)
			config, err := m.Load()

			if tt.wantErr {
				if err == nil {
					t.Error("Load() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error = %v", err)
				return
			}

			if len(config.Checks) != tt.wantChecks {
				t.Errorf("Load() got %v checks, want %v", len(config.Checks), tt.wantChecks)
			}
		})
	}
}

func TestManager_LoadNonExistentFile(t *testing.T) {
	m := NewManager("non-existent-file.yaml")
	_, err := m.Load()
	if err == nil {
		t.Error("Load() error = nil, want error for non-existent file")
	}
}

func TestManager_LoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	err := os.WriteFile(configPath, []byte("invalid: yaml: content"), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	m := NewManager(configPath)
	_, err = m.Load()
	if err == nil {
		t.Error("Load() error = nil, want error for invalid YAML")
	}
}
