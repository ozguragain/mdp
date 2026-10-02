package e2e

import "testing"

func TestPlainText(t *testing.T) {
	cases := []testCase{
		{
			id: "plain-document-golden",
			input: `This is plain text with no markup.

A second paragraph with punctuation, (parentheses) and [brackets].
Line two of the second paragraph.`,
			wantHTML: "<p>This is plain text with no markup.</p>\n" +
				"<p>A second paragraph with punctuation, (parentheses) and [brackets].\n" +
				"Line two of the second paragraph.</p>\n",
		},
		{
			id:       "plain-short-line",
			input:    "Hello, world.",
			wantHTML: "<p>Hello, world.</p>\n",
		},
		{
			id:       "plain-text-escaped",
			input:    `a < b > c & d "e"`,
			wantHTML: "<p>a &lt; b &gt; c &amp; d &#34;e&#34;</p>\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
