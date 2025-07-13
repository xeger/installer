#!/bin/bash

set -e

VERSION=${1:-dev}

echo "Building CrossnoKaye CLI for Darwin (version: $VERSION)"

# Clean and create build directories
rm -rf dist pkg-build
mkdir -p dist
mkdir -p pkg-build/usr/local/bin
mkdir -p pkg-build/scripts
mkdir -p pkg-build/tmp/ck-install

echo "Building binaries..."

# Build for darwin/amd64
echo "Building for darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -o dist/darwin-amd64/ck ./cmd/ck

# Build for darwin/arm64
echo "Building for darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -o dist/darwin-arm64/ck ./cmd/ck

echo "Creating PKG installer scripts..."

# Copy installer scripts
cp scripts/pkg/postinstall pkg-build/scripts/
cp scripts/pkg/preinstall pkg-build/scripts/
chmod +x pkg-build/scripts/postinstall
chmod +x pkg-build/scripts/preinstall

echo "Staging binaries for PKG..."
cp -r dist/* pkg-build/tmp/ck-install/

echo "Building PKG..."
pkgbuild \
  --root pkg-build \
  --scripts pkg-build/scripts \
  --identifier com.crossnokaye.ck \
  --version "$VERSION" \
  --install-location / \
  CrossnoKaye-CLI.pkg

echo "PKG build complete!"
ls -la CrossnoKaye-CLI.pkg