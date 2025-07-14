#!/usr/bin/env pwsh

param(
    [string]$Version = "dev"
)

Write-Host "Building CrossnoKaye CLI for Windows (version: $Version)"

# Clean and create build directories
Remove-Item -Path "dist" -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "dist"

Write-Host "Building binaries..."

# Build for windows/amd64
Write-Host "Building for windows/amd64..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-X github.com/crossnokaye/cli/internal/version.ReleaseTag=$Version" -o dist/windows-amd64/ck.exe ./cmd/ck

# Build for windows/arm64
Write-Host "Building for windows/arm64..."
$env:GOOS = "windows"
$env:GOARCH = "arm64"
go build -ldflags "-X github.com/crossnokaye/cli/internal/version.ReleaseTag=$Version" -o dist/windows-arm64/ck.exe ./cmd/ck

Write-Host "Building MSI installers..."

# Build MSI for x64
Write-Host "Building MSI for x64..."
wix build scripts/wix/installer.wxs -d Platform=x64 -o dist/CrossnoKaye-CLI-x64.msi

# Build MSI for ARM64
Write-Host "Building MSI for ARM64..."
wix build scripts/wix/installer.wxs -d Platform=arm64 -o dist/CrossnoKaye-CLI-arm64.msi

Write-Host "MSI build complete!"
Get-ChildItem dist/*.msi | Format-Table Name, Length, LastWriteTime
