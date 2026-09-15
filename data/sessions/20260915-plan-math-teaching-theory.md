---
id: 20260915-plan-math-teaching-theory
type: plan
topic: math-teaching-theory
date: 2026-09-15
tool: workbuddy
summary: 为 math-teaching-theory 排定 2 周冲刺计划（5 个里程碑 + 双周计划），并首次产出全局计划 master.md
misconceptions: []
outcomes:
  - 第 1 周先把 76 张卡片首轮全部评一遍，建立基线（当前覆盖率仅 5.3%）
  - 第 1 周末补一份含"数据分析"核心素养的课标材料并导入，清掉 core-literacy-answer-structure 的 Gap
  - 第 2 周重点转向案例分析：四类评析各手写一份评析稿，不能只做口述或选择题
  - spaced-repetition 降为维持性复习，不排新内容
cards_created: []
notes_updated: []
---
# 规划报告

**模式**：主题模式（`/study:plan math-teaching-theory`），同时产出全局计划。

**已确认输入**（学习者答复）：期限 **2 周内**；每周可投入 **>10 小时**。

## 一、读到的现状

- `mastery.json`：math-teaching-theory level **0**，唯一 evidence 是 2026-09-14 的 import（24 笔记 / 76 卡）。
- `review-log.jsonl`：85 张卡里只有 17 条评分记录。按主题拆——
  - math-teaching-theory：76 张卡中仅 **4 张**被评过（case-analysis-objectives c1/c2/c3、case-analysis-process-c1），覆盖率 **5.3%**；
  - spaced-repetition：9 张卡 **全部**评过，评分以 4（Easy/Good）为主，仅 fsrs-dsr-model-c3 出现过一次 rating 1（次日改判 4）。
- `/api/stats`：`due_now = 73`。
- 笔记 Gaps：只有 1 处——`core-literacy-answer-structure` 缺课标六大核心素养中的**数据分析**。
- `data/sessions/`：只有两条 import 记录，**没有 tutor / feynman / quiz / diagnose**。

**核心判断**：level 0 的真实含义是"**零测验证据**"。76 张卡只评过 4 张，无法区分"真的会"和"看着眼熟"。因此第 1 周的首要任务不是学新内容，而是把基线测出来——这也是把 M1 放在最前面的理由。

## 二、backwards design（从目标倒推）

终点是"能按简答题标准结构完整作答三类题"。倒推得到 5 个里程碑，每个绑定可检验标准：

| 里程碑 | 可检验标准 |
| --- | --- |
| M1 建立基线 | review-log 覆盖 76/76 张卡 |
| M2 课标模块成型 | 11 篇课标笔记无 Gaps；quiz 课标题 ≥ 80% |
| M3 教学知识与设计能套用 | 陌生课题 15 分钟内写出目标+重难点+五环节；quiz ≥ 80% |
| M4 案例分析专项过关 | 四类评析脱稿逐条列出并套用新案例；quiz ≥ 85% |
| M5 限时混合模拟 | 10 题 + 1 完整案例分析，正确率 ≥ 85%，mastery ≥ 4 |

## 三、排程理由

- **间隔与交错**：第 1 周新学的卡片在第 2 周按递增间隔回访；quiz 题目混合课标 / 教学知识 / 案例分析三类，不按模块孤立出题。
- **保底任务**：每天清空复习队列，明确写在每周计划第一条；队列 > 40 张时当天只清队列。
- **缓冲**：每周留约 20% 不排任务。
- **风险登记**：积压（73 张集中第 1 周）、材料缺口（数据分析）、主观题失分（必须手写评析稿）三条已写入计划文件。

## 四、产出

- 新建 `data/topics/math-teaching-theory/plan.md`（status: active）
- 新建 `data/plans/master.md`（首次产出全局计划）
- 新建本会话日志
