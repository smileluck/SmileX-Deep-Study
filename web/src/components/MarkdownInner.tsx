import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import remarkBreaks from 'remark-breaks'
import rehypeKatex from 'rehype-katex'
import type { Components } from 'react-markdown'
import 'katex/dist/katex.min.css'

// remark-math 的块级公式要求开/关 $$ 独占一行：独占一行的 $$...$$ 会被当成
// 行内公式，而跨行公式中开在 $$ 同一行的首行内容会被当作 meta 丢弃。
// 统一把 $$ 分隔符规范到独立行（代码块内不动）。
function normalizeDisplayMath(src: string): string {
  let inFence = false
  let inMath = false
  const out: string[] = []
  for (const line of src.split('\n')) {
    if (/^\s*```/.test(line)) {
      inFence = !inFence
      out.push(line)
      continue
    }
    if (inFence) {
      out.push(line)
      continue
    }
    if (inMath) {
      const close = line.match(/^(.*?)\$\$\s*$/)
      if (close) {
        if (close[1].trim()) out.push(close[1])
        out.push('$$')
        inMath = false
      } else {
        out.push(line)
      }
      continue
    }
    const single = line.match(/^\s*\$\$([^$]+)\$\$\s*$/)
    if (single) {
      out.push('$$', single[1].trim(), '$$')
      continue
    }
    const open = line.match(/^\s*\$\$(.+)$/)
    if (open && !open[1].includes('$$')) {
      out.push('$$', open[1])
      inMath = true
      continue
    }
    out.push(line)
  }
  return out.join('\n')
}

export default function MarkdownInner({
  children,
  components,
  softBreaks,
}: {
  children: string
  components?: Components
  softBreaks?: boolean
}) {
  return (
    <div className="prose-note text-[14.5px]">
      <ReactMarkdown
        remarkPlugins={softBreaks ? [remarkGfm, remarkMath, remarkBreaks] : [remarkGfm, remarkMath]}
        rehypePlugins={[[rehypeKatex, { strict: false }]]}
        components={components}
      >
        {normalizeDisplayMath(children)}
      </ReactMarkdown>
    </div>
  )
}
