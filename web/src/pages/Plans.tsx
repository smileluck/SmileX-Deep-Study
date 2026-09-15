import { useEffect, useState } from 'react'
import { Globe, Map as MapIcon } from 'lucide-react'
import { get, type Plan, type PlansResp } from '../api'
import Markdown from '../components/Markdown'

const statusBadge: Record<string, string> = {
  active: 'badge-primary',
  done: 'badge-success',
  paused: 'badge-ghost',
}

function milestones(body: string): { done: number; total: number } {
  const done = body.match(/- \[[xX]\]/g)?.length ?? 0
  const todo = body.match(/- \[ \]/g)?.length ?? 0
  return { done, total: done + todo }
}

const GLOBAL = '__global__'

export default function Plans() {
  const [data, setData] = useState<PlansResp | null>(null)
  const [tab, setTab] = useState<string>(GLOBAL)

  useEffect(() => {
    get<PlansResp>('/api/plans').then(setData).catch(() => setData({ master: null, topics: [] }))
  }, [])

  if (!data) return <div className="p-8 text-sm opacity-50">加载中…</div>

  const empty = !data.master && data.topics.length === 0

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

  // 全局进度 = 各主题计划里程碑的聚合（master.md 自身不维护 checkbox）
  const agg = data.topics.reduce(
    (acc, p) => {
      const ms = milestones(p.body)
      return { done: acc.done + ms.done, total: acc.total + ms.total }
    },
    { done: 0, total: 0 },
  )

  const hasMaster = !!data.master
  const activeTab = tab === GLOBAL && !hasMaster ? (data.topics[0]?.slug ?? GLOBAL) : tab
  const activePlan = data.topics.find((p) => p.slug === activeTab)

  // 全局计划正文里的 [主题](topics/<slug>/plan.md) 链接 → 切到对应 tab
  const masterLinkComponents = {
    a: ({ href, children }: { href?: string; children?: React.ReactNode }) => {
      const m = href?.match(/topics\/([^/]+)\/plan\.md/)
      if (m && data.topics.some((p) => p.slug === m[1])) {
        return (
          <button className="link link-primary" onClick={() => setTab(m[1])}>
            {children}
          </button>
        )
      }
      return <a href={href}>{children}</a>
    },
  }

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
        {data.topics.map((p) => {
          const ms = milestones(p.body)
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

      {activeTab === GLOBAL && data.master && (
        <div className="card mt-4 bg-base-100 shadow-sm">
          <div className="card-body gap-3 p-5">
            <div className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <MapIcon className="h-4 w-4" />
              </div>
              <div className="font-semibold">{String(data.master.fm.title ?? '全局学习计划')}</div>
              <span className="ml-auto text-xs opacity-50">
                更新于 {String(data.master.fm.updated ?? '—')}
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

            {data.topics.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {data.topics.map((p) => {
                  const ms = milestones(p.body)
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

            <Markdown components={masterLinkComponents}>{data.master.body}</Markdown>
          </div>
        </div>
      )}

      {activePlan && (
        <TopicPlanCard plan={activePlan} />
      )}
    </div>
  )
}

function TopicPlanCard({ plan: p }: { plan: Plan }) {
  const ms = milestones(p.body)
  const status = String(p.fm.status ?? '')
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
          目标：{String(p.fm.goal ?? '—')}
          {p.fm.horizon ? <span className="opacity-60"> · 周期：{String(p.fm.horizon)}</span> : null}
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
