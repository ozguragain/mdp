package e2e

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// missingInputName is a file argument that must never exist.
const missingInputName = "no-such-input.md"

// exitNonZero is a wantExit sentinel meaning "any non-zero exit code".
const exitNonZero = -1

type inputMode int

const (
	inputStdin       inputMode = iota // pipe testCase.input on stdin
	inputFile                         // pass a temporary file holding testCase.input
	inputMissingFile                  // pass a file argument that does not exist
)

type testCase struct {
	id       string    // unique, stable case identifier
	mode     inputMode // how the input reaches the CLI
	input    string    // markdown source (display text for inputMissingFile)
	wantHTML string    // expected stdout: the rendered HTML fragment
	wantDiag []string  // expected stderr lines, each including the "mdp: " prefix
	wantExit int       // expected exit code; exitNonZero accepts any non-zero
}

// runMDP runs the built CLI with input on stdin and returns stdout, stderr,
// and the exit code.
func runMDP(t *testing.T, input string, args []string) (stdout string, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(mdpBin, args...)
	cmd.Stdin = strings.NewReader(input)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	exitCode = 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			t.Fatalf("running mdp %v: %v", args, runErr)
		}
		exitCode = exitErr.ExitCode()
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// stderrLines splits stderr into lines; empty stderr becomes nil so it
// matches cases with no expected diagnostics.
func stderrLines(stderr string) []string {
	trimmed := strings.TrimSuffix(stderr, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// commandArgs returns the CLI arguments for the case's input mode.
func (tc testCase) commandArgs(t *testing.T) []string {
	t.Helper()

	switch tc.mode {
	case inputStdin:
		return nil
	case inputFile:
		inputPath := filepath.Join(t.TempDir(), "input.md")
		if err := os.WriteFile(inputPath, []byte(tc.input), 0o644); err != nil {
			t.Fatalf("write input file: %v", err)
		}
		return []string{inputPath}
	case inputMissingFile:
		return []string{missingInputName}
	}

	t.Fatalf("case %s: unknown input mode %d", tc.id, tc.mode)
	return nil
}

func (tc testCase) exitMatches(got int) bool {
	if tc.wantExit == exitNonZero {
		return got != 0
	}
	return got == tc.wantExit
}

func (tc testCase) exitExpectation() string {
	if tc.wantExit == exitNonZero {
		return "non-zero"
	}
	return strconv.Itoa(tc.wantExit)
}

func (tc testCase) inputCell() string {
	switch tc.mode {
	case inputStdin:
		return tc.input
	case inputFile:
		return tc.input + " (via file argument)"
	case inputMissingFile:
		return tc.input + " (missing file argument)"
	}
	return tc.input
}

// expectedOutputCell folds expected stdout and `stderr: ...` lines into the
// report's "expected output" cell.
func (tc testCase) expectedOutputCell() string {
	var b strings.Builder
	b.WriteString(tc.wantHTML)
	for _, line := range tc.wantDiag {
		b.WriteString("stderr: ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

// runCase runs one scenario through the CLI, asserts stdout/stderr/exit, and
// records the result for the artifact report.
func runCase(t *testing.T, tc testCase) {
	t.Helper()

	args := tc.commandArgs(t)
	stdout, stderr, exitCode := runMDP(t, tc.input, args)
	gotDiag := stderrLines(stderr)

	pass := true
	if stdout != tc.wantHTML {
		pass = false
		t.Errorf("stdout mismatch:\n got: %q\nwant: %q", stdout, tc.wantHTML)
	}
	if !slices.Equal(gotDiag, tc.wantDiag) {
		pass = false
		t.Errorf("stderr mismatch:\n got: %q\nwant: %q", gotDiag, tc.wantDiag)
	}
	if !tc.exitMatches(exitCode) {
		pass = false
		t.Errorf("exit code mismatch: got %d, want %s", exitCode, tc.exitExpectation())
	}

	record(caseRecord{
		id:       tc.id,
		input:    tc.inputCell(),
		expected: tc.expectedOutputCell(),
		actual:   stdout,
		pass:     pass,
	})
}
