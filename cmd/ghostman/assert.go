package main

import (
	"github.com/liamhendricks/ghostman/pkg/pipe"
	"github.com/spf13/cobra"
)

func newAssertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `assert <expression>`,
		Short: "Assert a condition on piped JSON (exits non-zero on failure)",
		Long: `Evaluates an assertion against piped JSON. Supported operators: == and !=
Example: ghostman request auth/login | ghostman assert '.status == "ok"'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := pipe.ReadJSON(cmd.InOrStdin())
			if err != nil {
				return err
			}
			if err := pipe.Assert(data, args[0]); err != nil {
				return err
			}
			// Pass-through: write original JSON to stdout
			cmd.OutOrStdout().Write(data) //nolint:errcheck
			return nil
		},
	}
	return cmd
}
