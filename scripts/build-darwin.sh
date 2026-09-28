#!/bin/bash
# Builds a macOS installer package (dist/xn.pkg) holding static amd64 and
# arm64 binaries, staged in /tmp/xn-install; the postinstall script copies
# the one matching the host to /usr/local/bin.

set -euo pipefail

VERSION=${1:-dev}
LDFLAGS="-X github.com/xeger/installer/internal/platform.version=$VERSION"

echo "Building xn for Darwin (version: $VERSION)"

rm -rf dist pkg-build
mkdir -p dist pkg-build/scripts pkg-build/root/xn-install

for arch in amd64 arm64; do
  echo "Building for darwin/$arch..."
  CGO_ENABLED=0 GOOS=darwin GOARCH=$arch go build -trimpath -ldflags "$LDFLAGS" -o "dist/darwin-$arch/xn" ./cmd/xn
done

echo "Creating PKG installer scripts..."
cp scripts/pkg/postinstall pkg-build/scripts/
chmod +x pkg-build/scripts/postinstall

echo "Staging binaries for PKG..."
cp -r dist/* pkg-build/root/xn-install/

echo "Building PKG..."
pkgbuild \
  --root pkg-build/root \
  --scripts pkg-build/scripts \
  --identifier net.xeger.xn \
  --version "${VERSION#v}" \
  --install-location /tmp \
  dist/xn.pkg

echo "PKG build complete!"
ls -la dist/xn.pkg
