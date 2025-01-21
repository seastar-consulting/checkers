package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd.Use != "checker" {
		t.Errorf("NewRootCommand().Use = %v, want %v", cmd.Use, "checker")
	}

	// Test default flag values
	configFlag := cmd.Flag("config")
	if configFlag == nil {
		t.Fatal("config flag not found")
	}
	if configFlag.DefValue != "checks.yaml" {
		t.Errorf("config flag default = %v, want %v", configFlag.DefValue, "checks.yaml")
	}

	verboseFlag := cmd.Flag("verbose")
	if verboseFlag == nil {
		t.Fatal("verbose flag not found")
	}
	if verboseFlag.DefValue != "false" {
		t.Errorf("verbose flag default = %v, want %v", verboseFlag.DefValue, "false")
	}

	timeoutFlag := cmd.Flag("timeout")
	if timeoutFlag == nil {
		t.Fatal("timeout flag not found")
	}
	if timeoutFlag.DefValue != defaultTimeout.String() {
		t.Errorf("timeout flag default = %v, want %v", timeoutFlag.DefValue, defaultTimeout.String())
	}
}

func TestRun(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	validConfig := `
checks:
  - name: test-check
    type: test
    command: echo '{"status":"success","output":"test output"}'
`

	invalidConfig := `
checks:
  - name: test-check
    type: test
    command: [invalid yaml`

	tests := []struct {
		name        string
		configYAML  string
		opts        *Options
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid config",
			configYAML: validConfig,
			opts: &Options{
				Verbose: true,
				Timeout: time.Second,
			},
			wantErr: false,
		},
		{
			name:       "invalid config path",
			configYAML: "",
			opts: &Options{
				ConfigFile: "nonexistent.yaml",
				Timeout:    time.Second,
			},
			wantErr:     true,
			errContains: "no such file",
		},
		{
			name:       "invalid yaml",
			configYAML: invalidConfig,
			opts: &Options{
				ConfigFile: "checks.yaml", // Set default config file
				Timeout:    time.Second,
			},
			wantErr:     true,
			errContains: "did not find expected ',' or ']'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file if content provided
			if tt.configYAML != "" {
				configPath := filepath.Join(tmpDir, tt.name+".yaml")
				err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
				if err != nil {
					t.Fatalf("failed to write test config: %v", err)
				}
				tt.opts.ConfigFile = configPath
			}

			var buf bytes.Buffer
			err := run(context.Background(), tt.opts, &buf)

			if tt.wantErr {
				if err == nil {
					t.Error("run() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("run() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("run() unexpected error = %v", err)
			}
		})
	}
}

func TestCommandExecution(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	config := `
checks:
  - name: test-check
    type: command
    command: echo '{"status":"success","output":"test output"}'
`

	err := os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Create a command with a buffer for output
	cmd := NewRootCommand()
	outBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)

	// Set command line arguments
	cmd.SetArgs([]string{
		"--config", configPath,
		"--verbose",
		"--timeout", "1s",
	})

	// Execute the command
	err = cmd.Execute()
	if err != nil {
		t.Errorf("command execution failed: %v", err)
		return
	}

	// Check output
	output := outBuf.String()
	expectedOutputs := []string{
		"test-check",
		"test output",
		"✅", // Check for success indicator
	}
	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("command output missing expected content %q, got: %s", expected, output)
		}
	}
}
