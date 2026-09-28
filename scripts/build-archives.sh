#!/bin/bash
# Builds static release archives for every supported platform:
#   dist/xn_<version>_<os>_<arch>.tar.gz, each containing xn (xn.exe on Windows).
# This is the same asset convention xn itself uses to install tools.

set -euo pipefail

VERSION=${1:-dev}
LDFLAGS="-X github.com/xeger/installer/internal/platform.version=$VERSION"

mkdir -p dist
for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
  os=${target%/*}
  arch=${target#*/}
  exe=xn
  [ "$os" = windows ] && exe=xn.exe
  stage=$(mktemp -d)
  echo "Building $os/$arch..."
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "$LDFLAGS" -o "$stage/$exe" ./cmd/xn
  tar -czf "dist/xn_${VERSION}_${os}_${arch}.tar.gz" -C "$stage" "$exe"
  rm -rf "$stage"
done

ls -la dist/*.tar.gz
