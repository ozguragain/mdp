#!/usr/bin/env bash
#
# build-binaries.sh — build the platform binaries the npm package ships.
#
# Usage:
#   scripts/build-binaries.sh                    # every platform
#   scripts/build-binaries.sh darwin/arm64       # just one
#
# Output: bin/mdp-<os>-<arch>, the exact names index.js resolves at runtime.
# Commit the results. The package is consumed as a git dependency, so it is
# installed as-is: the consumer's machine has no TinyGo to build Go with, and
# a git dependency cannot pull per-platform packages.
#
# Requires: tinygo on PATH.

set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd "$script_dir/.." && pwd)
cd "$repo_root"

if ! command -v tinygo >/dev/null 2>&1; then
  echo "build-binaries.sh: tinygo not found on PATH; install TinyGo (https://tinygo.org) or add it to PATH" >&2
  exit 1
fi

targets=${*:-"darwin/arm64 darwin/amd64 linux/amd64 linux/arm64"}

for target in $targets; do
  case "$target" in
    */*) ;;
    *)
      echo "build-binaries.sh: target must look like os/arch (e.g. darwin/arm64); got: $target" >&2
      exit 1
      ;;
  esac
  goos=${target%%/*}
  goarch=${target##*/}

  # npm's process.arch says x64 where Go says amd64; the filename has to match
  # what index.js computes, so normalise here.
  arch=$goarch
  [ "$arch" = "amd64" ] && arch=x64

  mkdir -p bin
  out="bin/mdp-${goos}-${arch}"
  echo "build-binaries.sh: building $goos/$goarch"
  GOOS="$goos" GOARCH="$goarch" tinygo build -no-debug -o "$out" ./cmd/cli
  chmod +x "$out"
  echo "build-binaries.sh: ${out} ($(wc -c <"$out" | tr -d '[:space:]') bytes)"
done
