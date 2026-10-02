package e2e

import "testing"

func TestDiagnostics(t *testing.T) {
	cases := []testCase{
		{
			id:       "diag-unclosed-fence-position",
			input:    "alpha\n\n```\nbody\nmore",
			wantHTML: "<p>alpha</p>\n<pre><code>body\nmore\n</code></pre>\n",
			wantDiag: []string{
				"mdp: 3:1: [unclosed-fence] fenced code block is never closed; the block runs to end of input",
			},
			wantExit: 0,
		},
		{
			id:       "diag-unclosed-code-span-position",
			input:    "alpha\n\nbeta `oops",
			wantHTML: "<p>alpha</p>\n<p>beta `oops</p>\n",
			wantDiag: []string{
				"mdp: 3:6: [unclosed-code-span] code span opened with \"`\" is never closed; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-unmatched-delimiter-position",
			input:    "alpha\n\nbe *ta",
			wantHTML: "<p>alpha</p>\n<p>be *ta</p>\n",
			wantDiag: []string{
				"mdp: 3:4: [unmatched-delimiter] unmatched \"*\" emphasis delimiter; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-unresolved-link-position",
			input:    "alpha\n\nsee [a](/b",
			wantHTML: "<p>alpha</p>\n<p>see [a](/b</p>\n",
			wantDiag: []string{
				"mdp: 3:5: [unresolved-link] could not parse link destination; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-unsafe-url-position",
			input:    "alpha\n\nsee [x](javascript:alert(1))",
			wantHTML: "<p>alpha</p>\n<p>see x</p>\n",
			wantDiag: []string{
				`mdp: 3:5: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept`,
			},
			wantExit: 0,
		},
		{
			id:       "diag-code-span-in-blockquote-column",
			input:    "intro\n\n> quoted `oops",
			wantHTML: "<p>intro</p>\n<blockquote>\n<p>quoted `oops</p>\n</blockquote>\n",
			wantDiag: []string{
				"mdp: 3:10: [unclosed-code-span] code span opened with \"`\" is never closed; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-nested-blockquote-column",
			input:    "> > `oops",
			wantHTML: "<blockquote>\n<blockquote>\n<p>`oops</p>\n</blockquote>\n</blockquote>\n",
			wantDiag: []string{
				"mdp: 1:5: [unclosed-code-span] code span opened with \"`\" is never closed; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-code-span-in-list-item-column",
			input:    "intro\n\n- item `oops",
			wantHTML: "<p>intro</p>\n<ul>\n<li>item `oops</li>\n</ul>\n",
			wantDiag: []string{
				"mdp: 3:8: [unclosed-code-span] code span opened with \"`\" is never closed; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-unresolved-link-in-list-item",
			input:    "- x\n- [a](/b",
			wantHTML: "<ul>\n<li>x</li>\n<li>[a](/b</li>\n</ul>\n",
			wantDiag: []string{
				"mdp: 2:3: [unresolved-link] could not parse link destination; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-unmatched-delimiter-in-blockquote",
			input:    "> x\n> **unmatched",
			wantHTML: "<blockquote>\n<p>x\n**unmatched</p>\n</blockquote>\n",
			wantDiag: []string{
				"mdp: 2:3: [unmatched-delimiter] unmatched \"**\" emphasis delimiter; kept as literal text",
			},
			wantExit: 0,
		},
		{
			id:       "diag-unsafe-url-in-list-item",
			input:    "ok line\n\n> `unclosed\n\n- [x](javascript:alert(1))\n\n**unmatched at line 7",
			wantHTML: "<p>ok line</p>\n<blockquote>\n<p>`unclosed</p>\n</blockquote>\n<ul>\n<li>x</li>\n</ul>\n<p>**unmatched at line 7</p>\n",
			wantDiag: []string{
				"mdp: 3:3: [unclosed-code-span] code span opened with \"`\" is never closed; kept as literal text",
				`mdp: 5:3: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept`,
				"mdp: 7:1: [unmatched-delimiter] unmatched \"**\" emphasis delimiter; kept as literal text",
			},
			wantExit: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
