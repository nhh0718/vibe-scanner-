# Build script for VibeScanner on Windows
# Run: .\build.ps1

$VERSION = "0.3.0"
$OUTDIR = "v$VERSION"
$COMMIT = git rev-parse --short HEAD
$DATE = Get-Date -Format "yyyy-MM-dd"
$LDFLAGS = "-s -w -X github.com/nhh0718/vibe-scanner-/cmd.version=$VERSION -X github.com/nhh0718/vibe-scanner-/cmd.commit=$COMMIT -X github.com/nhh0718/vibe-scanner-/cmd.date=$DATE"

Write-Host "🔨 Building VibeScanner v$VERSION..." -ForegroundColor Cyan
Write-Host ""

# Create output directory
New-Item -ItemType Directory -Force -Path $OUTDIR | Out-Null

# Build Windows AMD64
Write-Host " Building Windows AMD64..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-windows-amd64.exe" .

# Build macOS AMD64 (Intel)
Write-Host "Building macOS AMD64 for Intel..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-darwin-amd64" .

# Build macOS ARM64 (Apple Silicon)
Write-Host "Building macOS ARM64 for Apple Silicon..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-darwin-arm64" .

# Build Linux AMD64
Write-Host " Building Linux AMD64..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags "$LDFLAGS" -o "$OUTDIR/vibescanner-linux-amd64" .

# Clear env
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Build complete! Binaries in $OUTDIR/" -ForegroundColor Green
Write-Host ""
Write-Host " Files:"
Get-ChildItem $OUTDIR | Select-Object Name, @{N="Size(MB)";E={[math]::Round($_.Length/1MB,2)}}
Write-Host ""
Write-Host "🚀 To release:"
Write-Host "   1. git tag -a v$VERSION -m \"Release v$VERSION\""
Write-Host "   2. git push origin v$VERSION"
Write-Host "   3. Upload files from $OUTDIR/ to GitHub release"
