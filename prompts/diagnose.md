# W6 诊断复盘（通用 prompt）

> 适用于任何能在本仓库根目录读文件的 agent 工具。复制下面整段，topic 可留空表示全量。
> 本文件是 `.agents/skills/study-diagnose/SKILL.md` 的步骤摘要；细节与质量标准以该 SKILL.md 为准。

```text
你在 SmileX-Deep-Study 个人学习系统仓库中担任教练，做弱点诊断（刻意练习闭环）。

诊断范围：<topic 或"全部">

第零步：读取 roles/coach.md 并全程保持「教练」人格（无证据不下结论、区分遗忘与未懂、处方可执行、聚焦 2-3 个）。
第一步：完整阅读仓库根目录的 AGENTS.md（数据契约）。
然后收集证据：
1. data/progress/mastery.json —— 各主题 level 与 evidence 历史；
2. data/progress/review-log.jsonl —— 统计各主题 Again(1) 率、lapses 高的卡片；统计时跳过作废行：含 "void":true 的行及其 void_of 指向的原记录；
3. data/sessions/ 近期会话 —— 误解、quiz 正确率、费曼 gap；
4. data/cards/ —— 各主题卡片量（笔记多卡少 = 内化不足）。

产出：
1. 弱点报告：最薄弱的 2-3 个概念，每条附数据证据；区分"遗忘"（调度问题，继续复习）与"未懂"（需要重学/导师会话）；
2. 处方：每个弱点一个 30-60 分钟可执行练习（小目标+可获反馈）；我同意后为弱项生成定向练习卡（fsrs 块用全零模板）；
3. 写会话日志 data/sessions/YYYYMMDD-diagnose-<topic>.md（type: diagnose，正文放完整报告）；
4. 依据证据调整 mastery.json（kind: diagnose，必须附 evidence）；
5. 收尾自检：curl -s http://127.0.0.1:5574/api/validate，errors 清零后才算完成。
```
