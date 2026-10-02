package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// artifactDir holds the run's deterministic artifacts (contract in
// main_test.go).
const artifactDir = "artifact"

type caseRecord struct {
	id       string
	input    string // report "input" column text
	expected string // report "expected output" column text
	actual   string // actual rendered stdout, for artifact/rendered.html
	pass     bool
}

var (
	recordsMu sync.Mutex
	records   []caseRecord
)

func record(rec caseRecord) {
	recordsMu.Lock()
	defer recordsMu.Unlock()
	records = append(records, rec)
}

// recordsSortedByID sorts the recorded cases by id and rejects duplicates.
func recordsSortedByID() ([]caseRecord, error) {
	recordsMu.Lock()
	defer recordsMu.Unlock()

	sorted := slices.Clone(records)
	slices.SortStableFunc(sorted, func(a, b caseRecord) int {
		return strings.Compare(a.id, b.id)
	})
	for i := 1; i < len(sorted); i++ {
		if sorted[i-1].id == sorted[i].id {
			return nil, fmt.Errorf("duplicate case id %q", sorted[i].id)
		}
	}
	return sorted, nil
}

// writeArtifacts writes artifact/report.md and artifact/rendered.html from
// the recorded cases (contract in main_test.go).
func writeArtifacts() error {
	sorted, err := recordsSortedByID()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return err
	}
	if err := writeReport(filepath.Join(artifactDir, "report.md"), sorted); err != nil {
		return err
	}
	return writeRendered(filepath.Join(artifactDir, "rendered.html"), sorted)
}

func writeReport(path string, recs []caseRecord) error {
	var b strings.Builder
	b.WriteString("# mdp end-to-end report\n\n")
	b.WriteString("One row per executed case, sorted by case id. Deterministic: no timestamps, no absolute paths, no randomness.\n")
	b.WriteString("The \"expected output\" cell holds the expected stdout followed by `stderr: ...` lines for expected diagnostics; cell newlines appear as literal \\n.\n\n")
	b.WriteString("| case id | input | expected output | result |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, rec := range recs {
		result := "FAIL"
		if rec.pass {
			result = "PASS"
		}
		row := fmt.Sprintf("| %s | %s | %s | %s |\n",
			mdCell(rec.id), mdCell(rec.input), mdCell(rec.expected), result)
		b.WriteString(row)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeRendered(path string, recs []caseRecord) error {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n")
	b.WriteString("<html lang=\"en\">\n")
	b.WriteString("<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<title>mdp e2e rendered output</title>\n")
	b.WriteString("</head>\n")
	b.WriteString("<body>\n")
	b.WriteString("<!-- Deterministic: one section per executed case, sorted by case id; no timestamps, no absolute paths. -->\n")
	for _, rec := range recs {
		b.WriteString("<section data-case=\"" + rec.id + "\">\n")
		b.WriteString("<!-- case: " + rec.id + " -->\n")
		b.WriteString(rec.actual)
		if !strings.HasSuffix(rec.actual, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("</section>\n")
	}
	b.WriteString("</body>\n")
	b.WriteString("</html>\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// mdCell escapes a report cell so every table row stays on a single line.
func mdCell(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "|", `\|`)
	return s
}
