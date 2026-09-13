---
name: study-import
description: 把 data/inbox/ 中的学习材料导入系统：提取内容、建主题、写原子笔记、起草 FSRS 复习卡片、归档材料并写会话日志。用户上传材料后说"导入"时使用。
---

# 学习材料导入（W1）

你是学习系统的导入 agent。严格遵循仓库根目录 `AGENTS.md` 的文件格式契约。

## 步骤

1. **解析材料**：读取 `data/inbox/` 中用户指定的文件。PDF/DOCX/EPUB 等用你可用的解析能力提取文本；提取失败就如实报告，不要编造内容。
2. **确定主题**：用户给了 topic-slug 就用；否则根据材料内容起一个 kebab-case slug，在 `data/topics/<slug>/manifest.json` 建主题（字段：slug/name/goal/created/description，goal 可先留空问用户）。
3. **归档材料**：把文件移动到 `data/library/`，重命名为 `<materials.json 当前最大 id + 1>-<原文件名>`；在 `data/library/materials.json` 数组**追加** `{id, original_name, stored_name, topic, status: "imported", imported_at: <今天>}`。
4. **写原子笔记**：为材料中每个值得学的独立概念写一篇 `data/notes/<kebab-id>.md`——用自己的话重组，不是摘抄；用 `links` 关联相关笔记。
5. **起草卡片**：每篇笔记 2-5 张 `data/cards/<note-id>-cN.md`。好卡标准：问"为什么 / 怎么用 / 边界在哪 / 和 X 的区别"，避免逐字定义背诵。**frontmatter 的 `fsrs:` 块必须原样复制 AGENTS.md 中的全零模板**（due=今天，last_review: null）。
6. **写会话日志**：`data/sessions/YYYYMMDD-import-<topic>.md`（type: import），并在 `data/progress/mastery.json` 为该 topic 新建条目（level 0，evidence 记 kind: import）。
7. **报告**：列出新建/更新的全部文件，提醒用户到 Web UI（默认 http://127.0.0.1:8788）查看到期队列开始复习。
