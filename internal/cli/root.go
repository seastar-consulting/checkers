package cli

import (
	"context"
	"fmt"
	"github.com/seastar-consulting/checkers/types"
	"io"
	"sync"
	"time"

	"github.com/seastar-consulting/checkers/internal/config"
	"github.com/seastar-consulting/checkers/internal/executor"
	"github.com/seastar-consulting/checkers/internal/ui"
	"github.com/spf13/cobra"
)

const defaultTimeout = 30 * time.Second

// Options holds the command line options
type Options struct {
	ConfigFile string
	Verbose    bool
	Timeout    time.Duration
}

// NewRootCommand creates the root command for the CLI
func NewRootCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "checker",
		Short: "A CLI tool to read and process checks from a YAML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts, cmd.OutOrStdout())
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "checks.yaml", "config file path")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "enable verbose logging")
	cmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", defaultTimeout, "timeout for each check")

	return cmd
}

func run(ctx context.Context, opts *Options, stdout io.Writer) error {
	// Initialize components
	configMgr := config.NewManager(opts.ConfigFile)
	executor := executor.NewExecutor(opts.Timeout)
	formatter := ui.NewFormatter(opts.Verbose)

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create wait group for concurrent execution
	var wg sync.WaitGroup
	results := make([]types.CheckResult, len(cfg.Checks))

	// Execute checks concurrently
	for i, check := range cfg.Checks {
		wg.Add(1)
		go func(i int, check types.CheckItem) {
			defer wg.Done()

			// Execute check and get result
			result, err := executor.ExecuteCheck(ctx, check)
			if err != nil {
				results[i] = types.CheckResult{
					Name:   check.Name,
					Type:   check.Type,
					Status: types.Error,
					Error:  err.Error(),
				}
				return
			}

			results[i] = result
		}(i, check)
	}

	// Wait for all checks to complete
	wg.Wait()

	// Format and print results
	fmt.Fprint(stdout, formatter.FormatResults(results))

	return nil
}
