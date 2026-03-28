package main

import (
	"fmt"

	"github.com/liamhendricks/ghostman/pkg/pipe"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Extract a value from piped JSON using a gjson path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := pipe.ReadJSON(cmd.InOrStdin())
			if err != nil {
				return err
			}
			value, err := pipe.Extract(data, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
	return cmd
}
