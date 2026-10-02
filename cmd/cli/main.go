// Command mdp renders markdown to HTML: an HTML fragment on stdout and
// diagnostics as "line:col: [kind] message" lines on stderr.
//
//	mdp [file.md]   # without an argument, reads stdin
//
// Diagnostics do not affect the exit code; only input failures exit non-zero.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ozguragain/mdp/codegen"
	"github.com/ozguragain/mdp/inline"
	"github.com/ozguragain/mdp/lexer"
	"github.com/ozguragain/mdp/parser"
)

func main() {
	input, err := readInput()
	if err != nil {
		fmt.Fprintln(os.Stderr, "mdp:", err)
		os.Exit(1)
	}

	tokens := lexer.NewLexer(string(input)).Tokenize()
	p := parser.NewParser(tokens)
	doc := p.Parse()
	diags := p.Diagnostics()
	diags = append(diags, inline.Process(doc)...)
	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Span.Start.Offset < diags[j].Span.Start.Offset
	})
	for _, d := range diags {
		fmt.Fprintf(os.Stderr, "mdp: %s\n", d)
	}
	fmt.Print(codegen.RenderHTML(doc))
}

func readInput() ([]byte, error) {
	if len(os.Args) > 1 {
		return os.ReadFile(os.Args[1])
	}
	return io.ReadAll(os.Stdin)
}
