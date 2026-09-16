---
topic: spaced-repetition
goal: 吃透 FSRS 调度原理，能给别人讲明白为什么这样安排复习
horizon: 2 周（2026-09-15 至 2026-09-28，与 math-teaching-theory 并行）
created: 2026-09-15
updated: 2026-09-15
status: active
---
# 间隔重复与 FSRS — 并行精读路线图

## 现状快照（规划依据）

| 项目 | 数据 | 来源 |
| --- | --- | --- |
| mastery | level 0 | `progress/mastery.json`（唯一 evidence 是 2026-09-13 的 import） |
| 笔记 / 卡片 | 4 篇 / 9 张 | `data/notes/`、`data/cards/` |
| 复习覆盖率 | 9 / 9 = **100%** | `progress/review-log.jsonl`，评分以 Good/Easy 为主，仅 fsrs-dsr-model-c3 一次 rating 1（次日改判 4） |
| 笔记 Gaps | 1 处 | `fsrs-dsr-model`：幂函数衰减的具体形式说不清 |

**关键判断**：基线健康、体量小（4 篇笔记），但 level 0 = 零测验证据。目标是"能讲明白"，所以主线是费曼输出而非刷卡。

**排程方式**：与 math-teaching-theory（P0 主线）并行交错——本主题用每天 20-30 分钟的切片时间，不挤占数学的整块新内容时间。

## 里程碑

- [ ] **M1 模型理解**：4 篇笔记无 Gaps（重点补上 `fsrs-dsr-model` 的幂函数衰减形式）
- [ ] **M2 能讲清 D/S/R 交互与评分语义**：`/study:feynman spaced-repetition` 讲解 gap ≤ 1
- [ ] **M3 自测过关**：`/study:quiz spaced-repetition 5` 正确率 ≥ 80%，mastery ≥ 3

## 周计划

### 第 1 周（2026-09-15 起）—— 精读补 Gap

- **每日保底**：清空复习队列（本主题到期量小，并入每日队列一起清）
- 按笔记 order 顺序精读：testing-effect → fsrs-dsr-model → fsrs-rating-semantics → fsrs-vs-sm2
- 跑 `/study:tutor spaced-repetition`，重点攻克幂函数衰减形式 → 兑现 M1

### 第 2 周（2026-09-22 起）—— 输出检验

- **每日保底**：清空复习队列
- 跑 `/study:feynman spaced-repetition`：脱稿讲清"为什么 FSRS 这样安排复习" → 兑现 M2
- 跑 `/study:quiz spaced-repetition 5` → 兑现 M3；若 < 80%，针对错题回炉笔记再来一轮

## 缓冲与风险

- 本主题切片时间若被数学复习积压挤占，优先级让位于"每日清空队列"，M 节点顺延到当周缓冲时间。
- 风险：幂函数衰减形式涉及参数化细节，若导师会话一次讲不透，允许拆成两次，M1 顺延不超过 3 天。
