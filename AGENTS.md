# AGENTS.md

## Testing policy

These rules govern all testing work in this repository (mdp, a markdown → HTML compiler).

1. **Never write unit tests after you write code.** Tests written after the fact only
   confirm what the code already does instead of specifying behavior. Do not add them.

2. **Highly prefer E2E tests as the sole testing mechanism.** Drive the real CLI
   (markdown on stdin or a file argument → HTML fragment on stdout, diagnostics on
   stderr) and assert the full observable output: HTML, diagnostic lines, exit code.

3. **Use E2E tests to verify complex features work.** New behavior — grammar rules,
   URL sanitization, source-position mapping, diagnostics — is proven end to end,
   not by testing internals.

4. **At the end of E2E tests, produce a verifiable and repeatable artifact.** The
   artifact must be deterministic (no timestamps, no absolute paths, no randomness)
   so two runs diff byte-for-byte and a reviewer can verify exactly what was checked.

5. **If you must test a system in isolation, first write down all the ways it could
   fail, then write the code.** The failure-mode list comes before the implementation
   and shapes it. It is the only justification for an isolated test.

## Code readability

6. **Write readable code and annotate only where it is earned.** Do not put
   command-style comments on code that already reads clearly — readable code needs
   no narration. When readability is genuinely hard, or when you suspect you are
   overengineering, leave a short note that points out the difficulty or the
   suspicion and explains it; that note is enough.
