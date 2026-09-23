// Command mdp renders markdown to HTML.
//
// Usage:
//
//	mdp [file.md]   # without an argument, reads stdin
//
// The HTML fragment is written to stdout; errors go to stderr with a
// non-zero exit code.
package main

import (
	"fmt"
	"io"
	"os"

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
	doc := parser.NewParser(tokens).Parse()
	inline.Process(doc)
	fmt.Print(codegen.RenderHTML(doc))
}

// readInput returns the contents of the file given as the first argument,
// or everything read from stdin when no argument is provided.
func readInput() ([]byte, error) {
	if len(os.Args) > 1 {
		return os.ReadFile(os.Args[1])
	}
	return io.ReadAll(os.Stdin)
}
