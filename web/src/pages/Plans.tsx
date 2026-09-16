import { useEffect, useMemo, useState } from 'react'
import { Globe, Map as MapIcon } from 'lucide-react'
import { get, type Plan, type PlansResp } from '../api'
import Markdown from '../components/Markdown'

const statusBadge: Record<string, string> = {
  active: 'badge-primary',
  done: 'badge-success',
  paused: 'badge-ghost',
}

// 只统计「## 里程碑」段内（到下一个 ## 标题为止）的 checkbox，周计划里的任务 checkbox 不算里程碑
function milestones(body: string): { done: number; total: number } {
  let inSection = false
  const lines: string[] = []
  for (const line of body.split('\n')) {
    if (line.startsWith('## ')) {
      if (inSection) break
      inSection = line.slice(3).trim() === '里程碑'
      continue
    }
    if (inSection) lines.push(line)
  }
  const section = lines.join('\n')
  const done = section.match(/- \[[xX]\]/g)?.length ?? 0
  const todo = section.match(/- \[ \]/g)?.length ?? 0
  return { done, total: done + todo }
}

const GLOBAL = '__global__'

export default function Plans() {
  const [data, setData] = useState<PlansResp | null>(null)
  const [tab, setTab] = useState<string>(GLOBAL)
  const [err, setErr] = useState('')

  useEffect(() => {
    get<PlansResp>('/api/plans')
      .then(setData)
      .catch((e) => setErr(String(e.message ?? e)))
  }, [])

  const topics = useMemo(() => data?.topics ?? [], [data])
  const master = data?.master ?? null

  // 每个主题计划的里程碑统计只解析一次（tab 徽标 / 全局聚合 / 进度条共用）
  const msBySlug = useMemo(() => {
    const m = new Map<string, { done: number; total: number }>()
    for (const p of topics) m.set(p.slug, milestones(p.body))
    return m
  }, [topics])

  // 全局进度 = 各主题计划里程碑的聚合（master.md 自身不维护 checkbox）
  const agg = useMemo(
    () =>
      [...msBySlug.values()].reduce(
        (acc, ms) => ({ done: acc.done + ms.done, total: acc.total + ms.total }),
        { done: 0, total: 0 },
      ),
    [msBySlug],
  )

  // 全局计划正文里的 [主题](topics/<slug>/plan.md) 链接 → 切到对应 tab
  const masterLinkComponents = useMemo(
    () => ({
      a: ({ href, children }: { href?: string; children?: React.ReactNode }) => {
        const m = href?.match(/topics\/([^/]+)\/plan\.md/)
        if (m && topics.some((p) => p.slug === m[1])) {
          return (
            <button className="link link-primary" onClick={() => setTab(m[1])}>
              {children}
            </button>
          )
        }
        return <a href={href}>{children}</a>
      },
    }),
    [topics],
  )

  if (err) return <div className="p-8 text-sm text-error">{err}</div>
  if (!data) return <div className="p-8 text-sm opacity-50">加载中…</div>

  const empty = !master && topics.length === 0

  if (empty) {
    return (
      <div className="mx-auto max-w-4xl p-8">
        <h1 className="text-xl font-bold">计划</h1>
        <div className="mt-8 rounded-xl bg-base-200/50 p-6 text-sm opacity-60">
          还没有学习计划 —— 在 harness 中运行 <code>/study:plan</code> 生成学习计划（可带 topic 生成单主题路线图）。
        </div>
      </div>
    )
  }

  const hasMaster = !!master
  const activeTab = tab === GLOBAL && !hasMaster ? (topics[0]?.slug ?? GLOBAL) : tab
  const activePlan = topics.find((p) => p.slug === activeTab)

  return (
    <div className="mx-auto max-w-4xl p-8">
      <h1 className="text-xl font-bold">计划</h1>
      <p className="mt-1 text-sm opacity-60">
        W7 学习规划产出：全局学习计划 + 各主题路线图，里程碑进度实时从 plan.md 读取。
      </p>

      <div role="tablist" className="tabs tabs-boxed mt-6 w-fit bg-base-200/70">
        {hasMaster && (
          <button
            role="tab"
            className={`tab gap-1.5 ${activeTab === GLOBAL ? 'tab-active' : ''}`}
            onClick={() => setTab(GLOBAL)}
          >
            <Globe className="h-3.5 w-3.5" />
            全局
          </button>
        )}
        {topics.map((p) => {
          const ms = msBySlug.get(p.slug) ?? { done: 0, total: 0 }
          return (
            <button
              key={p.slug}
              role="tab"
              className={`tab gap-1.5 ${activeTab === p.slug ? 'tab-active' : ''}`}
              onClick={() => setTab(p.slug)}
            >
              {p.topic_name}
              {ms.total > 0 && (
                <span className="badge badge-xs badge-ghost tabular-nums">
                  {ms.done}/{ms.total}
                </span>
              )}
            </button>
          )
        })}
      </div>

      {activeTab === GLOBAL && master && (
        <div className="card mt-4 bg-base-100 shadow-sm">
          <div className="card-body gap-3 p-5">
            <div className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <MapIcon className="h-4 w-4" />
              </div>
              <div className="font-semibold">{master.fm.title ?? '全局学习计划'}</div>
              <span className="ml-auto text-xs opacity-50">
                更新于 {master.fm.updated ?? '—'}
              </span>
            </div>

            {agg.total > 0 && (
              <div className="flex items-center gap-3">
                <progress
                  className="progress progress-primary h-2.5 flex-1"
                  value={agg.done}
                  max={agg.total}
                />
                <span className="text-xs tabular-nums opacity-60">
                  总里程碑 {agg.done}/{agg.total}
                </span>
              </div>
            )}

            {topics.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {topics.map((p) => {
                  const ms = msBySlug.get(p.slug) ?? { done: 0, total: 0 }
                  return (
                    <button key={p.slug} className="btn btn-outline btn-xs" onClick={() => setTab(p.slug)}>
                      {p.topic_name}
                      {ms.total > 0 && (
                        <span className="tabular-nums opacity-60">
                          {ms.done}/{ms.total}
                        </span>
                      )}
                    </button>
                  )
                })}
              </div>
            )}

            <Markdown components={masterLinkComponents}>{master.body}</Markdown>
          </div>
        </div>
      )}

      {activePlan && (
        <TopicPlanCard plan={activePlan} ms={msBySlug.get(activePlan.slug) ?? { done: 0, total: 0 }} />
      )}
    </div>
  )
}

function TopicPlanCard({ plan: p, ms }: { plan: Plan; ms: { done: number; total: number } }) {
  const status = p.fm.status ?? ''
  return (
    <div className="card mt-4 bg-base-100 shadow-sm">
      <div className="card-body gap-3 p-5">
        <div className="flex items-start justify-between gap-2">
          <div>
            <div className="font-semibold">{p.topic_name}</div>
            <div className="font-mono text-xs opacity-50">{p.slug}</div>
          </div>
          <span className={`badge badge-sm shrink-0 ${statusBadge[status] ?? 'badge-ghost'}`}>
            {status || '—'}
          </span>
        </div>
        <p className="text-[13px] leading-relaxed opacity-70">
          目标：{p.fm.goal ?? '—'}
          {p.fm.horizon ? <span className="opacity-60"> · 周期：{p.fm.horizon}</span> : null}
        </p>
        {ms.total > 0 && (
          <div className="flex items-center gap-3">
            <progress
              className="progress progress-primary h-2.5 flex-1"
              value={ms.done}
              max={ms.total}
            />
            <span className="text-xs tabular-nums opacity-60">
              里程碑 {ms.done}/{ms.total}
            </span>
          </div>
        )}
        <Markdown>{p.body}</Markdown>
      </div>
    </div>
  )
}
