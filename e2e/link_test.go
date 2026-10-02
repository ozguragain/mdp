package e2e

import "testing"

func TestLinks(t *testing.T) {
	cases := []testCase{
		{
			id:       "link-basic",
			input:    "[t](/u)",
			wantHTML: "<p><a href=\"/u\">t</a></p>\n",
		},
		{
			id:       "link-title",
			input:    "[t](/u \"T\")",
			wantHTML: "<p><a href=\"/u\" title=\"T\">t</a></p>\n",
		},
		{
			id:       "link-nested-markup-label",
			input:    "[*t*](/u)",
			wantHTML: "<p><a href=\"/u\"><em>t</em></a></p>\n",
		},
		{
			id:       "link-balanced-parens-destination",
			input:    "[t](/a_(b))",
			wantHTML: "<p><a href=\"/a_(b)\">t</a></p>\n",
		},
		{
			id:       "link-nested-link-suppressed",
			input:    "[a [b](/c) d](/e)",
			wantHTML: "<p><a href=\"/e\">a [b](/c) d</a></p>\n",
		},
		{
			id:       "link-empty-destination",
			input:    "[a]()",
			wantHTML: "<p><a href=\"\">a</a></p>\n",
		},
		{
			id:       "link-broken-literal",
			input:    "[a](/b",
			wantHTML: "<p>[a](/b</p>\n",
			wantDiag: []string{
				"mdp: 1:1: [unresolved-link] could not parse link destination; kept as literal text",
			},
		},
		{
			id:       "link-bare-label-literal",
			input:    "[a]",
			wantHTML: "<p>[a]</p>\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
