export interface Root {
  /** Wrapper element tag. Defaults to "article". */
  tag?: string
  /** Extra class on the wrapper. */
  className?: string
}

export interface Options {
  /** Add a slug id to each heading. */
  headingIds?: boolean
  /** Add "<prefix>-<construct>" classes to rendered elements. */
  classPrefix?: string
  /** Wrap the rendered fragment in one element. */
  root?: Root
}

export interface Diagnostic {
  kind: string
  message: string
  line: number
  column: number
}

export interface CompileResult {
  /** The escaped HTML fragment. */
  html: string
  /** Raw YAML body from the head of the source, or "" when there was none. */
  frontmatter: string
  /** Everything the compiler recovered from, in source order. */
  diagnostics: Diagnostic[]
}

/** Compile markdown source to HTML plus its YAML frontmatter. */
export function compile(source: string, options?: Options): CompileResult

/**
 * Read only the YAML frontmatter from a source, without rendering anything.
 * One process spawn — enough for a post listing.
 */
export function readFrontmatter(source: string): string

/** Compile a markdown file. */
export function compileFile(path: string, options?: Options): CompileResult

/** Compile many documents, keyed however you like. */
export function compileAll(
  sources: Record<string, string>,
  options?: Options,
): Record<string, CompileResult>

/**
 * Split raw frontmatter text into a plain object. Minimal on purpose: scalars
 * and inline flow lists (`tags: [a, b]`); block lists and nested maps stay
 * out of scope.
 */
export function parseFrontmatter(text: string): Record<string, unknown>
