#!/usr/bin/env pwsh
# Builds Windows installers (dist/xn-x64.msi, dist/xn-arm64.msi) holding
# static binaries. The MSI installs xn under Program Files and adds it to PATH.

param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"

Write-Host "Building xn for Windows (version: $Version)"

Remove-Item -Path "dist" -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "dist" | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$ldflags = "-X github.com/xeger/installer/internal/platform.version=$Version"

foreach ($arch in "amd64", "arm64") {
    Write-Host "Building for windows/$arch..."
    $env:GOARCH = $arch
    go build -trimpath -ldflags $ldflags -o "dist/windows-$arch/xn.exe" ./cmd/xn
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

# MSI versions must be numeric (major.minor.build); non-release builds get 0.0.0.
$msiVersion = "0.0.0"
if ($Version -match '^v?(\d+\.\d+\.\d+)$') { $msiVersion = $Matches[1] }

foreach ($platform in "x64", "arm64") {
    Write-Host "Building MSI for $platform (version $msiVersion)..."
    wix build scripts/wix/installer.wxs -d Platform=$platform -d Version=$msiVersion -o "dist/xn-$platform.msi"
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

Write-Host "MSI build complete!"
Get-ChildItem dist/*.msi | Format-Table Name, Length, LastWriteTime
