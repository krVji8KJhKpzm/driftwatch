package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/driftwatch/internal/container"
	"github.com/driftwatch/internal/watch"
	"github.com/spf13/cobra"
)

var (
	watchInterval string
	watchFormat   string
	watchOutput   string
)

func init() {
	watchCmd := &cobra.Command{
		Use:   "watch [manifest]",
		Short: "Continuously watch for configuration drift",
		Args:  cobra.ExactArgs(1),
		RunE:  runWatch,
	}

	watchCmd.Flags().StringVarP(&watchInterval, "interval", "i", "30s", "polling interval (e.g. 10s, 1m)")
	watchCmd.Flags().StringVarP(&watchFormat, "format", "f", "text", "output format: text or json")
	watchCmd.Flags().StringVarP(&watchOutput, "output", "o", "", "write output to file instead of stdout")

	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	interval, err := time.ParseDuration(watchInterval)
	if err != nil {
		return fmt.Errorf("invalid interval %q: %w", watchInterval, err)
	}

	cfg := watch.Config{
		ManifestPath: args[0],
		Interval:     interval,
		Format:       watchFormat,
		Output:       watchOutput,
	}

	inspector := container.NewInspector(nil)
	runner := watch.NewRunner(inspector)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stderr, "Watching %s every %s. Press Ctrl+C to stop.\n", args[0], interval)

	if err := runner.Run(ctx, cfg); err != nil && err != context.Canceled {
		return err
	}
	return nil
}
