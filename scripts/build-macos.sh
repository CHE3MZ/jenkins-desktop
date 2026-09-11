#!/bin/bash
# Build Jenkins Desktop for macOS: ./scripts/build-macos.sh
# Extra args are passed through to `wails build` (e.g. -clean).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

for tool in go bun wails; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "error: required tool '$tool' not found on PATH" >&2
    exit 1
  fi
done

if [ ! -f build/darwin/icon.icns ]; then
  echo "warning: build/darwin/icon.icns is missing, the app will use the default Wails icon" >&2
fi

# Newer Xcode toolchains fail to link the Wails webview without this
# (undefined _OBJC_CLASS_$_UTType). Harmless on older toolchains.
export CGO_LDFLAGS="${CGO_LDFLAGS:-} -framework UniformTypeIdentifiers"

wails build "$@"

echo "built: $ROOT/build/bin/Jenkins Desktop.app"
