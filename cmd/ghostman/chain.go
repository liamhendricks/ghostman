package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/liamhendricks/ghostman/pkg/chain"
	"github.com/spf13/cobra"
)

func newChainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chain <name>",
		Short: "Execute a named chain of ghostman commands",
		Long: `Execute a named chain of ghostman commands defined in .ghostman/chain/<name>.md.

The chain file must contain a YAML frontmatter block with a 'steps' list.
Each step is a ghostman command string. Steps execute sequentially with
stdout of each step piped as stdin to the next. The chain aborts immediately
on the first step that exits with a non-zero exit code.

Example chain file (.ghostman/chain/login-and-fetch.md):
  ---
  steps:
    - ghostman request auth/login
    - ghostman --col auth set_collection .data.token
    - ghostman request users/list
  ---`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			chainPath := filepath.Join(".ghostman", "chain", name+".md")

			steps, err := chain.ParseChainFile(chainPath)
			if err != nil {
				return err
			}

			self, err := os.Executable()
			if err != nil {
				return fmt.Errorf("cannot locate ghostman binary: %w", err)
			}

			runErr := chain.RunChain(self, steps, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			if runErr != nil {
				var exitErr *exec.ExitError
				if errors.As(runErr, &exitErr) {
					os.Exit(exitErr.ExitCode())
				}
				return runErr
			}
			return nil
		},
	}
}
