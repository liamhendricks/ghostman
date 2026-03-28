package main

import (
	"fmt"

	"github.com/liamhendricks/ghostman/pkg/collection"
	"github.com/liamhendricks/ghostman/pkg/config"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available requests across all collections",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			roots, err := collection.Roots(cfg)
			if err != nil {
				return err
			}
			names, err := collection.List(roots)
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
			return nil
		},
	}
}
