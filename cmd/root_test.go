package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd.Use != "checkers" {
		t.Errorf("NewRootCommand().Use = %s, want checkers", cmd.Use)
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
    type: command
    command: echo '{"status":"success","output":"test output"}'
`

	validConfigWithTimeout := `
timeout: 5s
checks:
  - name: test-check
    type: command
    command: echo '{"status":"success","output":"test output"}'
`

	invalidConfig := `
checks:
  - name: test-check
    type: test
    command: [invalid yaml`

	timeoutConfig := `
checks:
  - name: slow-check
    type: command
    command: sleep 2 && echo '{"status":"success","output":"test output"}'
`

	multipleChecksConfig := `
checks:
  - name: slow-check-1
    type: command
    command: "sleep 1"
  - name: slow-check-2
    type: command
    command: "sleep 1"
`

	multipleSlowChecksConfig := `
checks:
  - name: slow-check-1
    type: command
    command: "sleep 3"
  - name: slow-check-2
    type: command
    command: "sleep 3"
  - name: slow-check-3
    type: command
    command: "sleep 3"
  - name: fast-check
    type: command
    command: "echo hello"
`

	tests := []struct {
		name        string
		configYAML  string
		opts        *Options
		wantErr     bool
		errContains string
		checkOutput func(t *testing.T, output string)
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
			name:       "valid config with timeout",
			configYAML: validConfigWithTimeout,
			opts: &Options{
				Verbose: true,
				Timeout: time.Second, // This should be overridden by config
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
				ConfigFile: "checks.yaml",
				Timeout:    time.Second,
			},
			wantErr:     true,
			errContains: "did not find expected ',' or ']'",
		},
		{
			name:       "check timeout",
			configYAML: timeoutConfig,
			opts: &Options{
				Verbose: true,
				Timeout: 500 * time.Millisecond,
			},
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
		{
			name:       "multiple checks with individual timeouts",
			configYAML: multipleChecksConfig,
			opts: &Options{
				Verbose: true,
				Timeout: 500 * time.Millisecond,
			},
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
		{
			name:       "all slow checks timeout",
			configYAML: multipleSlowChecksConfig,
			opts: &Options{
				Verbose: true,
				Timeout: 500 * time.Millisecond,
			},
			wantErr:     true,
			errContains: "context deadline exceeded",
			checkOutput: func(t *testing.T, output string) {
				// Verify that all slow checks show as timed out
				for i := 1; i <= 3; i++ {
					if !strings.Contains(output, fmt.Sprintf("slow-check-%d", i)) {
						t.Errorf("output missing timed out check: slow-check-%d", i)
					}
					if !strings.Contains(output, "command execution timed out") {
						t.Errorf("output missing timeout message for check: slow-check-%d", i)
					}
				}
			},
		},
		{
			name: "config timeout takes precedence when flag not set",
			opts: &Options{
				ConfigFile: "test-config.yaml",
			},
			configYAML: `
timeout: 5s
checks:
  - name: quick-check
    type: command
    command: "echo hello"
`,
			wantErr: false,
		},
		{
			name: "command-line timeout overrides config timeout",
			opts: &Options{
				ConfigFile: "test-config.yaml",
				Timeout:    2 * time.Second,
			},
			configYAML: `
timeout: 5s
checks:
  - name: quick-check
    type: command
    command: "echo hello"
`,
			wantErr: false,
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
			cmd := &cobra.Command{}
			cmd.SetContext(context.Background())
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.Flags().Bool("verbose", tt.opts.Verbose, "")
			cmd.Flags().String("config", tt.opts.ConfigFile, "")
			cmd.Flags().Duration("timeout", tt.opts.Timeout, "")
			err := run(cmd, tt.opts)

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

			// Additional checks for timeout tests
			output := buf.String()
			if tt.name == "check timeout" && !tt.wantErr {
				if !strings.Contains(output, "command execution timed out") {
					t.Errorf("expected timeout message in output, got: %s", output)
				}
			} else if tt.name == "multiple checks with individual timeouts" && !tt.wantErr {
				if !strings.Contains(output, "command execution timed out") {
					t.Errorf("expected timeout message in output, got: %s", output)
				}
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
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
