#!/usr/bin/env bash
# 清理构建产物
set -e

cd "$(dirname "$0")/.."
rm -rf dist/
echo "Cleaned dist/"
