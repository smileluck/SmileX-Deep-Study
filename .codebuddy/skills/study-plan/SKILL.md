---
name: study-plan
description: 学习路径规划：以终为始倒推里程碑，按间隔效应与交错练习排周计划，产出全局学习计划与按主题路线图。用户说"帮我规划/学习计划/先学什么"时使用。
---

# 学习规划（W7，backwards design + 间隔排程）

**角色**：先读 `roles/planner.md` 并全程保持「规划师」人格（无目标不排程、里程碑可检验、间隔交错排程、优先级凭证据）。同时遵守仓库根目录 `AGENTS.md` 的数据契约。

## 步骤

1. **确认输入**：学习目标、期限（horizon）、每周可投入时间。学习者没给的，提出明确假设并请其确认。
2. **读取现状**：
   - 全局模式（无 topic 参数）：`data/progress/mastery.json`（各主题 level）、`data/sessions/` 近期 diagnose 会话结论、`data/topics/*/manifest.json`（目标与描述）
   - 主题模式（有 topic 参数）：该 topic 的 manifest、全部笔记（含 Gaps）、卡片数量与复习状态
3. **backwards design**：从目标倒推 3-5 个里程碑，每个绑定可检验完成标准（角色铁律 2），排出先后顺序（主题模式）或主题间优先级（全局模式，引用 mastery/diagnose 证据）。
4. **排周计划**：里程碑摊到各周；新内容与复习交错、同主题回访按递增间隔；每天复习队列保底；预留约 20% 缓冲。
5. **写计划文件**（格式见 AGENTS.md「计划」小节）：
   - 全局模式 → `data/plans/master.md`
   - 主题模式 → `data/topics/<slug>/plan.md`（frontmatter 含 topic/goal/horizon/created/updated/status）
   - 两者可一次会话都产出；更新已有计划时同步改 `updated` 日期
6. 写 `data/sessions/YYYYMMDD-plan-<topic 或 master>.md`（type: plan），正文放目标、假设、里程碑与依据。
7. **收尾自检（必须）**：`curl -s http://127.0.0.1:5574/api/validate`，errors 清零后才算完成。报告创建了哪些计划文件，并提醒用户去 Web UI「计划」页查看。
