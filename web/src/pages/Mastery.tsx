import { useEffect, useMemo, useState } from 'react'
import { get, type MasteryResp } from '../api'
import { evidenceKindLabel } from '../labels'

const levelDesc = ['未接触', '有印象', '能复述要点', '能应用', '能关联迁移', '能讲授他人']

export default function Mastery() {
  const [data, setData] = useState<MasteryResp | null>(null)
  const [selected, setSelected] = useState<string | null>(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    get<MasteryResp>('/api/mastery')
      .then(setData)
      .catch((e) => setErr(String(e.message ?? e)))
  }, [])

  const topics = Object.keys(data?.mastery ?? {})
  const slugs = [...new Set([...topics, ...Object.keys(data?.per_topic ?? {})])]
  const current = selected ?? slugs[0]
  const m = data?.mastery[current]
  const stat = data?.per_topic[current] ?? { cards: 0, due: 0, new: 0, reviews: 0, again: 0 }
  const againRate = stat.reviews > 0 ? Math.round((stat.again / stat.reviews) * 100) : 0
  const evidence = useMemo(() => [...(m?.evidence ?? [])].reverse(), [m])

  if (err) return <div className="p-8 text-error">{err}</div>
  if (!data) return <div className="p-8 text-sm opacity-50">加载中…</div>

  return (
    <div className="mx-auto max-w-4xl p-8">
      <h1 className="text-xl font-bold">掌握度</h1>
      <p className="mt-1 text-sm opacity-60">
        Bloom 精熟模型：0-5 级 + 证据链。每次升降级都必须有自测/费曼/诊断的证据。
      </p>

      {slugs.length === 0 ? (
        <div className="mt-8 rounded-xl bg-base-200/50 p-6 text-sm opacity-60">
          还没有掌握度数据 —— 导入材料后由 harness 初始化，自测/费曼/诊断工作流持续更新。
        </div>
      ) : (
        <>
          <div className="card mt-6 bg-base-100 shadow-sm">
            <div className="card-body p-5">
              <table className="table table-sm">
                <thead>
                  <tr className="text-xs opacity-60">
                    <th>主题</th>
                    <th>等级</th>
                    <th className="text-right">卡片</th>
                    <th className="text-right">到期</th>
                    <th className="text-right">Again 率</th>
                  </tr>
                </thead>
                <tbody>
                  {slugs.map((s) => {
                    const mm = data.mastery[s]
                    const st = data.per_topic[s]
                    const rate = st && st.reviews > 0 ? Math.round((st.again / st.reviews) * 100) : 0
                    return (
                      <tr
                        key={s}
                        className={`cursor-pointer transition-colors hover:bg-base-200/60 ${current === s ? 'bg-primary/5' : ''}`}
                        onClick={() => setSelected(s)}
                      >
                        <td className="font-mono text-xs">{s}</td>
                        <td>
                          <span className="badge badge-primary badge-sm">L{mm?.level ?? 0}</span>
                          <span className="ml-2 text-xs opacity-50">{levelDesc[mm?.level ?? 0]}</span>
                        </td>
                        <td className="text-right tabular-nums">{st?.cards ?? 0}</td>
                        <td className="text-right tabular-nums">{st?.due ?? 0}</td>
                        <td className={`text-right tabular-nums ${rate >= 25 ? 'font-semibold text-error' : ''}`}>
                          {st?.reviews ? `${rate}%` : '—'}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>

          {current && (
            <div className="card mt-4 bg-base-100 shadow-sm">
              <div className="card-body p-5">
                <div className="flex items-center justify-between">
                  <h2 className="font-semibold">
                    {current} · 证据链
                  </h2>
                  <span className="text-xs opacity-50">更新于 {m?.updated ?? '—'}</span>
                </div>
                <div className="mt-1 flex items-center gap-3">
                  <progress className="progress progress-primary h-2.5 flex-1" value={(m?.level ?? 0) * 20} max={100} />
                  <span className="text-sm font-bold tabular-nums">L{m?.level ?? 0}</span>
                </div>
                <div className="text-xs opacity-50">{levelDesc[m?.level ?? 0]}</div>

                {evidence.length === 0 ? (
                  <p className="mt-3 text-sm opacity-55">暂无证据记录。</p>
                ) : (
                  <ul className="mt-3 flex flex-col">
                    {evidence.map((e) => (
                      <li key={`${e.date}-${e.kind}-${e.detail}`} className="flex items-center gap-3 border-b border-base-200 py-2 text-sm last:border-none">
                        <span className="w-24 shrink-0 font-mono text-xs opacity-50">{e.date}</span>
                        <span className="badge badge-ghost badge-sm shrink-0">{evidenceKindLabel[e.kind] ?? e.kind}</span>
                        <span className="min-w-0 flex-1 truncate opacity-75">{e.detail}</span>
                        <span
                          className={`w-8 shrink-0 text-right font-bold tabular-nums ${
                            (e.delta ?? 0) > 0 ? 'text-success' : (e.delta ?? 0) < 0 ? 'text-error' : 'opacity-40'
                          }`}
                        >
                          {(e.delta ?? 0) > 0 ? `+${e.delta}` : e.delta ?? 0}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}

                <div className="mt-4 grid grid-cols-4 gap-2 text-center text-xs">
                  <div className="rounded-lg bg-base-200/60 p-2">
                    <div className="font-bold tabular-nums">{stat.cards}</div>
                    <div className="opacity-50">卡片</div>
                  </div>
                  <div className="rounded-lg bg-base-200/60 p-2">
                    <div className="font-bold tabular-nums">{stat.due}</div>
                    <div className="opacity-50">当前到期</div>
                  </div>
                  <div className="rounded-lg bg-base-200/60 p-2">
                    <div className="font-bold tabular-nums">{stat.reviews}</div>
                    <div className="opacity-50">累计复习</div>
                  </div>
                  <div className={`rounded-lg p-2 ${againRate >= 25 ? 'bg-error/10' : 'bg-base-200/60'}`}>
                    <div className={`font-bold tabular-nums ${againRate >= 25 ? 'text-error' : ''}`}>{stat.reviews ? `${againRate}%` : '—'}</div>
                    <div className="opacity-50">Again 率</div>
                  </div>
                </div>
                {againRate >= 25 && stat.reviews >= 5 && (
                  <p className="mt-3 text-xs text-error">
                    Again 率偏高 —— 建议在 harness 里跑一次 /study:diagnose {current} 做弱点归因。
                  </p>
                )}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
