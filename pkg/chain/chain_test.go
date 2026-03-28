package chain

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testHelperBin is the path to the compiled helper binary used in RunChain tests.
var testHelperBin string

// helperSource is a small Go program used as a fake ghostman binary in tests.
// It supports the following "subcommands" (argv[0] after stripping "ghostman"):
//
//	upper        — reads stdin, writes it uppercased to stdout
//	cat          — copies stdin to stdout unchanged
//	echo <args>  — writes args joined with space to stdout
//	fail         — exits with code 1
const helperSource = `package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "helper: no subcommand")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "upper":
		data, _ := io.ReadAll(os.Stdin)
		fmt.Print(strings.ToUpper(string(data)))
	case "cat":
		io.Copy(os.Stdout, os.Stdin)
	case "echo":
		fmt.Println(strings.Join(os.Args[2:], " "))
	case "fail":
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "helper: unknown subcommand %q\n", os.Args[1])
		os.Exit(1)
	}
}
`

func TestMain(m *testing.M) {
	// Build the helper binary into a temp directory.
	tmpDir, err := os.MkdirTemp("", "chain-helper-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "helper.go")
	if err := os.WriteFile(srcFile, []byte(helperSource), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to write helper source: %v\n", err)
		os.Exit(1)
	}

	binPath := filepath.Join(tmpDir, "helper")
	cmd := exec.Command("go", "build", "-o", binPath, srcFile)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to build helper binary: %v\n", err)
		os.Exit(1)
	}

	testHelperBin = binPath
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// ParseChainFile tests
// ---------------------------------------------------------------------------

func TestParseChainFile_Valid(t *testing.T) {
	content := "---\nsteps:\n  - ghostman request auth/login\n  - ghostman --col auth set_collection .data.token\n---\n"
	f := writeTempFile(t, content)

	steps, err := ParseChainFile(f)
	if err != nil {
		t.Fatalf("ParseChainFile returned unexpected error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0] != "ghostman request auth/login" {
		t.Errorf("steps[0] = %q; want %q", steps[0], "ghostman request auth/login")
	}
	if steps[1] != "ghostman --col auth set_collection .data.token" {
		t.Errorf("steps[1] = %q; want %q", steps[1], "ghostman --col auth set_collection .data.token")
	}
}

func TestParseChainFile_NoFrontmatter(t *testing.T) {
	content := "# Just a markdown file\n\nNo frontmatter here.\n"
	f := writeTempFile(t, content)

	_, err := ParseChainFile(f)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no frontmatter found") {
		t.Errorf("error %q does not contain %q", err.Error(), "no frontmatter found")
	}
}

func TestParseChainFile_EmptySteps(t *testing.T) {
	content := "---\nsteps: []\n---\n"
	f := writeTempFile(t, content)

	_, err := ParseChainFile(f)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing or empty 'steps'") {
		t.Errorf("error %q does not contain %q", err.Error(), "missing or empty 'steps'")
	}
}

func TestParseChainFile_MissingStepsKey(t *testing.T) {
	content := "---\ntitle: my-chain\n---\n"
	f := writeTempFile(t, content)

	_, err := ParseChainFile(f)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing or empty 'steps'") {
		t.Errorf("error %q does not contain %q", err.Error(), "missing or empty 'steps'")
	}
}

func TestParseChainFile_FileNotFound(t *testing.T) {
	_, err := ParseChainFile("/nonexistent/path/chain.md")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "chain file not found") {
		t.Errorf("error %q does not contain %q", err.Error(), "chain file not found")
	}
}

// ---------------------------------------------------------------------------
// RunChain tests
// ---------------------------------------------------------------------------

func TestRunChain_PipesStdout(t *testing.T) {
	// Step 1: echo "hello" → stdout "hello\n"
	// Step 2: upper → reads "hello\n" from stdin, writes "HELLO\n" to stdout
	steps := []string{
		"ghostman echo hello",
		"ghostman upper",
	}

	var out bytes.Buffer
	err := RunChain(testHelperBin, steps, nil, &out, os.Stderr)
	if err != nil {
		t.Fatalf("RunChain returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "HELLO") {
		t.Errorf("expected output to contain HELLO, got %q", got)
	}
}

func TestRunChain_AbortsOnFailure(t *testing.T) {
	steps := []string{
		"ghostman echo first",
		"ghostman fail",
		"ghostman echo third", // should never be reached
	}

	var out bytes.Buffer
	err := RunChain(testHelperBin, steps, nil, &out, os.Stderr)
	if err == nil {
		t.Fatal("expected error from failing step, got nil")
	}

	var exitErr *exec.ExitError
	if !isExitError(err, &exitErr) {
		t.Errorf("expected *exec.ExitError, got %T: %v", err, err)
	}
}

func TestRunChain_EmptyStep(t *testing.T) {
	steps := []string{""}

	err := RunChain(testHelperBin, steps, nil, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for empty step, got nil")
	}
	if !strings.Contains(err.Error(), "empty step string") {
		t.Errorf("error %q does not contain %q", err.Error(), "empty step string")
	}
}

func TestRunChain_StripsGhostmanPrefix(t *testing.T) {
	// "ghostman echo hello" should invoke testHelperBin with args ["echo", "hello"]
	// The helper "echo hello" prints "hello\n"
	steps := []string{"ghostman echo hello"}

	var out bytes.Buffer
	err := RunChain(testHelperBin, steps, nil, &out, os.Stderr)
	if err != nil {
		t.Fatalf("RunChain returned unexpected error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "chain-*.md")
	if err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writeTempFile write: %v", err)
	}
	f.Close()
	return f.Name()
}

// isExitError checks if err (or any wrapped error) is an *exec.ExitError.
func isExitError(err error, target **exec.ExitError) bool {
	if err == nil {
		return false
	}
	// Direct type assertion first.
	if e, ok := err.(*exec.ExitError); ok {
		if target != nil {
			*target = e
		}
		return true
	}
	return false
}
