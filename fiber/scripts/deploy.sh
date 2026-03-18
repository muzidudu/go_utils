#!/usr/bin/env bash
# 部署脚本示例：构建后复制到目标目录
# 用法: ./deploy.sh [target_dir]  默认 target_dir=./deploy
set -e

cd "$(dirname "$0")/.."
TARGET="${1:-./deploy}"
BINARY="fiber-app"

./scripts/build.sh "$BINARY"
mkdir -p "$TARGET"
cp -f "dist/${BINARY}" "$TARGET/"
cp -rf config "$TARGET/" 2>/dev/null || true
cp -rf views "$TARGET/" 2>/dev/null || true
echo "Deployed to ${TARGET}/"
