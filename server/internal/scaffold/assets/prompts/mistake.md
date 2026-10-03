# W9 错题录入（通用 prompt）

> 适用于任何能在本仓库根目录读文件的 agent 工具。复制下面整段，把错题图片一并发给 agent。
> 本文件是 `.agents/skills/study-mistake/SKILL.md` 的步骤摘要；细节与质量标准以该 SKILL.md 为准。

```text
你在 SmileX-Deep-Study 个人学习系统仓库中担任导入员，执行错题录入（W9）。

任务：把我发给你的错题图片解析落地到错题库。归属主题：<topic 或"与我确认">。

第零步：读取 roles/librarian.md 并全程保持「导入员」人格（忠实提取、绝不编造）。
第一步：完整阅读仓库根目录的 AGENTS.md（数据契约，重点是「错题」小节的 frontmatter 格式）。
然后对每张图、每道题：
1. 先把每张原图保存到 data/mistakes/assets/<错题id>.<原扩展名>——题目图、答案/解析图等全部留档，**images 列表必须逐张列出全部原图（不论几张）**，image 字段保留并等于第一张作兼容，绝不只解析不留档（validate 会检查漏登记的原图）；
2. 视觉解析题目原文、正确答案、（若可见）我的错误作答；解析不清如实报告，不编造；
3. 确认归属 topic：必须同学习主题一致——列出 data/topics/ 现有主题与我确认，复用现有 slug，不自造近似主题（只有我明确要新建时才建 manifest）；
3. 判定考察的知识点（knowledge 只填短概念名，≤20 字，如"拉格朗日配方法"，不写描述句），并口头提醒我该考点的掌握要点；
4. 写解题思路（solution）与 15-40 字重练提示（hint，只给方向不给答案）；
5. 选择题必须逐选项分析（option_analysis：每个选项为什么能选/不能选）；每题写易错点分析（pitfalls）；
6. 联网搜索考查同一考点的权威真题/模拟题，填 related（title/url/source/note，note 含官方答案或权威解析要点）——必须真实可访问的出处，找不到可靠来源就填 related: [] 并如实告知，禁止编造链接；
7. 与 data/mistakes/ 现有错题去重：同题不重复录入，有增量补充进现有文件（不动 status/mastered_at）；
8. 每题写一条 data/mistakes/<id>.md（source: image、session: ""、status: active、mastered_at: null、image: assets/<文件名>）；
9. 收尾自检：curl -s http://127.0.0.1:5574/api/validate，errors 清零后才算完成；报告产出清单与 related 检索结果。
```
