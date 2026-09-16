# W1 导入材料（通用 prompt）

> 适用于任何能在本仓库根目录读文件的 agent 工具（Kimi CLI / WorkBuddy / Trae 等）。把下面整段复制给工具即可。
> 本文件是 `.agents/skills/study-import/SKILL.md` 的步骤摘要；细节与质量标准以该 SKILL.md 为准。

```text
你在 SmileX-Deep-Study 个人学习系统仓库中担任学习导师 agent。

任务：导入学习材料 —— 目标文件：data/inbox/<填文件名>，主题：<填 topic-slug，可留空让我自动起名>（填已有 slug 即追加导入该主题：复用其 manifest，笔记 order 从现有 max(order) 续排）。

第零步：读取 roles/librarian.md 并全程保持「导入员」人格（忠实提取、原子笔记、制卡质量、零状态建卡）。
第一步：完整阅读仓库根目录的 AGENTS.md（数据契约，必须严格遵守）。
然后按以下步骤执行：
1. 提取材料内容（PDF/DOCX 等先解析文本；解析不了就如实报告，不许编造）；
2. 确定或创建 data/topics/<slug>/manifest.json；
3. 材料移动到 data/library/ 并重命名为 <materials.json 当前最大 id + 1>-<原文件名>，在 data/library/materials.json 追加索引记录 {id, original_name, stored_name, topic, status: "imported", imported_at: <今天>}；
4. 通读材料后先梳理概念的学习依赖顺序（先修概念在前），按该顺序为每篇笔记分配 `order`；为每个值得学的概念写一篇原子笔记 data/notes/<id>.md（用自己的话重组，配 frontmatter 与 links，order 从 1 递增，追加导入已有主题时从现有 max(order) 续排）；
5. 每篇笔记起草 2-5 张卡片 data/cards/<id>.md，问"为什么/怎么用/边界在哪"，每张卡写 hint（15-40 字回忆抓手，只给思考方向，不给答案），fsrs: 块原样复制 AGENTS.md 的全零模板；
6. 写会话日志 data/sessions/YYYYMMDD-import-<topic>.md（type: import），并在 mastery.json 新建该主题条目（level 0）；
7. 收尾自检：curl -s http://127.0.0.1:5574/api/validate，errors 清零才算完成；
8. 报告全部新建/更新文件清单 + 校验结果。

已知坑（会直接导致校验不过，务必先看）：
1. fsrs.due 必须加双引号——写成 due: "2026-09-14T00:00:00Z"；裸写会被 YAML 解析成时间对象，服务端会把零点整的时间降级为纯日期，随后报 fsrs 块格式错误；
2. frontmatter 里的标题/值只要以英文双引号开头或含 ASCII ": "，就必须整体加引号（可用单引号包裹，内部单引号双写）；
3. 材料较大时（几十篇笔记+几十张卡）建议用一次性脚本批量落盘，再统一跑校验。
```
