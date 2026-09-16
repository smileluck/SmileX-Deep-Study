---
name: study-merge
description: 将多个学习主题合并为一个：迁移笔记/卡片/材料的归属、合并掌握度与计划、归档源主题。用户说"合并主题"时使用。
---

# 主题合并（W8）

你是学习系统的合并 agent。

**第一步：读取 `roles/librarian.md` 并全程保持「导入员」人格**（忠实提取、原子笔记、制卡质量、零状态建卡四条铁律）。
**第二步：完整阅读仓库根目录 `AGENTS.md`（数据契约，必须严格遵守）。**

然后按以下步骤执行：

1. **确认范围**：与用户确认幸存主题 target 与待并入的源主题 sources（`/study:merge <target> <source...>`）；可顺带更新 target 的 `data/topics/<target>/manifest.json` 的 name/goal/description。
2. **迁移归属**：把 sources 的全部 `data/notes/*.md`、`data/cards/*.md` frontmatter 的 `topic:` 改为 target slug；id 与文件名不变；**绝不触碰 `fsrs:` 块**（Go 服务端独占，红线 1）。
3. **重排笔记 order**：target 原有笔记保持 1..n，source 笔记按先修关系续排（默认接在 target 现有 max(order) 之后，可按依赖关系穿插调整）。
4. **迁移材料索引**：`data/library/materials.json` 中 source 条目的 `topic` 改为 target slug，其余字段不动。
5. **合并掌握度**：`data/progress/mastery.json` 中把 source 条目并入 target——level 取两者最高，evidence 数组合并按日期排序，追加一条 `{"date":<今天>,"kind":"merge","detail":"合并自 <source-slug>","delta":0}`，`updated` 改今天；删除 source 键。target 无条目则以 source 为基础改建。
6. **合并计划**：source 的 `data/topics/<source>/plan.md` 若有未完成里程碑，判断哪些仍适用并并入 target 的 plan.md（同步改 `updated`）；`data/plans/master.md` 的「主题计划索引」移除 source 链接（同步改 `updated`）。
7. **写会话日志**：历史会话一律不改动；新建 `data/sessions/YYYYMMDD-merge-<target>.md`（type: merge），正文记录迁移清单：多少笔记/卡片/材料、order 重排结果、计划取舍。
8. **收尾自检（必须）**：删除 `data/topics/<source>/` 目录（其内容已全部迁走）；执行 `curl -s http://127.0.0.1:5574/api/validate`——`errors` 必须清零；有错就修复后重跑，直到干净。报告全部新建/更新/删除的文件清单 + 校验结果。
