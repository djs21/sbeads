#!/usr/bin/env bash
set -euo pipefail

# Default settings
MODE="dev"
TARGET_OS=""
TARGET_ARCH=""

# Parse arguments
for arg in "$@"; do
  case "$arg" in
    --dev)
      MODE="dev"
      ;;
    --prod)
      MODE="prod"
      ;;
    --os=*)
      TARGET_OS="${arg#*=}"
      ;;
    --arch=*)
      TARGET_ARCH="${arg#*=}"
      ;;
    --help|-h)
      echo "sbeads Build Script"
      echo "Usage: ./build.sh [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --dev            Build development binary (Version: dev-<commit>)"
      echo "  --prod           Build production binary (Version: <git-tag>, stripped: -s -w -trimpath)"
      echo "  --os=<os>        Target operating system (linux, darwin, windows)"
      echo "  --arch=<arch>    Target architecture (amd64, arm64)"
      echo "  --help, -h       Show this help message"
      exit 0
      ;;
    *)
      echo "❌ Unknown option: $arg"
      echo "Run './build.sh --help' for usage."
      exit 1
      ;;
  esac
done

# Detect host OS
HOST_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$HOST_OS" in
  linux*)   DETECTED_OS="linux" ;;
  darwin*)  DETECTED_OS="darwin" ;;
  msys*|mingw*|cygwin*) DETECTED_OS="windows" ;;
  *)        DETECTED_OS="linux" ;;
esac

GOOS="${TARGET_OS:-$DETECTED_OS}"
GOARCH="${TARGET_ARCH:-$(go env GOARCH)}"

# Binary output path (in current root directory)
OUTPUT="sbeads"
if [ "$GOOS" = "windows" ]; then
  OUTPUT="sbeads.exe"
fi

# SQLite note: mattn/go-sqlite3 requires CGO. Keep CGO enabled.
export CGO_ENABLED=1

if [ "$MODE" = "dev" ]; then
  COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "local")
  VERSION="dev-${COMMIT}"
  LDFLAGS="-X main.Version=${VERSION}"
  echo "🔨 Building sbeads [DEV] (OS: ${GOOS}/${GOARCH}, Version: ${VERSION})..."
  GOOS="${GOOS}" GOARCH="${GOARCH}" go build -ldflags "${LDFLAGS}" -o "${OUTPUT}" ./cmd/bd
else
  VERSION=$(git describe --tags --exact-match 2>/dev/null || git describe --tags --always 2>/dev/null || echo "v1.0.0")
  LDFLAGS="-s -w -X main.Version=${VERSION}"
  echo "🚀 Building sbeads [PROD] (OS: ${GOOS}/${GOARCH}, Version: ${VERSION})..."
  GOOS="${GOOS}" GOARCH="${GOARCH}" go build -trimpath -ldflags "${LDFLAGS}" -o "${OUTPUT}" ./cmd/bd
fi

echo "✅ Build complete: ./${OUTPUT}"

# Run version check if compiled for host OS
if [ "$GOOS" = "$DETECTED_OS" ]; then
  echo -n "   Verified: "
  ./"${OUTPUT}" version
fi
