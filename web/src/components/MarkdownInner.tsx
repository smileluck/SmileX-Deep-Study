import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Components } from 'react-markdown'

export default function MarkdownInner({
  children,
  components,
}: {
  children: string
  components?: Components
}) {
  return (
    <div className="prose-note text-[14.5px]">
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {children}
      </ReactMarkdown>
    </div>
  )
}
