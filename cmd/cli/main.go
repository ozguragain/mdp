package main

import (
	"fmt"

	"github.com/ozguragain/mdp/lexer"
)

func main() {
	// Markdown input for testing the lexer
	markdown := `# Hi Lexer!
	This is first markdown text line. 
	
	- This is first list item.
	- This is second list item.

	` + "```go\n" +
		`func main() {
		fmt.Println("Hello, World!")
	}` + "\n```" + `

	> This is a blockquote.

	This is second markdown text line.
	`

	l := lexer.NewLexer(markdown)
	tokens := l.Tokenize()

	fmt.Println("---Detected Tokens---")
	for i, token := range tokens {
		fmt.Printf("[%02d] Type: %-15s | Meta: %-4s | Literal: %q\n", i, token.Type, token.Meta, token.Literal)
	}
}
