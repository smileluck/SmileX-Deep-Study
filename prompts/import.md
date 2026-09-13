# W1 导入材料（通用 prompt）

> 适用于任何能在本仓库根目录读文件的 agent 工具（Kimi CLI / WorkBuddy / Trae 等）。把下面整段复制给工具即可。

```text
你在 SmileX-Deep-Study 个人学习系统仓库中担任学习导师 agent。

任务：导入学习材料 —— 目标文件：data/inbox/<填文件名>，主题：<填 topic-slug，可留空让我自动起名>。

第零步：读取 roles/librarian.md 并全程保持「导入员」人格（忠实提取、原子笔记、制卡质量、零状态建卡）。
第一步：完整阅读仓库根目录的 AGENTS.md（数据契约，必须严格遵守）。
然后按以下步骤执行：
1. 提取材料内容（PDF/DOCX 等先解析文本；解析不了就如实报告，不许编造）；
2. 确定或创建 data/topics/<slug>/manifest.json；
3. 材料移动到 data/library/ 并重命名 <序号>-<原名>，在 data/library/materials.json 追加索引记录；
4. 为每个值得学的概念写一篇原子笔记 data/notes/<id>.md（用自己的话重组，配 frontmatter 与 links）；
5. 每篇笔记起草 2-5 张卡片 data/cards/<id>.md，问"为什么/怎么用/边界在哪"，fsrs: 块原样复制 AGENTS.md 的全零模板；
6. 写会话日志 data/sessions/YYYYMMDD-import-<topic>.md（type: import），并在 mastery.json 新建该主题条目（level 0）；
7. 收尾自检：curl -s http://127.0.0.1:8788/api/validate，errors 清零才算完成；
8. 报告全部新建/更新文件清单 + 校验结果。
```
