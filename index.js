// mdp — compile markdown to HTML by driving the Go mdp binary.
//
// This is a build-time tool. It shells out to a native binary, so it belongs
// in a build or server context (a Next.js page, a static site generator, a
// script) and never in browser code.
//
// The binaries are committed to the repository, one per platform, because a
// git dependency cannot pull optional platform packages and the build machine
// (a CI runner, a deploy host) has no TinyGo to produce one. Resolution order:
//   1. $MDP_BIN, if set
//   2. <this package>/bin/mdp-<platform>   (committed, what a git install uses)
//   3. <this package>/dist/mdp-<platform>  (a local build, not committed)
//
// Each document takes two spawns: one renders HTML and reports diagnostics,
// one reports the YAML frontmatter. That is a couple of milliseconds per post
// at build time; the alternative — teaching the Go side to emit JSON — costs
// ~200 KB of binary for nothing gained here.

import { spawnSync } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/** @typedef {{ headingIds?: boolean, classPrefix?: string, root?: { tag?: string, className?: string } }} Options */
/** @typedef {{ kind: string, message: string, line: number, column: number }} Diagnostic */
/** @typedef {{ html: string, frontmatter: string, diagnostics: Diagnostic[] }} CompileResult */

const packageDir = path.dirname(fileURLToPath(import.meta.url))
const platform = `${process.platform}-${process.arch}`

let resolved

function binaryPath() {
  if (resolved) return resolved

  const fromEnv = process.env.MDP_BIN
  if (fromEnv) {
    resolved = fromEnv
    return resolved
  }

  // Built with path.join rather than `new URL(...)`: bundlers treat the URL
  // form as a static asset request and fail when the binary is not in the
  // module graph.
  const candidates = [
    path.join(packageDir, 'bin', `mdp-${platform}`),
    path.join(packageDir, 'dist', `mdp-${platform}`),
  ]
  for (const candidate of candidates) {
    // The path is computed, so no bundler can scope it statically; without the
    // opt-out Turbopack traces the whole project into the server output.
    if (existsSync(/* turbopackIgnore: true */ candidate)) {
      resolved = candidate
      return resolved
    }
  }

  throw new Error(
    `mdp: no binary for ${platform}. Build one with scripts/build.sh, or point MDP_BIN at it.`,
  )
}

function flagsFor(options) {
  const flags = []
  if (options.headingIds) flags.push('-heading-ids')
  if (options.classPrefix) flags.push(`-class-prefix=${options.classPrefix}`)
  if (options.root?.tag) flags.push(`-root-tag=${options.root.tag}`)
  if (options.root?.className) flags.push(`-root-class=${options.root.className}`)
  return flags
}

// invoke runs the binary with the source on stdin and returns both streams.
// Diagnostics come back on stderr as "mdp: line:col: [kind] message".
function invoke(flags, source) {
  const res = spawnSync(
    /* turbopackIgnore: true */ binaryPath(),
    flags,
    {
      input: source,
      encoding: 'utf8',
      maxBuffer: 64 * 1024 * 1024,
    },
  )
  if (res.error) throw res.error
  if (res.status !== 0) {
    throw new Error(`mdp exited with ${res.status}: ${res.stderr.trim()}`)
  }
  return { stdout: res.stdout, stderr: res.stderr }
}

function parseDiagnostics(stderr) {
  const out = []
  for (const line of stderr.split('\n')) {
    const m = /^mdp: (\d+):(\d+): \[([^\]]+)\] (.*)$/.exec(line)
    if (m) {
      out.push({ line: Number(m[1]), column: Number(m[2]), kind: m[3], message: m[4] })
    }
  }
  return out
}

/**
 * Compile markdown source to HTML plus its YAML frontmatter.
 *
 * @param {string} source markdown, optionally opening with "---" frontmatter
 * @param {Options} [options]
 * @returns {CompileResult}
 */
export function compile(source, options = {}) {
  const flags = flagsFor(options)

  const rendered = invoke(flags, source)
  const meta = invoke([...flags, '-frontmatter'], source)

  return {
    html: rendered.stdout,
    frontmatter: meta.stdout,
    diagnostics: parseDiagnostics(rendered.stderr),
  }
}

/**
 * Compile markdown to an HTML fragment — `compile` without the frontmatter and
 * diagnostics, for callers that only want the markup.
 *
 * @param {string} source markdown, optionally opening with "---" frontmatter
 * @param {Options} [options]
 * @returns {string} the escaped HTML fragment
 */
export function compileHtml(source, options = {}) {
  return compile(source, options).html
}

/**
 * Read only the YAML frontmatter from a source, without rendering anything.
 * One process spawn — enough for a post listing.
 *
 * @param {string} source
 * @returns {string} the raw YAML body, or "" when there is none
 */
export function readFrontmatter(source) {
  return invoke(['-frontmatter'], source).stdout
}

/**
 * Compile a markdown file.
 *
 * @param {string} path
 * @param {Options} [options]
 * @returns {CompileResult}
 */
export function compileFile(path, options = {}) {
  return compile(readFileSync(path, 'utf8'), options)
}

/**
 * Compile many documents, keyed however you like. Two spawns per document —
 * fine for a blog build, where this is called once per post.
 *
 * @param {Record<string, string>} sources
 * @param {Options} [options]
 * @returns {Record<string, CompileResult>}
 */
export function compileAll(sources, options = {}) {
  const out = {}
  for (const [key, source] of Object.entries(sources)) out[key] = compile(source, options)
  return out
}

// scalar reads one YAML scalar: quoted text stays text, true/false become
// booleans, numerals become numbers, anything else stays a string.
function scalar(raw) {
  const unquoted = raw.replace(/^["'](.*)["']$/, '$1')
  if (raw !== unquoted) return unquoted
  if (unquoted === 'true') return true
  if (unquoted === 'false') return false
  if (unquoted !== '' && !Number.isNaN(Number(unquoted))) return Number(unquoted)
  return unquoted
}

// splitFlow splits a comma-separated list, honouring quotes so that
// ["a, b", "c"] stays two items rather than three.
function splitFlow(inner) {
  const items = []
  let current = ''
  let quote = null
  for (const ch of inner) {
    if (quote) {
      current += ch
      if (ch === quote) quote = null
    } else if (ch === '"' || ch === "'") {
      quote = ch
      current += ch
    } else if (ch === ',') {
      items.push(current.trim())
      current = ''
    } else {
      current += ch
    }
  }
  if (current.trim() !== '') items.push(current.trim())
  return items
}

/**
 * Split raw frontmatter text into a plain object. Deliberately minimal:
 * scalars and inline "flow" lists (`tags: [a, b]`). Block lists and nested
 * maps are out of scope — reach for a real YAML parser if you need those.
 *
 * @param {string} text the raw body mdp reports, without the "---" fences
 * @returns {Record<string, unknown>}
 */
export function parseFrontmatter(text) {
  const out = {}
  for (const line of text.split('\n')) {
    const at = line.indexOf(':')
    if (at <= 0) continue
    const key = line.slice(0, at).trim()
    const raw = line.slice(at + 1).trim()
    if (!key) continue

    if (raw.startsWith('[') && raw.endsWith(']')) {
      const inner = raw.slice(1, -1).trim()
      out[key] = inner === '' ? [] : splitFlow(inner).map(scalar)
    } else {
      out[key] = scalar(raw)
    }
  }
  return out
}
