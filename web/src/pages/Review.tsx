import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Lightbulb, PenLine, PartyPopper, Repeat } from 'lucide-react'
import { get, post, type MasteryResp, type QueueCard, type Topic } from '../api'

// 队列播放器：先评分后揭示（检索练习），四档自评走 FSRS，确认答案后手动点「下一张」；
// 自由回忆模式把答案写给 harness 事后批改。支持按主题过滤队列。
// mode=learn 只放待学新卡；mode=review 只放已学到期卡。
export default function QueuePlayer({ mode }: { mode: 'learn' | 'review' }) {
  const [queue, setQueue] = useState<QueueCard[] | null>(null)
  const [idx, setIdx] = useState(0)
  const [revealed, setRevealed] = useState(false)
  const [recallMode, setRecallMode] = useState(false)
  const [recallText, setRecallText] = useState('')
  const [recallSaved, setRecallSaved] = useState(false)
  // 已提交的评分档位：非 null 表示本卡已评分并揭示答案
  const [graded, setGraded] = useState<number | null>(null)
  // 提示默认隐藏，点击按钮或按 H 才显示（避免直接剧透回忆抓手）
  const [hintShown, setHintShown] = useState(false)
  const [busy, setBusy] = useState(false)
  const [done, setDone] = useState(0)
  const [err, setErr] = useState('')
  // 所选主题同步到 URL（/learn 或 /review?topic=xxx），刷新后保留；为空时移除参数
  const [searchParams, setSearchParams] = useSearchParams()
  const topic = searchParams.get('topic') ?? ''
  const setTopic = (t: string) => setSearchParams(t ? { topic: t } : {}, { replace: true })
  const [topics, setTopics] = useState<Topic[]>([])
  const [countByTopic, setCountByTopic] = useState<Record<string, number>>({})
  // 请求序号：切换主题/模式时递增，过期响应直接丢弃（竞态防护）
  const reqSeq = useRef(0)

  const load = useCallback(() => {
    const seq = ++reqSeq.current
    setQueue(null)
    setIdx(0)
    setRevealed(false)
    setRecallMode(false)
    setRecallText('')
    setRecallSaved(false)
    setGraded(null)
    setHintShown(false)
    setDone(0)
    setErr('')
    const q = `?mode=${mode}${topic ? `&topic=${encodeURIComponent(topic)}` : ''}`
    get<{ cards: QueueCard[] }>(`/api/review/queue${q}`)
      .then((r) => {
        if (seq === reqSeq.current) setQueue(r.cards)
      })
      .catch((e) => {
        if (seq === reqSeq.current) setErr(String(e.message ?? e))
      })
    get<MasteryResp>('/api/mastery')
      .then((r) => {
        if (seq !== reqSeq.current) return
        // 计数徽标按模式取数：learn 看待学新卡，review 看已学到期
        const m: Record<string, number> = {}
        for (const [k, v] of Object.entries(r.per_topic ?? {})) m[k] = mode === 'learn' ? v.new : v.due
        setCountByTopic(m)
      })
      .catch(() => {})
  }, [topic, mode])

  useEffect(load, [load])

  useEffect(() => {
    get<Topic[]>('/api/topics')
      .then(setTopics)
      .catch(() => {})
  }, [])

  const card = queue?.[idx]
  const total = queue?.length ?? 0

  const topicSlugs = Array.from(new Set([...topics.map((t) => t.slug), ...Object.keys(countByTopic)]))
  const picker =
    topicSlugs.length > 0 ? (
      <select
        className="select select-bordered select-sm max-w-48"
        value={topic}
        onChange={(e) => setTopic(e.target.value)}
        title={mode === 'learn' ? '选择要学习的主题' : '选择要复习的主题'}
      >
        <option value="">全部主题</option>
        {topicSlugs.map((s) => (
          <option key={s} value={s}>
            {topics.find((t) => t.slug === s)?.name || s}
            {countByTopic[s] ? `（${countByTopic[s]} ${mode === 'learn' ? '新卡' : '到期'}）` : ''}
          </option>
        ))}
      </select>
    ) : null

  // 提交评分：首评成功后揭示答案（不自动跳下一张）；已揭示时点其他档位 = 改判，
  // 服务端恢复评分前 fsrs 状态后按新档位重算（不重复计数）。失败不揭示、可重试
  const grade = useCallback(
    async (rating: number) => {
      if (!card || busy) return
      const isRegrade = revealed && graded !== null
      if (revealed && !isRegrade) return
      if (isRegrade && rating === graded) return
      setBusy(true)
      try {
        await post(isRegrade ? '/api/review/regrade' : '/api/review/grade', { id: card.id, rating })
        if (!isRegrade) {
          setDone((d) => d + 1)
          setRevealed(true)
        }
        setGraded(rating)
      } catch (e) {
        setErr(String((e as Error).message))
      } finally {
        setBusy(false)
      }
    },
    [card, busy, revealed, graded],
  )

  // 第二步：推进到下一张（或清空队列），并重置本卡状态
  const next = useCallback(() => {
    setRevealed(false)
    setRecallMode(false)
    setRecallText('')
    setRecallSaved(false)
    setGraded(null)
    setHintShown(false)
    if (idx + 1 >= total) {
      setQueue([])
    } else {
      setIdx((i) => i + 1)
    }
  }, [idx, total])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null
      // 可编辑的表单控件（输入框/下拉框）聚焦时放行按键，不触发快捷键
      if (
        (target instanceof HTMLTextAreaElement ||
          target instanceof HTMLInputElement ||
          target instanceof HTMLSelectElement) &&
        !target.disabled
      )
        return
      if ((e.key === 'h' || e.key === 'H') && card?.hint) {
        // H = 显示/隐藏提示
        setHintShown((v) => !v)
      } else if (['1', '2', '3', '4'].includes(e.key)) {
        // 1-4 评分；已评分（已揭示）时 = 改判
        grade(Number(e.key))
      } else if (revealed && (e.code === 'Space' || e.key === 'Enter')) {
        // 已评分：空格/回车 = 下一张
        e.preventDefault()
        next()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [revealed, grade, next, card])

  const saveRecall = async () => {
    if (!card || !recallText.trim()) return
    try {
      await post('/api/review/recall', { id: card.id, answer: recallText.trim() })
      setRecallSaved(true)
    } catch (e) {
      setErr(String((e as Error).message))
    }
  }

  if (err && !queue) return <div className="p-8 text-error">{err}</div>
  if (!queue)
    return (
      <div className="p-8">
        {picker}
        <div className="loading loading-dots loading-lg text-primary mt-4" />
      </div>
    )

  if (total === 0)
    return (
      <div className="flex h-full flex-col items-center justify-center gap-4 p-8 text-center">
        {picker}
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/15 text-success">
          <PartyPopper className="h-8 w-8" />
        </div>
        <h1 className="text-lg font-bold">
          {done > 0
            ? mode === 'learn'
              ? `队列完成，共学习 ${done} 张新卡`
              : `队列完成，共复习 ${done} 张`
            : mode === 'learn'
              ? topic
                ? '该主题没有待学新卡'
                : '没有待学新卡'
              : topic
                ? '该主题当前没有到期卡片'
                : '当前没有到期卡片'}
        </h1>
        <p className="max-w-md text-sm opacity-60">
          {mode === 'learn'
            ? '新卡来自 harness 的导入/诊断工作流。想补充材料时去「工作流」页跑导入，或回「仪表盘」看看全局。'
            : 'FSRS 会把下次复习安排在记忆临界点上。空档期可以去 harness 里跑一次自测或费曼（「工作流」页有现成命令），或回「仪表盘」看看全局。'}
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
        <div className="flex items-center gap-3">
          {picker}
          <span className="opacity-60">
            第 {idx + 1} / {total} 张{done > 0 ? ` · 已完成 ${done}` : ''}
          </span>
        </div>
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
            </div>
            <h2 className="mt-3 text-lg leading-relaxed font-medium whitespace-pre-wrap">{card.front}</h2>

            {card.hint &&
              (hintShown ? (
                <div className="mt-3 w-full rounded-xl bg-warning/10 p-4 text-sm leading-relaxed">
                  <span className="mr-2 text-xs font-semibold opacity-60">提示</span>
                  {card.hint}
                </div>
              ) : (
                <button className="btn btn-ghost btn-sm mt-3 gap-1.5" onClick={() => setHintShown(true)}>
                  <Lightbulb className="h-3.5 w-3.5" /> 显示提示 <kbd className="kbd kbd-sm">H</kbd>
                </button>
              ))}

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

            {revealed && (
              <>
                <div className="mt-4 w-full rounded-xl bg-secondary/8 p-5 text-[15px] leading-relaxed whitespace-pre-wrap">
                  {card.back}
                </div>
                <button className="btn btn-primary mt-4 gap-2" onClick={next}>
                  下一张 <kbd className="kbd kbd-sm">空格</kbd>
                </button>
              </>
            )}
          </div>
        </div>
      )}

      {/* 四档评分：评分后仍可按其他档位改判；高亮当前档位，其余变淡 */}
      <div className="mt-6 grid grid-cols-4 gap-3 transition-opacity">
        {[
          { r: 1, label: '重来', sub: '完全想不起', cls: 'btn-error' },
          { r: 2, label: '困难', sub: '想了很久', cls: 'btn-warning' },
          { r: 3, label: '良好', sub: '略有迟疑', cls: 'btn-primary' },
          { r: 4, label: '简单', sub: '秒答', cls: 'btn-success' },
        ].map(({ r, label, sub, cls }) => (
          <button
            key={r}
            className={`btn ${cls} flex flex-col gap-0 py-3 text-primary-content ${revealed && graded !== r ? 'opacity-30' : ''}`}
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
        {revealed
          ? '看答案后发现与记忆不符？直接改点其他档位即可改判'
          : '评分写入卡片 fsrs 块与 review-log · 由 go-fsrs 计算下次间隔'}
      </p>
    </div>
  )
}
