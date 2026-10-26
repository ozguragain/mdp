#!/usr/bin/env bash
#
# check-parity.sh — parity gate: a TinyGo-built mdp must produce byte-identical
# e2e artifacts to a Go-built mdp, otherwise this gate fails.
#
# Usage:
#   scripts/check-parity.sh        # exit 0 on parity PASS, non-zero on FAIL
#
# Why a shadow GOROOT? The e2e suite (e2e/main_test.go) builds the CLI itself
# with `go build -o <tmp>/mdp ./cmd/cli` in TestMain, and there is no hook to
# point it at a prebuilt binary. Worse, `go test` prepends GOROOT/bin to the
# PATH of the test binary, so a shim found earlier on PATH is silently
# bypassed: `exec.Command("go", ...)` resolves the real toolchain and both
# "toolchains" end up being the same build. A gate built that way compares Go
# against Go and can never fail.
#
# So instead of fighting PATH we control GOROOT. We build a shadow GOROOT
# directory that symlinks the real toolchain but replaces bin/go with a shim
# that routes exactly the `go build ./cmd/cli` invocation to the selected
# toolchain (real Go, or TinyGo). Every other `go` invocation is delegated to
# the real toolchain with the real GOROOT. Because GOROOT/bin is what gets
# prepended, the shim is guaranteed to win.
#
# Per toolchain we run `go test -count=1 ./e2e/...` once and snapshot the two
# deterministic artifacts the suite writes (contract in e2e/main_test.go):
#   e2e/artifact/report.md
#   e2e/artifact/rendered.html
# We then compare the two snapshots byte-for-byte. Identical => parity PASS.
#
# -count=1 is load-bearing: without it `go test` returns a cached result, the
# suite never runs, the artifacts stay truncated to empty, and the gate would
# compare two empty files and report a false PASS.
#
# Artifact end state: each run overwrites e2e/artifact/. At exit we restore the
# Go-run artifacts so e2e/artifact/ is left byte-identical to a normal
# `go test ./e2e/...` run (the Go toolchain is the reference). An empty
# snapshot never clobbers a good artifact. Snapshots live in a temp dir that is
# removed on exit.
#
# Stale-artifact safety: before each run we truncate the two artifact files to
# empty (a bash builtin). If a toolchain run fails to write its artifacts, its
# snapshot is empty and is therefore not mistaken for the other toolchain's
# output — a build failure cannot fake a parity match.
#
# Deterministic output: no timestamps, no absolute paths, no randomness. The
# PASS line names the two files; the FAIL diff is produced by `git diff
# --no-index` over the snapshot dirs using relative paths. Exit codes are read
# from files, not from stdout. Run it twice on the same break and the output is
# byte-identical.
#
# Requires: go, tinygo, git on PATH. Exits non-zero with a clear message if any
# is absent.

set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd "$script_dir/.." && pwd)
cd "$repo_root"

real_go=$(command -v go) || { echo "check-parity: go not found on PATH" >&2; exit 1; }
if ! command -v tinygo >/dev/null 2>&1; then
  echo "check-parity: tinygo not found on PATH; install TinyGo or add it to PATH" >&2
  exit 1
fi
if ! command -v git >/dev/null 2>&1; then
  echo "check-parity: git not found on PATH (needed for the diff summary)" >&2
  exit 1
fi

real_goroot=$("$real_go" env GOROOT)
artifact_dir="$repo_root/e2e/artifact"
artifacts=(report.md rendered.html)

tmp=$(mktemp -d)
mkdir -p "$artifact_dir"

# Preserve any pre-existing artifacts as a fallback restore source.
mkdir -p "$tmp/orig"
for f in "${artifacts[@]}"; do
  if [ -f "$artifact_dir/$f" ]; then
    cp "$artifact_dir/$f" "$tmp/orig/$f"
  fi
done

# --- shadow GOROOT -------------------------------------------------------
# A copy of the real toolchain whose bin/go is our routing shim, so that the
# GOROOT/bin PATH entry `go test` prepends resolves to the shim.
shadow="$tmp/goroot"
mkdir -p "$shadow/bin"
for entry in "$real_goroot"/*; do
  name=$(basename "$entry")
  [ "$name" = "bin" ] && continue
  ln -sfn "$entry" "$shadow/$name"
done
for entry in "$real_goroot/bin"/*; do
  name=$(basename "$entry")
  [ "$name" = "go" ] && continue
  ln -sfn "$entry" "$shadow/bin/$name"
done

cat >"$shadow/bin/go" <<'SHIM'
#!/usr/bin/env bash
# go shim for check-parity: route `go build ... ./cmd/cli` to the selected
# toolchain; delegate everything else to the real go under the real GOROOT.
set -euo pipefail
real_go="${MDP_REAL_GO:?go shim: MDP_REAL_GO not set}"
real_goroot="${MDP_REAL_GOROOT:?go shim: MDP_REAL_GOROOT not set}"
toolchain="${MDP_TOOLCHAIN:?go shim: MDP_TOOLCHAIN not set}"

if [ "${1:-}" = "build" ]; then
  for a in "$@"; do
    if [ "$a" = "./cmd/cli" ]; then
      shift
      case "$toolchain" in
        go)
          GOROOT="$real_goroot" exec "$real_go" build -ldflags "-s -w" "$@"
          ;;
        tinygo)
          GOROOT="$real_goroot" exec tinygo build "$@"
          ;;
        *)
          echo "go shim: unknown MDP_TOOLCHAIN '$toolchain'" >&2
          exit 2
          ;;
      esac
    fi
  done
fi

GOROOT="$real_goroot" exec "$real_go" "$@"
SHIM
chmod +x "$shadow/bin/go"

# shellcheck disable=SC2329  # restore is invoked indirectly via `trap restore EXIT`.
restore() {
  set +e
  mkdir -p "$artifact_dir"
  # Prefer the Go-run artifact and never let an empty snapshot clobber a good
  # one: a truncated file left by a failed run must not overwrite the original.
  for f in "${artifacts[@]}"; do
    if [ -s "$tmp/go/$f" ]; then
      cp "$tmp/go/$f" "$artifact_dir/$f"
    elif [ -s "$tmp/orig/$f" ]; then
      cp "$tmp/orig/$f" "$artifact_dir/$f"
    fi
  done
  # Quiet cleanup: some environments wrap `rm` in a recoverable-trash shim that
  # prints to stdout; silence it and never let cleanup failure fail the gate.
  rm -rf "$tmp" >/dev/null 2>&1 || true
}
trap restore EXIT

# run_toolchain <go|tinygo> <snapshot-dir>
# Runs the suite once for the toolchain and snapshots its artifacts. The suite
# exit code is written to <snapshot-dir>.rc (a file, immune to stdout noise).
run_toolchain() {
  local toolchain="$1"
  local outdir="$2"

  # Truncate stale artifacts (bash builtin; a failed run leaves them empty and
  # is therefore caught by the byte comparison instead of faking a match).
  for f in "${artifacts[@]}"; do
    : >"$artifact_dir/$f"
  done

  set +e
  GOROOT="$shadow" MDP_REAL_GO="$real_go" MDP_REAL_GOROOT="$real_goroot" MDP_TOOLCHAIN="$toolchain" \
    "$real_go" test -count=1 ./e2e/... >/dev/null 2>"$tmp/$toolchain.testlog"
  echo "$?" >"$tmp/$toolchain.rc"
  set -e

  mkdir -p "$outdir"
  for f in "${artifacts[@]}"; do
    if [ -f "$artifact_dir/$f" ]; then
      cp "$artifact_dir/$f" "$outdir/$f"
    fi
  done
}

run_toolchain go "$tmp/go"
run_toolchain tinygo "$tmp/tinygo"

go_rc=$(cat "$tmp/go.rc")
tinygo_rc=$(cat "$tmp/tinygo.rc")

parity_ok=true
for f in "${artifacts[@]}"; do
  if ! cmp -s "$tmp/go/$f" "$tmp/tinygo/$f"; then
    parity_ok=false
  fi
done

suites_ok=true
if [ "$go_rc" != "0" ] || [ "$tinygo_rc" != "0" ]; then
  suites_ok=false
fi

if [ "$parity_ok" = true ] && [ "$suites_ok" = true ]; then
  echo "check-parity: PASS — Go and TinyGo builds produced byte-identical e2e artifacts (report.md, rendered.html)."
  exit 0
fi

echo "check-parity: FAIL" >&2
if [ "$suites_ok" = false ]; then
  echo "  suite exit code: go=${go_rc} tinygo=${tinygo_rc} (non-zero means that toolchain's run failed)" >&2
fi
if [ "$parity_ok" = true ]; then
  echo "  artifacts are identical but a suite run failed — a build problem, not a toolchain parity gap" >&2
else
  echo "  artifacts differ between the Go and TinyGo builds (byte-for-byte):" >&2
  for f in "${artifacts[@]}"; do
    if ! cmp -s "$tmp/go/$f" "$tmp/tinygo/$f"; then
      echo "    - $f" >&2
    fi
  done
  echo "  unified diff (left=go, right=tinygo):" >&2
  (cd "$tmp" && git --no-pager diff --no-index --no-color go tinygo) >&2 || true
fi
exit 1
