import { useEffect, useState } from 'react'
import { NavLink, Route, Routes } from 'react-router-dom'
import {
  BookOpenText,
  FlaskConical,
  FolderOpen,
  Gauge,
  GraduationCap,
  NotebookPen,
  Repeat,
  ScrollText,
  Target,
} from 'lucide-react'
import { get, type Stats } from './api'
import Dashboard from './pages/Dashboard'
import Library from './pages/Library'
import Notes from './pages/Notes'
import Review from './pages/Review'
import Sessions from './pages/Sessions'
import Workflows from './pages/Workflows'
import Mastery from './pages/Mastery'

const nav = [
  { to: '/', label: '仪表盘', icon: Gauge, end: true },
  { to: '/review', label: '复习', icon: Repeat, end: false },
  { to: '/library', label: '资料库', icon: FolderOpen, end: false },
  { to: '/notes', label: '笔记', icon: NotebookPen, end: false },
  { to: '/sessions', label: '会话', icon: ScrollText, end: false },
  { to: '/workflows', label: '工作流', icon: FlaskConical, end: false },
  { to: '/mastery', label: '掌握度', icon: Target, end: false },
]

function DueBadge() {
  const [due, setDue] = useState(0)
  useEffect(() => {
    let alive = true
    const load = () =>
      get<Stats>('/api/stats')
        .then((s) => alive && setDue(s.due_now))
        .catch(() => {})
    load()
    const t = setInterval(load, 30_000)
    return () => {
      alive = false
      clearInterval(t)
    }
  }, [])
  if (!due) return null
  return (
    <span className="badge badge-error badge-sm ml-auto border-none text-error-content">{due}</span>
  )
}

export default function App() {
  return (
    <div className="flex h-full">
      <aside className="flex w-52 shrink-0 flex-col border-r border-base-300 bg-base-200/50">
        <div className="flex items-center gap-2 px-5 pt-6 pb-4">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-content">
            <GraduationCap className="h-5 w-5" />
          </div>
          <div>
            <div className="text-sm font-bold leading-tight">Deep Study</div>
            <div className="text-[11px] opacity-60">个人学习舱 · 零 LLM</div>
          </div>
        </div>
        <nav className="flex flex-col gap-0.5 px-3">
          {nav.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                `flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? 'bg-primary/10 font-semibold text-primary'
                    : 'text-base-content/75 hover:bg-base-200 hover:text-base-content'
                }`
              }
            >
              <Icon className="h-4 w-4" />
              {label}
              {to === '/review' && <DueBadge />}
            </NavLink>
          ))}
        </nav>
        <div className="mt-auto p-4 text-[11px] leading-relaxed opacity-50">
          <div className="flex items-center gap-1.5">
            <BookOpenText className="h-3.5 w-3.5" />
            文件即数据库 · AGENTS.md 即契约
          </div>
          <div className="mt-1">导师 = 你打开的 harness</div>
        </div>
      </aside>
      <main className="h-full flex-1 overflow-y-auto">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/review" element={<Review />} />
          <Route path="/library" element={<Library />} />
          <Route path="/notes" element={<Notes />} />
          <Route path="/notes/:id" element={<Notes />} />
          <Route path="/sessions" element={<Sessions />} />
          <Route path="/sessions/:id" element={<Sessions />} />
          <Route path="/workflows" element={<Workflows />} />
          <Route path="/mastery" element={<Mastery />} />
        </Routes>
      </main>
    </div>
  )
}
