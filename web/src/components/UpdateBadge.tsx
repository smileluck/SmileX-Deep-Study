import { useEffect, useState } from 'react'
import { ArrowUpCircle, ExternalLink, RefreshCw } from 'lucide-react'
import { get, post, type UpdateApplyResult, type UpdateCheck } from '../api'

type Phase =
  | { kind: 'idle' }
  | { kind: 'applying' }
  | { kind: 'done'; message: string }
  | { kind: 'failed'; error: string }

export default function UpdateBadge() {
  const [info, setInfo] = useState<UpdateCheck | null>(null)
  const [open, setOpen] = useState(false)
  const [phase, setPhase] = useState<Phase>({ kind: 'idle' })

  useEffect(() => {
    get<UpdateCheck>('/api/update/check')
      .then(setInfo)
      .catch(() => setInfo(null))
  }, [])

  if (!info) return null

  const apply = async () => {
    setPhase({ kind: 'applying' })
    try {
      const res = await post<UpdateApplyResult>('/api/update/apply')
      if (res.ok) setPhase({ kind: 'done', message: res.message ?? '更新完成，请重启生效' })
      else setPhase({ kind: 'failed', error: res.error ?? '更新失败' })
    } catch (e) {
      setPhase({ kind: 'failed', error: e instanceof Error ? e.message : '更新失败' })
    }
  }

  const manual = (
    <a
      href={info.manual_url}
      target="_blank"
      rel="noreferrer"
      className="link link-hover inline-flex items-center gap-1"
    >
      手动下载
      <ExternalLink className="h-3 w-3" />
    </a>
  )

  return (
    <div className="text-[11px] leading-relaxed">
      <button
        className="flex items-center gap-1.5 opacity-70 hover:opacity-100"
        onClick={() => setOpen(!open)}
      >
        {info.has_update && <span className="badge badge-warning badge-xs" />}
        <RefreshCw className="h-3 w-3" />
        {info.reason === 'dev build'
          ? '开发版'
          : info.has_update
            ? `${info.current} → ${info.latest} 可更新`
            : info.error
              ? '检查更新失败'
              : `已是最新 ${info.current}`}
      </button>

      {open && (
        <div className="mt-2 rounded-lg border border-base-300 bg-base-100 p-2.5">
          {phase.kind === 'done' ? (
            <div className="text-success">{phase.message}</div>
          ) : (
            <>
              {info.error && (
                <div className="opacity-80">
                  {info.error}，可{manual}
                </div>
              )}
              {info.has_update && (
                <>
                  {info.notes && (
                    <div className="mb-2 max-h-24 overflow-y-auto whitespace-pre-wrap opacity-70">
                      {info.notes}
                    </div>
                  )}
                  <button
                    className="btn btn-primary btn-xs w-full"
                    disabled={phase.kind === 'applying'}
                    onClick={apply}
                  >
                    <ArrowUpCircle className="h-3.5 w-3.5" />
                    {phase.kind === 'applying' ? '下载替换中…' : `立即更新到 ${info.latest}`}
                  </button>
                </>
              )}
              {!info.has_update && !info.error && <div className="opacity-60">无需更新</div>}
              {phase.kind === 'failed' && (
                <div className="mt-2 text-error">
                  {phase.error}，可{manual}
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}
