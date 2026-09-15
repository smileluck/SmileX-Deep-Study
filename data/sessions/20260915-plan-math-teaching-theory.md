---
id: 20260915-plan-math-teaching-theory
type: plan
topic: math-teaching-theory
date: 2026-09-15
tool: workbuddy
summary: 为 math-teaching-theory 排定 2 周冲刺计划（5 个里程碑 + 双周计划），首次产出全局计划 master.md；同日复核修正模块篇数并解除 validate 的 plan 类型报错
misconceptions: []
outcomes:
  - 第 1 周先把 76 张卡片首轮全部评一遍，建立基线（当前覆盖率仅 5.3%）
  - 第 1 周末补一份含"数据分析"核心素养的课标材料并导入，清掉 core-literacy-answer-structure 的 Gap
  - 第 2 周重点转向案例分析：四类评析各手写一份评析稿，不能只做口述或选择题
  - spaced-repetition 降为维持性复习，不排新内容
  - 复核修正模块篇数为实际的 课标 10 / 教学知识 8 / 教学设计 3 / 案例分析 3；validate 的 plan 类型报错经查是服务端二进制陈旧，数据侧零改动
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
| M2 课标模块成型 | 10 篇课标笔记无 Gaps；quiz 课标题 ≥ 80% |
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

## 五、复核与修正（同日二次执行）

重跑本工作流时逐项复核了计划里的引用数据，发现并修正两处问题：

**1. 模块篇数与实际笔记不符（已修正）**

原计划按 课标 11 / 教学知识 7 / 教学设计 3 / 案例分析 6 分周排任务，与 `data/notes/` 实际聚合对不上。逐篇点名的实际分布是：

| 模块 | 篇数 | 笔记 id |
| --- | --- | --- |
| 课标 | 10 | `hs-math-curriculum-nature`、`four-bases-curriculum-goal`、`curriculum-structure-basis`、`teaching-evaluation-principles`、`core-literacy-answer-structure`、`core-literacy-math-abstraction`、`core-literacy-logical-reasoning`、`core-literacy-mathematical-modeling`、`core-literacy-intuitive-imagination`、`core-literacy-mathematical-operation` |
| 教学知识 | 8 | `common-teaching-methods`、`concept-definition-methods`、`cultivating-math-thinking`、`math-thinking-methods`、`situation-problem-diversity`、`student-learning-modes`、`textbook-material-selection`、`theorem-teaching-stages` |
| 教学设计 | 3 | `teaching-design-objectives`、`teaching-design-set-operations`、`teaching-process-design` |
| 案例分析 | 3 | `case-analysis-objectives`、`case-analysis-process`、`error-attribution-and-improvement` |

合计 24 篇，与 import 记录一致。已把 plan.md 的 M2 标准、第 1 周「课标 10 篇 / 教学知识 8 篇」、第 2 周「案例分析 3 篇精读」改为实际篇数，并在现状快照里补了「模块分布」一行，便于后续核对。

**2. `/api/validate` 报 `type 非法: plan`——是运行中的服务端二进制陈旧，不是数据问题（已解除）**

- 现象：`curl /api/validate` 返回 `{"ok":false,"errors":[{"file":"sessions/20260915-plan-math-teaching-theory.md","issues":["type 非法: plan"]}]}`。
- 排查：`server/internal/api/validate.go` 第 33 行的 `sessionTypes` 白名单**已包含** `plan`（d573f37 提交加入），与 AGENTS.md 的会话类型约定一致；当时监听的进程用的是 `/private/var/folders/.../tmp.XXXX/deep-study`，即 `make dev` 在 `plan` 特性合并**之前**构建的临时二进制。
- 验证：用当前源码另建一份二进制到备用端口（5599）跑同一份 `data/`，`/api/validate` 返回 `{"ok":true,"errors":[]}`——证明数据合规，问题只在陈旧二进制。
- 处置：重跑 `scripts/sync-adapters.sh` + `go build -o deep-study ./server/cmd/server` 刷新仓库根二进制；服务端随后由新一轮 `make dev` 起在当前源码上，`/api/validate` 已 `ok:true`。
- 结论：**数据侧零改动**（本次只改 plan.md 的篇数与本节文字）；`type: plan` 是合法类型，不应为迁就旧二进制而降级成其他类型。
