import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { CalendarCheck, Layers, Play, TrendingUp } from 'lucide-react'
import { get, type Stats } from '../api'
import Heatmap from '../components/Heatmap'

const sessionTypeLabel: Record<string, string> = {
  tutor: '导师',
  feynman: '费曼',
  quiz: '自测',
  diagnose: '诊断',
  import: '导入',
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    get<Stats>('/api/stats').then(setStats).catch((e) => setErr(String(e.message ?? e)))
  }, [])

  if (err) return <div className="p-8 text-error">{err}</div>
  if (!stats) return <div className="p-8 opacity-60">加载中…</div>

  const cards = [
    { label: '今日到期', value: stats.due_now, icon: CalendarCheck, to: '/review', cta: '开始复习' },
    { label: '卡片总数', value: stats.total_cards, icon: Layers, to: '/notes', cta: '看笔记' },
    { label: '今日已复习', value: stats.reviews_today, icon: TrendingUp, to: '/mastery', cta: '掌握度' },
    { label: '连续天数', value: stats.streak, icon: CalendarCheck, to: '/sessions', cta: '会话记录' },
  ]

  return (
    <div className="mx-auto max-w-4xl p-8">
      <h1 className="text-xl font-bold">仪表盘</h1>
      <p className="mt-1 text-sm opacity-60">
        调度在这里，导师在 harness 里 —— 需要理解类帮助时去「工作流」页复制命令。
      </p>

      <div className="mt-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        {cards.map(({ label, value, icon: Icon, to, cta }) => (
          <Link key={label} to={to} className="card bg-base-100 shadow-sm transition-shadow hover:shadow-md">
            <div className="card-body gap-1 p-5">
              <div className="flex items-center justify-between text-xs opacity-60">
                {label}
                <Icon className="h-4 w-4" />
              </div>
              <div className="text-3xl font-bold tabular-nums">{value}</div>
              <div className="mt-1 text-[11px] text-primary">{cta} →</div>
            </div>
          </Link>
        ))}
      </div>

      <div className="card mt-6 bg-base-100 shadow-sm">
        <div className="card-body p-5">
          <h2 className="card-title text-sm">近 90 天复习热力图</h2>
          <div className="mt-2">
            <Heatmap data={stats.heatmap} />
          </div>
        </div>
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-2">
        <div className="card bg-base-100 shadow-sm">
          <div className="card-body p-5">
            <h2 className="card-title text-sm">掌握度</h2>
            {Object.keys(stats.mastery).length === 0 ? (
              <p className="mt-2 text-sm opacity-55">
                还没有主题。去{' '}
                <Link className="link link-primary" to="/library">
                  资料库
                </Link>{' '}
                上传材料，然后在 harness 里执行导入。
              </p>
            ) : (
              <ul className="mt-2 flex flex-col gap-2">
                {Object.entries(stats.mastery).map(([topic, m]) => (
                  <li key={topic} className="flex items-center gap-3 text-sm">
                    <span className="w-36 truncate font-medium">{topic}</span>
                    <progress
                      className="progress progress-primary h-2 flex-1"
                      value={(m.level ?? 0) * 20}
                      max={100}
                    />
                    <span className="w-12 text-right tabular-nums opacity-70">L{m.level ?? 0}/5</span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        <div className="card bg-base-100 shadow-sm">
          <div className="card-body p-5">
            <div className="flex items-center justify-between">
              <h2 className="card-title text-sm">最近会话</h2>
              <Link to="/sessions" className="link link-primary text-xs">
                全部 →
              </Link>
            </div>
            {!stats.recent_sessions?.length ? (
              <p className="mt-2 text-sm opacity-55">
                还没有会话。导入/导师/费曼/自测/诊断的记录都会出现在这里。
              </p>
            ) : (
              <ul className="mt-2 flex flex-col divide-y divide-base-200 text-sm">
                {stats.recent_sessions.map((s) => (
                  <li key={s.id} className="flex items-center gap-2 py-2">
                    <span className="badge badge-ghost badge-sm shrink-0">
                      {sessionTypeLabel[s.type] ?? s.type}
                    </span>
                    <span className="w-32 truncate">{s.topic || '—'}</span>
                    <span className="flex-1 truncate opacity-60">{s.summary}</span>
                    <span className="shrink-0 text-xs opacity-40">{s.date}</span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </div>

      {stats.due_now > 0 && (
        <Link
          to="/review"
          className="btn btn-primary mt-8 w-full gap-2 text-base font-normal shadow"
        >
          <Play className="h-4 w-4" /> 开始复习（{stats.due_now} 张到期）
        </Link>
      )}
    </div>
  )
}
