package e2e

import "testing"

func TestBlocks(t *testing.T) {
	cases := []testCase{
		{
			id:       "block-atx-basic",
			input:    "# H1",
			wantHTML: "<h1>H1</h1>\n",
		},
		{
			id:       "block-atx-closing-hashes",
			input:    "## H2 ##",
			wantHTML: "<h2>H2</h2>\n",
		},
		{
			id:       "block-atx-inline-spans",
			input:    "### *x* ###",
			wantHTML: "<h3><em>x</em></h3>\n",
		},
		{
			id:       "block-atx-no-space-literal",
			input:    "#no-space",
			wantHTML: "<p>#no-space</p>\n",
		},
		{
			id:       "block-atx-empty-hash-literal",
			input:    "#",
			wantHTML: "<p>#</p>\n",
		},
		{
			id:       "block-paragraph-soft-break",
			input:    "one\ntwo",
			wantHTML: "<p>one\ntwo</p>\n",
		},
		{
			id:       "block-emphasis-across-soft-break",
			input:    "*a\nb*",
			wantHTML: "<p><em>a\nb</em></p>\n",
		},
		{
			id:       "block-fence-language-class",
			input:    "```go\nx := 1\n```",
			wantHTML: "<pre><code class=\"language-go\">x := 1\n</code></pre>\n",
		},
		{
			id:       "block-fence-plain",
			input:    "```\nplain\n```",
			wantHTML: "<pre><code>plain\n</code></pre>\n",
		},
		{
			id:       "block-fence-unclosed-runs-to-eof",
			input:    "```\nruns to eof\nmore",
			wantHTML: "<pre><code>runs to eof\nmore\n</code></pre>\n",
			wantDiag: []string{
				"mdp: 1:1: [unclosed-fence] fenced code block is never closed; the block runs to end of input",
			},
		},
		{
			id: "block-blockquote-nested",
			input: `> one
> two
>
> > nested

after`,
			wantHTML: "<blockquote>\n<p>one\ntwo</p>\n<blockquote>\n<p>nested</p>\n</blockquote>\n</blockquote>\n<p>after</p>\n",
		},
		{
			id: "block-list-tight-bullets",
			input: `- a
- b`,
			wantHTML: "<ul>\n<li>a</li>\n<li>b</li>\n</ul>\n",
		},
		{
			id: "block-list-marker-variants",
			input: `- a
* b
+ c`,
			wantHTML: "<ul>\n<li>a</li>\n<li>b</li>\n<li>c</li>\n</ul>\n",
		},
		{
			id: "block-list-nested",
			input: `- a
  - b
  - c
- d`,
			wantHTML: "<ul>\n<li>\n<p>a</p>\n<ul>\n<li>b</li>\n<li>c</li>\n</ul>\n</li>\n<li>d</li>\n</ul>\n",
		},
		{
			id: "block-list-continuation-lines",
			input: `- a
  continued
- b

para`,
			wantHTML: "<ul>\n<li>a\ncontinued</li>\n<li>b</li>\n</ul>\n<p>para</p>\n",
		},
		{
			id: "block-list-blank-line-splits",
			input: `- a

- b`,
			wantHTML: "<ul>\n<li>a</li>\n</ul>\n<ul>\n<li>b</li>\n</ul>\n",
		},
		{
			id:       "block-atx-hash-in-content",
			input:    "# C# rocks",
			wantHTML: "<h1>C# rocks</h1>\n",
		},
		{
			id:       "block-crlf-normalized",
			input:    "# Title\r\nbody\r\n\r\n- item",
			wantHTML: "<h1>Title</h1>\n<p>body</p>\n<ul>\n<li>item</li>\n</ul>\n",
		},
		{
			id:       "block-fence-protects-block-markers",
			input:    "```\n# not a heading\n- not a list\n```",
			wantHTML: "<pre><code># not a heading\n- not a list\n</code></pre>\n",
		},
		{
			id:       "block-fence-info-first-word",
			input:    "```js {hl_lines=[1]}\ncode\n```",
			wantHTML: "<pre><code class=\"language-js\">code\n</code></pre>\n",
		},
		{
			id:       "block-fence-content-escaped",
			input:    "```go\nx := `<y>`\n```",
			wantHTML: "<pre><code class=\"language-go\">x := `&lt;y&gt;`\n</code></pre>\n",
		},
		{
			id:       "block-fence-empty",
			input:    "```\n```",
			wantHTML: "<pre><code></code></pre>\n",
		},
		{
			id: "block-list-item-fence",
			input: `- example
  ` + "```" + `
  x

  y
  ` + "```",
			wantHTML: "<ul>\n<li>\n<p>example</p>\n<pre><code>x\n\ny\n</code></pre>\n</li>\n</ul>\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
