# CLAUDE.md

本仓库是 SmileX-Deep-Study 个人学习系统。**请阅读并严格遵循仓库根目录的 `AGENTS.md`**——它是与 Web UI 共享的权威契约（目录结构、文件格式、读写红线、六条学习工作流）。

要点速览：

- 学习数据全部是纯文件（`data/` 下的 markdown + frontmatter + JSON），可直接读写；
- `cards/*.md` 的 `fsrs:` 块、`review-log.jsonl`、`recall-log.jsonl` 只有 Go 服务端可写，你只读；
- 六条工作流（导入/导师/费曼/复习/自测/诊断）的完整步骤见 `AGENTS.md`；
- 通用 prompt 模板在 `prompts/`，与 `.zcode/` 下的技能内容一致。
