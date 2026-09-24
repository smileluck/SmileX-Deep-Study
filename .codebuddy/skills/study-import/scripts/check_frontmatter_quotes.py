#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""frontmatter 体检：ASCII 双引号奇偶 + 会话 type 合法性。

背景（SKILL.md 坑记录 40）：YAML 双引号标量里若嵌套了 ASCII `"`，会在该处提前闭合，
go-yaml 却报一个完全看不出因果的错（如 `did not find expected '-' indicator`，
且行号常常指向上一行）。本脚本用「引号数为奇数」这一机械判据把这类问题直接揪出来。

用法：
    python check_frontmatter_quotes.py [data目录]      # 默认 ./data
退出码：0 = 干净；1 = 有问题（便于 CI / 收尾自检串联）。
"""
import os
import re
import sys

# 与 server/internal/api/validate.go 的 sessionTypes 保持一致
SESSION_TYPES = {"tutor", "feynman", "quiz", "diagnose", "import", "plan", "merge"}

FM = re.compile(r"^---\r?\n([\s\S]*?)\r?\n---")


def main() -> int:
    root = sys.argv[1] if len(sys.argv) > 1 else "data"
    if not os.path.isdir(root):
        print(f"目录不存在: {root}")
        return 1

    scanned = 0
    problems = []

    for dirpath, _dirs, files in os.walk(root):
        for fn in files:
            if not fn.endswith(".md"):
                continue
            path = os.path.join(dirpath, fn)
            rel = os.path.relpath(path, root).replace("\\", "/")
            try:
                text = open(path, "rb").read().decode("utf-8")
            except Exception as e:  # noqa: BLE001
                problems.append((rel, 0, f"读取失败: {e}"))
                continue

            m = FM.match(text)
            if not m:
                continue
            scanned += 1
            fm_body = m.group(1)
            id_val = None

            for i, line in enumerate(fm_body.split("\n"), start=1):
                line = line.rstrip("\r")
                n = line.count('"')
                if n % 2 == 1:
                    problems.append(
                        (rel, i, "ASCII 双引号数为奇数（%d）→ 疑似嵌套/未闭合: %s" % (n, line.strip()[:100]))
                    )
                mm = re.match(r"^id:\s*(.+?)\s*$", line)
                if mm:
                    id_val = mm.group(1).strip().strip('"').strip("'")

            # 会话日志：type 必须在枚举内
            if rel.startswith("sessions/"):
                mt = re.search(r"^type:\s*(.+?)\s*$", fm_body, re.M)
                if mt:
                    t = mt.group(1).strip().strip('"').strip("'")
                    if t not in SESSION_TYPES:
                        problems.append(
                            (rel, 0, f"session type 非法: {t!r}（合法: {'/'.join(sorted(SESSION_TYPES))}）")
                        )
                # id 必须与文件名一致（validate.go 会校验）
                stem = fn[:-3]
                if id_val and id_val != stem:
                    problems.append((rel, 0, f"frontmatter id ({id_val}) 与文件名 ({stem}) 不一致"))

    print(f"扫描（含 frontmatter 的）文件: {scanned}")
    print(f"问题数: {len(problems)}")
    for rel, i, msg in problems:
        loc = f"{rel}:{i}" if i else rel
        print(f"  {loc}  {msg}")

    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
