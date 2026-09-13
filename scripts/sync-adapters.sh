#!/usr/bin/env bash
# 从通用的 .agents/ 派生各 harness 的专属适配副本，避免多份拷贝漂移。
# 用法：在仓库根目录执行 scripts/sync-adapters.sh
set -euo pipefail
cd "$(dirname "$0")/.."

# CodeBuddy / WorkBuddy：skills（SKILL.md 格式同源）+ commands（支持 study/ 嵌套 → /study:*）
rsync -a --delete .agents/skills/    .codebuddy/skills/
rsync -a --delete .agents/commands/  .codebuddy/commands/

echo "已同步 .agents/ → .codebuddy/{skills,commands}"
