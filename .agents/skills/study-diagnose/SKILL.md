---
name: study-diagnose
description: 刻意练习诊断：综合掌握度、复习日志（Again率/lapses）与会话记录产出弱点报告，给出定向练习建议并生成练习卡。用户问"哪里薄弱/接下来学什么"时使用。
---

# 诊断复盘（W6，刻意练习闭环）

教练的核心工作是**找弱点 + 开处方**。你的诊断必须每条有数据支撑。

## 步骤

1. **收集证据**（有 topic 参数则聚焦，没有则全量）：
   - `data/progress/mastery.json`：各主题 level 与 evidence 历史
   - `data/progress/review-log.jsonl`：统计各 topic 的 Again(1) 率、lapses 高的卡片
   - `data/sessions/` 近期会话：misconceptions、quiz 正确率、费曼 gap
   - `data/cards/`：各 topic 卡片量（笔记多但卡少 = 内化不足）
2. **产出弱点报告**（正文输出给用户）：
   - 最薄弱的 2-3 个概念，每个附证据（如"card-x 最近 3 次评分 1,1,2"、"费曼时讲不清 Y"、"quiz 3/7"）
   - 区分两类：**遗忘**（曾经会，调度问题→继续复习即可）与**未懂**（从未真懂→需要重学/导师会话）
3. **开处方**（刻意练习要件：小目标 + 即时反馈）：
   - 每个弱点给一个 30-60 分钟的可执行练习（如"和导师过一遍 X 的三个为什么"）
   - 学习者同意后，为弱项生成定向练习卡（fsrs 块用全零模板）
4. 写 `data/sessions/YYYYMMDD-diagnose-<topic 或 all>.md`（type: diagnose），正文放完整报告。
5. 依据证据调整 mastery（kind: diagnose），必须附 evidence。
