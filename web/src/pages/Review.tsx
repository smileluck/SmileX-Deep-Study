import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Eye, PenLine, PartyPopper, Repeat } from 'lucide-react'
import { get, post, type QueueCard } from '../api'

// 复习播放器：先回忆后揭示（检索练习），四档自评走 FSRS；
// 自由回忆模式把答案写给 harness 事后批改。
export default function Review() {
  const [queue, setQueue] = useState<QueueCard[] | null>(null)
  const [idx, setIdx] = useState(0)
  const [revealed, setRevealed] = useState(false)
  const [recallMode, setRecallMode] = useState(false)
  const [recallText, setRecallText] = useState('')
  const [recallSaved, setRecallSaved] = useState(false)
  const [busy, setBusy] = useState(false)
  const [done, setDone] = useState(0)
  const [err, setErr] = useState('')

  const load = useCallback(() => {
    setQueue(null)
    setIdx(0)
    setDone(0)
    setErr('')
    get<{ cards: QueueCard[] }>('/api/review/queue')
      .then((r) => setQueue(r.cards))
      .catch((e) => setErr(String(e.message ?? e)))
  }, [])

  useEffect(load, [load])

  const card = queue?.[idx]
  const total = queue?.length ?? 0

  const grade = useCallback(
    async (rating: number) => {
      if (!card || busy) return
      setBusy(true)
      try {
        await post('/api/review/grade', { id: card.id, rating })
        setDone((d) => d + 1)
        setRevealed(false)
        setRecallMode(false)
        setRecallText('')
        setRecallSaved(false)
        if (idx + 1 >= total) {
          setQueue([])
        } else {
          setIdx((i) => i + 1)
        }
      } catch (e) {
        setErr(String((e as Error).message))
      } finally {
        setBusy(false)
      }
    },
    [card, busy, idx, total],
  )

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLTextAreaElement | null
      // 自由回忆模式下，仅在输入框可编辑时放行按键给它
      if (recallMode && target?.tagName === 'TEXTAREA' && !target.disabled) return
      if (e.code === 'Space' && !revealed) {
        e.preventDefault()
        setRevealed(true)
      } else if (revealed && ['1', '2', '3', '4'].includes(e.key)) {
        grade(Number(e.key))
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [revealed, grade, recallMode])

  const saveRecall = async () => {
    if (!card || !recallText.trim()) return
    try {
      await post('/api/review/recall', { id: card.id, answer: recallText.trim() })
      setRecallSaved(true)
    } catch (e) {
      setErr(String((e.message ?? e)))
    }
  }

  if (err && !queue) return <div className="p-8 text-error">{err}</div>
  if (!queue)
    return (
      <div className="p-8">
        <div className="loading loading-dots loading-lg text-primary" />
      </div>
    )

  if (total === 0)
    return (
      <div className="flex h-full flex-col items-center justify-center gap-4 p-8 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/15 text-success">
          <PartyPopper className="h-8 w-8" />
        </div>
        <h1 className="text-lg font-bold">
          {done > 0 ? `今日队列完成，共复习 ${done} 张` : '当前没有到期卡片'}
        </h1>
        <p className="max-w-md text-sm opacity-60">
          FSRS 会把下次复习安排在记忆临界点上。空档期可以去 harness 里跑一次自测或费曼
          （「工作流」页有现成命令），或回「仪表盘」看看全局。
        </p>
        <div className="mt-2 flex gap-2">
          <button className="btn btn-outline btn-sm" onClick={load}>
            <Repeat className="h-4 w-4" /> 刷新队列
          </button>
          <Link to="/workflows" className="btn btn-primary btn-sm">
            去自测 / 费曼
          </Link>
        </div>
      </div>
    )

  return (
    <div className="mx-auto flex h-full max-w-2xl flex-col p-8">
      <div className="flex items-center justify-between text-sm">
        <span className="opacity-60">
          第 {idx + 1} / {total} 张{done > 0 ? ` · 已完成 ${done}` : ''}
        </span>
        <div className="flex items-center gap-2">
          <label className="label cursor-pointer gap-1.5 text-xs">
            <input
              type="checkbox"
              className="checkbox checkbox-xs checkbox-primary"
              checked={recallMode}
              onChange={(e) => setRecallMode(e.target.checked)}
            />
            <PenLine className="h-3.5 w-3.5" /> 自由回忆模式
          </label>
        </div>
      </div>
      <progress className="progress progress-primary mt-3 h-1.5" value={idx} max={total} />

      {card && (
        <div className="card mt-6 flex-1 bg-base-100 shadow-sm">
          <div className="card-body flex-col items-start p-8">
            <div className="flex items-center gap-2">
              <span className="badge badge-ghost badge-sm">{card.topic || '未分类'}</span>
              {card.state_name === 'new' && (
                <span className="badge badge-primary badge-outline badge-sm">新卡</span>
              )}
            </div>
            <h2 className="mt-3 text-lg leading-relaxed font-medium whitespace-pre-wrap">{card.front}</h2>

            {recallMode && (
              <div className="mt-2 w-full">
                <textarea
                  className="textarea textarea-bordered min-h-24 w-full text-sm leading-relaxed"
                  placeholder="先别翻答案 —— 凭记忆写下你能想起的一切，稍后由自测/诊断工作流批改"
                  value={recallText}
                  onChange={(e) => setRecallText(e.target.value)}
                  disabled={recallSaved}
                />
                {!recallSaved ? (
                  <button className="btn btn-ghost btn-xs mt-1" onClick={saveRecall} disabled={!recallText.trim()}>
                    保存回忆答案
                  </button>
                ) : (
                  <span className="text-xs text-success">已记录，待导师批改 ✓</span>
                )}
              </div>
            )}

            {revealed ? (
              <div className="mt-4 w-full rounded-xl bg-secondary/8 p-5 text-[15px] leading-relaxed whitespace-pre-wrap">
                {card.back}
              </div>
            ) : (
              <button className="btn btn-primary btn-outline mt-6 gap-2" onClick={() => setRevealed(true)}>
                <Eye className="h-4 w-4" /> 显示答案 <kbd className="kbd kbd-sm">空格</kbd>
              </button>
            )}
          </div>
        </div>
      )}

      <div className={`mt-6 grid grid-cols-4 gap-3 transition-opacity ${revealed ? '' : 'pointer-events-none opacity-30'}`}>
        {[
          { r: 1, label: '重来', sub: '完全想不起', cls: 'btn-error' },
          { r: 2, label: '困难', sub: '想了很久', cls: 'btn-warning' },
          { r: 3, label: '良好', sub: '略有迟疑', cls: 'btn-primary' },
          { r: 4, label: '简单', sub: '秒答', cls: 'btn-success' },
        ].map(({ r, label, sub, cls }) => (
          <button
            key={r}
            className={`btn ${cls} flex flex-col gap-0 py-3 text-primary-content`}
            onClick={() => grade(r)}
            disabled={busy}
          >
            <span className="text-sm font-semibold">
              {label} <kbd className="kbd kbd-sm opacity-70">{r}</kbd>
            </span>
            <span className="text-[10px] font-normal opacity-80">{sub}</span>
          </button>
        ))}
      </div>
      {err && <div className="mt-3 text-sm text-error">{err}</div>}
      <p className="mt-3 text-center text-[11px] opacity-40">
        评分写入卡片 fsrs 块与 review-log · 由 go-fsrs 计算下次间隔
      </p>
    </div>
  )
}
