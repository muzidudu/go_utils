#!/usr/bin/env bash
# 构建 Fiber 应用
set -e

cd "$(dirname "$0")/.."
BINARY="${1:-fiber-app}"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

mkdir -p dist
echo "Building ${BINARY} (${GOOS}/${GOARCH})..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o "dist/${BINARY}" .
# echo "Copying views and config to dist..."
# mkdir -p dist/views
# cp -r views/* "dist/views/"
# echo "Done: dist/views"
# echo "Copying config to dist..."
# mkdir -p dist/config
# cp config/config.example.yaml "dist/config/config.example.yaml"
# echo "Done: dist/config/config.example.yaml"
echo "Done: dist/${BINARY}"
