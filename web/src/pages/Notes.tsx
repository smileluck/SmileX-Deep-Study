import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ChevronDown, ChevronRight, ChevronsDownUp, ChevronsUpDown, Link2, Network, Search, X } from 'lucide-react'
import { get, type NoteDetail, type NoteListItem } from '../api'
import Markdown from '../components/Markdown'

// 笔记详情：列表模式的详情页与图谱模式的浮层面板共用。
function NoteDetailView({ detail }: { detail: NoteDetail }) {
  return (
    <article className="mx-auto max-w-2xl p-8">
      <h1 className="text-xl font-bold">{detail.fm.title ?? detail.id}</h1>
      <div className="mt-2 flex flex-wrap items-center gap-2 text-xs opacity-60">
        <span className="badge badge-ghost">{detail.fm.topic ?? '未分类'}</span>
        {(detail.fm.tags ?? []).map((t) => (
          <span key={t} className="badge badge-outline badge-sm">#{t}</span>
        ))}
        <span className="font-mono text-[10px] opacity-60">{detail.id}</span>
      </div>

      {(detail.backlinks.length > 0 || detail.fm.links?.length) && (
        <div className="mt-3 flex flex-wrap items-center gap-1.5 text-xs">
          <Link2 className="h-3.5 w-3.5 opacity-50" />
          {detail.fm.links?.map((l) => (
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
  )
}

// 单主题笔记图谱：节点均匀分布在圆环上 + links 连线（个人规模足够）。
function NoteGraph({ notes, selected, onSelect }: {
  notes: NoteListItem[]
  selected?: string | null
  onSelect: (id: string) => void
}) {
  const W = 800
  const H = 560
  const [hover, setHover] = useState<string | null>(null)

  const pos = useMemo(() => {
    const p = new Map<string, { x: number; y: number; inner: boolean }>()
    const rOuter = Math.min(W, H) * 0.34
    const rInner = Math.min(W, H) * 0.23
    // 节点较多时内外两环交错，避免标签互相挤压。
    const stagger = notes.length > 10
    notes.forEach((n, i) => {
      if (notes.length === 1) {
        p.set(n.id, { x: W / 2, y: H / 2, inner: false })
        return
      }
      const inner = stagger && i % 2 === 1
      const r = inner ? rInner : rOuter
      const a = (i / notes.length) * Math.PI * 2 - Math.PI / 2
      p.set(n.id, { x: W / 2 + Math.cos(a) * r, y: H / 2 + Math.sin(a) * r, inner })
    })
    return p
  }, [notes])

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

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="h-full w-full">
      {edges.map((e) => {
        const a = pos.get(e.from)!
        const b = pos.get(e.to)!
        return (
          <line key={e.key} x1={a.x} y1={a.y} x2={b.x} y2={b.y} stroke="var(--color-base-content)" strokeOpacity={0.18} strokeWidth={1.2} />
        )
      })}
      {notes.map((n) => {
        const p = pos.get(n.id)!
        const isSel = n.id === selected
        // 内环节点的标签只在悬停/选中时显示，避免与外环标签重叠。
        const showLabel = !p.inner || isSel || hover === n.id
        return (
          <g
            key={n.id}
            onClick={() => onSelect(n.id)}
            onMouseEnter={() => setHover(n.id)}
            onMouseLeave={() => setHover((h) => (h === n.id ? null : h))}
            style={{ cursor: 'pointer' }}
          >
            <circle
              cx={p.x}
              cy={p.y}
              r={isSel ? 11 : 8}
              fill="var(--color-primary)"
              fillOpacity={isSel ? 1 : 0.55}
              stroke={isSel ? 'var(--color-base-content)' : 'none'}
              strokeWidth={1.5}
            />
            {showLabel && (
              <text x={p.x} y={p.y + 26} textAnchor="middle" fontSize={12} fill="var(--color-base-content)" fillOpacity={0.8}>
                {n.title.length > 14 ? n.title.slice(0, 14) + '…' : n.title}
              </text>
            )}
          </g>
        )
      })}
    </svg>
  )
}

export default function Notes() {
  const { id } = useParams()
  const [notes, setNotes] = useState<NoteListItem[] | null>(null)
  const [detail, setDetail] = useState<NoteDetail | null>(null)
  const [q, setQ] = useState('')
  const [tab, setTab] = useState<'list' | 'graph'>('list')
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  const [graphTopic, setGraphTopic] = useState<string | null>(null)
  const [graphSelected, setGraphSelected] = useState<string | null>(null)
  const [graphDetail, setGraphDetail] = useState<NoteDetail | null>(null)
  const [listErr, setListErr] = useState('')
  // 请求序号：快速切换笔记/图谱节点时递增，过期响应直接丢弃（竞态防护）
  const detailSeq = useRef(0)
  const graphSeq = useRef(0)

  const toggleGroup = (topic: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(topic)) {
        next.delete(topic)
      } else {
        next.add(topic)
      }
      return next
    })
  }

  useEffect(() => {
    get<NoteListItem[]>('/api/notes')
      .then(setNotes)
      .catch((e) => setListErr(String(e.message ?? e)))
  }, [])
  useEffect(() => {
    const seq = ++detailSeq.current
    if (id) {
      setDetail(null)
      get<NoteDetail>(`/api/notes/${id}`)
        .then((d) => {
          if (seq === detailSeq.current) setDetail(d)
        })
        .catch(() => {
          if (seq === detailSeq.current) setDetail(null)
        })
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

  const grouped = useMemo(() => {
    const m = new Map<string, NoteListItem[]>()
    for (const n of filtered) {
      const key = n.topic || ''
      const arr = m.get(key) ?? []
      arr.push(n)
      m.set(key, arr)
    }
    return [...m.entries()].sort(([a], [b]) => a.localeCompare(b))
  }, [filtered])

  // 图谱模式：默认选中第一个主题；当前主题被搜索过滤掉时回退。
  useEffect(() => {
    if (tab !== 'graph' || grouped.length === 0) return
    if (graphTopic === null || !grouped.some(([t]) => t === graphTopic)) {
      setGraphTopic(grouped[0][0])
    }
  }, [tab, grouped, graphTopic])

  // 切换主题 / 切回列表时关闭详情面板。
  useEffect(() => {
    setGraphSelected(null)
    setGraphDetail(null)
  }, [graphTopic, tab])

  // 图谱节点选中后拉取详情。
  useEffect(() => {
    const seq = ++graphSeq.current
    if (!graphSelected) {
      setGraphDetail(null)
      return
    }
    setGraphDetail(null)
    get<NoteDetail>(`/api/notes/${graphSelected}`)
      .then((d) => {
        if (seq === graphSeq.current) setGraphDetail(d)
      })
      .catch(() => {
        if (seq === graphSeq.current) setGraphDetail(null)
      })
  }, [graphSelected])

  const graphNotes = useMemo(() => {
    if (graphTopic === null) return []
    return grouped.find(([t]) => t === graphTopic)?.[1] ?? []
  }, [grouped, graphTopic])

  return (
    <div className="flex h-full">
      <div className="flex w-80 shrink-0 flex-col border-r border-base-300 bg-base-200/30">
        <div className="p-4 pb-2">
          <div className="flex items-center justify-between px-1">
            <h1 className="text-base font-bold">笔记</h1>
            {tab === 'list' && (
              <div className="flex items-center gap-0.5">
                <button
                  className="btn btn-ghost btn-xs btn-circle"
                  title="展开全部"
                  onClick={() => setCollapsed(new Set())}
                >
                  <ChevronsUpDown className="h-3.5 w-3.5" />
                </button>
                <button
                  className="btn btn-ghost btn-xs btn-circle"
                  title="折叠全部"
                  onClick={() => setCollapsed(new Set(grouped.map(([t]) => t)))}
                >
                  <ChevronsDownUp className="h-3.5 w-3.5" />
                </button>
              </div>
            )}
          </div>
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
            listErr ? (
              <div className="p-4 text-sm text-error">{listErr}</div>
            ) : !notes ? (
              <div className="p-4 text-sm opacity-50">加载中…</div>
            ) : filtered.length === 0 ? (
              <div className="p-4 text-sm opacity-50">
                暂无笔记。笔记由导入工作流（/study:import）生成。
              </div>
            ) : (
              <div className="flex flex-col gap-4">
                {grouped.map(([topic, list]) => {
                  const isCollapsed = !q.trim() && collapsed.has(topic)
                  return (
                    <section key={topic || '_uncategorized'}>
                      <button
                        className="flex w-full items-center gap-1.5 rounded-lg px-3 py-1 text-[11px] font-medium opacity-60 transition-opacity hover:opacity-100"
                        onClick={() => toggleGroup(topic)}
                      >
                        {isCollapsed ? <ChevronRight className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                        <span className="font-mono">{topic || '未分类'}</span>
                        <span>{list.length} 篇</span>
                      </button>
                      {!isCollapsed && (
                        <ul className="mt-0.5 flex flex-col gap-1">
                          {list.map((n) => (
                            <li key={n.id}>
                              <Link
                                to={`/notes/${n.id}`}
                                className={`block rounded-lg px-3 py-2 text-sm transition-colors ${
                                  id === n.id ? 'bg-primary/10 font-medium text-primary' : 'hover:bg-base-200'
                                }`}
                              >
                                <div className="flex items-center gap-1.5">
                                  {n.order != null && (
                                    <span className="badge badge-ghost badge-xs shrink-0">#{n.order}</span>
                                  )}
                                  <div className="truncate">{n.title || n.id}</div>
                                </div>
                                {n.cards > 0 && (
                                  <div className="mt-0.5 text-[11px] opacity-50">{n.cards} 卡</div>
                                )}
                                {n.gaps > 0 && (
                                  <div className="mt-0.5 text-[11px] text-warning/80">{n.gaps} 个待补 Gap</div>
                                )}
                              </Link>
                            </li>
                          ))}
                        </ul>
                      )}
                    </section>
                  )
                })}
              </div>
            )
          ) : listErr ? (
            <div className="p-4 text-sm text-error">{listErr}</div>
          ) : !notes ? (
            <div className="p-4 text-sm opacity-50">加载中…</div>
          ) : grouped.length === 0 ? (
            <div className="p-4 text-sm opacity-50">
              暂无笔记。笔记由导入工作流（/study:import）生成。
            </div>
          ) : (
            <ul className="flex flex-col gap-1">
              {grouped.map(([topic, list]) => (
                <li key={topic || '_uncategorized'}>
                  <button
                    className={`block w-full rounded-lg px-3 py-2 text-left text-sm transition-colors ${
                      graphTopic === topic ? 'bg-primary/10 font-medium text-primary' : 'hover:bg-base-200'
                    }`}
                    onClick={() => setGraphTopic(topic)}
                  >
                    <span className="font-mono">{topic || '未分类'}</span>
                    <span className="ml-2 text-[11px] opacity-50">{list.length} 篇</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {tab === 'graph' ? (
        <div className="relative flex-1">
          {graphTopic === null || graphNotes.length === 0 ? (
            <div className="flex h-full items-center justify-center text-sm opacity-50">从左侧选择一个主题</div>
          ) : (
            <div className="h-full p-6">
              <NoteGraph notes={graphNotes} selected={graphSelected} onSelect={(nid) => setGraphSelected(nid)} />
            </div>
          )}
          {graphSelected && (
            <div className="absolute inset-y-0 right-0 w-[26rem] max-w-full overflow-y-auto border-l border-base-300 bg-base-100 shadow-xl">
              <button
                className="btn btn-ghost btn-sm btn-circle absolute right-3 top-3 z-10"
                onClick={() => setGraphSelected(null)}
                aria-label="关闭"
              >
                <X className="h-4 w-4" />
              </button>
              {!graphDetail ? (
                <div className="p-8 text-sm opacity-50">加载中…</div>
              ) : (
                <NoteDetailView detail={graphDetail} />
              )}
            </div>
          )}
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto">
          {!id ? (
            <div className="flex h-full items-center justify-center text-sm opacity-50">从左侧选择一篇笔记</div>
          ) : !detail ? (
            <div className="p-8 text-sm opacity-50">加载中…</div>
          ) : (
            <NoteDetailView detail={detail} />
          )}
        </div>
      )}
    </div>
  )
}
