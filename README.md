# SmileX-Deep-Study · 个人学习舱

> **文件即数据库 · AGENTS.md 即契约 · harness 即导师 · Go 单二进制即驾驶舱。系统零 LLM 接入。**

把学习闭环拆成两半：**理解类工作**（导入提取、苏格拉底导师、费曼追问、出题批改、弱点诊断）交给你已有的 agent harness（ZCode / Trae / Kimi CLI / WorkBuddy）；**调度类工作**（FSRS 间隔重复、统计、管理界面）由本系统的 Go 单二进制完成。所有数据都是纯 Markdown + JSON 文件，任何工具都能直接读写。

理论依据与设计差距分析见 [docs/00-theory-gap-analysis.md](docs/00-theory-gap-analysis.md)，架构与数据契约见 [docs/01-architecture.md](docs/01-architecture.md)。

## 快速开始

```bash
# 1. 构建前端（首次或前端有改动时）
cd web && pnpm install && pnpm build && cd ..

# 2. 构建单二进制（会把 web/dist 内嵌进去）
go build -o deep-study ./server/cmd/server

# 3. 在仓库根目录运行（工作流页的通用 prompt 功能依赖根目录下的 prompts/）
./deep-study
# → http://127.0.0.1:8788
```

开发模式：`./deep-study`（8788）+ `cd web && pnpm dev`（5173，已配 /api 代理）。

自定义：`./deep-study -addr 0.0.0.0:8788 -data /path/to/data`

## 六条工作流

| 工作流 | 执行者 | 用法 |
|---|---|---|
| W1 导入 | UI 上传 + harness | 网页「资料库」拖入文件 → 在 harness 里执行 `/study:import <文件名> [topic]` |
| W2 精读导师 | harness | `/study:tutor <topic>` —— 苏格拉底对话，先提问后讲解，阶梯提示 |
| W3 费曼内化 | harness | `/study:feynman <topic|note>` —— 你讲它追问，gap 写回笔记 |
| W4 间隔复习 | Web UI（零 LLM） | 「复习」页：先回忆后揭示，四档评分，FSRS 调度 |
| W5 检索自测 | harness | `/study:quiz <topic> [n]` —— 全新题目 + 批改 + 掌握度回写 |
| W6 诊断复盘 | harness | `/study:diagnose [topic]` —— 弱点报告 + 定向练习（刻意练习闭环） |

## 四家 harness 怎么接

| 工具 | 接入方式 |
|---|---|
| **ZCode** | 开箱即用：`.zcode/skills/` 与 `.zcode/commands/` 已就位，直接 `/study:*` |
| **Kimi CLI** | 原生读 `AGENTS.md`；在仓库根目录打开后发 `prompts/` 里的通用模板即可 |
| **WorkBuddy** | 无 `CODEBUDDY.md` 时自动加载 `AGENTS.md`；`.codebuddy/rules/` 已就位 |
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

红线规则（详见 `AGENTS.md`）：`cards/*.md` 的 `fsrs:` 块与两个日志只有 Go 服务端可写；日志只追加；掌握度变更必须附证据。

## 技术栈

- **后端**：Go + gin + [go-fsrs](https://github.com/open-spaced-repetition/go-fsrs)（FSRS 官方实现，目标记忆率 0.90）+ yaml.Node 定向改写 frontmatter（保留 agent 写入的额外字段）
- **前端**：Vite + React 19 + TypeScript + Tailwind 4 + daisyUI 5，`go:embed` 内嵌进二进制
- **部署**：一个静态二进制 + `data/` 目录，无运行时依赖

## 目录结构

```
├── AGENTS.md                  # 权威契约（agent 必读）
├── CLAUDE.md / .trae/ / .codebuddy/   # 各家薄适配
├── .zcode/                    # ZCode skills + 斜杠命令
├── prompts/                   # 通用 prompt 模板
├── server/                    # Go 后端（cmd/ + internal/{api,store,fsrsx}）
├── web/                       # 前端源码
├── docs/                      # 理论差距分析 + 架构文档
└── gui-test-screenshots/      # GUI 测试证据截图
```
