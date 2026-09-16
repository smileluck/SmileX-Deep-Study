import { lazy, Suspense } from 'react'
import type { Components } from 'react-markdown'

// react-markdown + remark-gfm 是主 chunk 的最大头，按需加载；
// 加载期间先以纯文本兜底，避免内容闪烁空白。
const MarkdownInner = lazy(() => import('./MarkdownInner'))

export default function Markdown({
  children,
  components,
}: {
  children: string
  components?: Components
}) {
  return (
    <Suspense fallback={<div className="text-[14.5px] whitespace-pre-wrap">{children}</div>}>
      <MarkdownInner components={components}>{children}</MarkdownInner>
    </Suspense>
  )
}
