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
- `math-teaching-theory` — 数学学科教学论（高级中学）（2026-09-14 首次导入；2026-09-16 追加材料 3 → 32 笔记 / 100 卡）
  - 覆盖：课标理念 10 篇 / 教学知识 8 篇 / 教学设计 4 篇 / 案例分析 9 篇 / 错误归因 1 篇 / 《集合的基本运算》范例
  - 材料 3《高中数学教资-教学设计案例分析模板》补上了材料 2 的三处缺口：提问四类型、教师三角色的独立优缺点模板、教学目标四维与行为主体规范
  - 已知缺口：缺课标六大核心素养中的"数据分析"，待补材料

## 运行环境（2026-09-15 踩过）

- 起服务：`make dev` → Go 后端 `127.0.0.1:5574` + Vite 前端 `:5573`。Go 服务用 `make dev` 的临时二进制，**源码改了必须重启 `make dev` 才生效**。
- **validate 报 `type 非法: xxx` 先怀疑二进制陈旧**：`sessionTypes` 白名单在 `server/internal/api/validate.go`。2026-09-15 遇到 `type 非法: plan`，源码里 `plan` 早已合法，根因是 `make dev` 起的是特性合并前的旧构建。判定法：用当前源码另建二进制跑备用端口 + 同一份 `data/` 再 curl `/api/validate`，若 `ok:true` 即证明数据无问题。
- **2026-09-16 再次撞上陈旧构建**：5574 的 `checked.materials` 恒为 1（实际 3），同法对照新二进制报 3。**`checked` 计数不符也属于陈旧构建的信号，不是数据错误**——只看 `errors` 是否为空。
- go 不在默认 PATH：用 `/opt/homebrew/bin/go`。
- **Vite 只监听 IPv6 `[::1]:5573`**：用 IPv4 `curl 127.0.0.1:5573` 会 connection refused，而走 `HTTP_PROXY` 时表现为 `502 Bad Gateway`——这不是前端挂了，浏览器访问 `localhost:5573` 正常。要绕开代理诊断本地端口，用 Python `socket` 直连或 `curl --noproxy '*'`。
- 接口路径注意：主题计划是 `/api/plans/<slug>`（不是 `/api/topics/<slug>/plan`）。

## 工作习惯

- 大批量产出（>20 文件）用一次性 Python 脚本落盘，再统一跑校验；不要逐文件手写。
- PDF 提取用 `pdfplumber`（`/Users/smilex/.workbuddy-ai/binaries/python/versions/3.13.12/bin/python3`），环境里没有 pdftotext / mutool / pymupdf。
- 写计划/报告时**引用数据必须逐条点名核对**（篇数、覆盖率、due 数），2026-09-15 就是靠逐篇点名发现模块篇数错了 4 处。
- **追加导入既有主题时，补写既有笔记必须在正文里标出来源材料**（如 `**补充（材料 3）：xxx**` 或末尾「来源标注」段）。2026-09-16 复查发现 4 篇补写笔记里 2 篇没标，材料 2/3 的内容分不清，只能回头重新提取材料 2 全文核对。
- **判断材料该并入既有主题还是新建**：同一场考试、同一考点群 → 并入（拆开会把同一考点的 mastery 和复习队列切碎）；考点群不同 → 新建。并入的代价是主题变胖，但系统只有 merge 没有 split，拆分成本高。
- 核对「某条目到底出自哪份材料」的可靠方法：用 pdfplumber 把对照材料提全文存临时文件，再检索关键词定出处——比凭印象判断可靠。
