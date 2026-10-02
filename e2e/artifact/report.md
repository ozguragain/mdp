# mdp end-to-end report

One row per executed case, sorted by case id. Deterministic: no timestamps, no absolute paths, no randomness.
The "expected output" cell holds the expected stdout followed by `stderr: ...` lines for expected diagnostics; cell newlines appear as literal \n.

| case id | input | expected output | result |
| --- | --- | --- | --- |
| block-atx-basic | # H1 | <h1>H1</h1>\n | PASS |
| block-atx-closing-hashes | ## H2 ## | <h2>H2</h2>\n | PASS |
| block-atx-empty-hash-literal | # | <p>#</p>\n | PASS |
| block-atx-hash-in-content | # C# rocks | <h1>C# rocks</h1>\n | PASS |
| block-atx-inline-spans | ### *x* ### | <h3><em>x</em></h3>\n | PASS |
| block-atx-no-space-literal | #no-space | <p>#no-space</p>\n | PASS |
| block-blockquote-nested | > one\n> two\n>\n> > nested\n\nafter | <blockquote>\n<p>one\ntwo</p>\n<blockquote>\n<p>nested</p>\n</blockquote>\n</blockquote>\n<p>after</p>\n | PASS |
| block-crlf-normalized | # Title\nbody\n\n- item | <h1>Title</h1>\n<p>body</p>\n<ul>\n<li>item</li>\n</ul>\n | PASS |
| block-emphasis-across-soft-break | *a\nb* | <p><em>a\nb</em></p>\n | PASS |
| block-fence-content-escaped | ```go\nx := `<y>`\n``` | <pre><code class="language-go">x := `&lt;y&gt;`\n</code></pre>\n | PASS |
| block-fence-empty | ```\n``` | <pre><code></code></pre>\n | PASS |
| block-fence-info-first-word | ```js {hl_lines=[1]}\ncode\n``` | <pre><code class="language-js">code\n</code></pre>\n | PASS |
| block-fence-language-class | ```go\nx := 1\n``` | <pre><code class="language-go">x := 1\n</code></pre>\n | PASS |
| block-fence-plain | ```\nplain\n``` | <pre><code>plain\n</code></pre>\n | PASS |
| block-fence-protects-block-markers | ```\n# not a heading\n- not a list\n``` | <pre><code># not a heading\n- not a list\n</code></pre>\n | PASS |
| block-fence-unclosed-runs-to-eof | ```\nruns to eof\nmore | <pre><code>runs to eof\nmore\n</code></pre>\nstderr: mdp: 1:1: [unclosed-fence] fenced code block is never closed; the block runs to end of input\n | PASS |
| block-list-blank-line-splits | - a\n\n- b | <ul>\n<li>a</li>\n</ul>\n<ul>\n<li>b</li>\n</ul>\n | PASS |
| block-list-continuation-lines | - a\n  continued\n- b\n\npara | <ul>\n<li>a\ncontinued</li>\n<li>b</li>\n</ul>\n<p>para</p>\n | PASS |
| block-list-item-fence | - example\n  ```\n  x\n\n  y\n  ``` | <ul>\n<li>\n<p>example</p>\n<pre><code>x\n\ny\n</code></pre>\n</li>\n</ul>\n | PASS |
| block-list-marker-variants | - a\n* b\n+ c | <ul>\n<li>a</li>\n<li>b</li>\n<li>c</li>\n</ul>\n | PASS |
| block-list-nested | - a\n  - b\n  - c\n- d | <ul>\n<li>\n<p>a</p>\n<ul>\n<li>b</li>\n<li>c</li>\n</ul>\n</li>\n<li>d</li>\n</ul>\n | PASS |
| block-list-tight-bullets | - a\n- b | <ul>\n<li>a</li>\n<li>b</li>\n</ul>\n | PASS |
| block-paragraph-soft-break | one\ntwo | <p>one\ntwo</p>\n | PASS |
| cli-exit-zero-with-diagnostics | [x](javascript:alert(1)) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept\n | PASS |
| cli-file-argument-renders | # File arg\n\nplain *em* (via file argument) | <h1>File arg</h1>\n<p>plain <em>em</em></p>\n | PASS |
| cli-missing-file-nonzero-exit | no-such-input.md (missing file argument) | stderr: mdp: open no-such-input.md: no such file or directory\n | PASS |
| codespan-double-backtick-holds-backtick | ``a`b`` | <p><code>a`b</code></p>\n | PASS |
| codespan-escape-not-interpreted | `a\\*b` | <p><code>a\\*b</code></p>\n | PASS |
| codespan-markup-is-literal | `x*y` | <p><code>x*y</code></p>\n | PASS |
| codespan-trims-edge-spaces | ` x ` | <p><code>x</code></p>\n | PASS |
| codespan-unmatched-backtick-literal | unmatched ` backtick | <p>unmatched ` backtick</p>\nstderr: mdp: 1:11: [unclosed-code-span] code span opened with "`" is never closed; kept as literal text\n | PASS |
| diag-code-span-in-blockquote-column | intro\n\n> quoted `oops | <p>intro</p>\n<blockquote>\n<p>quoted `oops</p>\n</blockquote>\nstderr: mdp: 3:10: [unclosed-code-span] code span opened with "`" is never closed; kept as literal text\n | PASS |
| diag-code-span-in-list-item-column | intro\n\n- item `oops | <p>intro</p>\n<ul>\n<li>item `oops</li>\n</ul>\nstderr: mdp: 3:8: [unclosed-code-span] code span opened with "`" is never closed; kept as literal text\n | PASS |
| diag-nested-blockquote-column | > > `oops | <blockquote>\n<blockquote>\n<p>`oops</p>\n</blockquote>\n</blockquote>\nstderr: mdp: 1:5: [unclosed-code-span] code span opened with "`" is never closed; kept as literal text\n | PASS |
| diag-unclosed-code-span-position | alpha\n\nbeta `oops | <p>alpha</p>\n<p>beta `oops</p>\nstderr: mdp: 3:6: [unclosed-code-span] code span opened with "`" is never closed; kept as literal text\n | PASS |
| diag-unclosed-fence-position | alpha\n\n```\nbody\nmore | <p>alpha</p>\n<pre><code>body\nmore\n</code></pre>\nstderr: mdp: 3:1: [unclosed-fence] fenced code block is never closed; the block runs to end of input\n | PASS |
| diag-unmatched-delimiter-in-blockquote | > x\n> **unmatched | <blockquote>\n<p>x\n**unmatched</p>\n</blockquote>\nstderr: mdp: 2:3: [unmatched-delimiter] unmatched "**" emphasis delimiter; kept as literal text\n | PASS |
| diag-unmatched-delimiter-position | alpha\n\nbe *ta | <p>alpha</p>\n<p>be *ta</p>\nstderr: mdp: 3:4: [unmatched-delimiter] unmatched "*" emphasis delimiter; kept as literal text\n | PASS |
| diag-unresolved-link-in-list-item | - x\n- [a](/b | <ul>\n<li>x</li>\n<li>[a](/b</li>\n</ul>\nstderr: mdp: 2:3: [unresolved-link] could not parse link destination; kept as literal text\n | PASS |
| diag-unresolved-link-position | alpha\n\nsee [a](/b | <p>alpha</p>\n<p>see [a](/b</p>\nstderr: mdp: 3:5: [unresolved-link] could not parse link destination; kept as literal text\n | PASS |
| diag-unsafe-url-in-list-item | ok line\n\n> `unclosed\n\n- [x](javascript:alert(1))\n\n**unmatched at line 7 | <p>ok line</p>\n<blockquote>\n<p>`unclosed</p>\n</blockquote>\n<ul>\n<li>x</li>\n</ul>\n<p>**unmatched at line 7</p>\nstderr: mdp: 3:3: [unclosed-code-span] code span opened with "`" is never closed; kept as literal text\nstderr: mdp: 5:3: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept\nstderr: mdp: 7:1: [unmatched-delimiter] unmatched "**" emphasis delimiter; kept as literal text\n | PASS |
| diag-unsafe-url-position | alpha\n\nsee [x](javascript:alert(1)) | <p>alpha</p>\n<p>see x</p>\nstderr: mdp: 3:5: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept\n | PASS |
| inline-em-asterisk | *em* | <p><em>em</em></p>\n | PASS |
| inline-em-strong-triple-run | ***x*** | <p><em><strong>x</strong></em></p>\n | PASS |
| inline-em-underscore | _em_ | <p><em>em</em></p>\n | PASS |
| inline-escape-asterisk | \\*x\\* | <p>*x*</p>\n | PASS |
| inline-escape-backslash | a\\\\b | <p>a\\b</p>\n | PASS |
| inline-intraword-emphasis | un*frigging*believable | <p>un<em>frigging</em>believable</p>\n | PASS |
| inline-intraword-underscore-literal | snake_case_name | <p>snake_case_name</p>\n | PASS |
| inline-nested-em-stack | *a *b* c* | <p><em>a <em>b</em> c</em></p>\n | PASS |
| inline-quadruple-run-literal | ****x**** | <p>****x****</p>\n | PASS |
| inline-spaced-delimiters-literal | a * b * c | <p>a * b * c</p>\n | PASS |
| inline-strong-asterisk | **strong** | <p><strong>strong</strong></p>\n | PASS |
| inline-strong-underscore | __strong__ | <p><strong>strong</strong></p>\n | PASS |
| inline-strong-wrapping-em | **a *b* c** | <p><strong>a <em>b</em> c</strong></p>\n | PASS |
| inline-unequal-run-closes-literal | *foo** | <p>*foo**</p>\nstderr: mdp: 1:1: [unmatched-delimiter] unmatched "*" emphasis delimiter; kept as literal text\n | PASS |
| inline-unequal-run-opens-literal | **foo* | <p>**foo*</p>\nstderr: mdp: 1:1: [unmatched-delimiter] unmatched "**" emphasis delimiter; kept as literal text\n | PASS |
| link-balanced-parens-destination | [t](/a_(b)) | <p><a href="/a_(b)">t</a></p>\n | PASS |
| link-bare-label-literal | [a] | <p>[a]</p>\n | PASS |
| link-basic | [t](/u) | <p><a href="/u">t</a></p>\n | PASS |
| link-broken-literal | [a](/b | <p>[a](/b</p>\nstderr: mdp: 1:1: [unresolved-link] could not parse link destination; kept as literal text\n | PASS |
| link-empty-destination | [a]() | <p><a href="">a</a></p>\n | PASS |
| link-nested-link-suppressed | [a [b](/c) d](/e) | <p><a href="/e">a [b](/c) d</a></p>\n | PASS |
| link-nested-markup-label | [*t*](/u) | <p><a href="/u"><em>t</em></a></p>\n | PASS |
| link-title | [t](/u "T") | <p><a href="/u" title="T">t</a></p>\n | PASS |
| plain-document-golden | This is plain text with no markup.\n\nA second paragraph with punctuation, (parentheses) and [brackets].\nLine two of the second paragraph. | <p>This is plain text with no markup.</p>\n<p>A second paragraph with punctuation, (parentheses) and [brackets].\nLine two of the second paragraph.</p>\n | PASS |
| plain-short-line | Hello, world. | <p>Hello, world.</p>\n | PASS |
| plain-text-escaped | a < b > c & d "e" | <p>a &lt; b &gt; c &amp; d &#34;e&#34;</p>\n | PASS |
| url-href-ampersand-escaped | [a](https://e.com/x?y=1&z=2) | <p><a href="https://e.com/x?y=1&amp;z=2">a</a></p>\n | PASS |
| url-href-quote-escaped | [x](/a"b) | <p><a href="/a&#34;b">x</a></p>\n | PASS |
| url-safe-fragment | [e](#frag) | <p><a href="#frag">e</a></p>\n | PASS |
| url-safe-http | [a](http://e.com) | <p><a href="http://e.com">a</a></p>\n | PASS |
| url-safe-https | [a](https://e.com) | <p><a href="https://e.com">a</a></p>\n | PASS |
| url-safe-mailto | [a](mailto:a@b.c) | <p><a href="mailto:a@b.c">a</a></p>\n | PASS |
| url-safe-protocol-relative | [f](//host/p) | <p><a href="//host/p">f</a></p>\n | PASS |
| url-safe-relative | [d](/r) | <p><a href="/r">d</a></p>\n | PASS |
| url-unsafe-blob | [x](blob:https://a/b) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "blob:https://a/b"; link label is kept\n | PASS |
| url-unsafe-data | [x](data:text/html,<h1>x</h1>) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "data:text/html,<h1>x</h1>"; link label is kept\n | PASS |
| url-unsafe-entity-encoded | [x](&#106;avascript:alert(1)) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "&#106;avascript:alert(1)"; link label is kept\n | PASS |
| url-unsafe-file | [x](file:///etc/passwd) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "file:///etc/passwd"; link label is kept\n | PASS |
| url-unsafe-javascript | [x](javascript:alert(1)) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept\n | PASS |
| url-unsafe-javascript-mixed-case | [x](JaVaScRiPt:alert(1)) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "JaVaScRiPt:alert(1)"; link label is kept\n | PASS |
| url-unsafe-vbscript | [x](vbscript:msgbox) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "vbscript:msgbox"; link label is kept\n | PASS |
