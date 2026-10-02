package e2e

import "testing"

func TestURLSanitization(t *testing.T) {
	cases := []testCase{
		{
			id:       "url-unsafe-javascript",
			input:    "[x](javascript:alert(1))",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept`,
			},
		},
		{
			id:       "url-unsafe-javascript-mixed-case",
			input:    "[x](JaVaScRiPt:alert(1))",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "JaVaScRiPt:alert(1)"; link label is kept`,
			},
		},
		{
			id:       "url-unsafe-data",
			input:    "[x](data:text/html,<h1>x</h1>)",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "data:text/html,<h1>x</h1>"; link label is kept`,
			},
		},
		{
			id:       "url-unsafe-vbscript",
			input:    "[x](vbscript:msgbox)",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "vbscript:msgbox"; link label is kept`,
			},
		},
		{
			id:       "url-unsafe-file",
			input:    "[x](file:///etc/passwd)",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "file:///etc/passwd"; link label is kept`,
			},
		},
		{
			id:       "url-unsafe-blob",
			input:    "[x](blob:https://a/b)",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "blob:https://a/b"; link label is kept`,
			},
		},
		{
			id:       "url-unsafe-entity-encoded",
			input:    "[x](&#106;avascript:alert(1))",
			wantHTML: "<p>x</p>\n",
			wantDiag: []string{
				`mdp: 1:1: [unsafe-url] dropped href with unsafe URL "&#106;avascript:alert(1)"; link label is kept`,
			},
		},
		{
			id:       "url-safe-https",
			input:    "[a](https://e.com)",
			wantHTML: "<p><a href=\"https://e.com\">a</a></p>\n",
		},
		{
			id:       "url-safe-http",
			input:    "[a](http://e.com)",
			wantHTML: "<p><a href=\"http://e.com\">a</a></p>\n",
		},
		{
			id:       "url-safe-mailto",
			input:    "[a](mailto:a@b.c)",
			wantHTML: "<p><a href=\"mailto:a@b.c\">a</a></p>\n",
		},
		{
			id:       "url-safe-relative",
			input:    "[d](/r)",
			wantHTML: "<p><a href=\"/r\">d</a></p>\n",
		},
		{
			id:       "url-safe-fragment",
			input:    "[e](#frag)",
			wantHTML: "<p><a href=\"#frag\">e</a></p>\n",
		},
		{
			id:       "url-safe-protocol-relative",
			input:    "[f](//host/p)",
			wantHTML: "<p><a href=\"//host/p\">f</a></p>\n",
		},
		{
			id:       "url-href-ampersand-escaped",
			input:    "[a](https://e.com/x?y=1&z=2)",
			wantHTML: "<p><a href=\"https://e.com/x?y=1&amp;z=2\">a</a></p>\n",
		},
		{
			id:       "url-href-quote-escaped",
			input:    "[x](/a\"b)",
			wantHTML: "<p><a href=\"/a&#34;b\">x</a></p>\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			runCase(t, tc)
		})
	}
}
