package main

import (
	"github.com/liamhendricks/ghostman/pkg/env"
	"github.com/liamhendricks/ghostman/pkg/pipe"
	"github.com/spf13/cobra"
)

func newSetEnvCmd() *cobra.Command {
	var envFile string
	cmd := &cobra.Command{
		Use:   "set_env <key> <path>",
		Short: "Extract a JSON value and write it to a .env file",
		Long: `Reads JSON from stdin, extracts the value at <path>, and writes <key>=<value> to the specified .env file.
Example: ghostman request auth/login | ghostman set_env --env-file .ghostman/.env token .data.token`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, path := args[0], args[1]
			data, err := pipe.ReadJSON(cmd.InOrStdin())
			if err != nil {
				return err
			}
			value, err := pipe.Extract(data, path)
			if err != nil {
				return err
			}
			if err := env.UpdateEnv(envFile, key, value); err != nil {
				return err
			}
			// Pass-through: write original JSON to stdout
			cmd.OutOrStdout().Write(data) //nolint:errcheck
			return nil
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", "", "path to .env file (required)")
	cmd.MarkFlagRequired("env-file") //nolint:errcheck
	return cmd
}
