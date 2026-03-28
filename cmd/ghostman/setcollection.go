package main

import (
	"fmt"
	"path/filepath"

	"github.com/liamhendricks/ghostman/pkg/collection"
	"github.com/liamhendricks/ghostman/pkg/pipe"
	"github.com/spf13/cobra"
)

func newSetCollectionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set_collection <path>",
		Short: "Extract a JSON value and write it to a collection's vars file",
		Long: `Reads JSON from stdin, extracts the value at <path>, derives a key from
the last segment of <path>, and writes key=value to the collection's vars.yaml.
The original JSON is passed through to stdout for further chaining.
Example: ghostman request auth/login | ghostman --col auth set_collection .data.token`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			colName, _ := cmd.Flags().GetString("col")
			if colName == "" {
				return fmt.Errorf("--col flag is required for set_collection")
			}

			data, err := pipe.ReadJSON(cmd.InOrStdin())
			if err != nil {
				return err
			}

			value, err := pipe.Extract(data, path)
			if err != nil {
				return err
			}

			key := collection.DeriveKey(path)
			collectionDir := filepath.Join(".ghostman", colName)

			if err := collection.UpdateColVars(collectionDir, key, value); err != nil {
				return err
			}

			// Pass-through: write original JSON to stdout
			cmd.OutOrStdout().Write(data) //nolint:errcheck
			return nil
		},
	}
}
