package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// colFlag holds the collection name provided via the --col persistent flag.
var colFlag string

// newRootCmd creates the root cobra command.
// Subcommands will be added in Phase 2.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "ghostman",
		Short: "Terminal-native API client — local-first alternative to Postman",
		Long: `ghostman is a terminal-native API client and testing tool.
Requests, collections, and environments are stored as markdown files
with YAML frontmatter. Every command is composable via Unix pipes.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("no command specified — run 'ghostman --help' for usage")
		},
	}

	// Version flag
	root.Version = "0.1.0-dev"

	// Persistent flag for collection name (used by set_collection)
	root.PersistentFlags().StringVar(&colFlag, "col", "", "collection name for set_collection")

	root.AddCommand(newRequestCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newGetCmd())
	root.AddCommand(newAssertCmd())
	root.AddCommand(newSetEnvCmd())
	root.AddCommand(newSetCollectionCmd())
	root.AddCommand(newChainCmd())
	root.AddCommand(newScriptCmd())

	return root
}
