package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/driftwatch/internal/tag"
)

var defaultTagStore = "driftwatch-tags.json"

func init() {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage named tags for drift snapshots",
	}

	addCmd := &cobra.Command{
		Use:   "add <name> <snapshot-path>",
		Short: "Tag a snapshot with a name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, _ := cmd.Flags().GetString("note")
			store, _ := cmd.Flags().GetString("store")
			if err := tag.Save(store, args[0], args[1], note); err != nil {
				return fmt.Errorf("tag add: %w", err)
			}
			fmt.Printf("Tagged snapshot %q as %q\n", args[1], args[0])
			return nil
		},
	}
	addCmd.Flags().String("note", "", "optional note for the tag")
	addCmd.Flags().String("store", defaultTagStore, "path to tag store file")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, _ := cmd.Flags().GetString("store")
			tags, err := tag.List(store)
			if err != nil {
				return fmt.Errorf("tag list: %w", err)
			}
			if len(tags) == 0 {
				fmt.Println("No tags found.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tSNAPSHOT\tCREATED\tNOTE")
			for _, t := range tags {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					t.Name, t.SnapshotPath,
					t.CreatedAt.Format(time.RFC3339), t.Note)
			}
			return w.Flush()
		},
	}
	listCmd.Flags().String("store", defaultTagStore, "path to tag store file")

	delCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a tag by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, _ := cmd.Flags().GetString("store")
			if err := tag.Delete(store, args[0]); err != nil {
				return fmt.Errorf("tag delete: %w", err)
			}
			fmt.Printf("Deleted tag %q\n", args[0])
			return nil
		},
	}
	delCmd.Flags().String("store", defaultTagStore, "path to tag store file")

	tagCmd.AddCommand(addCmd, listCmd, delCmd)
	rootCmd.AddCommand(tagCmd)
}
