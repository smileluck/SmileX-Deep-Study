---
id: fsrs-vs-sm2
title: FSRS 与 SM-2 的本质区别
topic: spaced-repetition
order: 4
tags: [fsrs, sm2, comparison]
source: library/1-fsrs-入门材料.md
links: [fsrs-dsr-model]
created: 2026-09-13
---
SM-2（Anki 传统算法）是启发式规则：ease 因子乘间隔倍率，同一套规则对待所有卡片，没有对"记忆本身"建模。FSRS 则显式刻画记忆衰减曲线（D/S/R 三变量），并且能拿你的完整复习历史去优化参数。

实测收益：相同目标记忆率下，FSRS 比 SM-2 少约 20-25% 的复习量。代价是需要积累复习数据（冷启动阶段参数未拟合），以及算法实现更复杂——但对使用者来说这些都被调度器封装了。

一句话：SM-2 回答"下次隔多久"，FSRS 回答"你的记忆现在什么状态，所以下次隔多久"。

## Gaps
