# 项目长期记忆 — SmileX-Deep-Study

## 数据契约红线（不可违）

- 只有 Go 服务端可写：`cards/*.md` 的 `fsrs:` 块、`review-log.jsonl`、`recall-log.jsonl`。Agent 只读。
- 日志类文件只追加，不改历史行。不删笔记/卡片/会话；废弃卡片把 `topic` 改成 `_archived`。
- 建卡必须带全零 `fsrs:` 块；建笔记必须有完整 frontmatter（id 与文件名一致、title/topic/created）。
- 任何写文件的工作流收尾必须跑 `curl -s http://127.0.0.1:5574/api/validate` 并把 `errors` 清零。

## 两个必须记住的 YAML 坑（2026-09-14 踩过）

1. `fsrs.due` **必须加双引号**：`due: "2026-09-14T00:00:00Z"`。
   裸写会被 YAML 解析成时间对象，服务端 `normalizeTimes` 对零点整的时间降级为纯日期，导致 `fsrs 块格式错误`。
2. frontmatter 的值只要以 `"` 开头或含 ASCII `: `，就必须整体加引号；含双引号时用单引号包裹（内部单引号双写），如 `title: '"四基"课程目标：…'`。

## 已建立的主题

- `spaced-repetition` — 间隔重复与 FSRS（2026-09-13 导入，4 笔记 / 9 卡）
- `math-teaching-theory` — 数学学科教学论（高级中学）（2026-09-14 导入，24 笔记 / 76 卡）
  - 覆盖：课标理念 / 教学知识 / 教学设计与案例分析 / 《集合的基本运算》范例
  - 已知缺口：缺课标六大核心素养中的"数据分析"，待补材料

## 工作习惯

- 大批量产出（>20 文件）用一次性 Python 脚本落盘，再统一跑校验；不要逐文件手写。
- PDF 提取用 `pdfplumber`（`/Users/smilex/.workbuddy-ai/binaries/python/versions/3.13.12/bin/python3`），环境里没有 pdftotext / mutool / pymupdf。
