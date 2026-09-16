# SmileX-Deep-Study · 个人学习舱

[English](README.md) | **简体中文**

> **文件即数据库 · AGENTS.md 即契约 · harness 即导师 · Go 单二进制即驾驶舱。系统零 LLM 接入。**

把学习闭环拆成两半：**理解类工作**（导入提取、苏格拉底导师、费曼追问、出题批改、弱点诊断）交给你已有的 agent harness（ZCode / Trae / Kimi CLI / WorkBuddy）；**调度类工作**（FSRS 间隔重复、统计、管理界面）由本系统的 Go 单二进制完成。所有数据都是纯 Markdown + JSON 文件，任何工具都能直接读写。

理论依据与设计差距分析见 [docs/00-theory-gap-analysis.md](docs/00-theory-gap-analysis.md)，架构与数据契约见 [docs/01-architecture.md](docs/01-architecture.md)。

## 快速开始

```bash
# 1. 构建前端（首次或前端有改动时）
cd web && pnpm install && pnpm build && cd ..

# 2. 构建单二进制（会把 web/dist 内嵌进去）
go build -o deep-study ./server/cmd/server

# 3. 在任意目录运行——启动时自动创建 data/ 并铺出 harness 工作区文件
#    （AGENTS.md/.agents/.codebuddy/roles/prompts，只建缺失不覆盖，-scaffold=false 可关）。
#    在非空目录运行时工作区会收进新建的 deepstudy/ 子目录，二进制留在原地，
#    二次运行仍是同一条 ./deep-study；想指定位置就用 -data 显式传数据目录。
./deep-study
# → http://127.0.0.1:5574
```

开发模式：`./deep-study`（5574）+ `cd web && pnpm dev`（5573，已配 /api 代理）。

自定义：`./deep-study -addr 0.0.0.0:5574 -data /path/to/data`

## 八条工作流

| 工作流 | 执行者 | 用法 |
|---|---|---|
| W1 导入 | UI 上传 + harness | 网页「资料库」拖入文件 → 在 harness 里执行 `/study:import <文件名> [topic]` |
| W2 精读导师 | harness | `/study:tutor <topic>` —— 苏格拉底对话，先提问后讲解，阶梯提示 |
| W3 费曼内化 | harness | `/study:feynman <topic|note>` —— 你讲它追问，gap 写回笔记 |
| W4 间隔复习 | Web UI（零 LLM） | 「复习」页：先回忆后揭示，四档评分，FSRS 调度 |
| W5 检索自测 | harness | `/study:quiz <topic> [n]` —— 全新题目 + 批改 + 掌握度回写 |
| W6 诊断复盘 | harness | `/study:diagnose [topic]` —— 弱点报告 + 定向练习（刻意练习闭环） |
| W7 学习规划 | harness | `/study:plan [topic]` —— 以终为始倒推里程碑，排出全局/主题路线图 |
| W8 主题合并 | harness | `/study:merge <目标> <源>...` —— 迁移笔记/卡片归属，合并掌握度与计划 |

## 四家 harness 怎么接

| 工具 | 接入方式 |
|---|---|
| **任何遵循通用格式的 agent** | `.agents/skills/`（SKILL.md 标准）+ `.agents/commands/` + `AGENTS.md`，开箱即用 |
| **ZCode** | 原生发现 `.agents/`（`.zcode/` 的通用回退路径），直接 `/study:*` |
| **Kimi CLI** | 原生读 `AGENTS.md`；在仓库根目录打开后发 `prompts/` 里的通用模板即可 |
| **WorkBuddy** | 完整适配：`CODEBUDDY.md` 默认全量加载 + `.codebuddy/{rules,skills,commands}/`（`/study:*` 命令与技能就位） |
| **Trae** | 设置 → Rules → 勾选“包含 AGENTS.md”；`.trae/rules/deep-study.mdc` 已就位 |

“工作流”页面有为每个工作流准备的**一键复制命令与通用 prompt**（任何工具可用）。

## 数据都在哪

```
data/
├── inbox/     上传的原始材料        ├── notes/     原子笔记（含 ## Gaps）
├── library/   归档材料 + 索引       ├── cards/     卡片（内嵌 FSRS 调度状态）
├── topics/    主题清单              ├── sessions/  导师/费曼/自测/诊断会话日志
└── progress/  mastery.json + 复习/回忆日志（追加式）
```

红线规则（详见 `AGENTS.md`）：`cards/*.md` 的 `fsrs:` 块与两个日志只有 Go 服务端可写；日志只追加；掌握度变更必须附证据；**凡写文件的工作流收尾必须跑 `curl -s http://127.0.0.1:5574/api/validate` 把 errors 清零**（硬校验：schema、fsrs 块、日期格式、证据链）。agent 的行为边界由 `roles/` 下的角色定义约束（导师/测验官/初学者/教练/导入员），技能只定义流程——软约束 + 硬校验双层保证产出质量。

## 技术栈

- **后端**：Go + gin + [go-fsrs](https://github.com/open-spaced-repetition/go-fsrs)（FSRS 官方实现，目标记忆率 0.90）+ yaml.Node 定向改写 frontmatter（保留 agent 写入的额外字段）
- **前端**：Vite + React 19 + TypeScript + Tailwind 4 + daisyUI 5，`go:embed` 内嵌进二进制
- **部署**：一个静态二进制 + `data/` 目录，无运行时依赖；启动时自动铺出 harness 工作区文件（AGENTS.md/.agents/.codebuddy/roles/prompts，只建缺失），未传 `-data` 且在非空目录运行时工作区自动收进 `deepstudy/` 子目录（二进制留在原地，二次运行命令不变），任意目录都能直接用 WorkBuddy 等工具打开

## 目录结构

```
├── AGENTS.md                  # 权威契约（agent 必读）
├── CLAUDE.md / CODEBUDDY.md / .trae/ / .codebuddy/   # 各家适配（.codebuddy 副本由 scripts/sync-adapters.sh 从 .agents/ 同步）
├── .agents/                   # 通用 agent 资产：skills（SKILL.md）+ 斜杠命令
├── roles/                     # 角色定义（身份+纪律）：导师/测验官/初学者/教练/导入员
├── prompts/                   # 通用 prompt 模板
├── server/                    # Go 后端（cmd/ + internal/{api,store,fsrsx}）
├── web/                       # 前端源码
├── docs/                      # 理论差距分析 + 架构文档
└── gui-test-screenshots/      # GUI 测试证据截图
```
