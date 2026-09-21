# -*- coding: utf-8 -*-
"""修复 GFM 表格被竖线拆坏的问题（配合 check_math_render.mjs 使用）。

    python -X utf8 fix_table_pipes.py <topic> [notesDir] [--apply] [--widen]

三类修法：

① 表格行 $...$ 数学区间内的裸 |（如 $|z|$、$\left|F_n\right|$）
   → 换成 \\vert。KaTeX 中等价，且不含 | 字符，GFM 就不会拆列。
   **必须写成 "\\vert " 带尾随空格**，否则 `$|H_a|$` 会变成 `\\vertH_a`
   而被当成控制字 \\vertH → KaTeX 报 undefined control sequence。

② 表格行数学区间**之外**、且两侧都紧贴非空白字符的 |（如 `ln|x|`、`E(|X|)`）
   → 转义成 \\|（GFM 规范写法，渲染为字面 |）。
   真正的列分隔符在本项目里一律写成 " | "（两侧有空格），据此区分，不要误伤分隔符。

③ 数据行格子数确实多于表头的表（内容里本来就多写了一列）
   → 默认只报告，不动手；确认后用 --widen 把表头与分隔行补到最大列数
   （GFM 会丢弃多出的格子，不补表头就丢内容）。**注意区分 ① 与 ③**：
   若某行多出的格子其实来自 ①，应修 ① 而不是加宽表头。
"""
import os
import re
import sys

ARGS = [a for a in sys.argv[1:] if not a.startswith('--')]
TOPIC = ARGS[0] if ARGS else 'signals-and-systems'
HERE = os.path.dirname(os.path.abspath(__file__))
NOTES = ARGS[1] if len(ARGS) > 1 else os.path.join(HERE, '..', '..', '..', '..', 'data', 'notes')
NOTES = os.path.abspath(NOTES)
APPLY = '--apply' in sys.argv
WIDEN = '--widen' in sys.argv

SEP = re.compile(r'^\s*\|[\s:|-]+\|\s*$')
ROW = re.compile(r'^\s*\|')
CELL = re.compile(r'(?<!\\)\|')


def split_cells(line):
    s = line.strip()
    if s.startswith('|'):
        s = s[1:]
    if s.endswith('|'):
        s = s[:-1]
    return CELL.split(s)


def fix_row(line):
    """返回 (新行, 数学内替换数, 数学外转义数)。"""
    out = []
    in_math = False
    n_math = 0
    n_esc = 0
    i = 0
    while i < len(line):
        ch = line[i]
        if ch == '\\' and i + 1 < len(line):
            out.append(line[i:i + 2])
            i += 2
            continue
        if ch == '$':
            in_math = not in_math
            out.append(ch)
            i += 1
            continue
        if ch == '|':
            nxt = line[i + 1] if i + 1 < len(line) else ''
            prv = line[i - 1] if i > 0 else ''
            if in_math:
                out.append(r'\vert ' if (nxt.isalpha() or nxt == '\\') else r'\vert')
                n_math += 1
            elif prv.strip() and nxt.strip():
                out.append(r'\|')          # 数学外、两侧紧贴 → 字面竖线
                n_esc += 1
            else:
                out.append(ch)             # 真正的列分隔符 " | "
            i += 1
            continue
        out.append(ch)
        i += 1
    return ''.join(out), n_math, n_esc


def main():
    if not os.path.isdir(NOTES):
        print('找不到笔记目录:', NOTES)
        return
    fixable, widenables = [], []
    for fn in sorted(os.listdir(NOTES)):
        if not fn.endswith('.md'):
            continue
        p = os.path.join(NOTES, fn)
        txt = open(p, encoding='utf-8').read()
        if 'topic: "%s"' % TOPIC not in txt[:800]:
            continue
        lines = txt.split('\n')
        out = list(lines)
        a = b = c = 0
        i = 0
        while i < len(lines):
            if not (ROW.match(lines[i]) and i + 1 < len(lines) and SEP.match(lines[i + 1])):
                i += 1
                continue
            j = i
            while j < len(lines) and ROW.match(lines[j]):
                if SEP.match(lines[j]):        # 分隔行 |---|---| 原样保留
                    j += 1
                    continue
                new, nm, ne = fix_row(lines[j])
                if nm or ne:
                    out[j] = new
                a += nm
                b += ne
                j += 1
            ncol = len(split_cells(out[i]))
            maxcol = max(len(split_cells(out[t])) for t in range(i, j))
            if maxcol > ncol:
                c += 1
                if WIDEN:
                    head = split_cells(out[i])
                    while len(head) < maxcol:
                        head.append('要点' if len(head) >= 2 else '补充')
                    cr = '\r' if out[i].endswith('\r') else ''
                    out[i] = '| ' + ' | '.join(x.strip() for x in head) + ' |' + cr
                    out[i + 1] = '|' + '---|' * maxcol + cr
            i = j
        if a or b or c:
            fixable.append((fn[:-3], a, b))
            if c:
                widenables.append((fn[:-3], c))
            if APPLY:
                open(p, 'w', encoding='utf-8', newline='\n').write('\n'.join(out))

    print('主题:', TOPIC, ' 模式:', 'APPLY' if APPLY else 'DRY-RUN', ' 加宽表头:', WIDEN)
    print('数学内 → \\vert:', sum(x[1] for x in fixable),
          ' 数学外 → \\|:', sum(x[2] for x in fixable),
          ' 需人工确认的表:', sum(x[1] for x in widenables))
    for nid, a, b in fixable:
        print('  %-58s \\vert %d  \\| %d' % (nid, a, b))
    if widenables:
        print('\n下列表的表头列数少于数据行，GFM 会丢弃多余格子（请人工确认是加宽表头还是上面的 ①/② 问题）:')
        for nid, c in widenables:
            print('  %-58s %d 个表' % (nid, c))
        if not WIDEN:
            print('  → 确认后加 --widen 生效')
    if not APPLY:
        print('\n（DRY-RUN，未写盘；加 --apply 生效）')


if __name__ == '__main__':
    main()
