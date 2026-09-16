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
4. **写原子笔记**：通读材料后先梳理概念的学习依赖顺序（先修概念在前），按该顺序为每篇笔记分配 `order`（frontmatter 字段，主题内从 1 递增；追加导入已有主题时从该主题现有 max(order) 续排）。为材料中每个值得学的独立概念写一篇 `data/notes/<kebab-id>.md`（按角色纪律：自己的话重组、原子化、带 links）。
5. **起草卡片**：每篇笔记 2-5 张 `data/cards/<note-id>-cN.md`，质量标准见角色铁律 3；每张卡必须写 `hint`（15-40 字回忆抓手，只给思考方向，不给答案）；frontmatter 的 `fsrs:` 块原样复制 AGENTS.md 的全零模板。
6. **写会话日志**：`data/sessions/YYYYMMDD-import-<topic>.md`（type: import），并在 `data/progress/mastery.json` 为该 topic 新建条目（level 0，evidence 记 kind: import）。
7. **收尾自检（必须）**：执行 `curl -s http://127.0.0.1:5574/api/validate`——`errors` 必须清零；有错就修复后重跑，直到干净。
8. **报告**：列出新建/更新的全部文件 + 校验结果，提醒用户到 Web UI（默认 http://127.0.0.1:5574）查看到期队列开始复习。

## 已知坑（会直接导致校验不过，务必先看）

1. **`fsrs.due` 必须加双引号**。写成 `due: 2026-09-14T00:00:00Z`（裸标量）时 YAML 会解析成时间对象，而服务端 `normalizeTimes` 对"零点整"的时间会降级成纯日期 `2026-09-14`，随后 `fsrsx.FromMap` 报 `fsrs 块格式错误: parsing time "2026-09-14" as ...`。正确写法：`due: "2026-09-14T00:00:00Z"`。
2. **frontmatter 里的标题/值只要以 `"` 开头或含 ASCII `: `，就必须整体加引号**。例：`title: "四基"课程目标：…` 会让 YAML 在第 1 行报 `did not find expected key`。写成 `title: '"四基"课程目标：…'`（单引号包裹，内部单引号需双写）。
3. **批量产出建议用一次性脚本**。材料较大时（几十篇笔记 + 几十张卡）逐文件手写既慢又容易漏字段；用 Python 把 `NOTES` 数据结构一次性落盘，再统一跑校验。落盘后仍要执行第 7 步。
