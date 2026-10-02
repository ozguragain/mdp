package e2e

import "testing"

func TestInlineGrammar(t *testing.T) {
	cases := []testCase{
		{
			id:       "inline-em-asterisk",
			input:    "*em*",
			wantHTML: "<p><em>em</em></p>\n",
		},
		{
			id:       "inline-em-underscore",
			input:    "_em_",
			wantHTML: "<p><em>em</em></p>\n",
		},
		{
			id:       "inline-strong-asterisk",
			input:    "**strong**",
			wantHTML: "<p><strong>strong</strong></p>\n",
		},
		{
			id:       "inline-strong-underscore",
			input:    "__strong__",
			wantHTML: "<p><strong>strong</strong></p>\n",
		},
		{
			id:       "inline-em-strong-triple-run",
			input:    "***x***",
			wantHTML: "<p><em><strong>x</strong></em></p>\n",
		},
		{
			id:       "inline-quadruple-run-literal",
			input:    "****x****",
			wantHTML: "<p>****x****</p>\n",
		},
		{
			id:       "inline-intraword-emphasis",
			input:    "un*frigging*believable",
			wantHTML: "<p>un<em>frigging</em>believable</p>\n",
		},
		{
			id:       "inline-intraword-underscore-literal",
			input:    "snake_case_name",
			wantHTML: "<p>snake_case_name</p>\n",
		},
		{
			id:       "inline-nested-em-stack",
			input:    "*a *b* c*",
			wantHTML: "<p><em>a <em>b</em> c</em></p>\n",
		},
		{
			id:       "inline-strong-wrapping-em",
			input:    "**a *b* c**",
			wantHTML: "<p><strong>a <em>b</em> c</strong></p>\n",
		},
		{
			id:       "inline-unequal-run-opens-literal",
			input:    "**foo*",
			wantHTML: "<p>**foo*</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unmatched-delimiter] unmatched "**" emphasis delimiter; kept as literal text`,
			},
		},
		{
			id:       "inline-unequal-run-closes-literal",
			input:    "*foo**",
			wantHTML: "<p>*foo**</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unmatched-delimiter] unmatched "*" emphasis delimiter; kept as literal text`,
			},
		},
		{
			id:       "inline-spaced-delimiters-literal",
			input:    "a * b * c",
			wantHTML: "<p>a * b * c</p>\n",
		},
		{
			id:       "inline-escape-asterisk",
			input:    `\*x\*`,
			wantHTML: "<p>*x*</p>\n",
		},
		{
			id:       "inline-escape-backslash",
			input:    `a\\b`,
			wantHTML: "<p>a\\b</p>\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
