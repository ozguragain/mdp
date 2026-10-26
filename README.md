# mdp

Markdown to HTML, by driving the Go `mdp` compiler. A build-time tool: it shells
out to a native binary, so it belongs in a build or server context, never in
browser code.

```bash
npm install @ozguragain/mdp
```

```js
import { compileHtml } from '@ozguragain/mdp'

const html = compileHtml('# Hello **world**')
// "<h1>Hello <strong>world</strong></h1>\n"
```

## Supported syntax

A small, explicit grammar — not CommonMark. Anything outside it renders as
literal text.

**Blocks**

- ATX headings, `#` through `######`
- paragraphs, separated by a blank line
- fenced code blocks, backticks only, with an optional info string
- blockquotes, `>`
- unordered lists, `-`, `*` or `+`
- ordered lists, `1.` or `1)`, starting at the number written
- thematic breaks, `---`, `***` or `___`
- pipe tables with a delimiter row, where `:---`, `:---:` and `---:` set column alignment
- frontmatter between `---` fences at the head of a document, reported verbatim
  rather than parsed — `readFrontmatter` returns it as a string

**Inline**

- text, with backslash escapes for the delimiter characters
- emphasis, `*…*` or `_…_`
- strong, `**…**` or `__…__`
- strikethrough, `~~…~~`
- inline code, `` `…` ``
- links, `[text](url "title")`
- images, `![alt](src "title")`
- autolinks, `<https://example.com>` and `<user@example.com>`
- hard breaks: two trailing spaces, or a trailing `\`, before the newline

A link or image whose URL is unsafe — `javascript:`, `vbscript:`, `data:` —
keeps its text and loses the target, with an `unsafe-url` diagnostic.

**Not supported**, each landing in the output as literal text: setext headings,
indented code blocks, `~~~` fences, reference links, task lists, footnotes, raw
HTML and entity decoding.

## API

| function | returns |
| --- | --- |
| `compileHtml(source, options?)` | the HTML fragment |
| `compile(source, options?)` | `{ html, frontmatter, diagnostics }` |
| `compileFile(path, options?)` | `{ html, frontmatter, diagnostics }` |
| `compileAll(sources, options?)` | the same, keyed by input key |
| `readFrontmatter(source)` | the raw YAML body, or `""` |
| `parseFrontmatter(text)` | scalars and `[a, b]` flow lists as an object |

Options: `headingIds`, `classPrefix`, `root: { tag, className }` to wrap the
fragment in one element.

Diagnostics — the places the compiler recovered from — come back in
`compile().diagnostics`, in source order, with `line`, `column`, `kind` and
`message`. `compileHtml()` drops them.

## Platform binaries

The package ships `bin/mdp-darwin-arm64` and `bin/mdp-linux-x64`. On any other
platform, build one in the [mdp repository](https://github.com/ozguragain/mdp)
with `scripts/build-binaries.sh` and point `MDP_BIN` at it.
