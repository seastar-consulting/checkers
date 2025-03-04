package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/seastar-consulting/checkers/types"
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

func TestConcurrentExecution(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "concurrent-test.yaml")

	// Create a config with multiple checks that have measurable execution times
	config := `
checks:
  - name: concurrent-check-1
    type: command
    command: "sleep 0.5 && echo '{\"status\":\"success\",\"output\":\"check 1\"}'"
  - name: concurrent-check-2
    type: command
    command: "sleep 0.5 && echo '{\"status\":\"success\",\"output\":\"check 2\"}'"
  - name: concurrent-check-3
    type: command
    command: "sleep 0.5 && echo '{\"status\":\"success\",\"output\":\"check 3\"}'"
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

	// Set command line arguments with a timeout that would fail if checks run sequentially
	cmd.SetArgs([]string{
		"--config", configPath,
		"--verbose",
		"--timeout", "1s", // This should be enough for concurrent execution but not for sequential
	})

	// Record start time
	start := time.Now()

	// Execute the command
	err = cmd.Execute()
	if err != nil {
		t.Errorf("command execution failed: %v", err)
		return
	}

	// Check execution time
	executionTime := time.Since(start)
	if executionTime >= 1500*time.Millisecond {
		t.Errorf("checks appear to run sequentially, took %v", executionTime)
	}

	// Check output for all checks
	output := outBuf.String()
	for i := 1; i <= 3; i++ {
		checkName := fmt.Sprintf("concurrent-check-%d", i)
		if !strings.Contains(output, checkName) {
			t.Errorf("output missing check %s", checkName)
		}
		if !strings.Contains(output, fmt.Sprintf("check %d", i)) {
			t.Errorf("output missing result for check %d", i)
		}
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

func TestOutputFormat(t *testing.T) {
	tests := []struct {
		name         string
		format       string
		wantInStdout bool
		wantJSON     bool
	}{
		{
			name:         "pretty format goes to stdout",
			format:       "pretty",
			wantInStdout: true,
			wantJSON:     false,
		},
		{
			name:         "json format goes to stdout",
			format:       "json",
			wantJSON:     true,
			wantInStdout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for test files
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "checks.yaml")

			// Create a minimal config file
			configContent := `
checks:
  - name: test-check
    type: command
    command: echo "test output"
`
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Create buffers for stdout and stderr
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			// Create and configure the command
			cmd := NewRootCommand()
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			cmd.SetArgs([]string{
				"--config", configPath,
				"--output", tt.format,
			})

			// Run the command
			if err := cmd.Execute(); err != nil {
				t.Fatalf("cmd.Execute() error = %v", err)
			}

			// Check stdout
			gotStdout := stdout.String()
			if tt.wantInStdout {
				if gotStdout == "" {
					t.Error("Expected output in stdout, got empty string")
				}

				if tt.wantJSON {
					// Verify JSON structure
					var output types.JSONOutput
					if err := json.Unmarshal([]byte(gotStdout), &output); err != nil {
						t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, gotStdout)
					}

					// Verify results
					if len(output.Results) != 1 || output.Results[0].Name != "test-check" {
						t.Errorf("Expected one result with name 'test-check', got: %+v", output.Results)
					}

					// Verify metadata
					if output.Metadata.Version != "v1.2.3-test" {
						t.Errorf("Expected version v1.2.3-test in metadata, got: %s", output.Metadata.Version)
					}
					if output.Metadata.DateTime == "" {
						t.Error("Expected datetime in metadata")
					}
					if output.Metadata.OS == "" {
						t.Error("Expected OS info in metadata")
					}
				} else {
					if !strings.Contains(gotStdout, "test-check") {
						t.Errorf("Expected pretty output in stdout, got: %s", gotStdout)
					}
				}
			}

			// Check stderr - should only contain debug/error messages if any
			gotStderr := stderr.String()
			if strings.Contains(gotStderr, "test-check") {
				t.Errorf("Found check output in stderr, should be in stdout. Stderr: %s", gotStderr)
			}
		})
	}
}

func TestOutputFile(t *testing.T) {
	tests := []struct {
		name           string
		outputFlag     string
		fileFlag       string
		expectedFormat string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "file with json extension",
			fileFlag:       "output.json",
			expectedFormat: "json",
		},
		{
			name:           "file with txt extension",
			fileFlag:       "output.txt",
			expectedFormat: "pretty",
		},
		{
			name:           "file with log extension",
			fileFlag:       "output.log",
			expectedFormat: "pretty",
		},
		{
			name:           "file with out extension",
			fileFlag:       "output.out",
			expectedFormat: "pretty",
		},
		{
			name:           "file with no extension",
			fileFlag:       "output",
			expectedFormat: "pretty",
		},
		{
			name:        "file with unsupported extension",
			fileFlag:    "output.csv",
			wantErr:     true,
			errContains: "unsupported file extension",
		},
		{
			name:           "output flag takes precedence over file extension (json)",
			outputFlag:     "json",
			fileFlag:       "output.txt",
			expectedFormat: "json",
		},
		{
			name:           "output flag takes precedence over file extension (pretty)",
			outputFlag:     "pretty",
			fileFlag:       "output.json",
			expectedFormat: "pretty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for test files
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "checks.yaml")
			outputPath := filepath.Join(tmpDir, tt.fileFlag)

			// Create a minimal config file
			configContent := `
checks:
  - name: test-check
    type: command
    command: echo "test output"
`
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Create buffers for stdout and stderr
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			// Create and configure the command
			cmd := NewRootCommand()
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			// Build command arguments
			args := []string{
				"--config", configPath,
				"--file", outputPath,
			}

			// Add output flag if specified
			if tt.outputFlag != "" {
				args = append(args, "--output", tt.outputFlag)
			}

			cmd.SetArgs(args)

			// Run the command
			err := cmd.Execute()

			// Check for expected errors
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("Expected error containing %q, got %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("cmd.Execute() error = %v", err)
			}

			// Check that stdout is empty (output should go to file)
			gotStdout := stdout.String()
			if gotStdout != "" {
				t.Errorf("Expected empty stdout when using --file, got: %s", gotStdout)
			}

			// Check that the file was created
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Fatalf("Output file was not created: %v", err)
			}

			// Read the file content
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			fileContent := string(content)
			if fileContent == "" {
				t.Error("Expected content in output file, got empty string")
			}

			// Verify the format of the content
			if tt.expectedFormat == "json" {
				// Verify JSON structure
				var output types.JSONOutput
				if err := json.Unmarshal(content, &output); err != nil {
					t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, fileContent)
				}

				// Verify results
				if len(output.Results) != 1 || output.Results[0].Name != "test-check" {
					t.Errorf("Expected one result with name 'test-check', got: %+v", output.Results)
				}
			} else {
				// Pretty format
				if !strings.Contains(fileContent, "test-check") {
					t.Errorf("Expected pretty output in file, got: %s", fileContent)
				}
			}
		})
	}
}
