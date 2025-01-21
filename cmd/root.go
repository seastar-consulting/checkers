package cmd

import (
	"context"
	"fmt"
	"io"
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
			return run(cmd.Context(), opts, cmd.OutOrStdout())
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

	// Process checks from config
	var results []types.CheckResult
	for _, checkItem := range cfg.Checks {
		result, err := executor.ExecuteCheck(ctx, checkItem)
		if err != nil {
			return fmt.Errorf("failed to execute check %s: %w", checkItem.Name, err)
		}

		results = append(results, result)
	}

	// Format and write all results
	output := formatter.FormatResults(results)
	_, err = fmt.Fprint(stdout, output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
