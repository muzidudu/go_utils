#!/usr/bin/env bash
# 本地运行 Fiber 应用（开发模式）
set -e

cd "$(dirname "$0")/.."

# 确保 config 存在
if [[ ! -f config/config.yaml ]]; then
  echo "Warning: config/config.yaml not found, using defaults"
fi

go run .
