// Package diag defines the diagnostics reported alongside the AST. They
// never abort processing: the pipeline keeps its literal-text recovery and
// reports what was recovered.
package diag

import (
	"fmt"

	"github.com/ozguragain/mdp/ast"
)

// Kind classifies a diagnostic.
type Kind string

const (
	// UnsafeURL: link URL failed the allowlist; the href was dropped.
	UnsafeURL Kind = "unsafe-url"
	// UnmatchedDelimiter: emphasis delimiter that found no closing run.
	UnmatchedDelimiter Kind = "unmatched-delimiter"
	// UnclosedCodeSpan: backtick run that found no closing run.
	UnclosedCodeSpan Kind = "unclosed-code-span"
	// UnresolvedLink: `[label](...` whose destination could not be parsed.
	UnresolvedLink Kind = "unresolved-link"
	// UnclosedFence: fenced code block running to EOF, fence missing.
	UnclosedFence Kind = "unclosed-fence"
)

// Diagnostic is a positioned problem found anywhere in the pipeline.
type Diagnostic struct {
	Kind    Kind
	Message string
	Span    ast.Span
}

// String formats the diagnostic as "line:col: [kind] message".
func (d Diagnostic) String() string {
	return fmt.Sprintf("%d:%d: [%s] %s", d.Span.Start.Line, d.Span.Start.Column, d.Kind, d.Message)
}
