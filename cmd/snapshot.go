package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var snapshotFile string

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Save or compare drift snapshots",
}

var snapshotSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Run drift detection and save results to a snapshot file",
	RunE: func(cmd *cobra.Command, args []string) error {
		results, err := runDetection(cmd)
		if err != nil {
			return err
		}
		if err := snapshot.Save(snapshotFile, results); err != nil {
			return fmt.Errorf("failed to save snapshot: %w", err)
		}
		fmt.Fprintf(os.Stdout, "snapshot saved to %s\n", snapshotFile)
		return nil
	},
}

var snapshotDiffCmd = &cobra.Command{
	Use:   "diff <previous-snapshot>",
	Short: "Compare current drift results against a previous snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prevPath := args[0]

		prev, err := snapshot.Load(prevPath)
		if err != nil {
			return fmt.Errorf("failed to load previous snapshot: %w", err)
		}

		results, err := runDetection(cmd)
		if err != nil {
			return err
		}

		curr := &snapshot.Snapshot{Results: results}
		entries := snapshot.Diff(prev, curr)
		fmt.Fprint(os.Stdout, snapshot.Summary(entries))
		return nil
	},
}

func init() {
	snapshotSaveCmd.Flags().StringVarP(&snapshotFile, "output", "o", "drift-snapshot.json", "path to save the snapshot")
	addDetectionFlags(snapshotSaveCmd)
	addDetectionFlags(snapshotDiffCmd)

	snapshotCmd.AddCommand(snapshotSaveCmd)
	snapshotCmd.AddCommand(snapshotDiffCmd)
	rootCmd.AddCommand(snapshotCmd)
}
