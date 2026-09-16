import { useEffect, useRef, useState } from 'react'
import { Check, Copy } from 'lucide-react'

export default function CopyButton({ text, label }: { text: string; label?: string }) {
  const [ok, setOk] = useState(false)
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(
    () => () => {
      if (timer.current) clearTimeout(timer.current)
    },
    [],
  )

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      // 剪贴板 API 不可用时（非安全上下文）退化方案：textarea 选区 + execCommand。
      // execCommand 已废弃，但仍是该场景下唯一可编程复制的兜底；本应用跑在
      // 127.0.0.1（安全上下文）时不会走到这里。
      const ta = document.createElement('textarea')
      ta.value = text
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    setOk(true)
    if (timer.current) clearTimeout(timer.current)
    timer.current = setTimeout(() => setOk(false), 1600)
  }
  return (
    <button className={`btn btn-xs ${label ? '' : 'btn-square'} btn-ghost gap-1`} onClick={copy} title="复制">
      {ok ? <Check className="h-3.5 w-3.5 text-success" /> : <Copy className="h-3.5 w-3.5" />}
      {label ? label : null}
      {ok && label ? '已复制' : null}
    </button>
  )
}
