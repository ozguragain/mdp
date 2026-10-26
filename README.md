# @ozguragain/mdp

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

`bin/mdp-darwin-arm64` and `bin/mdp-linux-x64` are committed. On anything else,
build a binary with `scripts/build-binaries.sh` and commit it, or point
`MDP_BIN` at one.

## Installing from the repository instead

Every release can also be consumed straight from a commit archive, with no
registry and no login:

```bash
npm i https://codeload.github.com/ozguragain/mdp/tar.gz/<commit>
```
