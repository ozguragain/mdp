# mdp end-to-end report

One row per executed case, sorted by case id. Deterministic: no timestamps, no absolute paths, no randomness.
The "expected output" cell holds the expected stdout followed by `stderr: ...` lines for expected diagnostics; cell newlines appear as literal \n.

| case id | input | expected output | result |
| --- | --- | --- | --- |
| autolink-email | <user@example.com> | <p><a href="mailto:user@example.com">user@example.com</a></p>\n | PASS |
| autolink-plain-angle-text-is-literal | <foo> | <p>&lt;foo&gt;</p>\n | PASS |
| autolink-unclosed-is-literal | <https://example.com | <p>&lt;https://example.com</p>\n | PASS |
| autolink-unsafe-scheme-keeps-text | <javascript:alert(1)> | <p>javascript:alert(1)</p>\nstderr: mdp: 1:1: [unsafe-url] dropped href with unsafe URL "javascript:alert(1)"; link label is kept\n | PASS |
| autolink-uri | <https://example.com/a> | <p><a href="https://example.com/a">https://example.com/a</a></p>\n | PASS |
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
| frontmatter-body-only | ---\ntitle: x\n---\nbody | <p>body</p>\n | PASS |
| frontmatter-closed-by-dots | ---\ntitle: x\n...\n\nbody | <p>body</p>\n | PASS |
| frontmatter-empty-body | ---\n---\nbody | <p>body</p>\n | PASS |
| frontmatter-is-not-rendered | ---\ntitle: x\n---\n\n# H | <h1>H</h1>\n | PASS |
| frontmatter-only-at-start-of-document | x\n\n---\ntitle: y\n--- | <p>x</p>\n<hr>\n<p>title: y</p>\n<hr>\n | PASS |
| frontmatter-unclosed-is-thematic-break | ---\ntitle: x | <hr>\n<p>title: x</p>\n | PASS |
| hardbreak-backslash | one\\\ntwo | <p>one<br>\ntwo</p>\n | PASS |
| hardbreak-single-space-stays-soft | one \ntwo | <p>one \ntwo</p>\n | PASS |
| hardbreak-three-spaces | a   \nb | <p>a<br>\nb</p>\n | PASS |
| hardbreak-trailing-spaces | one  \ntwo | <p>one<br>\ntwo</p>\n | PASS |
| image-alt-is-literal-text | ![a *b*](/x.png) | <p><img src="/x.png" alt="a *b*"></p>\n | PASS |
| image-basic | ![alt](/x.png) | <p><img src="/x.png" alt="alt"></p>\n | PASS |
| image-empty-alt | ![](/x.png) | <p><img src="/x.png" alt=""></p>\n | PASS |
| image-in-paragraph | before ![a](/p.jpg) after | <p>before <img src="/p.jpg" alt="a"> after</p>\n | PASS |
| image-title | ![alt](/x.png "title") | <p><img src="/x.png" alt="alt" title="title"></p>\n | PASS |
| image-unsafe-src-keeps-alt-text | ![x](javascript:alert(1)) | <p>x</p>\nstderr: mdp: 1:1: [unsafe-url] dropped img src with unsafe URL "javascript:alert(1)"; alt text is kept\n | PASS |
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
| ordered-basic | 1. one\n2. two | <ol>\n<li>one</li>\n<li>two</li>\n</ol>\n | PASS |
| ordered-continuation-line | 1. one\n   continued\n2. two | <ol>\n<li>one\ncontinued</li>\n<li>two</li>\n</ol>\n | PASS |
| ordered-digits-without-marker-are-text | 1.no-space | <p>1.no-space</p>\n | PASS |
| ordered-marker-change-ends-the-list | 1. one\n- two | <ol>\n<li>one</li>\n</ol>\n<ul>\n<li>two</li>\n</ul>\n | PASS |
| ordered-paren-marker | 1) one | <ol>\n<li>one</li>\n</ol>\n | PASS |
| ordered-start-keeps-its-number | 3. three\n4. four | <ol start="3">\n<li>three</li>\n<li>four</li>\n</ol>\n | PASS |
| plain-document-golden | This is plain text with no markup.\n\nA second paragraph with punctuation, (parentheses) and [brackets].\nLine two of the second paragraph. | <p>This is plain text with no markup.</p>\n<p>A second paragraph with punctuation, (parentheses) and [brackets].\nLine two of the second paragraph.</p>\n | PASS |
| plain-short-line | Hello, world. | <p>Hello, world.</p>\n | PASS |
| plain-text-escaped | a < b > c & d "e" | <p>a &lt; b &gt; c &amp; d &#34;e&#34;</p>\n | PASS |
| strike-basic | ~~gone~~ | <p><del>gone</del></p>\n | PASS |
| strike-holds-other-inline | ~~a **b** c~~ | <p><del>a <strong>b</strong> c</del></p>\n | PASS |
| strike-single-tilde-literal | ~x~ | <p>~x~</p>\n | PASS |
| strike-triple-tilde-literal | ~~~x~~~ | <p>~~~x~~~</p>\n | PASS |
| strike-unclosed-literal | ~~open | <p>~~open</p>\n | PASS |
| table-basic | \| a \| b \|\n\| --- \| --- \|\n\| 1 \| 2 \| | <table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n | PASS |
| table-cells-parse-inline | \| *a* \|\n\| --- \|\n\| `b` \| | <table>\n<thead>\n<tr>\n<th><em>a</em></th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td><code>b</code></td>\n</tr>\n</tbody>\n</table>\n | PASS |
| table-column-alignment | \| l \| c \| r \|\n\| :-- \| :-: \| --: \|\n\| 1 \| 2 \| 3 \| | <table>\n<thead>\n<tr>\n<th align="left">l</th>\n<th align="center">c</th>\n<th align="right">r</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td align="left">1</td>\n<td align="center">2</td>\n<td align="right">3</td>\n</tr>\n</tbody>\n</table>\n | PASS |
| table-header-only-has-no-tbody | \| a \|\n\| --- \| | <table>\n<thead>\n<tr>\n<th>a</th>\n</tr>\n</thead>\n</table>\n | PASS |
| table-short-row-is-padded | \| a \| b \|\n\| --- \| --- \|\n\| 1 \| | <table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td></td>\n</tr>\n</tbody>\n</table>\n | PASS |
| table-without-delimiter-row-stays-text | a \| b\nmore | <p>a \| b\nmore</p>\n | PASS |
| thematic-break-dashes | a\n\n---\n\nb | <p>a</p>\n<hr>\n<p>b</p>\n | PASS |
| thematic-break-four-runs | ---- | <hr>\n | PASS |
| thematic-break-spaced-dashes | - - - | <hr>\n | PASS |
| thematic-break-stars | *** | <hr>\n | PASS |
| thematic-break-two-dashes-are-text | -- | <p>--</p>\n | PASS |
| thematic-break-underscores | ___ | <hr>\n | PASS |
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
