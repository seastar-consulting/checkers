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
		checkNames  []string
	}{
		{
			name: "valid config with command",
			configYAML: `
checks:
  - name: test-check
    type: test
    command: echo "test"
`,
			wantErr:    false,
			wantChecks: 1,
			checkNames: []string{"test-check"},
		},
		{
			name: "valid config with parameters",
			configYAML: `
checks:
  - name: test-check
    type: test
    parameters:
      key: value
`,
			wantErr:    false,
			wantChecks: 1,
			checkNames: []string{"test-check"},
		},
		{
			name: "valid config with items",
			configYAML: `
checks:
  - name: test-check
    type: test
    items:
      - key: value1
      - key: value2
`,
			wantErr:    false,
			wantChecks: 2,
			checkNames: []string{"test-check: 1", "test-check: 2"},
		},
		{
			name: "invalid_missing_required_field",
			configYAML: `
checks:
  - name: test-check
    type: test
`,
			wantErr:     true,
			errContains: "must have exactly one of 'command', 'parameters', or 'items' fields",
		},
		{
			name: "invalid_command_and_parameters",
			configYAML: `
checks:
  - name: test-check
    type: test
    command: echo "test"
    parameters:
      key: value
`,
			wantErr:     true,
			errContains: "cannot have multiple of 'command', 'parameters', and 'items' fields",
		},
		{
			name: "invalid_command_and_items",
			configYAML: `
checks:
  - name: test-check
    type: test
    command: echo "test"
    items:
      - key: value
`,
			wantErr:     true,
			errContains: "cannot have multiple of 'command', 'parameters', and 'items' fields",
		},
		{
			name: "invalid_parameters_and_items",
			configYAML: `
checks:
  - name: test-check
    type: test
    parameters:
      key: value
    items:
      - key: value
`,
			wantErr:     true,
			errContains: "cannot have multiple of 'command', 'parameters', and 'items' fields",
		},
		{
			name: "invalid_all_three_fields",
			configYAML: `
checks:
  - name: test-check
    type: test
    command: echo "test"
    parameters:
      key: value
    items:
      - key: value
`,
			wantErr:     true,
			errContains: "cannot have multiple of 'command', 'parameters', and 'items' fields",
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
		{
			name: "invalid empty item parameters",
			configYAML: `
checks:
  - name: test-check
    type: test
    items:
      - {}
`,
			wantErr:     true,
			errContains: "must have parameters",
		},
		{
			name: "valid config with items and name template",
			configYAML: `
checks:
  - name: "Check binary: {{ .name }}"
    type: test
    items:
      - path: else
        name: git
      - name: docker
`,
			wantErr:    false,
			wantChecks: 2,
			checkNames: []string{"Check binary: git", "Check binary: docker"},
		},
		{
			name: "invalid template syntax",
			configYAML: `
checks:
  - name: "Check binary: {{ .name"
    type: test
    items:
      - name: git
`,
			wantErr:     true,
			errContains: "invalid template in check name",
		},
		{
			name: "missing template field",
			configYAML: `
checks:
  - name: "Check binary: {{ .missing }}"
    type: test
    items:
      - name: git
`,
			wantErr:     true,
			errContains: "failed to render check name template",
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

			if tt.checkNames != nil {
				for i, want := range tt.checkNames {
					if i >= len(config.Checks) {
						t.Errorf("Load() missing check at index %d, want name %s", i, want)
						continue
					}
					if got := config.Checks[i].Name; got != want {
						t.Errorf("Load() check[%d].Name = %v, want %v", i, got, want)
					}
				}
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
