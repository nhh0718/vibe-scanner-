#!/bin/bash
# Build script for VibeScanner release
# Usage: ./build-release.sh [version]
# If no version provided, reads from latest git tag

set -e

# Version: from argument, git tag, or "dev"
VERSION="${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo 'dev')}"
VERSION="${VERSION#v}"  # Strip leading 'v'

OUTDIR="dist"
MODULE="github.com/nhh0718/vibe-scanner-"
LDFLAGS="-s -w -X ${MODULE}/cmd.version=${VERSION}"

echo "Building VibeScanner v${VERSION}..."
echo ""

# Build web dashboard if needed
if [ ! -d "web/dist" ]; then
    echo "Building web dashboard..."
    cd web && npm ci && npm run build && cd ..
fi

mkdir -p "${OUTDIR}"

# Build all platforms
PLATFORMS="windows/amd64 darwin/amd64 darwin/arm64 linux/amd64"
for PLATFORM in ${PLATFORMS}; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    EXT=""; [ "${GOOS}" = "windows" ] && EXT=".exe"
    OUTPUT="${OUTDIR}/vibescanner-${GOOS}-${GOARCH}${EXT}"

    echo "Building ${GOOS}/${GOARCH}..."
    GOOS="${GOOS}" GOARCH="${GOARCH}" go build -ldflags "${LDFLAGS}" -o "${OUTPUT}" .
done

echo ""
echo "Build complete! Binaries in ${OUTDIR}/"
ls -lh "${OUTDIR}/"
echo ""
echo "To release:"
echo "  git tag -a v${VERSION} -m 'Release v${VERSION}'"
echo "  git push origin v${VERSION}"
echo "  # GitHub Actions will create release automatically"
