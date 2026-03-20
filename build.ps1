# Build script for VibeScanner on Windows
# Usage: .\build.ps1 [version]
# If no version provided, reads from latest git tag

param([string]$Ver)
$ErrorActionPreference = 'Stop'

if (-not $Ver) {
    $Ver = (git describe --tags --abbrev=0 2>$null) -replace '^v',''
    if (-not $Ver) { $Ver = "dev" }
}

$OUTDIR = "dist"
$MODULE = "github.com/nhh0718/vibe-scanner-"
$LDFLAGS = "-s -w -X $MODULE/cmd.version=$Ver"

Write-Host "Building VibeScanner v$Ver..." -ForegroundColor Cyan
Write-Host ""

# Build web dashboard if needed
if (-not (Test-Path "web/dist")) {
    Write-Host "Building web dashboard..." -ForegroundColor Yellow
    Push-Location web
    npm ci
    npm run build
    Pop-Location
}

New-Item -ItemType Directory -Force -Path $OUTDIR | Out-Null

$targets = @(
    @{ GOOS="windows"; GOARCH="amd64"; EXT=".exe" },
    @{ GOOS="darwin";  GOARCH="amd64"; EXT="" },
    @{ GOOS="darwin";  GOARCH="arm64"; EXT="" },
    @{ GOOS="linux";   GOARCH="amd64"; EXT="" }
)

foreach ($t in $targets) {
    $output = "$OUTDIR/vibescanner-$($t.GOOS)-$($t.GOARCH)$($t.EXT)"
    Write-Host "Building $($t.GOOS)/$($t.GOARCH)..." -ForegroundColor Yellow
    $env:GOOS = $t.GOOS; $env:GOARCH = $t.GOARCH
    go build -ldflags "$LDFLAGS" -o $output .
}

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Build complete! Binaries in $OUTDIR/" -ForegroundColor Green
Get-ChildItem $OUTDIR | Select-Object Name, @{N="Size(MB)";E={[math]::Round($_.Length/1MB,2)}}
