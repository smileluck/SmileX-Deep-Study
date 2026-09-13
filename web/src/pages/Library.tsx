import { useCallback, useEffect, useRef, useState } from 'react'
import { FileUp, FolderTree, Inbox, Terminal } from 'lucide-react'
import { fname, fmtDate, fmtSize, fsize, get, type MaterialsResp, type Topic } from '../api'
import CopyButton from '../components/CopyButton'

export default function Library() {
  const [data, setData] = useState<MaterialsResp | null>(null)
  const [topics, setTopics] = useState<Topic[]>([])
  const [dragging, setDragging] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [msg, setMsg] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const load = useCallback(() => {
    get<MaterialsResp>('/api/materials').then(setData).catch(() => setData({ inbox: [], library: [], index: [] }))
    get<Topic[]>('/api/topics').then(setTopics).catch(() => setTopics([]))
  }, [])
  useEffect(load, [load])

  const upload = async (files: FileList | File[]) => {
    const fd = new FormData()
    for (const f of files) fd.append('files', f)
    setUploading(true)
    setMsg('')
    try {
      const res = await fetch('/api/materials/upload', { method: 'POST', body: fd })
      const j = await res.json()
      if (!res.ok) throw new Error(j.error ?? res.statusText)
      setMsg(`已上传 ${j.saved.length} 个文件到 inbox，去 harness 执行导入命令即可`)
      load()
    } catch (e) {
      setMsg(`上传失败：${(e as Error).message}`)
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
          onChange={(e) => e.target.files && upload(e.target.files)}
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
      {msg && <div className="mt-3 text-sm text-primary">{msg}</div>}

      {data && (data.inbox?.length ?? 0) > 0 && (
        <section className="mt-8">
          <h2 className="flex items-center gap-2 text-sm font-semibold">
            <Inbox className="h-4 w-4" /> 收件箱（待导入 {data.inbox.length}）
          </h2>
          <ul className="mt-3 flex flex-col gap-2">
            {data.inbox.map((f) => (
              <li key={fname(f)} className="card bg-base-100 p-4 shadow-sm">
                <div className="flex items-center gap-3">
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-medium">{fname(f)}</div>
                    <div className="text-xs opacity-50">
                      {fmtSize(fsize(f))} · {fmtDate(f.mtime ?? f.Mtime)}
                    </div>
                  </div>
                  <code className="hidden rounded-lg bg-base-200 px-2.5 py-1.5 text-xs sm:block">
                    /study:import {fname(f)}
                  </code>
                  <CopyButton text={`/study:import ${fname(f)}`} label="复制命令" />
                </div>
              </li>
            ))}
          </ul>
          <p className="mt-2 text-xs opacity-50">
            在仓库根目录打开 ZCode 执行上面的命令；其他工具去「工作流」页复制对应 prompt。
          </p>
        </section>
      )}

      {topics.length > 0 && (
        <section className="mt-8">
          <h2 className="flex items-center gap-2 text-sm font-semibold">
            <FolderTree className="h-4 w-4" /> 学习主题（{topics.length}）
          </h2>
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            {topics.map((t) => (
              <div key={t.slug} className="card bg-base-100 p-4 shadow-sm">
                <div className="text-sm font-semibold">{t.name || t.slug}</div>
                <div className="mt-0.5 font-mono text-[11px] opacity-45">{t.slug}</div>
                {t.goal && <div className="mt-2 text-xs opacity-70">🎯 {t.goal}</div>}
                <div className="mt-3 flex gap-2 text-xs">
                  <span className="badge badge-ghost">{t.notes ?? 0} 笔记</span>
                  <span className="badge badge-ghost">{t.cards ?? 0} 卡片</span>
                </div>
              </div>
            ))}
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
                      <tr key={fname(f)}>
                        <td className="max-w-56 truncate">{fname(f)}</td>
                        <td>—</td>
                        <td className="text-xs opacity-60">{fmtDate(f.mtime ?? f.Mtime)}</td>
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
