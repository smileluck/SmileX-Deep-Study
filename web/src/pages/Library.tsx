import { useCallback, useEffect, useRef, useState } from 'react'
import { FileUp, FolderTree, Inbox, Pause, Play, Terminal } from 'lucide-react'
import { fmtDate, fmtSize, get, post, type MaterialsResp, type Topic } from '../api'
import CopyButton from '../components/CopyButton'

export default function Library() {
  const [data, setData] = useState<MaterialsResp | null>(null)
  const [topics, setTopics] = useState<Topic[]>([])
  const [dragging, setDragging] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [msg, setMsg] = useState<{ kind: 'ok' | 'err'; text: string } | null>(null)
  const [err, setErr] = useState('')
  const [importTopic, setImportTopic] = useState<Record<string, string>>({})
  const [statusBusy, setStatusBusy] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const load = useCallback(() => {
    setErr('')
    get<MaterialsResp>('/api/materials')
      .then(setData)
      .catch((e) => setErr(`材料列表加载失败：${e.message ?? e}`))
    get<Topic[]>('/api/topics')
      .then(setTopics)
      .catch((e) => setErr(`主题列表加载失败：${e.message ?? e}`))
  }, [])
  useEffect(load, [load])

  const toggleTopicStatus = async (t: Topic) => {
    const next = t.status === 'paused' ? 'active' : 'paused'
    setStatusBusy(t.slug)
    setMsg(null)
    try {
      await post(`/api/topics/${t.slug}/status`, { status: next })
      load()
    } catch (e) {
      setMsg({ kind: 'err', text: `操作失败：${(e as Error).message}` })
    } finally {
      setStatusBusy(null)
    }
  }

  // MAX_UPLOAD_MB 与服务端 maxUploadBytes（1GB）保持一致。
  // 超限文件必须在前端拦下：直接发送的话，服务端预检 413 时请求体仍在传输，
  // 连接会被 RST，fetch 只报 "Failed to fetch"，用户看不到真实原因。
  const MAX_UPLOAD_MB = 1024

  const upload = async (files: FileList | File[]) => {
    const list = [...files]
    const tooBig = list.filter((f) => f.size > MAX_UPLOAD_MB * 1024 * 1024)
    if (tooBig.length) {
      setMsg({ kind: 'err', text: `以下文件超过 1GB 上限：${tooBig.map((f) => `${f.name}（${(f.size / 1024 / 1024 / 1024).toFixed(2)}GB）`).join('、')}` })
      return
    }
    const fd = new FormData()
    for (const f of list) fd.append('files', f)
    setUploading(true)
    setMsg(null)
    try {
      const res = await fetch('/api/materials/upload', { method: 'POST', body: fd })
      // 错误响应不一定是 JSON（如网关 502），解析失败时给通用提示
      const j = await res.json().catch(() => null) as { saved?: unknown[]; error?: string } | null
      if (!res.ok) throw new Error(j?.error ?? `上传失败（HTTP ${res.status}）`)
      setMsg({ kind: 'ok', text: `已上传 ${j?.saved?.length ?? 0} 个文件到 inbox，去 harness 执行导入命令即可` })
      load()
    } catch (e) {
      const m = (e as Error).message
      setMsg({ kind: 'err', text: m === 'Failed to fetch' ? '上传失败：连接中断（文件过大或网络异常），请检查文件大小后重试' : `上传失败：${m}` })
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="mx-auto max-w-4xl p-8">
      <h1 className="text-xl font-bold">资料库</h1>
      <p className="mt-1 text-sm opacity-60">材料是学习的入口：上传 → harness 导入 → 生成笔记与卡片。</p>

      <div
        className={`mt-6 rounded-2xl border-2 border-dashed p-10 text-center transition-colors ${
          dragging ? 'border-primary bg-primary/5' : 'border-base-300'
        }`}
        onDragOver={(e) => {
          e.preventDefault()
          setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(e) => {
          e.preventDefault()
          setDragging(false)
          if (e.dataTransfer.files.length) upload(e.dataTransfer.files)
        }}
        onClick={() => inputRef.current?.click()}
        role="button"
      >
        <input
          ref={inputRef}
          type="file"
          multiple
          className="hidden"
          onChange={(e) => {
            if (e.target.files?.length) upload(e.target.files)
            // 清空 value，否则再次选择同一文件不会触发 onChange
            e.target.value = ''
          }}
        />
        {uploading ? (
          <div className="loading loading-dots loading-md text-primary" />
        ) : (
          <>
            <FileUp className="mx-auto h-8 w-8 text-primary/70" />
            <div className="mt-3 text-sm font-medium">拖拽文件到此处，或点击选择</div>
            <div className="mt-1 text-xs opacity-50">PDF / DOCX / EPUB / Markdown / TXT · 落盘到 data/inbox/</div>
          </>
        )}
      </div>
      {msg && (
        <div className={`mt-3 text-sm ${msg.kind === 'err' ? 'text-error' : 'text-primary'}`}>{msg.text}</div>
      )}
      {err && <div className="mt-3 text-sm text-error">{err}</div>}

      {data && (data.inbox?.length ?? 0) > 0 && (
        <section className="mt-8">
          <h2 className="flex items-center gap-2 text-sm font-semibold">
            <Inbox className="h-4 w-4" /> 收件箱（待导入 {data.inbox.length}）
          </h2>
          <ul className="mt-3 flex flex-col gap-2">
            {data.inbox.map((f) => {
              const name = f.name
              const slug = importTopic[name] ?? ''
              const cmd = `/study:import ${name}${slug ? ` ${slug}` : ''}`
              return (
                <li key={name} className="card bg-base-100 p-4 shadow-sm">
                  <div className="flex items-center gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-sm font-medium">{name}</div>
                      <div className="text-xs opacity-50">
                        {fmtSize(f.size)} · {fmtDate(f.mtime)}
                      </div>
                    </div>
                    <select
                      className="select select-bordered select-xs max-w-32"
                      value={slug}
                      onChange={(e) => setImportTopic((m) => ({ ...m, [name]: e.target.value }))}
                      title="选择导入到哪个主题"
                    >
                      <option value="">新主题</option>
                      {topics
                        .filter((t) => t.status !== 'paused')
                        .map((t) => (
                          <option key={t.slug} value={t.slug}>
                            {t.name || t.slug}
                          </option>
                        ))}
                    </select>
                    <code className="hidden rounded-lg bg-base-200 px-2.5 py-1.5 text-xs sm:block">{cmd}</code>
                    <CopyButton text={cmd} label="复制命令" />
                  </div>
                </li>
              )
            })}
          </ul>
          <p className="mt-2 text-xs opacity-50">
            在仓库根目录打开 ZCode 执行上面的命令；其他工具去「工作流」页复制对应 prompt。
            多个材料可陆续导入同一主题：在条目右侧选择已有主题，命令会自动带上该 slug，笔记序号在主题内续排；选「新主题」则由 harness 起名新建。
          </p>
        </section>
      )}

      {topics.length > 0 && (
        <section className="mt-8">
          <h2 className="flex items-center gap-2 text-sm font-semibold">
            <FolderTree className="h-4 w-4" /> 学习主题（{topics.length}）
          </h2>
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            {topics.map((t) => {
              const paused = t.status === 'paused'
              return (
                <div key={t.slug} className={`card relative bg-base-100 p-4 shadow-sm ${paused ? 'opacity-60' : ''}`}>
                  <div className="flex items-center gap-2">
                    <div className="text-sm font-semibold">{t.name || t.slug}</div>
                    {paused && <span className="badge badge-warning badge-xs">已搁置</span>}
                  </div>
                  <div className="mt-0.5 font-mono text-[11px] opacity-45">{t.slug}</div>
                  {t.goal && <div className="mt-2 text-xs opacity-70">🎯 {t.goal}</div>}
                  <div className="mt-3 flex gap-2 text-xs">
                    <span className="badge badge-ghost">{t.notes ?? 0} 笔记</span>
                    <span className="badge badge-ghost">{t.cards ?? 0} 卡片</span>
                  </div>
                  <button
                    className="btn btn-ghost btn-xs absolute right-2 bottom-2 gap-1"
                    onClick={() => toggleTopicStatus(t)}
                    disabled={statusBusy === t.slug}
                    title={paused ? '恢复该主题，重新进入学习/复习队列' : '搁置该主题，暂时退出学习/复习队列'}
                  >
                    {paused ? <Play className="h-3 w-3" /> : <Pause className="h-3 w-3" />}
                    {paused ? '恢复' : '搁置'}
                  </button>
                </div>
              )
            })}
          </div>
        </section>
      )}

      {data && ((data.index?.length ?? 0) > 0 || (data.library?.length ?? 0) > 0) && (
        <section className="mt-8">
          <h2 className="text-sm font-semibold">已归档材料</h2>
          <div className="card mt-3 bg-base-100 shadow-sm">
            <table className="table table-sm">
              <thead>
                <tr className="text-xs opacity-60">
                  <th>文件</th>
                  <th>主题</th>
                  <th>导入时间</th>
                </tr>
              </thead>
              <tbody>
                {(data.index?.length ?? 0) > 0
                  ? data.index.map((m) => (
                      <tr key={m.id}>
                        <td className="max-w-56 truncate">{m.stored_name}</td>
                        <td className="font-mono text-xs">{m.topic || '—'}</td>
                        <td className="text-xs opacity-60">{m.imported_at}</td>
                      </tr>
                    ))
                  : data.library.map((f) => (
                      <tr key={f.name}>
                        <td className="max-w-56 truncate">{f.name}</td>
                        <td>—</td>
                        <td className="text-xs opacity-60">{fmtDate(f.mtime)}</td>
                      </tr>
                    ))}
              </tbody>
            </table>
          </div>
          <p className="mt-2 flex items-center gap-1 text-xs opacity-50">
            <Terminal className="h-3 w-3" /> 手动导入模板见 prompts/import.md
          </p>
        </section>
      )}
    </div>
  )
}
