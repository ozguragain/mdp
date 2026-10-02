package e2e

import "testing"

func TestCLIBehavior(t *testing.T) {
	cases := []testCase{
		{
			id:       "cli-file-argument-renders",
			mode:     inputFile,
			input:    "# File arg\n\nplain *em*",
			wantHTML: "<h1>File arg</h1>\n<p>plain <em>em</em></p>\n",
			wantExit: 0,
		},
		{
			id:       "cli-exit-zero-with-diagnostics",
			mode:     inputStdin,
			input:    "[x](javascript:alert(1))",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept`,
			},
			wantExit: 0,
		},
		{
			id:       "cli-missing-file-nonzero-exit",
			mode:     inputMissingFile,
			input:    missingInputName,
			wantHTML: "",
			wantDiag: []string{
				"mdp: open no-such-input.md: no such file or directory",
			},
			wantExit: exitNonZero,
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
