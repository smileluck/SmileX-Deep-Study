---
name: study-diagnose
description: 刻意练习诊断：综合掌握度、复习日志（Again率/lapses）与会话记录产出弱点报告，给出定向练习建议并生成练习卡。用户问"哪里薄弱/接下来学什么"时使用。
---

# 诊断复盘（W6，刻意练习闭环）

**角色**：先读 `roles/coach.md` 并全程保持「教练」人格（无证据不下结论、区分遗忘与未懂、处方可执行、聚焦 2-3 个）。同时遵守仓库根目录 `AGENTS.md` 的数据契约。

## 步骤

1. **收集证据**（有 topic 参数则聚焦，没有则全量）：
   - `data/progress/mastery.json`：各主题 level 与 evidence 历史
   - `data/progress/review-log.jsonl`：统计各 topic 的 Again(1) 率、lapses 高的卡片；统计时跳过作废行——含 `"void":true` 的行及其 `void_of` 指向的原记录（AGENTS.md 红线 2）
   - `data/sessions/` 近期会话：misconceptions、quiz 正确率、费曼 gap
   - `data/cards/`：各 topic 卡片量（笔记多但卡少 = 内化不足）
2. **产出弱点报告**（正文输出给用户，每条附数据证据——角色铁律 1）：
   - 最薄弱的 2-3 个概念
   - 每条区分：**遗忘**（调度问题→继续复习）与**未懂**（需要重学/导师会话）
3. **开处方**（刻意练习要件：小目标 + 即时反馈）：每个弱点一个 30-60 分钟可执行练习；学习者同意后为弱项生成定向练习卡（fsrs 块用全零模板）。
4. 写 `data/sessions/YYYYMMDD-diagnose-<topic 或 all>.md`（type: diagnose），正文放完整报告。
5. 依据证据调整 mastery（kind: diagnose），必须附 evidence。
6. **收尾自检（必须）**：`curl -s http://127.0.0.1:5574/api/validate`，errors 清零后才算完成。
