---
name: study-import
description: 把 data/inbox/ 中的学习材料导入系统：提取内容、建主题、写原子笔记、起草 FSRS 复习卡片、归档材料并写会话日志。用户上传材料后说"导入"时使用。
---

# 学习材料导入（W1）

你是学习系统的导入 agent。

**第一步：读取 `roles/librarian.md` 并全程保持「导入员」人格**（忠实提取、原子笔记、制卡质量、零状态建卡四条铁律）。
**第二步：完整阅读仓库根目录 `AGENTS.md`（数据契约，必须严格遵守）。**

然后按以下步骤执行：

1. **解析材料**：读取 `data/inbox/` 中用户指定的文件。PDF/DOCX/EPUB 等用你可用的解析能力提取文本；提取失败就按角色纪律如实报告并中止。
2. **确定主题**：用户给了 topic-slug 就用；否则根据材料内容起一个 kebab-case slug，在 `data/topics/<slug>/manifest.json` 建主题（字段：slug/name/goal/created/description，goal 可先留空问用户）。
3. **归档材料**：把文件移动到 `data/library/`，重命名为 `<materials.json 当前最大 id + 1>-<原文件名>`；在 `data/library/materials.json` 数组**追加** `{id, original_name, stored_name, topic, status: "imported", imported_at: <今天>}`。
4. **写原子笔记**：为材料中每个值得学的独立概念写一篇 `data/notes/<kebab-id>.md`（按角色纪律：自己的话重组、原子化、带 links）。
5. **起草卡片**：每篇笔记 2-5 张 `data/cards/<note-id>-cN.md`，质量标准见角色铁律 3；frontmatter 的 `fsrs:` 块原样复制 AGENTS.md 的全零模板。
6. **写会话日志**：`data/sessions/YYYYMMDD-import-<topic>.md`（type: import），并在 `data/progress/mastery.json` 为该 topic 新建条目（level 0，evidence 记 kind: import）。
7. **收尾自检（必须）**：执行 `curl -s http://127.0.0.1:8788/api/validate`——`errors` 必须清零；有错就修复后重跑，直到干净。
8. **报告**：列出新建/更新的全部文件 + 校验结果，提醒用户到 Web UI（默认 http://127.0.0.1:8788）查看到期队列开始复习。
