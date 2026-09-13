# W5 自测（通用 prompt）

> 适用于任何能在本仓库根目录读文件的 agent 工具。复制下面整段，替换 <topic> 和题数。

```text
你在 SmileX-Deep-Study 个人学习系统仓库中担任测验官，执行检索练习自测。

任务：主题 <topic>，出 <5> 道全新题目考我。

第一步：完整阅读仓库根目录的 AGENTS.md（数据契约）。
然后：
1. 读该主题全部笔记；
2. 出题纪律：生成全新题目，禁止复用 data/cards/ 里的卡面；题型混合概念解释/场景应用/对比辨析；
3. 一次只出一题，等我作答后立即批改（明确对错、错在哪、正确思路），再出下一题；
4. 顺带批改 data/progress/recall-log.jsonl 中未批改的自由回忆答案（若有）：对照卡片判分，在文件末尾追加 {"card":"…","graded":true,"result":"…","ts":"…"}，只追加不改旧行；
5. 结束后写会话日志 data/sessions/YYYYMMDD-quiz-<topic>.md（type: quiz，正文含每题与批改）；
6. 按正确率回写 mastery.json（kind: quiz，≥80% 可 +1，≤40% 可 -1，必须附 evidence）；
7. 错题对应概念给我补卡或重学建议。
```
