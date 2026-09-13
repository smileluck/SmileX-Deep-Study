# 学习成长理论体系差距分析

> 调研日期：2026-09-13。本文回答一个问题：**为什么市面上没有一个适合个人的、覆盖完整学习闭环的系统**，以及 SmileX-Deep-Study 如何补位。

## 一、六个体系对比总表

| 体系 | 核心机制 | 对软件系统的要求 | 现有实现的缺口 |
|---|---|---|---|
| DeepTutor（港大 HKUDS） | 脚手架引导、苏格拉底对话、主动学习、知识追踪、误解诊断、形成性反馈 | LLM 对话引擎、学习者画像、RAG 检索 | 必须自接 LLM API；**无间隔重复调度**（有闪卡无算法）；RAG 基础设施对个人偏重 |
| Anki / FSRS | D-S-R 记忆模型（难度/稳定性/可提取性），按目标记忆率排期 | 每卡持久化状态 + 完整复习日志 + 可替换调度器 | 只管调度不管创作：制卡手工负担重；卡片与源材料、笔记脱节 |
| Zettelkasten / Smart Notes | 原子笔记 + 密集链接 + 用自己的话阐述 | 快速捕获收件箱、纯 Markdown、双向链接、反链 | 无复习调度、无反馈渠道；易陷入"收藏家谬误"（囤积不加工） |
| 刻意练习（Ericsson） | 明确小目标 + 即时反馈 + 能力边缘任务 + 心理表征打磨 | 弱点诊断 → 定向任务生成 → 客观即时反馈 | 需要教练；自学最难自制的正是诊断与反馈这一环 |
| Bloom 2σ / 精熟学习 | 一对一辅导 + 掌握检查 + 矫正循环直至达标 | 分技能诊断、掌握阈值、矫正回路、可变进度 | 一对一辅导贵；掌握检查无配套矫正则沦为走过场 |
| 检索练习 / 必要难度（Bjork、Roediger & Karpicke） | 测试本身强化记忆（一周后 61% vs 重读 40%）；间隔、交错、先生成后看答案 | 主动回忆 UI（绝不先展示答案）、从材料生成题目、交错排队 | 学生天然偏好重读（流畅性错觉），工具必须强制"先回忆再揭示" |

## 二、逐个体检

### 1. DeepTutor——导师对话的天花板，但缺两条腿

- 出品：港大数据智能实验室（HKUDS），Apache 2.0 开源，[GitHub](https://github.com/HKUDS/DeepTutor) / [官网](https://deeptutor.info/) / [论文 arXiv:2604.26962](https://arxiv.org/html/2604.26962v1)。
- 论文实际主张的教学法（非营销话术）：**脚手架**（Wood, Bruner & Ross 1976：给提示不给答案，且有"脚手架密度"自检）、**苏格拉底多轮对话**、**主动学习**（Freeman et al. 2014）、**知识追踪**（Corbett & Anderson 1994 等）、**误解诊断**（Smith, DiSessa & Roschelle 1994）、**形成性反馈**（Shute 2008）。
- 架构：静态知识接地（知识图谱 + 稠密向量双索引 + RRF 融合）+ 动态记忆（Trace Forest 三层记忆代理），约 35 家 LLM 供应商绑定，可全套本地模型。
- **关键缺口**（我们的机会）：
  1. 必须配置 LLM profile（API key 或本地模型），个人有成本与运维负担；
  2. **没有间隔重复**：闪卡存在，但全项目找不到任何调度算法（FSRS/SM-2 均无）；
  3. 重型 RAG（LlamaIndex/FAISS/GraphRAG 可选装）对"一个人用"是杀鸡用牛刀；
  4. 记忆三层是它自家格式，不是用户可直接编辑的纯文件。

### 2. Anki / FSRS——调度最优，创作与语境缺席

- FSRS 用 DSR 三变量建模记忆，按目标记忆率（如 90%）排期，基准测试显示比 SM-2 少约 20-25% 复习量（[awesome-fsrs](https://github.com/open-spaced-repetition/awesome-fsrs)）。
- Anki 模型：Note（数据）→ 模板生成 Card → Deck 分层；四级评分 Again/Hard/Good/Easy；现代 Anki 已内置 FSRS（[官方 FAQ](https://faqs.ankiweb.net/what-spaced-repetition-algorithm)）。
- 缺口：制卡全手工（最大弃坑原因）；卡片脱离源材料与笔记语境，复习变成孤立事实；"ease hell" 等评分语义问题。
- 对本项目的输入：**调度引擎直接复用官方实现**（Go 用 [go-fsrs](https://github.com/open-spaced-repetition/go-fsrs)），把制卡工作交给导入工作流（harness 从材料自动起草）。

### 3. Zettelkasten / Smart Notes——内化正确，闭环缺失

- 三类笔记：闪念（捕获）/ 文献（读过什么）/ 永久（一个自足的想法）；原子性 + 密集链接；阐述（用自己的话重述）本身就是思考步骤（[Ness Labs 摘要](https://nesslabs.com/how-to-take-smart-notes)）。
- 缺口：没有复习调度（学了就忘），没有反馈渠道（写了没人指出错误），工具折腾容易替代真正的写作。
- 对本项目的输入：**原子笔记即文件**（Markdown + frontmatter + links），费曼工作流往笔记写回 `## Gaps`，让卡片从笔记中生长。

### 4. 刻意练习——知道原理，缺一个教练

- 四要件：明确定义的小目标、即时且有信息量的反馈、能力边缘的任务、心理表征的持续打磨（Ericsson 1993）。
- 缺口：反馈与任务设计通常依赖教练；没有外部结构时练习退化为无效重复（平台期）。
- 对本项目的输入：**诊断工作流**（harness 读 mastery + 复习日志 + 会话记录 → 弱点报告 → 定向生成练习卡），把"教练"外置给 harness。

### 5. Bloom 2σ / 精熟学习——标准固定、时间可变

- 一对一辅导 + 精熟学习使平均学生提升两个标准差（Bloom 1984，[原文](https://web.mit.edu/5.95/readings/bloom-two-sigma.pdf)）；机制是"教 → 形成性检查 → 矫正 → 再测直至达标"。
- 缺口：一对一贵（Bloom 的"问题"就是如何规模化）；现代复现效应缩水；掌握检查若无高质量矫正就是走过场。
- 对本项目的输入：**掌握度模型**（主题 0-5 级 + 证据链），自测结果回写，诊断工作流驱动矫正循环。

### 6. 检索练习 / 必要难度——反直觉但证据最硬

- Roediger & Karpicke 2006：一周后保留率重复测试 ~61% vs 重复学习 ~40%（[PubMed](https://pubmed.ncbi.nlm.nih.gov/16507066/)）；Bjork 的"必要难度"：间隔、交错、测试、生成。
- 缺口：学习者系统性地误判学习效果（流畅性错觉），主动偏好重读——**工具必须替用户做对的事**。
- 对本项目的输入：复习播放器强制"先回忆后揭示"；自测工作流生成**新题**（不复用复习卡，避免再认冒充回忆）；到期队列按主题交错。

## 三、综合差距结论

完整学习闭环是：**材料 → 理解（导师对话/费曼）→ 内化（原子笔记）→ 巩固（FSRS 调度）→ 诊断（弱点 → 定向练习）**。

- DeepTutor 覆盖"理解"，缺"巩固"与纯文件知识库；
- Anki 覆盖"巩固"，缺"理解"与"内化"；
- Zettelkasten 覆盖"内化"，缺"巩固"与"诊断"；
- 刻意练习/Bloom/检索练习给出"诊断"与"练习设计"的原理，但没有个人可用的载体。

且所有 AI 方案都要求"系统内接 LLM"。**SmileX-Deep-Study 的补位方案**：把 LLM 能力外置给已有的编码 agent harness（ZCode/Trae/Kimi/WorkBuddy），系统本身只做三件事——纯文件数据契约（harness 可直接读写）、FSRS 调度与统计（纯数学，Go 单二进制）、管理工作流与界面的驾驶舱。零 API 成本，四家工具全兼容（AGENTS.md 是事实标准，[agents.md](https://agents.md)）。

## 四、四家 harness 兼容性核实（2026-09）

| 工具 | 项目级指令支持 | 核实结果 |
|---|---|---|
| ZCode | `<repo>/AGENTS.md` 原生加载；`.agents/skills/`（SKILL.md，通用格式，`.zcode/` 同样支持）与 `.agents/commands/`（.md 斜杠命令） | 本机 + [官方文档](https://zcode.z.ai/en/docs/agents)核实 |
| Kimi CLI | 原生只发现 AGENTS.md（CLAUDE.md 不读，[issue #2401](https://github.com/MoonshotAI/kimi-cli/issues/2401)） | [官方文档](https://moonshotai.github.io/kimi-cli/en/guides/getting-started.html)核实 |
| Trae | `.trae/rules/*.mdc`（frontmatter: alwaysApply/description/globs）；AGENTS.md 需在设置中开启导入 | [官方文档](https://docs.trae.ai/ide/rules)核实 |
| WorkBuddy (腾讯 CodeBuddy 系) | `.codebuddy/rules/<name>/RULE.mdc`；`CODEBUDDY.md` 默认全量加载；**无 CODEBUDDY.md 时自动加载 AGENTS.md** | [官方文档](https://www.workbuddy.ai/docs/ide/User-guide/Rules)核实 |

→ 结论：**AGENTS.md 作唯一权威契约**，其余三家放薄适配文件指向它。

主要来源：[HKUDS/DeepTutor](https://github.com/HKUDS/DeepTutor) · [deeptutor.info](https://deeptutor.info/) · [arXiv:2604.26962](https://arxiv.org/html/2604.26962v1) · [khanmigo.ai](https://www.khanmigo.ai/)（对照：Khanmigo 苏格拉底式不给答案） · [go-fsrs](https://github.com/open-spaced-repetition/go-fsrs) · [ts-fsrs](https://github.com/open-spaced-repetition/ts-fsrs) · [Anki FAQ](https://faqs.ankiweb.net/what-spaced-repetition-algorithm) · [agents.md](https://agents.md)
