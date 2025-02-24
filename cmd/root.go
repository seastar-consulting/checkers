package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
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
		Short:         "A CLI tool to run developer workstation diagnostics",
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

	cmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "checks.yaml", "config file path")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "enable verbose logging")
	cmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", defaultTimeout, "timeout for each check")

	cmd.PersistentFlags().StringVarP(&outputFormatStr, "output", "o", string(types.OutputFormatPretty),
		fmt.Sprintf("output format. One of: %s", strings.Join(supportedFormats, ", ")))

	// Parse the output format before running the command
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		opts.OutputFormat = types.OutputFormat(outputFormatStr)
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
	if opts.OutputFormat == types.OutputFormatJSON {
		// Sort results by name for consistent output
		sortedResults := make([]types.CheckResult, len(results))
		copy(sortedResults, results)
		sort.Slice(sortedResults, func(i, j int) bool {
			return sortedResults[i].Name < sortedResults[j].Name
		})

		// Get system information
		osInfo := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
		metadata := types.OutputMetadata{
			DateTime: time.Now().Format(time.RFC3339),
			Version:  version.GetVersion(),
			OS:       osInfo,
		}
		output = formatter.FormatResultsJSON(sortedResults, metadata)
	} else {
		output = formatter.FormatResults(results)
	}

	// Write output to stdout for both formats
	if _, err := cmd.OutOrStdout().Write([]byte(output)); err != nil {
		// Always show critical errors, even in non-verbose mode
		fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] Failed to write output: %v\n", err)
		return fmt.Errorf("output error: %w", err)
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
