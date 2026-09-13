---
id: fsrs-rating-semantics
title: 四档评分的语义与目标记忆率
topic: spaced-repetition
tags: [fsrs, rating, retention]
source: library/1-fsrs-入门材料.md
links: [fsrs-dsr-model]
created: 2026-09-13
---
四档评分对应四条调度路径：

- **Again**：完全想不起来 → 进入重学流程，稳定性大幅下降；
- **Hard**：吃力地想起 → 间隔增长放缓；
- **Good**：正常想起 → 大多数复习应有的评分；
- **Easy**：不假思索 → 间隔加速增长。

评分是 FSRS 唯一的学习信号，所以语义纪律很重要：**拿不准评 Good，慎用 Easy**。评分膨胀（什么都点 Easy）会让调度器高估记忆稳定性，之后集中崩盘成一片 Again。

目标记忆率（Request Retention）是全局旋钮：0.90 是复习量与遗忘风险的平衡点；备考期可临时调到 0.95（复习量明显增加）。

## Gaps
