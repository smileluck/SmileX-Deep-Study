import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Link2, Network, Search } from 'lucide-react'
import { get, type NoteDetail, type NoteListItem } from '../api'
import Markdown from '../components/Markdown'

// 简版笔记图谱：圆形布局 + links 连线（个人规模足够）。
function NoteGraph({ notes, selected, onSelect }: {
  notes: NoteListItem[]
  selected?: string
  onSelect: (id: string) => void
}) {
  const W = 560
  const H = 420
  const byTopic = useMemo(() => {
    const m = new Map<string, NoteListItem[]>()
    for (const n of notes) {
      const arr = m.get(n.topic || '') ?? []
      arr.push(n)
      m.set(n.topic || '', arr)
    }
    return m
  }, [notes])

  const pos = useMemo(() => {
    const topics = [...byTopic.keys()]
    const p = new Map<string, { x: number; y: number }>()
    topics.forEach((t, ti) => {
      const list = byTopic.get(t)!
      const ta = (ti / Math.max(1, topics.length)) * Math.PI * 2
      const ring = topics.length > 1 ? 0.62 : 0
      const cx = W / 2 + Math.cos(ta) * ring * W * 0.38
      const cy = H / 2 + Math.sin(ta) * ring * H * 0.36
      list.forEach((n, ni) => {
        const na = (ni / Math.max(1, list.length)) * Math.PI * 2 + ti
        const r = list.length > 1 ? 42 + Math.min(list.length, 8) * 6 : 0
        p.set(n.id, {
          x: Math.max(24, Math.min(W - 24, cx + Math.cos(na) * r)),
          y: Math.max(20, Math.min(H - 20, cy + Math.sin(na) * r)),
        })
      })
    })
    return p
  }, [byTopic])

  const edges = useMemo(() => {
    const seen = new Set<string>()
    const out: { from: string; to: string; key: string }[] = []
    for (const n of notes) {
      for (const l of n.links ?? []) {
        const key = [n.id, l].sort().join('→')
        if (!seen.has(key) && pos.has(l)) {
          seen.add(key)
          out.push({ from: n.id, to: l, key })
        }
      }
    }
    return out
  }, [notes, pos])

  const topicColors = ['var(--color-primary)', 'var(--color-secondary)', 'var(--color-accent)', 'var(--color-info)', 'var(--color-success)']
  const topicList = [...byTopic.keys()]

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="w-full rounded-xl bg-base-200/40">
      {edges.map((e) => {
        const a = pos.get(e.from)!
        const b = pos.get(e.to)!
        return (
          <line key={e.key} x1={a.x} y1={a.y} x2={b.x} y2={b.y} stroke="var(--color-base-content)" strokeOpacity={0.18} strokeWidth={1.2} />
        )
      })}
      {[...byTopic.entries()].map(([topic, list], ti) => (
        <g key={topic}>
          {topic && <text x={pos.get(list[0].id)!.x} y={Math.max(12, pos.get(list[0].id)!.y - 46)} textAnchor="middle" fontSize={11} fill={topicColors[ti % topicColors.length]} opacity={0.75}>{topic}</text>}
          {list.map((n) => {
            const p = pos.get(n.id)!
            const isSel = n.id === selected
            return (
              <g key={n.id} onClick={() => onSelect(n.id)} style={{ cursor: 'pointer' }}>
                <circle
                  cx={p.x}
                  cy={p.y}
                  r={isSel ? 9 : 6}
                  fill={topicColors[ti % topicColors.length]}
                  fillOpacity={isSel ? 1 : 0.55}
                  stroke={isSel ? 'var(--color-base-content)' : 'none'}
                  strokeWidth={1.5}
                />
                <text x={p.x} y={p.y + 20} textAnchor="middle" fontSize={9.5} fill="var(--color-base-content)" fillOpacity={0.75}>
                  {n.title.length > 10 ? n.title.slice(0, 10) + '…' : n.title}
                </text>
              </g>
            )
          })}
        </g>
      ))}
    </svg>
  )
}

export default function Notes() {
  const { id } = useParams()
  const nav = useNavigate()
  const [notes, setNotes] = useState<NoteListItem[] | null>(null)
  const [detail, setDetail] = useState<NoteDetail | null>(null)
  const [q, setQ] = useState('')
  const [tab, setTab] = useState<'list' | 'graph'>('list')

  useEffect(() => {
    get<NoteListItem[]>('/api/notes').then(setNotes).catch(() => setNotes([]))
  }, [])
  useEffect(() => {
    if (id) {
      setDetail(null)
      get<NoteDetail>(`/api/notes/${id}`).then(setDetail).catch(() => setDetail(null))
    } else {
      setDetail(null)
    }
  }, [id])

  const filtered = useMemo(() => {
    if (!notes) return []
    if (!q.trim()) return notes
    const s = q.trim().toLowerCase()
    return notes.filter(
      (n) =>
        n.title.toLowerCase().includes(s) ||
        n.id.toLowerCase().includes(s) ||
        n.topic.toLowerCase().includes(s) ||
        (n.tags ?? []).some((t) => String(t).toLowerCase().includes(s)),
    )
  }, [notes, q])

  return (
    <div className="flex h-full">
      <div className="flex w-80 shrink-0 flex-col border-r border-base-300 bg-base-200/30">
        <div className="p-4 pb-2">
          <h1 className="px-1 text-base font-bold">笔记</h1>
          <label className="input input-sm mt-3 flex items-center gap-2 rounded-lg bg-base-100 px-3">
            <Search className="h-3.5 w-3.5 opacity-40" />
            <input
              className="grow bg-transparent text-sm outline-none"
              placeholder="搜索标题 / 主题 / 标签"
              value={q}
              onChange={(e) => setQ(e.target.value)}
            />
          </label>
          <div className="tabs tabs-box tabs-xs mt-3 self-center">
            <button className={`tab ${tab === 'list' ? 'tab-active' : ''}`} onClick={() => setTab('list')}>列表</button>
            <button className={`tab ${tab === 'graph' ? 'tab-active' : ''}`} onClick={() => setTab('graph')}>
              <span className="flex items-center gap-1"><Network className="h-3 w-3" /> 图谱</span>
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto px-3 pb-4">
          {tab === 'list' ? (
            !notes ? (
              <div className="p-4 text-sm opacity-50">加载中…</div>
            ) : filtered.length === 0 ? (
              <div className="p-4 text-sm opacity-50">
                暂无笔记。笔记由导入工作流（/study:import）生成。
              </div>
            ) : (
              <ul className="flex flex-col gap-1">
                {filtered.map((n) => (
                  <li key={n.id}>
                    <Link
                      to={`/notes/${n.id}`}
                      className={`block rounded-lg px-3 py-2 text-sm transition-colors ${
                        id === n.id ? 'bg-primary/10 font-medium text-primary' : 'hover:bg-base-200'
                      }`}
                    >
                      <div className="truncate">{n.title || n.id}</div>
                      <div className="mt-0.5 flex items-center gap-1.5 text-[11px] opacity-50">
                        <span className="font-mono">{n.topic || '未分类'}</span>
                        {n.cards > 0 && <span>· {n.cards} 卡</span>}
                      </div>
                    </Link>
                  </li>
                ))}
              </ul>
            )
          ) : (
            notes && <NoteGraph notes={filtered} selected={id} onSelect={(nid) => nav(`/notes/${nid}`)} />
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {!id ? (
          <div className="flex h-full items-center justify-center text-sm opacity-50">从左侧选择一篇笔记</div>
        ) : !detail ? (
          <div className="p-8 text-sm opacity-50">加载中…</div>
        ) : (
          <article className="mx-auto max-w-2xl p-8">
            <h1 className="text-xl font-bold">{String(detail.fm.title ?? detail.id)}</h1>
            <div className="mt-2 flex flex-wrap items-center gap-2 text-xs opacity-60">
              <span className="badge badge-ghost">{String(detail.fm.topic ?? '未分类')}</span>
              {(Array.isArray(detail.fm.tags) ? detail.fm.tags : []).map((t) => (
                <span key={String(t)} className="badge badge-outline badge-sm">#{String(t)}</span>
              ))}
              <span className="font-mono text-[10px] opacity-60">{detail.id}</span>
            </div>

            {(detail.backlinks.length > 0 || (detail.fm.links as string[] | undefined)?.length) && (
              <div className="mt-3 flex flex-wrap items-center gap-1.5 text-xs">
                <Link2 className="h-3.5 w-3.5 opacity-50" />
                {(detail.fm.links as string[] | undefined)?.map((l) => (
                  <Link key={l} to={`/notes/${l}`} className="badge badge-sm badge-primary badge-outline">
                    → {l}
                  </Link>
                ))}
                {detail.backlinks.map((b) => (
                  <Link key={b} to={`/notes/${b}`} className="badge badge-sm badge-secondary badge-outline">
                    ← {b}
                  </Link>
                ))}
              </div>
            )}

            <div className="mt-6">
              <Markdown>{detail.body}</Markdown>
            </div>

            {detail.cards.length > 0 && (
              <section className="mt-10">
                <h2 className="text-sm font-semibold">关联卡片（{detail.cards.length}）</h2>
                <ul className="mt-3 flex flex-col gap-2">
                  {detail.cards.map((c) => (
                    <li key={String(c.id)} className="rounded-xl bg-base-100 p-3.5 text-sm shadow-sm">
                      <div className="font-medium">{String(c.front)}</div>
                      <div className="mt-1.5 whitespace-pre-wrap text-xs opacity-65">{String(c.back)}</div>
                      <div className="mt-2 flex items-center gap-2 text-[10px] opacity-45">
                        <span className="font-mono">{String(c.id)}</span>
                        <span>· {String(c.state_name ?? '')}</span>
                      </div>
                    </li>
                  ))}
                </ul>
              </section>
            )}
          </article>
        )}
      </div>
    </div>
  )
}
