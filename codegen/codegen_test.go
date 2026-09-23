package codegen_test

import (
	"testing"

	"github.com/ozguragain/mdp/codegen"
	"github.com/ozguragain/mdp/inline"
	"github.com/ozguragain/mdp/lexer"
	"github.com/ozguragain/mdp/parser"
)

// render runs the full markdown -> HTML pipeline over src.
func render(src string) string {
	tokens := lexer.NewLexer(src).Tokenize()
	doc := parser.NewParser(tokens).Parse()
	inline.Process(doc)
	return codegen.RenderHTML(doc)
}

func TestRenderHTML(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "empty document renders nothing",
			src:  "",
			want: "",
		},
		{
			name: "heading",
			src:  "# Title & More",
			want: "<h1>Title &amp; More</h1>\n",
		},
		{
			name: "paragraph keeps soft line breaks",
			src:  "one\ntwo",
			want: "<p>one\ntwo</p>\n",
		},
		{
			name: "text is escaped",
			src:  `a < b > c & d "e"`,
			want: "<p>a &lt; b &gt; c &amp; d &#34;e&#34;</p>\n",
		},
		{
			name: "code block with language",
			src:  "```go\nx := `<y>`\n```",
			want: "<pre><code class=\"language-go\">x := `&lt;y&gt;`\n</code></pre>\n",
		},
		{
			name: "code block without language",
			src:  "```\nplain\n```",
			want: "<pre><code>plain\n</code></pre>\n",
		},
		{
			name: "empty code block",
			src:  "```\n```",
			want: "<pre><code></code></pre>\n",
		},
		{
			name: "language is the first info word",
			src:  "```js {hl_lines=[1]}\ncode\n```",
			want: "<pre><code class=\"language-js\">code\n</code></pre>\n",
		},
		{
			name: "blockquote",
			src:  "> quoted",
			want: "<blockquote>\n<p>quoted</p>\n</blockquote>\n",
		},
		{
			name: "nested blockquote",
			src:  ">> deep",
			want: "<blockquote>\n<blockquote>\n<p>deep</p>\n</blockquote>\n</blockquote>\n",
		},
		{
			name: "tight list",
			src:  "- a\n- b",
			want: "<ul>\n<li>a</li>\n<li>b</li>\n</ul>\n",
		},
		{
			name: "nested list",
			src:  "- top\n  - deep",
			want: "<ul>\n<li>\n<p>top</p>\n<ul>\n<li>deep</li>\n</ul>\n</li>\n</ul>\n",
		},
		{
			name: "mixed document",
			src:  "# Hi\n\ntext\n\n- a\n",
			want: "<h1>Hi</h1>\n<p>text</p>\n<ul>\n<li>a</li>\n</ul>\n",
		},
		{
			name: "emphasis and strong",
			src:  "*a* **b**",
			want: "<p><em>a</em> <strong>b</strong></p>\n",
		},
		{
			name: "code span",
			src:  "run `x` now",
			want: "<p>run <code>x</code> now</p>\n",
		},
		{
			name: "code span content is escaped",
			src:  "`a<b>&c`",
			want: "<p><code>a&lt;b&gt;&amp;c</code></p>\n",
		},
		{
			name: "link",
			src:  "[t](/u)",
			want: "<p><a href=\"/u\">t</a></p>\n",
		},
		{
			name: "link with title",
			src:  `[t](/u "T")`,
			want: "<p><a href=\"/u\" title=\"T\">t</a></p>\n",
		},
		{
			name: "link label keeps inline markup",
			src:  "[*t*](/u)",
			want: "<p><a href=\"/u\"><em>t</em></a></p>\n",
		},
		{
			name: "href is ampersand escaped",
			src:  "[a](/x?a=1&b=2)",
			want: "<p><a href=\"/x?a=1&amp;b=2\">a</a></p>\n",
		},
		{
			name: "href quotes cannot break out of the attribute",
			src:  `[a](/x"onmouseover="y)`,
			want: "<p><a href=\"/x&#34;onmouseover=&#34;y\">a</a></p>\n",
		},
		{
			name: "unsafe URL renders the label only",
			src:  "[x](javascript:alert(1))",
			want: "<p>x</p>\n",
		},
		{
			name: "entity encoded scheme renders the label only",
			src:  "[x](&#106;avascript:x)",
			want: "<p>x</p>\n",
		},
		{
			name: "soft break survives inline parsing",
			src:  "*a*\nb",
			want: "<p><em>a</em>\nb</p>\n",
		},
		{
			name: "emphasis may span a soft break",
			src:  "*a\nb*",
			want: "<p><em>a\nb</em></p>\n",
		},
		{
			name: "code span in heading",
			src:  "# use `x`",
			want: "<h1>use <code>x</code></h1>\n",
		},
		{
			name: "inline markup in list item",
			src:  "- *a*",
			want: "<ul>\n<li><em>a</em></li>\n</ul>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(tt.src); got != tt.want {
				t.Errorf("render(%q)\n got: %q\nwant: %q", tt.src, got, tt.want)
			}
		})
	}
}
