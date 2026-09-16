// 极简 API client：生产同源、开发走 vite proxy。
export async function get<T>(path: string): Promise<T> {
  const res = await fetch(path)
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error ?? res.statusText)
  return res.json()
}

export async function post<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error ?? res.statusText)
  return res.json()
}

// ---------- 类型 ----------

export interface Stats {
  total_cards: number
  due_now: number
  new_cards: number
  reviews_today: number
  learned_today: number
  streak: number
  heatmap: { date: string; count: number; learned?: number }[]
  recent_sessions: { id: string; type: string; topic: string; date: string; summary: string }[] | null
  mastery: Record<string, { level: number; updated: string }>
}

export interface Workflow {
  id: string
  name: string
  icon: string
  runner: string
  command: string | null
  prompt_file: string | null
  description: string
  outputs: string[]
  ui_hint: string
}

export interface MaterialsResp {
  inbox: { Name?: string; name?: string; size?: number; Size?: number; mtime?: string; Mtime?: string }[]
  library: { Name?: string; name?: string; size?: number; Size?: number; mtime?: string; Mtime?: string }[]
  index: { id: number; original_name: string; stored_name: string; topic: string; status: string; imported_at: string }[]
}

export interface Topic {
  slug: string
  name?: string
  goal?: string
  notes?: number
  cards?: number
  description?: string
}

export interface NoteListItem {
  id: string
  title: string
  topic: string
  tags: string[] | null
  links: string[] | null
  created: string
  cards: number
  order?: number
}

export interface NoteDetail {
  id: string
  fm: Record<string, unknown>
  body: string
  backlinks: string[]
  cards: Record<string, unknown>[]
}

export interface SessionListItem {
  id: string
  type: string
  topic: string
  date: string
  tool: string
  summary: string
}

export interface SessionDetail {
  id: string
  fm: Record<string, unknown>
  body: string
}

export interface QueueCard {
  id: string
  front: string
  back: string
  body: string
  topic: string
  note: string | null
  type: string
  state_name?: string
  due?: string
}

export interface Plan {
  slug: string
  topic_name: string
  fm: Record<string, unknown>
  body: string
}

export interface PlansResp {
  master: { fm: Record<string, unknown>; body: string } | null
  topics: Plan[]
}

export interface MasteryResp {
  mastery: Record<
    string,
    { level: number; updated: string; evidence?: { date: string; kind: string; detail: string; delta?: number }[] }
  >
  per_topic: Record<string, { cards: number; due: number; new: number; reviews: number; again: number }>
}

// ---------- 工具 ----------

export function fmtSize(n?: number): string {
  if (!n && n !== 0) return ''
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}

export function fname(f: { Name?: string; name?: string }): string {
  return f.Name ?? f.name ?? ''
}

export function fsize(f: { Size?: number; size?: number }): number | undefined {
  return f.Size ?? f.size
}

export function fmtDate(d?: string): string {
  if (!d) return ''
  return d.slice(0, 10)
}
