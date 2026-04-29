package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/driftwatch/internal/pinned"
	"github.com/spf13/cobra"
)

var pinnedStorePath string

func init() {
	pinnedCmd := &cobra.Command{
		Use:   "pinned",
		Short: "Manage pinned (approved) drift entries",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all pinned drift entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := pinned.Load(pinnedStorePath)
			if err != nil {
				return err
			}
			if len(store.Entries) == 0 {
				fmt.Println("No pinned entries.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tIMAGE\tPINNED AT\tCOMMENT")
			for _, e := range store.Entries {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					e.Name,
					e.Image,
					e.PinnedAt.Format(time.RFC3339),
					e.Comment,
				)
			}
			return w.Flush()
		},
	}

	var unpinName string
	unpinCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a pinned entry by container name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if unpinName == "" {
				return fmt.Errorf("--name is required")
			}
			store, err := pinned.Load(pinnedStorePath)
			if err != nil {
				return err
			}
			if !pinned.Unpin(store, unpinName) {
				return fmt.Errorf("no pinned entry found for %q", unpinName)
			}
			if err := pinned.Save(pinnedStorePath, store); err != nil {
				return err
			}
			fmt.Printf("Removed pinned entry for %q\n", unpinName)
			return nil
		},
	}
	unpinCmd.Flags().StringVar(&unpinName, "name", "", "Container name to unpin")

	pinnedCmd.PersistentFlags().StringVar(&pinnedStorePath, "store", ".driftwatch-pinned.json", "Path to pinned store file")
	pinnedCmd.AddCommand(listCmd, unpinCmd)
	rootCmd.AddCommand(pinnedCmd)
}
