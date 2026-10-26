#!/usr/bin/env bash
#
# build.sh — build the mdp CLI with TinyGo into dist/ for local use.
#
# Usage:
#   scripts/build.sh              # build for the host platform
#   scripts/build.sh os/arch      # e.g. scripts/build.sh darwin/arm64
#
# Output:
#   dist/mdp-<os>-<arch>          # e.g. dist/mdp-darwin-arm64
#
# index.js falls back to dist/mdp-<platform> when bin/mdp-<platform> is absent,
# so a host build here is enough to exercise the package locally. Distributable
# binaries go through scripts/build-binaries.sh instead.
#
# Toolchain: TinyGo (the size-optimized toolchain this project ships with).
# The binary is the same CLI the e2e suite drives: markdown on stdin or a file
# argument -> HTML fragment on stdout, diagnostics on stderr.
#
# Cross-compilation: TinyGo infers the host target from its own toolchain. When
# a target other than the host is requested we attempt a best-effort cross build
# by setting GOOS/GOARCH; TinyGo only supports a limited set of targets, so an
# unsupported os/arch surfaces as a TinyGo error rather than a silently wrong
# binary. The host build is the well-supported, tested path.
#
# Requires: tinygo on PATH. Exits non-zero with a clear message if it is absent.

set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd "$script_dir/.." && pwd)
cd "$repo_root"

if ! command -v tinygo >/dev/null 2>&1; then
  echo "build.sh: tinygo not found on PATH; install TinyGo (https://tinygo.org) or add it to PATH" >&2
  exit 1
fi

# Host platform, used as the default target and to decide host vs cross build.
host_os=$(go env GOOS)
host_arch=$(go env GOARCH)

target="${1:-${host_os}/${host_arch}}"
case "$target" in
  */*) ;;
  *)
    echo "build.sh: target must look like os/arch (e.g. ${host_os}/${host_arch}); got: $target" >&2
    exit 1
    ;;
esac
os=${target%%/*}
arch=${target##*/}
if [ -z "$os" ] || [ -z "$arch" ]; then
  echo "build.sh: target must look like os/arch (e.g. ${host_os}/${host_arch}); got: $target" >&2
  exit 1
fi

# npm's process.arch says x64 where Go says amd64; index.js looks binaries up by
# the npm spelling, so normalise.
[ "$arch" = "amd64" ] && arch=x64

mkdir -p dist
out="dist/mdp-${os}-${arch}"

if [ "$os" = "$host_os" ] && [ "$arch" = "$host_arch" ]; then
  tinygo build -o "$out" ./cmd/cli
else
  # Best-effort cross build; TinyGo supports a limited target set.
  GOOS="$os" GOARCH="$arch" tinygo build -o "$out" ./cmd/cli
fi

size=$(wc -c < "$out" | tr -d '[:space:]')
echo "build.sh: built ${out} (${os}/${arch}, ${size} bytes)"
