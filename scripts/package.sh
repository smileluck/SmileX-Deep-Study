#!/usr/bin/env bash
# 一键打包：前端构建 → go:embed 内嵌 dist → 产出单文件二进制（无运行时依赖）。
# 用法：
#   scripts/package.sh                 # 本机平台，产出 ./deep-study
#   scripts/package.sh linux/amd64     # 交叉编译，产出 ./deep-study-linux-amd64
set -euo pipefail
cd "$(dirname "$0")/.."

need() { command -v "$1" >/dev/null 2>&1 || { echo "缺少命令：$1" >&2; exit 1; }; }
need go
need pnpm

OUT="deep-study"
EXTRA_LDFLAGS="-s -w"
if [[ $# -ge 1 ]]; then
  TARGET="$1"
  GOOS_VAL="${TARGET%%/*}"
  GOARCH_VAL="${TARGET##*/}"
  if [[ "$GOOS_VAL" == "$TARGET" || -z "$GOARCH_VAL" ]]; then
    echo "目标平台格式应为 <goos>/<goarch>，例如 linux/amd64" >&2
    exit 1
  fi
  export GOOS="$GOOS_VAL" GOARCH="$GOARCH_VAL"
  OUT="deep-study-${GOOS}-${GOARCH}"
fi

echo "→ 安装前端依赖"
(cd web && pnpm install --frozen-lockfile)

echo "→ 构建前端（web/dist）"
(cd web && pnpm build)

echo "→ 编译 Go 二进制（内嵌 web/dist）→ $OUT"
CGO_ENABLED=0 go build -trimpath -ldflags="$EXTRA_LDFLAGS" -o "$OUT" ./server/cmd/server

echo ""
echo "打包完成：./$OUT"
echo "运行方式：./$OUT   # 然后访问 http://127.0.0.1:8788"
