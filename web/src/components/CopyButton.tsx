import { useState } from 'react'
import { Check, Copy } from 'lucide-react'

export default function CopyButton({ text, label }: { text: string; label?: string }) {
  const [ok, setOk] = useState(false)
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      // 剪贴板不可用时退化为选区复制
      const ta = document.createElement('textarea')
      ta.value = text
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    setOk(true)
    setTimeout(() => setOk(false), 1600)
  }
  return (
    <button className={`btn btn-xs ${label ? '' : 'btn-square'} btn-ghost gap-1`} onClick={copy} title="复制">
      {ok ? <Check className="h-3.5 w-3.5 text-success" /> : <Copy className="h-3.5 w-3.5" />}
      {label ? label : null}
      {ok && label ? '已复制' : null}
    </button>
  )
}
