package chain

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/yuin/goldmark"
	gmparser "github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

// chainFrontmatter holds the parsed YAML frontmatter fields for a chain file.
type chainFrontmatter struct {
	Steps []string `yaml:"steps"`
}

// ParseChainFile reads a chain definition file at path and returns the list of
// steps from its YAML frontmatter. The file must contain a --- delimited
// frontmatter block with a non-empty "steps" list.
//
// Errors:
//   - "chain file not found: <path>" if the file does not exist
//   - "no frontmatter found" if the file has no --- delimiters
//   - "missing or empty 'steps' list in frontmatter" if steps is absent or empty
func ParseChainFile(path string) ([]string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("chain file not found: %s", path)
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	ctx := gmparser.NewContext()
	var buf bytes.Buffer
	if err := md.Convert(src, &buf, gmparser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	d := frontmatter.Get(ctx)
	if d == nil {
		return nil, fmt.Errorf("%s: no frontmatter found", path)
	}

	var fm chainFrontmatter
	if err := d.Decode(&fm); err != nil {
		return nil, fmt.Errorf("%s: invalid frontmatter: %w", path, err)
	}

	if len(fm.Steps) == 0 {
		return nil, fmt.Errorf("%s: missing or empty 'steps' list in frontmatter", path)
	}

	return fm.Steps, nil
}

// RunChain executes steps sequentially, piping stdout of step N as stdin to step N+1.
// binaryPath is the path to the ghostman binary to invoke for each step.
// stdin is connected to the first step; stdout and stderr from the final step flow
// to the provided writers.
//
// If a step string starts with "ghostman", that prefix is stripped before invoking
// binaryPath (e.g. "ghostman request auth/login" → binaryPath ["request", "auth/login"]).
//
// RunChain aborts on the first non-zero exit and returns the *exec.ExitError from the
// failing step, preserving its exit code. Non-exit errors (e.g. binary not found) are
// wrapped with the step number.
func RunChain(binaryPath string, steps []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	var currentStdin io.Reader = stdin

	for i, step := range steps {
		argv := strings.Fields(step)
		if len(argv) == 0 {
			return fmt.Errorf("step %d: empty step string", i+1)
		}

		// Strip the "ghostman" prefix — we invoke via binaryPath directly.
		if argv[0] == "ghostman" {
			argv = argv[1:]
			if len(argv) == 0 {
				return fmt.Errorf("step %d: step has no arguments after 'ghostman'", i+1)
			}
		}

		isLast := i == len(steps)-1

		var stepStdout io.Writer
		var buf bytes.Buffer
		if isLast {
			stepStdout = stdout
		} else {
			stepStdout = &buf
		}

		cmd := exec.Command(binaryPath, argv...)
		cmd.Stdin = currentStdin
		cmd.Stdout = stepStdout
		cmd.Stderr = stderr
		// cmd.Dir is intentionally left empty to inherit the working directory.

		if err := cmd.Run(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return exitErr
			}
			return fmt.Errorf("step %d %q: %w", i+1, step, err)
		}

		if !isLast {
			currentStdin = &buf
		}
	}

	return nil
}
