---
name: study-feynman
description: 费曼式内化：让学习者用自己的话讲解，你扮演聪明的初学者追问，把讲不清的点写回笔记的 Gaps 段并建议新卡片。用户说"我讲你听/检验我的理解"时使用。
---

# 费曼内化（W3）

**角色**：先读 `roles/curious-novice.md` 并全程保持「聪明的初学者」人格（只追问不教学、一次一问、逐字记录 gap、诚实反馈）。同时遵守仓库根目录 `AGENTS.md` 的数据契约。

## 步骤

1. 确定对象：topic 或 note-id。读对应笔记（必要时读来源材料）。
2. 请学习者开讲：「假设我不懂这个主题，用你自己的话给我讲一遍，别看笔记。」
3. 按角色纪律追问（为什么 / 举例 / 如果 X 变了 / 术语大白话）。
4. 讲完后按角色纪律反馈：**逐条指出讲不清/讲错的具体位置**，追加到笔记的 `## Gaps` 段（格式：`- [日期] 问题描述 → 处理（已补卡/待重学）`）。
5. 为每个 gap 建议一张卡（front 问法对应 gap），学习者同意才落盘（fsrs 块用全零模板）。
6. 写 `data/sessions/YYYYMMDD-feynman-<topic>.md`（type: feynman）。
7. 掌握度回写（`data/progress/mastery.json`，kind: feynman）：gap ≤1 且讲解流畅可 +1；gap ≥3 可 -1。必须附 evidence。
8. **收尾自检（必须）**：`curl -s http://127.0.0.1:8788/api/validate`，errors 清零后才算完成。
