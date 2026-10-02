// Artifact contract: TestMain rewrites e2e/artifact/report.md (one row per
// case, sorted by id: input, expected output, PASS/FAIL) and
// e2e/artifact/rendered.html (the combined output). Both are deterministic —
// no timestamps, absolute paths, or randomness — so two runs of
// `go test ./e2e/...` diff byte-for-byte.

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// mdpBin is the CLI binary built once in TestMain; every case runs it.
var mdpBin string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "mdp-e2e-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: create temp dir:", err)
		os.Exit(1)
	}

	mdpBin = filepath.Join(tmpDir, "mdp")

	buildCmd := exec.Command("go", "build", "-o", mdpBin, "./cmd/cli")
	buildCmd.Dir = ".."
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: build ./cmd/cli: %v\n%s", err, buildOut)
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	code := m.Run()

	// Write artifacts even on failure so a failing run still gets its report.
	if err := writeArtifacts(); err != nil {
		fmt.Fprintln(os.Stderr, "e2e: write artifacts:", err)
		code = 1
	}

	os.RemoveAll(tmpDir)
	os.Exit(code)
}
