package e2e

import "testing"

func TestCodeSpans(t *testing.T) {
	cases := []testCase{
		{
			id:       "codespan-markup-is-literal",
			input:    "`x*y`",
			wantHTML: "<p><code>x*y</code></p>\n",
		},
		{
			id:       "codespan-double-backtick-holds-backtick",
			input:    "``a`b``",
			wantHTML: "<p><code>a`b</code></p>\n",
		},
		{
			id:       "codespan-trims-edge-spaces",
			input:    "` x `",
			wantHTML: "<p><code>x</code></p>\n",
		},
		{
			id:       "codespan-unmatched-backtick-literal",
			input:    "unmatched ` backtick",
			wantHTML: "<p>unmatched ` backtick</p>\n",
			wantDiag: []string{
				"mdp: 1:11: [unclosed-code-span] code span opened with \"`\" is never closed; kept as literal text",
			},
		},
		{
			id:       "codespan-escape-not-interpreted",
			input:    "`a\\*b`",
			wantHTML: "<p><code>a\\*b</code></p>\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
