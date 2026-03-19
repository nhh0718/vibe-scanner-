# Build script for VibeScanner on Windows
# Run: .\build.ps1

$VERSION = "0.7.0"
$OUTDIR = "v$VERSION"
$COMMIT = git rev-parse --short HEAD
$DATE = Get-Date -Format "yyyy-MM-dd"
$LDFLAGS = "-s -w -X github.com/nhh0718/vibe-scanner-/cmd.version=$VERSION -X github.com/nhh0718/vibe-scanner-/cmd.commit=$COMMIT -X github.com/nhh0718/vibe-scanner-/cmd.date=$DATE"

Write-Host "Building VibeScanner v$VERSION..." -ForegroundColor Cyan
Write-Host ""

New-Item -ItemType Directory -Force -Path $OUTDIR | Out-Null

Write-Host "Building Windows AMD64..." -ForegroundColor Yellow
$env:GOOS = "windows"; $env:GOARCH = "amd64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-windows-amd64.exe" .

Write-Host "Building macOS AMD64..." -ForegroundColor Yellow
$env:GOOS = "darwin"; $env:GOARCH = "amd64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-darwin-amd64" .

Write-Host "Building macOS ARM64..." -ForegroundColor Yellow
$env:GOOS = "darwin"; $env:GOARCH = "arm64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-darwin-arm64" .

Write-Host "Building Linux AMD64..." -ForegroundColor Yellow
$env:GOOS = "linux"; $env:GOARCH = "amd64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-linux-amd64" .

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Build complete! Binaries in $OUTDIR/" -ForegroundColor Green
Write-Host ""
Get-ChildItem $OUTDIR | Select-Object Name, @{N="Size(MB)";E={[math]::Round($_.Length/1MB,2)}}
