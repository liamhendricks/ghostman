package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newScriptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "script <file.go>",
		Short: "Run a Go script as a pipe step",
		Long: `Run a user-written Go script as a pipe step in a chain.

The script receives JSON on stdin and writes its output to stdout.
Scripts import github.com/liamhendricks/ghostman/pkg/script for the
Handler interface, Data type, and Run/RunFunc helpers.

The script is executed via 'go run' from the current working directory,
which must contain a go.mod (typically the project root). This allows
scripts to import ghostman's pkg/script package automatically.

Example:
  echo '{"name":"world"}' | ghostman script .ghostman/scripts/example.go`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptPath, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolving script path: %w", err)
			}
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				return fmt.Errorf("script not found: %s", args[0])
			}

			c := exec.Command("go", "run", scriptPath)
			c.Stdin = cmd.InOrStdin()
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.ErrOrStderr()

			if err := c.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}
			return nil
		},
	}
}
