#!/usr/bin/env bash
# 一键调试：同时启动 Go 后端（http://127.0.0.1:8788）和 Vite 前端（http://localhost:5173，/api 代理已配置）。
# Ctrl+C 会同时退出两个进程。
# 用法：在仓库根目录执行 scripts/dev.sh
set -euo pipefail
cd "$(dirname "$0")/.."

need() { command -v "$1" >/dev/null 2>&1 || { echo "缺少命令：$1" >&2; exit 1; }; }
need go
need pnpm

if [[ ! -d web/node_modules ]]; then
  echo "→ 首次运行，安装前端依赖…"
  (cd web && pnpm install)
fi

BIN_DIR="$(mktemp -d)"
BIN="$BIN_DIR/deep-study-dev"
BACK_PID=""
cleanup() {
  [[ -n "$BACK_PID" ]] && kill "$BACK_PID" 2>/dev/null || true
  rm -rf "$BIN_DIR"
}
trap cleanup EXIT INT TERM

echo "→ 构建并启动 Go 后端 http://127.0.0.1:8788"
go build -o "$BIN" ./server/cmd/server
"$BIN" &
BACK_PID=$!

echo "→ 启动 Vite 前端 http://localhost:5173"
cd web
pnpm dev
