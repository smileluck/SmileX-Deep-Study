---
id: 20260915-plan-master
type: plan
topic: master
date: 2026-09-15
tool: kimi-code
summary: 应学习者要求重排优先级为两主题并行：spaced-repetition 从维持性复习升级为 P0 并行（每日切片），新建其主题计划；master.md 与数学计划同步更新
misconceptions: []
outcomes:
  - 新建 data/topics/spaced-repetition/plan.md（3 个里程碑：M1 清 Gap / M2 费曼 gap ≤ 1 / M3 quiz ≥ 80% 且 mastery ≥ 3）
  - master.md 优先级改为「两主题并行」，每周节奏拆出 SR 10% 切片时间
  - math-teaching-theory/plan.md 补充并行主题说明，里程碑不变
cards_created: []
notes_updated: []
---
# 规划报告

**模式**：全局模式（重排优先级），同时新建 spaced-repetition 主题计划。

**已确认输入**（学习者答复）：期限 2 周、每周 >10 小时（沿用今日早些时候确认）；本轮变更诉求为「两主题并行」——spaced-repetition 也排精读/自测新内容，与数学交错进行。

## 一、现状依据

- `mastery.json`：两主题均 level 0，唯一 evidence 均为 import——零测验证据。
- `review-log.jsonl`：math-teaching-theory 覆盖 4/76（5.3%）；spaced-repetition 覆盖 9/9（100%），评分以 Good/Easy 为主，基线健康。
- spaced-repetition 仅 1 处笔记 Gap：`fsrs-dsr-model` 幂函数衰减形式说不清；体量小（4 篇笔记），适合切片并行。

## 二、重排决策

- math-teaching-theory 保持 P0 主线（2 周期限、占分重、基线未建立），整块新内容时间不变。
- spaced-repetition 从「维持性复习」升级为 P0 并行：每日 20-30 分钟切片，路线为 tutor（清 Gap）→ feynman（输出检验）→ quiz（≥80% 且 mastery ≥ 3）。
- 风险对冲：切片时间被积压挤占时让位于每日清空队列；数学里程碑不变。

## 三、产出文件

- 新建 `data/topics/spaced-repetition/plan.md`
- 更新 `data/plans/master.md`（优先级与每周节奏、索引）
- 更新 `data/topics/math-teaching-theory/plan.md`（缓冲与风险补并行说明）
