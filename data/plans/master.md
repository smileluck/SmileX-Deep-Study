---
id: master
title: 全局学习计划
created: 2026-09-15
updated: 2026-09-15
---
# 全局学习计划

## 现状（排序依据）

| 主题 | mastery | 笔记 / 卡片 | 复习覆盖率 | 期限 |
| --- | --- | --- | --- | --- |
| math-teaching-theory | 0 | 24 / 76 | 4 / 76 = **5.3%** | 2 周内 |
| spaced-repetition | 0 | 4 / 9 | 9 / 9 = **100%** | 无 |

- `math-teaching-theory`：唯一 evidence 是 2026-09-14 的 import（level 0）。76 张卡里 review-log 只覆盖 4 张（case-analysis-objectives c1-c3、case-analysis-process-c1），**基线未建立**；当前 due_now 73 张。
- `spaced-repetition`：9 张卡全部至少评过一次，评分以 4（Easy/Good）为主，仅 1 次 rating 1（fsrs-dsr-model-c3，次日已改判 4）。基线健康，但因**没有任何 quiz / feynman 证据**，level 仍停在 0。

## 优先级与顺序

- **P0 — math-teaching-theory**：有明确期限（2 周）、占分最重、基线覆盖率仅 5.3%。全部新内容投入此主题，路径为「清 76 张卡建立基线 → 课标 → 教学知识 → 教学设计 → 案例分析 → 混合模拟」。
- **P1 — spaced-repetition**：卡片已 100% 建过基线且评分健康，属于**支撑复习系统本身的元知识**。不排新内容，只保留每日复习队列；若 review-log 中该主题出现 Again 率上升或 lapses 累积，再升级为 P0 重学。
- 依据：以上排序引用 `progress/mastery.json` 的 level 与 evidence 字段，以及 `review-log.jsonl` 的卡片覆盖率——不凭感觉决定先学什么。

## 每周节奏

按每周可投入时间 T（当前 2 周窗口内假定 >10h）分配：

- 复习队列保底 **25%**——每天清空，不可被新内容挤掉
- 新内容 / 精读 / 手写练习 **35%**
- 自测 + 费曼 + 诊断 **20%**
- 缓冲 **20%**——消化积压与补漏

## 主题计划索引

- [math-teaching-theory](topics/math-teaching-theory/plan.md) — 2 周冲刺，P0，active
- spaced-repetition — 暂无独立计划（维持性复习）；需要时跑 `/study:plan spaced-repetition` 生成
