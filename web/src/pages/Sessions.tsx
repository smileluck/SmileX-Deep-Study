import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { get, type SessionDetail, type SessionListItem } from '../api'
import Markdown from '../components/Markdown'

const typeBadge: Record<string, string> = {
  tutor: 'badge-primary',
  feynman: 'badge-secondary',
  quiz: 'badge-accent',
  diagnose: 'badge-warning',
  import: 'badge-info',
}
const typeLabel: Record<string, string> = {
  tutor: '导师',
  feynman: '费曼',
  quiz: '自测',
  diagnose: '诊断',
  import: '导入',
}

export default function Sessions() {
  const { id } = useParams()
  const [list, setList] = useState<SessionListItem[] | null>(null)
  const [detail, setDetail] = useState<SessionDetail | null>(null)

  useEffect(() => {
    get<SessionListItem[]>('/api/sessions').then(setList).catch(() => setList([]))
  }, [id])
  useEffect(() => {
    if (id) {
      setDetail(null)
      get<SessionDetail>(`/api/sessions/${id}`).then(setDetail).catch(() => setDetail(null))
    } else {
      setDetail(null)
    }
  }, [id])

  if (id) {
    return (
      <div className="mx-auto max-w-2xl p-8">
        <Link to="/sessions" className="btn btn-ghost btn-sm -ml-3 gap-1 text-sm">
          <ArrowLeft className="h-4 w-4" /> 返回列表
        </Link>
        {!detail ? (
          <div className="mt-8 text-sm opacity-50">加载中…</div>
        ) : (
          <article className="mt-4">
            <h1 className="text-lg font-bold">{String(detail.fm.summary ?? detail.id)}</h1>
            <div className="mt-2 flex flex-wrap items-center gap-2 text-xs">
              <span className={`badge badge-sm ${typeBadge[String(detail.fm.type)] ?? 'badge-ghost'}`}>
                {typeLabel[String(detail.fm.type)] ?? String(detail.fm.type)}
              </span>
              <span className="badge badge-ghost badge-sm">{String(detail.fm.topic ?? '')}</span>
              <span className="opacity-50">{String(detail.fm.date ?? '')}</span>
              <span className="font-mono text-[10px] opacity-40">via {String(detail.fm.tool ?? '?')}</span>
            </div>
            {Array.isArray(detail.fm.misconceptions) && detail.fm.misconceptions.length > 0 && (
              <div className="mt-3 rounded-xl bg-warning/10 p-3 text-xs">
                <span className="font-semibold">发现的误解：</span>
                {(detail.fm.misconceptions as string[]).map((m, i) => (
                  <div key={i} className="mt-1">· {m}</div>
                ))}
              </div>
            )}
            <div className="mt-6">
              <Markdown>{detail.body}</Markdown>
            </div>
          </article>
        )}
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl p-8">
      <h1 className="text-xl font-bold">会话记录</h1>
      <p className="mt-1 text-sm opacity-60">
        导入 / 导师 / 费曼 / 自测 / 诊断的全部落盘记录 —— 诊断工作流的数据来源。
      </p>
      {!list ? (
        <div className="mt-8 text-sm opacity-50">加载中…</div>
      ) : list.length === 0 ? (
        <div className="mt-8 rounded-xl bg-base-200/50 p-6 text-sm opacity-60">
          还没有会话。在 harness 里跑一次 <code className="rounded bg-base-300 px-1">/study:import</code> 或
          <code className="mx-1 rounded bg-base-300 px-1">/study:tutor</code> 试试。
        </div>
      ) : (
        <ul className="mt-6 flex flex-col gap-2">
          {list.map((s) => (
            <li key={s.id}>
              <Link
                to={`/sessions/${s.id}`}
                className="card block bg-base-100 p-4 shadow-sm transition-shadow hover:shadow-md"
              >
                <div className="flex items-center gap-2.5">
                  <span className={`badge badge-sm shrink-0 ${typeBadge[s.type] ?? 'badge-ghost'}`}>
                    {typeLabel[s.type] ?? s.type}
                  </span>
                  <span className="w-32 shrink-0 truncate font-mono text-xs">{s.topic || '—'}</span>
                  <span className="min-w-0 flex-1 truncate text-sm opacity-75">{s.summary || s.id}</span>
                  <span className="shrink-0 text-xs opacity-40">{s.date}</span>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
