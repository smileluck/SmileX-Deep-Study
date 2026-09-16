# SmileX-Deep-Study 架构与数据契约

> 一句话：**文件即数据库，AGENTS.md 即契约，harness 即导师，Go 单二进制 Web UI 即驾驶舱。系统零 LLM 接入。**

## 一、设计原则

1. **纯文件存储**：全部学习数据是 Markdown + YAML frontmatter + JSON/JSONL，无数据库。任何 agent harness（ZCode/Trae/Kimi/WorkBuddy）都能直接读写，Web UI 与 harness 操作同一份文件。
2. **LLM 外置**：理解类工作（导入提取、苏格拉底导师、费曼追问、出题批改、弱点诊断）由用户打开的 harness 执行；系统只提供契约、模板和落盘格式。
3. **调度内置**：FSRS 间隔重复是纯数学，由 Go 服务端独占计算（`go-fsrs`），agent 不得伪造调度字段。
4. **单一部署单元**：`go build` 产出静态二进制，前端产物 `go:embed` 内嵌；部署 = 二进制 + `data/` 目录。二进制启动时自动铺出 harness 契约文件（AGENTS.md/.agents/.codebuddy/roles/prompts，只建缺失不覆盖）；未显式传 `-data` 且当前目录是非空目录时，会先新建 `deepstudy/` 子目录把工作区和数据收进去，不污染当前目录（二进制本身留在原地，二次运行命令不变）。
5. **实时读盘**：服务端每个请求实时扫描 `data/`，不缓存——因为 harness 随时在进程外改文件。

## 二、目录地图

```
SmileX-Deep-Study/
├── AGENTS.md                          # 权威契约（agent 必读）
├── CLAUDE.md / .trae/rules/ / .codebuddy/rules/   # 薄适配，均指向 AGENTS.md
├── .agents/skills/ + .agents/commands/  # 通用 agent 技能与斜杠命令（ZCode 原生兼容）
├── prompts/                           # 通用 prompt（任何工具可复制）
├── data/
│   ├── inbox/                         # UI 上传的原始材料
│   ├── library/                       # 已导入材料 + materials.json 索引
│   ├── topics/<slug>/manifest.json    # 学习主题
│   ├── topics/<slug>/plan.md          # 主题学习路线图（W7 产出，agent 读写）
│   ├── plans/master.md                # 全局学习计划（W7 产出，agent 读写）
│   ├── notes/<id>.md                  # 原子笔记
│   ├── cards/<id>.md                  # 卡片（一卡一文件，内嵌 FSRS 状态）
│   ├── sessions/<id>.md               # 会话日志（tutor/feynman/quiz/diagnose/import/plan）
│   └── progress/
│       ├── mastery.json               # 掌握度（0-5 + 证据链）
│       ├── review-log.jsonl           # 复习日志（追加式）
│       └── recall-log.jsonl           # 自由回忆答案（待 agent 批改）
├── server/  (cmd/ + internal/)        # Go 后端，go.mod 在仓库根
├── web/                               # Vite + React 前端，构建产物 web/dist 被 embed
└── docs/                              # 本文档与差距分析
```

## 三、文件格式契约（权威定义在 AGENTS.md，此处为设计说明）

### 笔记 `data/notes/<id>.md`

```markdown
---
id: fsrs-memory-model
title: FSRS 的记忆三变量模型
topic: spaced-repetition
order: 1                        # 主题内递进序号（先修概念在前），新增笔记取该 topic max(order)+1
tags: [fsrs, memory]
source: library/1-fsrs-guide.pdf
links: [spacing-effect]
created: 2026-09-13
---
正文（原子化：一个笔记只讲一个想法）……

## Gaps
- [2026-09-13] 说不清 difficulty 与 stability 的交互 → 已补卡 card-003
```

### 卡片 `data/cards/<id>.md`

```markdown
---
id: card-001
note: fsrs-memory-model        # 来源笔记，可空（手工卡）
topic: spaced-repetition
type: basic                    # basic | cloze
front: Stability（稳定性）的定义是什么？
back: |
  可提取性 R 从 100% 衰减到目标阈值（如 90%）所需的天数。
created: 2026-09-13
fsrs:                          # ★ 只有 Go 服务端可写
  due: "2026-09-13T00:00:00Z"  # ★ 必须加引号：裸写会被 YAML 当成时间对象，零点整会被服务端降级为纯日期而解析失败
  stability: 0
  difficulty: 0
  elapsed_days: 0
  scheduled_days: 0
  reps: 0
  lapses: 0
  state: 0                     # 0=New 1=Learning 2=Review 3=Relearning
  last_review: null
---
（可选正文，如 cloze 的上下文）
```

### 会话 `data/sessions/<id>.md`

```markdown
---
id: 20260913-tutor-fsrs
type: tutor          # tutor | feynman | quiz | diagnose | import | plan
topic: spaced-repetition
date: 2026-09-13
tool: zcode          # 产生本次会话的 harness
summary: 澄清了 stability 与 difficulty 的关系
misconceptions: ["以为难度高=间隔短"]
outcomes: ["决定补 2 张卡"]
cards_created: [card-004]
notes_updated: [fsrs-memory-model]
---
会话正文 / 对话要点 / 题目与批改结果……
```

### 掌握度 `data/progress/mastery.json`

```json
{
  "spaced-repetition": {
    "level": 2,
    "evidence": [
      {"date": "2026-09-13", "kind": "quiz", "detail": "5/7 正确", "delta": 1}
    ],
    "updated": "2026-09-13"
  }
}
```

level 0-5（0=未接触 … 5=能讲授）。每次变更必须附 evidence（kind: quiz | review | feynman | diagnose | import）。

### 日志（追加式，永不改写）

- `review-log.jsonl`：`{"card":"card-001","rating":3,"ts":"...","state_before":0,"state_after":1,"fsrs_before":{...}}`
  - `fsrs_before`：评分前完整 fsrs 状态，供改判（`POST /api/review/regrade`）还原调度状态。
  - 改判不改历史行：追加一条作废行 `{"card":"card-001","ts":"...","void":true,"void_of":"<被作废记录的 ts>"}` 和一条新评分行；统计端须跳过作废行及 `void_of` 指向的记录。
- `recall-log.jsonl`：`{"card":"card-001","answer":"学习者凭记忆写的答案","ts":"..."}`（复习播放器"自由回忆"模式产生，由 quiz/diagnose 工作流批改后清写 mastery）

### 材料 `data/library/materials.json`

```json
[{"id": 1, "original_name": "fsrs-guide.pdf", "stored_name": "1-fsrs-guide.pdf",
  "topic": "spaced-repetition", "status": "imported", "imported_at": "2026-09-13"}]
```

### 主题 `data/topics/<slug>/manifest.json`

```json
{"slug": "spaced-repetition", "name": "间隔重复", "goal": "吃透 FSRS 并能给人讲明白",
 "created": "2026-09-13", "description": "…"}
```

### 计划（W7 产出，agent 读写，UI 只读）

- 主题计划 `data/topics/<slug>/plan.md`：frontmatter `topic`（= slug）/ `goal` / `horizon` / `created` / `updated` / `status`（active|done|paused）；正文「## 里程碑」用 `- [ ]`/`- [x]` 清单，每个里程碑绑定可检验完成标准；「## 周计划」按间隔效应与交错练习排布。
- 全局计划 `data/plans/master.md`：frontmatter `id: master` / `title` / `created` / `updated`；正文含主题优先级（引用 mastery/diagnose 证据）、每周节奏、主题计划索引。全局计划不维护里程碑——「计划」页的总进度由各主题计划的 `- [x]` 聚合得出。

## 四、七条工作流与职责边界

| # | 工作流 | 执行者 | 输入 → 输出 |
|---|---|---|---|
| W1 导入 | UI 上传 + harness 执行 `/study:import` | UI + harness | inbox 文件 → topic/notes/cards(草案)，材料归档 library |
| W2 精读导师 | harness `/study:tutor` | harness | 材料笔记 → sessions/tutor-*.md（苏格拉底对话+误解记录） |
| W3 费曼内化 | harness `/study:feynman` | harness | 学习者口述 → 笔记 `## Gaps` 更新 + 新卡草案 |
| W4 复习 | Web UI（零 LLM） | Go | 到期队列 → 四档评分 → FSRS 更新 + review-log |
| W5 自测 | harness `/study:quiz` | harness | 笔记 → 新题（不复用卡片）→ 批改 → mastery 回写 |
| W6 诊断 | harness `/study:diagnose` | harness | mastery+日志+会话 → 弱点报告 → 定向练习卡 |
| W7 规划 | harness `/study:plan` | harness | 目标+mastery/诊断证据 → plans/master.md + topics/<slug>/plan.md（里程碑+周计划） |

## 五、Go 后端 API

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/materials/upload | multipart 上传 → data/inbox/ |
| GET  | /api/validate | 数据契约硬校验：扫描 cards/notes/sessions/mastery/materials，返回 errors（schema/fsrs 块/日期/证据链违规）与 warnings（空正文）；agent 工作流收尾必须清零 errors |
| GET  | /api/materials | inbox + library 列表（含 materials.json） |
| GET  | /api/topics | 主题列表（含卡片/笔记计数） |
| GET  | /api/review/queue | 扫描 cards/，返回到期队列（按主题交错排序；支持 `?topic=<slug>` 只复习指定主题） |
| POST | /api/review/grade | {id, rating:1-4} → go-fsrs 重写 frontmatter + 追加日志 |
| POST | /api/review/recall | {id, answer} → 追加 recall-log.jsonl |
| GET  | /api/notes / /api/notes/:id | 笔记列表（含反链）/ 详情 |
| POST | /api/cards | UI 手动建卡（fsrs 初始化为 New） |
| GET  | /api/sessions | 会话列表（解析 frontmatter） |
| GET  | /api/mastery | mastery.json + 各主题卡片健康度 |
| GET  | /api/plans / /api/plans/:slug | 全局学习计划 + 各主题路线图（只读，供「计划」页渲染） |
| GET  | /api/stats | 仪表盘：今日到期/总卡数/连续天数/90天热力图/最近会话 |
| GET  | /* | go:embed 的 SPA 静态文件（history 路由 fallback index.html） |

评分映射：1=Again 2=Hard 3=Good 4=Easy（go-fsrs Rating 枚举）。

## 六、技术栈

- 后端：Go + gin + go-fsrs + gopkg.in/yaml.v3（frontmatter 经 yaml.Node 定向更新，保留 agent 写入的任意额外字段）
- 前端：Vite + React 19 + TypeScript + Tailwind 4 + daisyUI + react-markdown + react-router + lucide-react
- 部署：`make build`（web/dist + 脚手架资产一起 `go:embed`）→ `./deep-study` 运行即得 `http://127.0.0.1:5574`；启动时自动创建 `data/` 并铺出 AGENTS.md/.agents/.codebuddy/roles/prompts（只建缺失，`-scaffold=false` 可关）；未传 `-data` 且在非空目录运行时，自动收进 `deepstudy/` 子目录（二进制留在原地，二次运行命令不变）
