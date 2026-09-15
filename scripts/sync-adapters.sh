#!/usr/bin/env bash
# 从通用的 .agents/ 派生各 harness 的专属适配副本，避免多份拷贝漂移；
# 并同步 server/internal/scaffold/assets/——打包进二进制的 harness 工作区脚手架
# （启动时铺到数据目录上一级，只建缺失不覆盖）。assets 内用无点目录名
# （agents/ codebuddy/），运行时写盘再还原成 .agents/ .codebuddy/。
# 用法：在仓库根目录执行 scripts/sync-adapters.sh（make build/cross 会自动调用）
set -euo pipefail
cd "$(dirname "$0")/.."

# CodeBuddy / WorkBuddy：skills（SKILL.md 格式同源）+ commands（支持 study/ 嵌套 → /study:*）
rsync -a --delete .agents/skills/    .codebuddy/skills/
rsync -a --delete .agents/commands/  .codebuddy/commands/

# 脚手架资产（go:embed 不能引用包外文件与点开头路径，故需在包内放副本）
A=server/internal/scaffold/assets
cp AGENTS.md CLAUDE.md "$A/"
rsync -a --delete .agents/    "$A/agents/"
rsync -a --delete .codebuddy/ "$A/codebuddy/"
rsync -a --delete roles/      "$A/roles/"
rsync -a --delete prompts/    "$A/prompts/"

echo "已同步 .agents/ → .codebuddy/{skills,commands}，并更新 scaffold/assets/"
