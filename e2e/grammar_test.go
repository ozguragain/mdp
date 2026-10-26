package e2e

import "testing"

// TestGrammar covers the block and inline rules added for blog-sized
// documents: images, tables, ordered lists, thematic breaks, strikethrough,
// hard breaks, autolinks and frontmatter.
func TestGrammar(t *testing.T) {
	cases := []testCase{
		// --- images ---
		{
			id:       "image-basic",
			input:    "![alt](/x.png)",
			wantHTML: "<p><img src=\"/x.png\" alt=\"alt\"></p>\n",
		},
		{
			id:       "image-title",
			input:    "![alt](/x.png \"title\")",
			wantHTML: "<p><img src=\"/x.png\" alt=\"alt\" title=\"title\"></p>\n",
		},
		{
			id:       "image-alt-is-literal-text",
			input:    "![a *b*](/x.png)",
			wantHTML: "<p><img src=\"/x.png\" alt=\"a *b*\"></p>\n",
		},
		{
			id:       "image-empty-alt",
			input:    "![](/x.png)",
			wantHTML: "<p><img src=\"/x.png\" alt=\"\"></p>\n",
		},
		{
			id:       "image-unsafe-src-keeps-alt-text",
			input:    "![x](javascript:alert(1))",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped img src with unsafe URL "javascript:alert(1)"; alt text is kept`,
			},
		},
		{
			id:       "image-in-paragraph",
			input:    "before ![a](/p.jpg) after",
			wantHTML: "<p>before <img src=\"/p.jpg\" alt=\"a\"> after</p>\n",
		},

		// --- strikethrough ---
		{
			id:       "strike-basic",
			input:    "~~gone~~",
			wantHTML: "<p><del>gone</del></p>\n",
		},
		{
			id:       "strike-holds-other-inline",
			input:    "~~a **b** c~~",
			wantHTML: "<p><del>a <strong>b</strong> c</del></p>\n",
		},
		{
			id:       "strike-single-tilde-literal",
			input:    "~x~",
			wantHTML: "<p>~x~</p>\n",
		},
		{
			id:       "strike-triple-tilde-literal",
			input:    "~~~x~~~",
			wantHTML: "<p>~~~x~~~</p>\n",
		},
		{
			id:       "strike-unclosed-literal",
			input:    "~~open",
			wantHTML: "<p>~~open</p>\n",
		},

		// --- hard line breaks ---
		{
			id:       "hardbreak-trailing-spaces",
			input:    "one  \ntwo",
			wantHTML: "<p>one<br>\ntwo</p>\n",
		},
		{
			id:       "hardbreak-backslash",
			input:    "one\\\ntwo",
			wantHTML: "<p>one<br>\ntwo</p>\n",
		},
		{
			id:       "hardbreak-single-space-stays-soft",
			input:    "one \ntwo",
			wantHTML: "<p>one \ntwo</p>\n",
		},
		{
			id:       "hardbreak-three-spaces",
			input:    "a   \nb",
			wantHTML: "<p>a<br>\nb</p>\n",
		},

		// --- autolinks ---
		{
			id:       "autolink-uri",
			input:    "<https://example.com/a>",
			wantHTML: "<p><a href=\"https://example.com/a\">https://example.com/a</a></p>\n",
		},
		{
			id:       "autolink-email",
			input:    "<user@example.com>",
			wantHTML: "<p><a href=\"mailto:user@example.com\">user@example.com</a></p>\n",
		},
		{
			id:       "autolink-plain-angle-text-is-literal",
			input:    "<foo>",
			wantHTML: "<p>&lt;foo&gt;</p>\n",
		},
		{
			id:       "autolink-unsafe-scheme-keeps-text",
			input:    "<javascript:alert(1)>",
			wantHTML: "<p>javascript:alert(1)</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept`,
			},
		},
		{
			id:       "autolink-unclosed-is-literal",
			input:    "<https://example.com",
			wantHTML: "<p>&lt;https://example.com</p>\n",
		},

		// --- thematic breaks ---
		{
			id:       "thematic-break-dashes",
			input:    "a\n\n---\n\nb",
			wantHTML: "<p>a</p>\n<hr>\n<p>b</p>\n",
		},
		{
			id:       "thematic-break-stars",
			input:    "***",
			wantHTML: "<hr>\n",
		},
		{
			id:       "thematic-break-underscores",
			input:    "___",
			wantHTML: "<hr>\n",
		},
		{
			id:       "thematic-break-spaced-dashes",
			input:    "- - -",
			wantHTML: "<hr>\n",
		},
		{
			id:       "thematic-break-four-runs",
			input:    "----",
			wantHTML: "<hr>\n",
		},
		{
			id:       "thematic-break-two-dashes-are-text",
			input:    "--",
			wantHTML: "<p>--</p>\n",
		},

		// --- ordered lists ---
		{
			id: "ordered-basic",
			input: `1. one
2. two`,
			wantHTML: "<ol>\n<li>one</li>\n<li>two</li>\n</ol>\n",
		},
		{
			id: "ordered-start-keeps-its-number",
			input: `3. three
4. four`,
			wantHTML: "<ol start=\"3\">\n<li>three</li>\n<li>four</li>\n</ol>\n",
		},
		{
			id:       "ordered-paren-marker",
			input:    "1) one",
			wantHTML: "<ol>\n<li>one</li>\n</ol>\n",
		},
		{
			id: "ordered-continuation-line",
			input: `1. one
   continued
2. two`,
			wantHTML: "<ol>\n<li>one\ncontinued</li>\n<li>two</li>\n</ol>\n",
		},
		{
			id: "ordered-marker-change-ends-the-list",
			input: `1. one
- two`,
			wantHTML: "<ol>\n<li>one</li>\n</ol>\n<ul>\n<li>two</li>\n</ul>\n",
		},
		{
			id:       "ordered-digits-without-marker-are-text",
			input:    "1.no-space",
			wantHTML: "<p>1.no-space</p>\n",
		},

		// --- tables ---
		{
			id: "table-basic",
			input: `| a | b |
| --- | --- |
| 1 | 2 |`,
			wantHTML: "<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n" +
				"<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n",
		},
		{
			id: "table-column-alignment",
			input: `| l | c | r |
| :-- | :-: | --: |
| 1 | 2 | 3 |`,
			wantHTML: "<table>\n<thead>\n<tr>\n" +
				"<th align=\"left\">l</th>\n<th align=\"center\">c</th>\n<th align=\"right\">r</th>\n" +
				"</tr>\n</thead>\n<tbody>\n<tr>\n" +
				"<td align=\"left\">1</td>\n<td align=\"center\">2</td>\n<td align=\"right\">3</td>\n" +
				"</tr>\n</tbody>\n</table>\n",
		},
		{
			id: "table-header-only-has-no-tbody",
			input: `| a |
| --- |`,
			wantHTML: "<table>\n<thead>\n<tr>\n<th>a</th>\n</tr>\n</thead>\n</table>\n",
		},
		{
			id: "table-short-row-is-padded",
			input: `| a | b |
| --- | --- |
| 1 |`,
			wantHTML: "<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n" +
				"<tbody>\n<tr>\n<td>1</td>\n<td></td>\n</tr>\n</tbody>\n</table>\n",
		},
		{
			id: "table-cells-parse-inline",
			input: `| *a* |
| --- |
| ` + "`b`" + ` |`,
			wantHTML: "<table>\n<thead>\n<tr>\n<th><em>a</em></th>\n</tr>\n</thead>\n" +
				"<tbody>\n<tr>\n<td><code>b</code></td>\n</tr>\n</tbody>\n</table>\n",
		},
		{
			id: "table-without-delimiter-row-stays-text",
			input: `a | b
more`,
			wantHTML: "<p>a | b\nmore</p>\n",
		},

		// --- frontmatter ---
		{
			id: "frontmatter-is-not-rendered",
			input: `---
title: x
---

# H`,
			wantHTML: "<h1>H</h1>\n",
		},
		{
			id: "frontmatter-body-only",
			input: `---
title: x
---
body`,
			wantHTML: "<p>body</p>\n",
		},
		{
			id: "frontmatter-closed-by-dots",
			input: `---
title: x
...

body`,
			wantHTML: "<p>body</p>\n",
		},
		{
			id: "frontmatter-unclosed-is-thematic-break",
			input: `---
title: x`,
			wantHTML: "<hr>\n<p>title: x</p>\n",
		},
		{
			id: "frontmatter-only-at-start-of-document",
			input: `x

---
title: y
---`,
			wantHTML: "<p>x</p>\n<hr>\n<p>title: y</p>\n<hr>\n",
		},
		{
			id: "frontmatter-empty-body",
			input: `---
---
body`,
			wantHTML: "<p>body</p>\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
