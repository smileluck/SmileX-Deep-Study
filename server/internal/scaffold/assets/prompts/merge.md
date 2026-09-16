# W8 合并主题（通用 prompt）

> 适用于任何能在本仓库根目录读文件的 agent 工具（Kimi CLI / WorkBuddy / Trae 等）。把下面整段复制给工具即可。
> 本文件是 `.agents/skills/study-merge/SKILL.md` 的步骤摘要；细节与质量标准以该 SKILL.md 为准。

```text
你在 SmileX-Deep-Study 个人学习系统仓库中担任学习导师 agent。

任务：合并学习主题 —— 幸存主题 target：<填目标 topic-slug>，待并入的源主题 sources：<填源 topic-slug，可多个>。

第零步：读取 roles/librarian.md 并全程保持「导入员」人格（忠实提取、原子笔记、制卡质量、零状态建卡）。
第一步：完整阅读仓库根目录的 AGENTS.md（数据契约，必须严格遵守）。
然后按以下步骤执行：
1. 与我确认 target 与 sources；可顺带更新 target manifest 的 name/goal/description；
2. 迁移归属：把 sources 的全部 data/notes/*.md、data/cards/*.md frontmatter 的 topic: 改为 target slug；id 与文件名不变；绝不触碰 fsrs: 块（Go 服务端独占）；
3. 重排笔记 order：target 原有笔记保持 1..n，source 笔记按先修关系续排（默认接在 max(order) 之后，可按依赖关系穿插调整）；
4. data/library/materials.json 中 source 条目的 topic 改为 target slug；
5. mastery.json：source 条目并入 target——level 取两者最高，evidence 合并按日期排序，追加一条 {"date":<今天>,"kind":"merge","detail":"合并自 <source-slug>","delta":0}，updated 改今天；删除 source 键（target 无条目则以 source 为基础改建）；
6. 计划：source 的 plan.md 未完成里程碑判断取舍后并入 target 的 plan.md（同步改 updated）；data/plans/master.md 索引移除 source 链接（同步改 updated）；
7. 历史会话不改动；写合并会话 data/sessions/YYYYMMDD-merge-<target>.md（type: merge，正文记录迁移清单与计划取舍）；
8. 删除 data/topics/<source>/ 目录；收尾自检：curl -s http://127.0.0.1:5574/api/validate，errors 清零才算完成；报告全部改动文件清单 + 校验结果。
```
