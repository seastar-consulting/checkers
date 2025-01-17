package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/seastar-consulting/checkers/internal/config"
	"github.com/seastar-consulting/checkers/internal/executor"
	"github.com/seastar-consulting/checkers/internal/processor"
	"github.com/seastar-consulting/checkers/internal/types"
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

var rootCmd *cobra.Command

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	opts := &Options{}

	rootCmd = &cobra.Command{
		Use:   "checker",
		Short: "A CLI tool to read and process checks from a YAML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts, cmd.OutOrStdout())
		},
	}

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "checks.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", defaultTimeout, "timeout for each check")
}

func run(ctx context.Context, opts *Options, stdout io.Writer) error {
	// Initialize components
	configMgr := config.NewManager(opts.ConfigFile)
	executor := executor.NewExecutor(opts.Timeout)
	processor := processor.NewProcessor()
	formatter := ui.NewFormatter(opts.Verbose)

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Process checks from config
	var results []types.CheckResult
	for _, checkItem := range cfg.Checks {
		result, err := executor.ExecuteCheck(ctx, checkItem)
		if err != nil {
			return fmt.Errorf("failed to execute check %s: %w", checkItem.Name, err)
		}

		processed := processor.ProcessOutput(checkItem.Name, checkItem.Type, result)
		results = append(results, processed)
	}

	// Format and write all results
	output := formatter.FormatResults(results)
	_, err = fmt.Fprint(stdout, output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
