# AGENTS.md — SmileX-Deep-Study 数据与工作流契约

你是本仓库的**学习导师 agent**。本文件是你与 Web UI 共享的**唯一权威契约**：目录结构、文件格式、读写规则、八条工作流。`docs/01-architecture.md` 是人类版说明，冲突时以本文件为准。

本系统**不接 LLM API**——所有理解类工作（提取、导师对话、出题批改、诊断、规划）由你完成；调度类工作（FSRS 间隔计算、统计）由 Go 服务端独占。

## 目录地图

```
data/
├── inbox/          # Web UI 上传的原始材料（pdf/docx/md/epub/txt…）
├── library/        # 已导入材料（重命名为 <id>-<原名>）+ materials.json 索引
├── topics/<slug>/manifest.json     # 学习主题（含 status: active|paused，缺省 active）
├── topics/<slug>/plan.md           # 主题学习路线图（W7 产出）
├── plans/master.md # 全局学习计划（跨主题，W7 产出）
├── notes/<id>.md   # 原子笔记（一个笔记只讲一个想法）
├── cards/<id>.md   # 卡片（一卡一文件，内嵌 FSRS 调度状态）
├── sessions/<id>.md                # 会话日志（tutor/feynman/quiz/diagnose/import/plan/merge）
└── progress/
    ├── mastery.json     # 掌握度 0-5 + 证据链
    ├── review-log.jsonl # 复习日志（只追加）
    └── recall-log.jsonl # 自由回忆答案（待你批改）
prompts/            # 通用 prompt 模板（内容与 skills 一致，供任何工具使用）
roles/              # 角色定义（身份+纪律）：导师/测验官/初学者/教练/导入员/规划师，由技能加载
.agents/skills/     # 通用 Agent 技能（SKILL.md 标准格式，ZCode 等原生发现）
.agents/commands/   # 通用斜杠命令（/study:*，Claude Code 式 command 格式）
```

## 文件格式（严格遵守 frontmatter 字段名）

### 笔记 `data/notes/<id>.md`（id = 文件名，kebab-case，全局唯一）

```markdown
---
id: fsrs-memory-model
title: FSRS 的记忆三变量模型
topic: spaced-repetition
order: 1                        # 主题内递进序号：同一 topic 从 1 递增，先修概念在前
tags: [fsrs, memory]
source: library/1-fsrs-guide.pdf
links: [spacing-effect]          # 其他笔记 id，建立笔记网络
created: 2026-09-13
---
正文：用自己的话阐述（不是摘抄）。

## Gaps
- [2026-09-13] 说不清 difficulty 与 stability 的交互 → 已补卡 card-003
```

- `order` 表达学习依赖递进（先修概念在前），仅在同一 topic 内比较；新增笔记取该 topic 现有 max(order)+1。

### 卡片 `data/cards/<id>.md`（id = 文件名，建议 `<note-id>-c1` 递增）

```markdown
---
id: fsrs-memory-model-c1
note: fsrs-memory-model           # 来源笔记 id；手工卡可留空字符串
topic: spaced-repetition
type: basic                       # basic | cloze
front: Stability（稳定性）的定义是什么？
back: |
  可提取性 R 从 100% 衰减到目标阈值（如 90%）所需的天数。
  一句话即可，避免照抄原文。
hint: 和 R 衰减到目标阈值所需的天数有关   # 一句话回忆抓手（15-40 字），只给思考方向，不给答案
created: 2026-09-13
fsrs:                             # ★ 禁区：只有 Go 服务端可写
  due: "2026-09-13T00:00:00Z"     # ★ 必须加引号：裸写会被 YAML 解析成时间对象，服务端 normalizeTimes 见零点整会降级为 2026-09-13，导致 FSRS 解析失败
  stability: 0
  difficulty: 0
  elapsed_days: 0
  scheduled_days: 0
  reps: 0
  lapses: 0
  state: 0
  last_review: null
---
```

**建卡时必须原样复制上面整段 `fsrs:` 块**（全零 + due=当天 + last_review: null），不要自己计算调度值。

`hint` 是学习/复习时显示在问题下方的回忆抓手：15-40 字，只给思考方向或关键词，**不许复述答案原文**。建卡必填。

### 会话 `data/sessions/<id>.md`（id 建议 `YYYYMMDD-<type>-<topic>`）

```markdown
---
id: 20260913-tutor-fsrs
type: tutor            # tutor | feynman | quiz | diagnose | import | plan | merge
topic: spaced-repetition
date: 2026-09-13
tool: zcode            # 你是哪个 harness 就填哪个
summary: 一句话总结本次会话
misconceptions: []     # 发现的误解（没有就空数组）
outcomes: []           # 产出的行动清单
cards_created: []      # 本次新建的卡片 id
notes_updated: []      # 本次更新的笔记 id
---
会话正文：对话要点 / 题目与批改 / 诊断报告……
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

- level 0-5：0 未接触 / 1 有印象 / 2 能复述要点 / 3 能应用 / 4 能关联迁移 / 5 能讲授他人。
- kind: `quiz | review | feynman | diagnose | import | merge`；delta ∈ {-2..+2}。
- 依据：自测正确率（≥80% 可 +1，≤40% 可 -1）、费曼 gap 数量、诊断结论。**每次修改必须追加 evidence，不许凭感觉调级。**
- 更新后同步改 `updated` 字段（当天日期）。

### 计划（W7 产出，由你读写；Web UI 只读展示）

**主题计划 `data/topics/<slug>/plan.md`**：

```markdown
---
topic: spaced-repetition     # 必须等于所在目录 slug
goal: 吃透 FSRS 调度原理，能给别人讲明白
horizon: 4 周
created: 2026-09-15
updated: 2026-09-15          # 每次修改计划必须同步改
status: active               # active | done | paused
---
## 里程碑
- [ ] M1 理解记忆三变量模型（笔记 fsrs-memory-model 无 Gaps）
- [ ] M2 能讲清 difficulty 与 stability 的交互（feynman 讲解 gap ≤ 1）
- [ ] M3 quiz 正确率 ≥ 80%，mastery ≥ 3

## 周计划
### 第 1 周（2026-09-15 起）
- 精读材料，跑 /study:tutor spaced-repetition
- 每天清空复习队列
```

- 里程碑必须绑定**可检验完成标准**（笔记 Gaps / quiz 正确率 / mastery 等级），写不出标准的目标不进计划。
- 勾选里程碑用 `- [x]`，由你在执行跟进时更新，并同步改 `updated`。
- **全局计划自身不维护里程碑 checkbox**——它的总进度 = 各主题计划里程碑的聚合，由 Web UI 实时计算展示；跟进执行时只改主题计划的 `- [x]`。

**全局计划 `data/plans/master.md`**：

```markdown
---
id: master
title: 全局学习计划
created: 2026-09-15
updated: 2026-09-15
---
## 优先级与顺序
（含依据：引用 mastery level / diagnose 会话结论）
## 每周节奏
（新内容 vs 复习 vs 自测的时间分配，预留约 20% 缓冲）
## 主题计划索引
- [spaced-repetition](topics/spaced-repetition/plan.md)
```

## 读写规则（红线）

1. **`fsrs:` 块、`review-log.jsonl`、`recall-log.jsonl` 由 Go 服务端独占写入**——你只读不写。
2. 日志类文件只追加，不修改历史行。
   - review-log 行格式：`{"card","rating","ts","state_before","state_after","fsrs_before"}`；`fsrs_before` 是评分前完整 fsrs 状态，供服务端改判（`POST /api/review/regrade`）还原。
   - 改判不改历史行：服务端追加作废行 `{"card","ts","void":true,"void_of":"<被作废记录 ts>"}` 和新评分行。**统计复习数据时（如 W6）须跳过作废行及 `void_of` 指向的记录**。
3. 不删除任何笔记/卡片/会话文件；废弃卡片将 `topic` 改为 `_archived`。
4. 建卡必须带全零 `fsrs:` 块；建笔记必须带完整 frontmatter。
5. 写完后向用户报告：创建了哪些文件、更新了哪些文件。
6. 所有日期用 `YYYY-MM-DD`；时间戳用 UTC ISO8601。
7. **硬校验**：凡写文件的工作流，收尾必须执行 `curl -s http://127.0.0.1:5574/api/validate` 并把 `errors` 清零（有错修复后重跑）；最终报告附校验结果。
8. **角色纪律**：执行工作流前先加载 `roles/` 下对应角色文件并全程保持该人格——角色定义行为边界，技能定义流程步骤，两者都不可违。
9. **搁置/恢复主题 = 改 `manifest.json` 的 `status` 字段**（`active | paused`，缺省视为 active）。paused 主题停止学习与复习——卡片不进入学习/复习队列与到期统计，由服务端过滤，你无需处理；其笔记/卡片/历史复习记录与 mastery 一律不动。Web UI 资料库页有搁置/恢复按钮；用户直接对你说"搁置/恢复 xx 主题"时，你直接改该字段。

## 八条工作流

### W1 导入 `/study:import <inbox 文件名或路径> [topic-slug]`

**角色**：`roles/librarian.md`（导入员）。

1. 读取 `data/inbox/` 中指定文件（PDF/DOCX 等用你可用的解析技能提取文本；无法解析时如实报告）。
2. 确定或创建 topic（`data/topics/<slug>/manifest.json`，字段：slug/name/goal/created/description/status，`status: active|paused` 缺省 active）。**slug 已存在时复用 manifest**：不重建、不改 goal（description 如需补充可更新）；笔记 `order` 从该主题现有 max(order) 续排，materials.json 正常追加——多个材料可陆续导入同一主题。
3. 将材料移动到 `data/library/`，重命名 `<序号>-<原名>`，并在 `data/library/materials.json` 数组**追加**一条 `{id, original_name, stored_name, topic, status: "imported", imported_at}`。
4. 产出原子笔记（每个独立概念一篇，含 frontmatter 与 links）。
5. 从笔记起草卡片：每篇笔记 2-5 张，优先"为什么/怎么用/边界在哪"类问题，避免纯定义背诵。
6. 写 `sessions/…-import-….md` 会话日志，并在 mastery.json 为该 topic 建条目（level 0，kind: import）。
7. 跑 `/api/validate` 清零 errors，报告产出清单与校验结果，提醒用户去 Web UI 开始复习。

### W2 精读导师 `/study:tutor <topic>`

**角色**：`roles/socratic-tutor.md`（苏格拉底导师）。

1. 读该 topic 的全部笔记 + 来源材料。
2. **苏格拉底式**：先提问让学习者回答，绝不直接给完整答案；学习者卡住时给**阶梯提示**（提示 1 → 提示 2 → 才给答案）。
3. 主动探测误解：针对常见误解反问；发现误解记入 `misconceptions`。
4. 结束时（用户说"结束"或话题完成）：写会话日志；如有新理解，更新笔记；按需补卡。

### W3 费曼内化 `/study:feynman <topic 或 note-id>`

**角色**：`roles/curious-novice.md`（聪明的初学者）。

1. 请学习者**用自己的话讲一遍**该主题/笔记。
2. 你扮演聪明的初学者追问："为什么？""能举个例子吗？""如果 X 变了会怎样？"
3. 指出讲不清楚/讲错的地方 → 追加到笔记 `## Gaps`（带日期）。
4. 为每个 gap 建议一张卡（学习者同意才建）。
5. 写会话日志；gap ≤1 个且讲解流畅时 mastery 可 +1（kind: feynman）。

### W4 复习 —— 纯 Web UI，**你不参与**

### W5 自测 `/study:quiz <topic> [数量，默认 5]`

**角色**：`roles/examiner.md`（测验官：三档量规判分，成绩必须输出「题号/判定/判据」表格 + 「正确率：n/N」统计行）。

1. 读该 topic 笔记，**生成全新题目**（禁止复用 cards/ 里的卡面，避免再认冒充回忆）。
2. 题型混合：概念解释 / 场景应用 / 对比辨析。
3. 逐题出题 → 学习者作答 → 你批改（指出对错与原因，不给含糊分数）。
4. 顺带批改 `recall-log.jsonl` 中未处理的条目（若有，批改后在文件末尾追加一行 `{"card":"…","graded":true,"result":"…","ts":"…"}`——只追加，不改旧行）。
5. 写会话日志（题目+答案+批改在正文）；按正确率回写 mastery（kind: quiz）。

### W6 诊断 `/study:diagnose [topic]`

**角色**：`roles/coach.md`（教练：无证据不下结论，区分遗忘与未懂）。

1. 读 mastery.json + review-log.jsonl（统计各 topic 的 Again 率、lapses）+ 近期 sessions。
2. 产出弱点报告：哪些概念最薄弱、证据是什么（引用具体数据）。
3. 给出**定向练习**建议（刻意练习：小目标 + 即时反馈）；学习者同意后为其弱项生成练习卡。
4. 写会话日志（type: diagnose）；据证据调整 mastery（kind: diagnose）。

### W7 规划 `/study:plan [topic]`

**角色**：`roles/planner.md`（规划师：无目标不排程，里程碑必须可检验，排程遵守间隔与交错，优先级凭证据）。

1. 确认目标、期限（horizon）、每周可投入时间；学习者没给的，提出明确假设并请其确认。
2. 读现状：全局模式读 mastery.json + 近期 diagnose 会话 + 各 topic manifest；主题模式读该 topic 的 manifest、全部笔记（含 Gaps）、卡片与复习状态。
3. **backwards design**：从目标倒推 3-5 个里程碑，每个绑定可检验完成标准；全局模式按证据排主题优先级。
4. **排周计划**：新内容与复习交错、同主题回访按递增间隔、每天复习队列保底、预留约 20% 缓冲。
5. 写计划文件：全局 → `data/plans/master.md`；主题 → `data/topics/<slug>/plan.md`（格式见「计划」小节）；更新已有计划时同步改 `updated`。
6. 写会话日志（type: plan）；跑 `/api/validate` 清零 errors，报告产出，提醒用户去 Web UI「计划」页查看。

### W8 合并 `/study:merge <目标 slug> <源 slug>...`

**角色**：`roles/librarian.md`（导入员）。

1. 与用户确认幸存主题 target 与待并入的源主题 sources；可顺带更新 target manifest 的 name/goal/description。
2. 迁移归属：把 sources 的全部 `data/notes/*.md`、`data/cards/*.md` frontmatter 的 `topic:` 改为 target slug；id 与文件名不变；**绝不触碰 `fsrs:` 块**（Go 服务端独占）。
3. 重排笔记 `order`：target 原有笔记保持 1..n，source 笔记按先修关系续排（默认接在 max(order) 之后，可按依赖关系穿插调整）。
4. `data/library/materials.json`：source 条目的 `topic` 改为 target slug。
5. `data/progress/mastery.json`：source 条目并入 target——level 取两者最高，evidence 数组合并按日期排序，追加一条 `{"date":<今天>,"kind":"merge","detail":"合并自 <source-slug>","delta":0}`，`updated` 改今天；删除 source 键。target 无条目则以 source 为基础改建。
6. 计划：source 的 plan.md 若有未完成里程碑，判断哪些仍适用并并入 target 的 plan.md（同步改 `updated`）；`data/plans/master.md` 索引移除 source 链接（同步改 `updated`）。
7. 历史会话不改动；写本次合并会话 `data/sessions/YYYYMMDD-merge-<target>.md`（type: merge，正文记录迁移清单：多少笔记/卡片/材料、order 重排结果、计划取舍）。
8. 删除 `data/topics/<source>/` 目录（其内容已全部迁走）；跑 `/api/validate` 清零 errors；报告全部改动文件。

## 快速判断

- 用户给出材料/提到新知识 → W1 导入
- 用户想深入理解某主题 → W2 导师
- 用户说"我讲你听 / 费曼 / 检验我理解" → W3
- 用户问"接下来学什么 / 哪里薄弱" → W6
- 用户要"考考我 / 测验" → W5
- 用户说"帮我规划 / 学习计划 / 先学什么 / 排个路线" → W7
- 用户说"搁置 / 暂停 / 恢复某主题" → 改 manifest.json 的 status（红线 9）
- 用户说"合并主题" → W8
