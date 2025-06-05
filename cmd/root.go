package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/seastar-consulting/checkers/internal/config"
	"github.com/seastar-consulting/checkers/internal/executor"
	"github.com/seastar-consulting/checkers/internal/ui"
	"github.com/seastar-consulting/checkers/internal/version"
	"github.com/seastar-consulting/checkers/types"
	"github.com/spf13/cobra"
)

const defaultTimeout = 30 * time.Second

// Options holds the command line options
type Options struct {
	ConfigFile   string
	Verbose      bool
	Timeout      time.Duration
	OutputFormat types.OutputFormat
	OutputFile   string
}

var (
	// debugLog is used for debug messages
	debugLog = log.New(io.Discard, "[DEBUG] ", log.Ltime)
	// errorLog is used for error messages
	errorLog        = log.New(io.Discard, "[ERROR] ", log.Ltime)
	rootCmd         *cobra.Command
	outputFormatStr string
)

// ErrChecksFailure indicates that one or more checks have failed
var ErrChecksFailure = fmt.Errorf("one or more checks failed")

func init() {
	rootCmd = NewRootCommand()
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// NewRootCommand creates and returns a new root command
func NewRootCommand() *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:           "checkers",
		Short:         "A CLI tool to run diagnostics",
		Version:       version.Version,
		SilenceUsage:  true, // Don't show usage on errors not related to usage
		SilenceErrors: true, // We handle error output ourselves
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate output format
			if !opts.OutputFormat.IsValid() {
				supported := make([]string, 0, len(types.SupportedOutputFormats()))
				for _, f := range types.SupportedOutputFormats() {
					supported = append(supported, string(f))
				}
				return fmt.Errorf("invalid output format: %s (supported formats: %s)", opts.OutputFormat, strings.Join(supported, ", "))
			}
			return run(cmd, opts)
		},
	}

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.Usage()
		return err
	})

	// Convert supported formats to string slice
	supportedFormats := make([]string, 0, len(types.SupportedOutputFormats()))
	for _, f := range types.SupportedOutputFormats() {
		supportedFormats = append(supportedFormats, string(f))
	}

	// Create a map of file extensions to output formats
	formatExtensions := map[string]types.OutputFormat{
		".json": types.OutputFormatJSON,
		".html": types.OutputFormatHTML,
		".txt":  types.OutputFormatPretty,
		".log":  types.OutputFormatPretty,
		".out":  types.OutputFormatPretty,
	}

	cmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "checks.yaml", "config file path")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "enable verbose logging")
	cmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", defaultTimeout, "timeout for each check")

	cmd.PersistentFlags().StringVarP(&outputFormatStr, "output", "o", string(types.OutputFormatPretty),
		fmt.Sprintf("output format. One of: %s", strings.Join(supportedFormats, ", ")))
	cmd.PersistentFlags().StringVarP(&opts.OutputFile, "file", "f", "",
		"output file path. Format will be determined by file extension (.json for JSON, .html for HTML, any other for pretty)")

	// Parse the output format before running the command
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// First set the output format from the --output flag
		opts.OutputFormat = types.OutputFormat(outputFormatStr)

		// If output file is specified but --output flag was not explicitly set,
		// determine format from file extension
		if opts.OutputFile != "" && !cmd.Flags().Changed("output") {
			ext := strings.ToLower(filepath.Ext(opts.OutputFile))

			// Check if extension maps to a specific format
			if format, exists := formatExtensions[ext]; exists {
				opts.OutputFormat = format
			} else if ext != "" {
				// If extension exists but is not in formatExtensions
				// Build a list of all supported extensions
				var supportedExts []string
				for extension := range formatExtensions {
					supportedExts = append(supportedExts, extension)
				}
				return fmt.Errorf("unsupported file extension: %s (supported extensions: %s)", ext, strings.Join(supportedExts, ", "))
			} else {
				// No extension, use pretty format
				opts.OutputFormat = types.OutputFormatPretty
			}
			// Update outputFormatStr to match the determined format
			outputFormatStr = string(opts.OutputFormat)
		}

		if !opts.OutputFormat.IsValid() {
			return fmt.Errorf("invalid output format: %s", outputFormatStr)
		}
		return nil
	}

	return cmd
}

func run(cmd *cobra.Command, opts *Options) error {
	// Configure loggers based on verbose flag
	if opts.Verbose {
		debugLog.SetOutput(cmd.ErrOrStderr())
		errorLog.SetOutput(cmd.ErrOrStderr())
	} else {
		// In non-verbose mode, discard all logs
		debugLog.SetOutput(io.Discard)
		errorLog.SetOutput(io.Discard)
	}

	startTime := time.Now()
	defer func() {
		totalRuntime := time.Since(startTime)
		debugLog.Printf("Total runtime: %v", totalRuntime)
		if opts.Timeout > 0 && totalRuntime > opts.Timeout*3/2 {
			// Always show performance warnings, even in non-verbose mode
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Performance warning: Total runtime (%v) exceeded timeout (%v) by more than 50%%\n", totalRuntime, opts.Timeout)
		}
	}()

	// Initialize components
	configMgr := config.NewManager(opts.ConfigFile)

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		// Always show critical errors, even in non-verbose mode
		fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] Failed to load configuration file '%s': %v\n", opts.ConfigFile, err)
		return fmt.Errorf("configuration error: %w", err)
	}

	// Determine timeout
	timeout := opts.Timeout
	if !cmd.Flags().Changed("timeout") && cfg.Timeout != nil {
		timeout = *cfg.Timeout
		debugLog.Printf("Using timeout from configuration file: %v", timeout)
	}

	// Create a context with timeout for all checks
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	executor := executor.NewExecutor(timeout)
	formatter := ui.NewFormatter(opts.Verbose)

	// Create channels for results and errors
	type checkResult struct {
		result types.CheckResult
		err    error
		item   types.CheckItem
	}
	resultChan := make(chan checkResult, len(cfg.Checks))

	debugLog.Printf("Starting execution of %d checks", len(cfg.Checks))

	// Start all checks concurrently
	for _, checkItem := range cfg.Checks {
		checkItem := checkItem // Create new variable for goroutine
		go func() {
			debugLog.Printf("Executing check: %s", checkItem.Name)
			result, err := executor.ExecuteCheck(ctx, checkItem)
			resultChan <- checkResult{result: result, err: err, item: checkItem}
		}()
	}

	// Collect results
	var results []types.CheckResult
	var timedOutChecks []types.CheckItem
	var failedChecks []string
	remainingChecks := len(cfg.Checks)

	for remainingChecks > 0 {
		select {
		case <-ctx.Done():
			debugLog.Printf("Global timeout reached after %v", time.Since(startTime))
			// Add timeout results for all remaining checks
			for _, check := range cfg.Checks {
				found := false
				for _, res := range results {
					if res.Name == check.Name {
						found = true
						break
					}
				}
				if !found {
					results = append(results, types.CheckResult{
						Name:   check.Name,
						Type:   check.Type,
						Status: types.Error,
						Output: "check execution timed out",
					})
					timedOutChecks = append(timedOutChecks, check)
					failedChecks = append(failedChecks, check.Name)
					debugLog.Printf("Check '%s' timed out", check.Name)
				}
			}
			remainingChecks = 0
		case res := <-resultChan:
			remainingChecks--
			if res.err == context.DeadlineExceeded {
				timedOutChecks = append(timedOutChecks, res.item)
				results = append(results, types.CheckResult{
					Name:   res.item.Name,
					Type:   res.item.Type,
					Status: types.Error,
					Output: "check execution timed out",
				})
				failedChecks = append(failedChecks, res.item.Name)
				debugLog.Printf("Check '%s' timed out", res.item.Name)
			} else if res.err != nil {
				results = append(results, types.CheckResult{
					Name:   res.item.Name,
					Type:   res.item.Type,
					Status: types.Error,
					Output: fmt.Sprintf("check failed: %v", res.err),
				})
				failedChecks = append(failedChecks, res.item.Name)
				debugLog.Printf("Check '%s' failed: %v", res.item.Name, res.err)
			} else if res.result.Status != types.Success {
				failedChecks = append(failedChecks, res.item.Name)
				results = append(results, res.result)
				debugLog.Printf("Check '%s' failed with status: %s", res.item.Name, res.result.Status)
			} else {
				results = append(results, res.result)
				debugLog.Printf("Check '%s' completed successfully", res.item.Name)
			}
		}
	}

	// Format and write all results
	var output string

	// Sort results by name for consistent output
	sortedResults := make([]types.CheckResult, len(results))
	copy(sortedResults, results)
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Name < sortedResults[j].Name
	})

	// Get system information once
	osInfo := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	metadata := types.OutputMetadata{
		DateTime: time.Now().Format(time.RFC3339),
		Version:  version.GetVersion(),
		OS:       osInfo,
	}

	// Map output formats to their respective formatting functions
	formatFuncs := map[types.OutputFormat]ui.FormatFunc{
		types.OutputFormatJSON:   formatter.FormatResultsJSON,
		types.OutputFormatHTML:   formatter.FormatResultsHTML,
		types.OutputFormatPretty: formatter.FormatResultsPretty,
	}

	// Get the appropriate formatting function and execute it
	if formatFunc, ok := formatFuncs[opts.OutputFormat]; ok {
		output = formatFunc(sortedResults, metadata)
	} else {
		// Fallback to pretty format if format is not supported
		output = formatter.FormatResultsPretty(sortedResults, metadata)
	}

	// Write output to stdout or file
	if opts.OutputFile != "" {
		// Create parent directories if they don't exist
		dir := filepath.Dir(opts.OutputFile)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] Failed to create directory for output file: %v\n", err)
				return fmt.Errorf("output error: %w", err)
			}
		}

		// Write to file
		if err := os.WriteFile(opts.OutputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] Failed to write to output file '%s': %v\n", opts.OutputFile, err)
			return fmt.Errorf("output error: %w", err)
		}
		debugLog.Printf("Output written to file: %s", opts.OutputFile)
	} else {
		// Write output to stdout
		if _, err := cmd.OutOrStdout().Write([]byte(output)); err != nil {
			// Always show critical errors, even in non-verbose mode
			fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] Failed to write output: %v\n", err)
			return fmt.Errorf("output error: %w", err)
		}
	}

	if len(timedOutChecks) > 0 {
		// Show summary in non-verbose mode
		if !opts.Verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] %d checks timed out\n", len(timedOutChecks))
		}
		return context.DeadlineExceeded
	}

	if len(failedChecks) > 0 {
		// Show detailed failures only in verbose mode
		debugLog.Printf("%d checks failed: %v", len(failedChecks), failedChecks)
		// Show summary in non-verbose mode
		if !opts.Verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] %d checks failed\n", len(failedChecks))
		}
		return ErrChecksFailure
	}

	debugLog.Printf("All checks completed successfully")
	return nil
}
