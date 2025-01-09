package cli

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

// Options represents the command line options
type Options struct {
	ConfigFile string
	Timeout    time.Duration
	Verbose    bool
}

// NewRootCommand creates the root command for the CLI
func NewRootCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "checkers",
		Short: "Run system checks",
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
	// Create config manager
	configMgr := config.NewManager(opts.ConfigFile)

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		return err
	}

	// Create processor
	processor := processor.NewProcessor()

	// Create executor
	executor := executor.NewExecutor(opts.Timeout, processor)

	// Execute checks
	results := make([]types.CheckResult, 0, len(cfg.Checks))
	checkTypes := make(map[string]string)

	for _, check := range cfg.Checks {
		result := executor.ExecuteCheck(check)
		results = append(results, result)
		checkTypes[check.Name] = check.Type
	}

	// Format and print results
	formatter := ui.NewFormatter(opts.Verbose)
	output := formatter.FormatResults(results, checkTypes)
	fmt.Fprint(stdout, output)

	return nil
}
