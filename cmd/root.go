package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/seastar-consulting/checkers/internal/config"
	"github.com/seastar-consulting/checkers/internal/executor"
	"github.com/seastar-consulting/checkers/internal/ui"
	"github.com/seastar-consulting/checkers/types"
	"github.com/spf13/cobra"
)

const defaultTimeout = 30 * time.Second

// Options holds the command line options
type Options struct {
	ConfigFile string
	Verbose    bool
	Timeout    time.Duration
}

var rootCmd *cobra.Command

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// NewRootCommand creates and returns a new root command
func NewRootCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "checkers",
		Short: "A CLI tool to run developer workstation diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "checks.yaml", "config file path")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "enable verbose logging")
	cmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", defaultTimeout, "timeout for each check")

	return cmd
}

func init() {
	rootCmd = NewRootCommand()
}

func run(cmd *cobra.Command, opts *Options) error {
	startTime := time.Now()
	defer func() {
		totalRuntime := time.Since(startTime)
		fmt.Printf("DEBUG: Total runtime: %v\n", totalRuntime)
		if opts.Timeout > 0 && totalRuntime > opts.Timeout*3/2 {
			fmt.Printf("WARNING: Total runtime (%v) exceeded timeout (%v) by more than 50%%\n", totalRuntime, opts.Timeout)
		}
	}()

	// Initialize components
	configMgr := config.NewManager(opts.ConfigFile)

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine timeout
	timeout := opts.Timeout
	if !cmd.Flags().Changed("timeout") && cfg.Timeout != nil {
		timeout = *cfg.Timeout
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
	
	// Start all checks concurrently
	for _, checkItem := range cfg.Checks {
		checkItem := checkItem // Create new variable for goroutine
		go func() {
			result, err := executor.ExecuteCheck(ctx, checkItem)
			resultChan <- checkResult{result: result, err: err, item: checkItem}
		}()
	}

	// Collect results
	var results []types.CheckResult
	var timedOutChecks []types.CheckItem
	remainingChecks := len(cfg.Checks)

	for remainingChecks > 0 {
		select {
		case <-ctx.Done():
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
						Output: "command execution timed out",
					})
					timedOutChecks = append(timedOutChecks, check)
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
					Output: "command execution timed out",
				})
			} else if res.err != nil {
				results = append(results, types.CheckResult{
					Name:   res.item.Name,
					Type:   res.item.Type,
					Status: types.Error,
					Output: fmt.Sprintf("check failed: %v", res.err),
				})
			} else {
				results = append(results, res.result)
			}
		}
	}

	// Format and write all results
	output := formatter.FormatResults(results)
	_, err = fmt.Fprint(cmd.OutOrStdout(), output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if len(timedOutChecks) > 0 {
		return context.DeadlineExceeded
	}
	return nil
}
