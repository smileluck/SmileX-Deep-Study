import { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  ChevronRight,
  ExternalLink,
  Eye,
  ImageIcon,
  Lightbulb,
  Play,
  RotateCcw,
  ScrollText,
  X,
} from 'lucide-react'
import { get, post, fmtDate, type Mistake, type MistakesResp, type Topic } from '../api'

type Filter = 'active' | 'mastered' | 'all'

const filterLabel: Record<Filter, string> = {
  active: '待攻克',
  mastered: '已攻克',
  all: '全部',
}

export default function Mistakes() {
  const [data, setData] = useState<MistakesResp | null>(null)
  const [topics, setTopics] = useState<Topic[]>([])
  const [err, setErr] = useState('')
  const [filter, setFilter] = useState<Filter>('active')
  const [params, setParams] = useSearchParams()
  const topic = params.get('topic') ?? ''
  const setTopic = (t: string) => {
    if (t) params.set('topic', t)
    else params.delete('topic')
    setParams(params, { replace: true })
  }
  const [selected, setSelected] = useState<Mistake | null>(null)
  const [busy, setBusy] = useState('')
  // 重练模式：纯前端状态，除「标记攻克」外不写任何接口
  const [drill, setDrill] = useState<Mistake[] | null>(null)
  const [idx, setIdx] = useState(0)
  const [revealed, setRevealed] = useState(false)
  // 原图灯箱：点击查看大图
  const [viewer, setViewer] = useState<string | null>(null)

  const load = () => {
    get<MistakesResp>('/api/mistakes')
      .then(setData)
      .catch((e) => setErr(String(e.message ?? e)))
  }
  useEffect(() => {
    load()
    get<Topic[]>('/api/topics')
      .then(setTopics)
      .catch(() => {})
  }, [])

  const topicName = useMemo(() => {
    const m: Record<string, string> = {}
    for (const t of topics) m[t.slug] = t.name || t.slug
    return m
  }, [topics])

  const mistakes = data?.mistakes ?? []
  // 主题筛选与学习主题同源（/api/topics）；错题里出现的未知 slug 兜底列出，避免被筛没
  const topicOptions = useMemo(() => {
    const slugs = topics.map((t) => t.slug)
    const extra = [...new Set(mistakes.map((m) => m.topic))].filter((t) => !slugs.includes(t))
    return [...slugs, ...extra.sort()]
  }, [topics, mistakes])
  const filtered = mistakes.filter(
    (m) => (filter === 'all' || m.status === filter) && (!topic || m.topic === topic),
  )
  const groups = useMemo(() => {
    const g = new Map<string, Mistake[]>()
    for (const m of filtered) {
      if (!g.has(m.topic)) g.set(m.topic, [])
      g.get(m.topic)!.push(m)
    }
    return [...g.entries()].sort((a, b) => a[0].localeCompare(b[0]))
  }, [filtered])

  const setStatus = (m: Mistake, status: 'active' | 'mastered') => {
    setBusy(m.id)
    post<Mistake>(`/api/mistakes/${m.id}/status`, { status })
      .then((updated) => {
        setSelected((s) => (s && s.id === updated.id ? updated : s))
        load()
      })
      .catch((e) => setErr(String(e.message ?? e)))
      .finally(() => setBusy(''))
  }

  const startDrill = () => {
    const list = filtered.filter((m) => m.status === 'active')
    if (!list.length) return
    setDrill(list)
    setIdx(0)
    setRevealed(false)
  }
  const exitDrill = () => {
    setDrill(null)
    setRevealed(false)
  }
  // 重练中就地更新该题状态，避免整页重载打断练习
  const drillSetStatus = (m: Mistake, status: 'active' | 'mastered') => {
    if (!drill) return
    setBusy(m.id)
    post<Mistake>(`/api/mistakes/${m.id}/status`, { status })
      .then((updated) => {
        setDrill(drill.map((d) => (d.id === updated.id ? updated : d)))
        load()
      })
      .catch((e) => setErr(String(e.message ?? e)))
      .finally(() => setBusy(''))
  }

  if (err) return <div className="p-8 text-error">{err}</div>
  if (!data) return <div className="p-8 text-sm opacity-50">加载中…</div>

  // ---------- 重练模式 ----------
  if (drill) {
    const m = drill[Math.min(idx, drill.length - 1)]
    const last = idx >= drill.length - 1
    return (
      <div className="mx-auto max-w-2xl p-8">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">错题重练</h1>
          <button className="btn btn-ghost btn-sm gap-1" onClick={exitDrill}>
            <X className="h-4 w-4" /> 退出重练
          </button>
        </div>
        <p className="mt-1 text-sm opacity-60">
          纯回忆练习：不写复习日志、不影响 FSRS 调度。
        </p>
        <div className="mt-2 text-xs opacity-50 tabular-nums">
          第 {idx + 1} / {drill.length} 题
        </div>
        <progress className="progress progress-primary mt-1 h-1.5 w-full" value={idx + 1} max={drill.length} />

        <div className="card mt-4 bg-base-100 shadow-sm">
          <div className="card-body p-5">
            <div className="flex items-center gap-2 text-xs">
              <span className="badge badge-ghost badge-sm">{topicName[m.topic] ?? m.topic}</span>
              <SourceBadge m={m} />
              {m.knowledge && <span className="text-info/70">知识点：{m.knowledge}</span>}
              <span className="ml-auto opacity-40">{fmtDate(m.created)}</span>
            </div>
            <p className="mt-3 whitespace-pre-wrap text-sm font-medium leading-relaxed">{m.question}</p>
            {m.hint && <HintToggle key={m.id} hint={m.hint} />}
            <MistakeImages images={mistakeImages(m)} onView={setViewer} />

            {!revealed ? (
              <button className="btn btn-primary btn-sm mt-5 gap-1 self-start" onClick={() => setRevealed(true)}>
                <Eye className="h-4 w-4" /> 显示答案
              </button>
            ) : (
              <div className="mt-5 flex flex-col gap-3 border-t border-base-200 pt-4 text-sm">
                <div>
                  <div className="text-xs font-semibold opacity-50">参考答案</div>
                  <p className="mt-1 whitespace-pre-wrap leading-relaxed">{m.answer}</p>
                </div>
                {m.my_answer && (
                  <div className="rounded-lg bg-error/5 p-3">
                    <div className="text-xs font-semibold text-error/70">我当时的答案</div>
                    <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-80">{m.my_answer}</p>
                  </div>
                )}
                {m.analysis && (
                  <div className="rounded-lg bg-info/5 p-3">
                    <div className="text-xs font-semibold text-info/70">错因分析</div>
                    <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-80">{m.analysis}</p>
                  </div>
                )}
                <MistakeAnalysis m={m} />
              </div>
            )}
          </div>
        </div>

        <div className="mt-4 flex items-center gap-2">
          <button
            className="btn btn-ghost btn-sm gap-1"
            disabled={idx === 0}
            onClick={() => {
              setIdx(idx - 1)
              setRevealed(false)
            }}
          >
            <ArrowLeft className="h-4 w-4" /> 上一题
          </button>
          <button
            className="btn btn-ghost btn-sm gap-1"
            disabled={last}
            onClick={() => {
              setIdx(idx + 1)
              setRevealed(false)
            }}
          >
            下一题 <ArrowRight className="h-4 w-4" />
          </button>
          <div className="ml-auto flex items-center gap-2">
            {m.status === 'active' && (
              <button
                className="btn btn-success btn-sm gap-1"
                disabled={busy === m.id}
                onClick={() => drillSetStatus(m, 'mastered')}
              >
                <CheckCircle2 className="h-4 w-4" /> 标记攻克
              </button>
            )}
            {last ? (
              <button className="btn btn-primary btn-sm" onClick={exitDrill}>
                完成，返回列表
              </button>
            ) : null}
          </div>
        </div>
        {viewer && <ImageLightbox url={viewer} onClose={() => setViewer(null)} />}
      </div>
    )
  }

  // ---------- 列表模式 ----------
  const activeCount = data.summary?.active ?? mistakes.filter((m) => m.status === 'active').length
  const masteredCount = data.summary?.mastered ?? mistakes.filter((m) => m.status === 'mastered').length
  const drillable = filtered.filter((m) => m.status === 'active').length

  return (
    <div className="mx-auto max-w-3xl p-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold">错题库</h1>
          <p className="mt-1 text-sm opacity-60">
            quiz 判错与图片录入的错题集中在这里 —— 逐个攻克，别让它们过夜。
          </p>
        </div>
        <button className="btn btn-primary btn-sm gap-1" disabled={drillable === 0} onClick={startDrill}>
          <Play className="h-4 w-4" /> 开始重练{topic ? `（${topicName[topic] ?? topic}）` : ''}
        </button>
      </div>

      <div className="stats mt-5 w-full bg-base-100 shadow-sm">
        <div className="stat py-3">
          <div className="stat-title text-xs">待攻克</div>
          <div className="stat-value text-2xl tabular-nums text-error">{activeCount}</div>
        </div>
        <div className="stat py-3">
          <div className="stat-title text-xs">已攻克</div>
          <div className="stat-value text-2xl tabular-nums text-success">{masteredCount}</div>
        </div>
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="tabs tabs-boxed w-fit bg-base-200/60">
          {(['active', 'mastered', 'all'] as Filter[]).map((f) => (
            <button key={f} className={`tab ${filter === f ? 'tab-active' : ''}`} onClick={() => setFilter(f)}>
              {filterLabel[f]}
            </button>
          ))}
        </div>
        <select
          className="select select-bordered select-sm w-44"
          value={topic}
          onChange={(e) => setTopic(e.target.value)}
        >
          <option value="">全部主题</option>
          {topicOptions.map((t) => (
            <option key={t} value={t}>
              {topicName[t] ?? t}
            </option>
          ))}
        </select>
      </div>

      {mistakes.length === 0 ? (
        <div className="mt-8 rounded-xl bg-base-200/50 p-6 text-sm opacity-60">
          暂无错题。quiz 判错或把错题图片发给 Agent 后会出现在这里。
        </div>
      ) : groups.length === 0 ? (
        <div className="mt-8 rounded-xl bg-base-200/50 p-6 text-sm opacity-60">
          「{topic ? `${topicName[topic] ?? topic} · ` : ''}{filterLabel[filter]}」下没有错题。
        </div>
      ) : (
        groups.map(([topic, list]) => (
          <section key={topic} className="mt-6">
            <h2 className="flex items-center gap-2 text-sm font-semibold">
              {topicName[topic] ?? topic}
              <span className="badge badge-ghost badge-sm">{list.length}</span>
            </h2>
            <ul className="mt-2 flex flex-col gap-2">
              {list.map((m) => {
                const imgs = mistakeImages(m)
                return (
                  <li key={m.id} className="card bg-base-100 shadow-sm">
                    <div
                      className="cursor-pointer p-4 transition-colors hover:bg-base-200/40"
                      onClick={() => setSelected(m)}
                    >
                      <div className="flex items-start gap-2">
                        <span className="min-w-0 flex-1 whitespace-pre-wrap text-sm leading-relaxed">
                          {m.question}
                        </span>
                        <ChevronRight className="mt-0.5 h-4 w-4 shrink-0 opacity-50" />
                      </div>
                      <div className="mt-2 flex flex-wrap items-center gap-2 text-xs">
                        <SourceBadge m={m} />
                        {m.knowledge && <span className="text-info/70">知识点：{m.knowledge}</span>}
                        <span className={`badge badge-sm ${m.status === 'active' ? 'badge-error badge-outline' : 'badge-success badge-outline'}`}>
                          {m.status === 'active' ? '待攻克' : '已攻克'}
                        </span>
                        {imgs.length > 0 && (
                          <span className="badge badge-ghost badge-sm gap-1">
                            <ImageIcon className="h-3 w-3" /> 图 {imgs.length}
                          </span>
                        )}
                        <span className="opacity-40">{fmtDate(m.created)}</span>
                        <button
                          className={`btn btn-xs ml-auto gap-1 ${m.status === 'active' ? 'btn-success btn-outline' : 'btn-ghost'}`}
                          disabled={busy === m.id}
                          onClick={(e) => {
                            e.stopPropagation()
                            setStatus(m, m.status === 'active' ? 'mastered' : 'active')
                          }}
                        >
                          {m.status === 'active' ? (
                            <>
                              <CheckCircle2 className="h-3.5 w-3.5" /> 标记攻克
                            </>
                          ) : (
                            <>
                              <RotateCcw className="h-3.5 w-3.5" /> 重新激活
                            </>
                          )}
                        </button>
                      </div>
                    </div>
                  </li>
                )
              })}
            </ul>
          </section>
        ))
      )}
      {selected && (
        <>
          <div className="fixed inset-0 z-40 bg-black/30" onClick={() => setSelected(null)} />
          <aside className="fixed right-0 top-0 z-40 flex h-full w-1/2 min-w-[24rem] flex-col border-l border-base-300 bg-base-100 shadow-2xl">
            <div className="flex items-center gap-2 border-b border-base-200 px-6 py-4 text-xs">
              <span className="badge badge-ghost badge-sm">{topicName[selected.topic] ?? selected.topic}</span>
              <SourceBadge m={selected} />
              {selected.knowledge && <span className="text-info/70">知识点：{selected.knowledge}</span>}
              <span className={`badge badge-sm ${selected.status === 'active' ? 'badge-error badge-outline' : 'badge-success badge-outline'}`}>
                {selected.status === 'active' ? '待攻克' : '已攻克'}
              </span>
              <span className="opacity-40">{fmtDate(selected.created)}</span>
              <button className="btn btn-circle btn-ghost btn-sm ml-auto" onClick={() => setSelected(null)}>
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto px-6 py-5">
              <div className="flex flex-col gap-3 text-sm">
                <p className="whitespace-pre-wrap font-medium leading-relaxed">{selected.question}</p>
                {selected.hint && <HintToggle key={selected.id} hint={selected.hint} />}
                <MistakeImages images={mistakeImages(selected)} onView={setViewer} />
                <div className="border-t border-base-200 pt-3">
                  <div className="text-xs font-semibold opacity-50">参考答案</div>
                  <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-85">{selected.answer}</p>
                </div>
                {selected.my_answer && (
                  <div className="rounded-lg bg-error/5 p-3">
                    <div className="text-xs font-semibold text-error/70">我当时的答案</div>
                    <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-80">{selected.my_answer}</p>
                  </div>
                )}
                {selected.analysis && (
                  <div className="rounded-lg bg-info/5 p-3">
                    <div className="text-xs font-semibold text-info/70">错因分析</div>
                    <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-80">{selected.analysis}</p>
                  </div>
                )}
                <MistakeAnalysis m={selected} />
              </div>
            </div>
            <div className="border-t border-base-200 px-6 py-3">
              <button
                className={`btn btn-sm gap-1 ${selected.status === 'active' ? 'btn-success' : 'btn-ghost'}`}
                disabled={busy === selected.id}
                onClick={() => setStatus(selected, selected.status === 'active' ? 'mastered' : 'active')}
              >
                {selected.status === 'active' ? (
                  <>
                    <CheckCircle2 className="h-4 w-4" /> 标记攻克
                  </>
                ) : (
                  <>
                    <RotateCcw className="h-4 w-4" /> 重新激活
                  </>
                )}
              </button>
            </div>
          </aside>
        </>
      )}
      {viewer && <ImageLightbox url={viewer} onClose={() => setViewer(null)} />}
    </div>
  )
}

function SourceBadge({ m }: { m: Mistake }) {
  if (m.source === 'quiz')
    return (
      <span className="badge badge-ghost badge-sm gap-1" title={m.session ?? ''}>
        <ScrollText className="h-3 w-3" /> quiz{m.session ? ` · ${m.session}` : ''}
      </span>
    )
  return (
    <span className="badge badge-ghost badge-sm gap-1">
      <ImageIcon className="h-3 w-3" /> image
    </span>
  )
}

function assetUrl(image: string): string {
  const name = image.replace(/^assets\//, '')
  return `/api/mistakes/asset?name=${encodeURIComponent(name)}`
}

// 一题可多图（题目图、答案图等）：image 单图字段与 images 列表字段合并
function mistakeImages(m: Mistake): string[] {
  return [...new Set([...(m.image ? [m.image] : []), ...(m.images ?? [])])]
}

function MistakeImages({ images, onView }: { images: string[]; onView: (url: string) => void }) {
  if (!images.length) return null
  return (
    <div className="mt-2 flex flex-wrap gap-3">
      {images.map((image) => {
        const url = assetUrl(image)
        return (
          <button key={image} className="group block w-fit cursor-zoom-in" title="查看原图" onClick={() => onView(url)}>
            <img
              src={url}
              alt="错题原图"
              loading="lazy"
              className="max-h-28 rounded-lg border border-base-300 object-contain transition-opacity group-hover:opacity-80"
            />
            <span className="mt-1 flex items-center gap-1 text-xs opacity-40 group-hover:opacity-70">
              <ImageIcon className="h-3 w-3" /> 查看原图
            </span>
          </button>
        )
      })}
    </div>
  )
}

function ImageLightbox({ url, onClose }: { url: string; onClose: () => void }) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-6"
      onClick={onClose}
    >
      <button className="btn btn-circle btn-ghost btn-sm absolute right-4 top-4 text-white" onClick={onClose}>
        <X className="h-5 w-5" />
      </button>
      <img
        src={url}
        alt="错题原图"
        className="max-h-full max-w-full rounded-lg object-contain"
        onClick={(e) => e.stopPropagation()}
      />
    </div>
  )
}

// 提示默认收起，点「查看提示」才展开（按错题 id 作 key，切题自动重置）
function HintToggle({ hint }: { hint: string }) {
  const [shown, setShown] = useState(false)
  if (!shown)
    return (
      <button
        className="mt-2 flex items-center gap-1 text-xs text-primary/70 transition-colors hover:text-primary"
        onClick={() => setShown(true)}
      >
        <Lightbulb className="h-3 w-3" /> 查看提示
      </button>
    )
  return (
    <p className="mt-2 flex items-center gap-1 text-xs opacity-50">
      <Lightbulb className="h-3 w-3" /> {hint}
    </p>
  )
}

// W9 录入的分析块：解题思路 / 逐选项分析 / 易错点 / 同考点真题
function MistakeAnalysis({ m }: { m: Mistake }) {
  if (!m.solution && !m.option_analysis && !m.pitfalls && !m.related?.length) return null
  return (
    <>
      {m.solution && (
        <div>
          <div className="text-xs font-semibold opacity-50">解题思路</div>
          <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-85">{m.solution}</p>
        </div>
      )}
      {m.option_analysis && (
        <div>
          <div className="text-xs font-semibold opacity-50">逐选项分析</div>
          <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-85">{m.option_analysis}</p>
        </div>
      )}
      {m.pitfalls && (
        <div className="rounded-lg bg-warning/5 p-3">
          <div className="text-xs font-semibold text-warning/70">易错点</div>
          <p className="mt-1 whitespace-pre-wrap leading-relaxed opacity-80">{m.pitfalls}</p>
        </div>
      )}
      {m.related && m.related.length > 0 && (
        <div>
          <div className="text-xs font-semibold opacity-50">同考点真题</div>
          <ul className="mt-1 flex flex-col gap-1.5">
            {m.related.map((r, i) => (
              <li key={i} className="text-xs leading-relaxed">
                <a
                  href={r.url}
                  target="_blank"
                  rel="noreferrer"
                  className="link link-primary inline-flex items-center gap-1 font-medium"
                >
                  {r.title} <ExternalLink className="h-3 w-3" />
                </a>
                {r.source && <span className="opacity-40">（{r.source}）</span>}
                {r.note && <p className="opacity-60">{r.note}</p>}
              </li>
            ))}
          </ul>
        </div>
      )}
    </>
  )
}
