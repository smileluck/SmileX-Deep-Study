import { useEffect, useState } from 'react'
import { get, type Stats } from './api'

// 共享 stats 轮询：模块级单例 store，多个订阅者（如两个 DueBadge）共用一个 30s 定时器；
// 页面隐藏时暂停轮询，重新可见时立即补拉一次。
let stats: Stats | null = null
const listeners = new Set<(s: Stats | null) => void>()
let timer: ReturnType<typeof setInterval> | null = null

function load() {
  get<Stats>('/api/stats')
    .then((s) => {
      stats = s
      listeners.forEach((l) => l(s))
    })
    .catch(() => {})
}

function onVisible() {
  if (document.visibilityState === 'visible') load()
}

function startPolling() {
  if (timer) return
  load()
  timer = setInterval(() => {
    if (document.visibilityState === 'hidden') return
    load()
  }, 30_000)
  document.addEventListener('visibilitychange', onVisible)
}

function stopPolling() {
  if (timer) clearInterval(timer)
  timer = null
  document.removeEventListener('visibilitychange', onVisible)
}

export function useStats(): Stats | null {
  const [s, setS] = useState<Stats | null>(stats)
  useEffect(() => {
    listeners.add(setS)
    setS(stats)
    startPolling()
    return () => {
      listeners.delete(setS)
      if (listeners.size === 0) stopPolling()
    }
  }, [])
  return s
}
