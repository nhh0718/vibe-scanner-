#!/bin/bash
# Build script for VibeScanner release
# Creates binaries for all platforms in v{VERSION} folder

VERSION="0.3.0"
OUTDIR="v${VERSION}"
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date +%Y-%m-%d)"

echo "🔨 Building VibeScanner v${VERSION}..."
echo ""

# Create output directory
mkdir -p ${OUTDIR}

# Build Windows AMD64
echo "📦 Building Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${OUTDIR}/vibescanner-windows-amd64.exe .

# Build macOS AMD64 (Intel)
echo "📦 Building macOS AMD64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${OUTDIR}/vibescanner-darwin-amd64 .

# Build macOS ARM64 (Apple Silicon M1/M2/M3)
echo "📦 Building macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${OUTDIR}/vibescanner-darwin-arm64 .

# Build Linux AMD64
echo "📦 Building Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${OUTDIR}/vibescanner-linux-amd64 .

echo ""
echo "✅ Build complete! Binaries in ${OUTDIR}/"
echo ""
echo "📋 Files:"
ls -lh ${OUTDIR}/
echo ""
echo "🚀 To release:"
echo "   1. git tag -a v${VERSION} -m \"Release v${VERSION}\""
echo "   2. git push origin v${VERSION}"
echo "   3. Upload files from ${OUTDIR}/ to GitHub release"
