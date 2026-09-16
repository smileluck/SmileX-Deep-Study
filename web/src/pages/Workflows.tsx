import { useEffect, useRef, useState } from 'react'
import {
  FileQuestion,
  GraduationCap,
  Inbox,
  Lightbulb,
  Map,
  Merge,
  MonitorSmartphone,
  Repeat,
  Stethoscope,
  Terminal,
} from 'lucide-react'
import { get, type Workflow } from '../api'
import CopyButton from '../components/CopyButton'

const icons: Record<string, typeof Inbox> = {
  inbox: Inbox,
  'graduation-cap': GraduationCap,
  lightbulb: Lightbulb,
  repeat: Repeat,
  'file-question': FileQuestion,
  stethoscope: Stethoscope,
  map: Map,
  merge: Merge,
}

export default function Workflows() {
  const [flows, setFlows] = useState<Workflow[]>([])
  const [expanded, setExpanded] = useState<string | null>(null)
  const [promptText, setPromptText] = useState('')
  // 请求序号：快速展开/收起不同工作流时递增，过期的 prompt 响应直接丢弃
  const promptSeq = useRef(0)

  useEffect(() => {
    get<Workflow[]>('/api/workflows').then(setFlows).catch(() => setFlows([]))
  }, [])

  const togglePrompt = async (f: Workflow) => {
    const seq = ++promptSeq.current
    if (expanded === f.id) {
      setExpanded(null)
      return
    }
    setExpanded(f.id)
    setPromptText('')
    if (f.prompt_file) {
      const name = f.prompt_file.split('/').pop()!.replace(/\.md$/, '')
      const res = await fetch(`/api/prompts/${name}`)
      const text = res.ok ? await res.text() : ''
      if (seq !== promptSeq.current) return
      if (res.ok) setPromptText(text)
    }
  }

  return (
    <div className="mx-auto max-w-4xl p-8">
      <h1 className="text-xl font-bold">工作流</h1>
      <p className="mt-1 text-sm opacity-60">
        八条学习工作流：理解类在 harness 里跑（复制命令即可），调度类在本界面完成。
      </p>

      <div className="mt-6 grid gap-4 lg:grid-cols-2">
        {flows.map((f) => {
          const Icon = icons[f.icon] ?? Terminal
          const isUI = f.runner.includes('Web UI')
          return (
            <div key={f.id} className="card bg-base-100 shadow-sm">
              <div className="card-body gap-3 p-5">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2.5">
                    <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${isUI ? 'bg-accent/20 text-accent-content' : 'bg-primary/10 text-primary'}`}>
                      <Icon className="h-4 w-4" />
                    </div>
                    <div className="font-semibold">{f.name}</div>
                  </div>
                  <span className={`badge badge-sm ${isUI ? 'badge-accent' : 'badge-primary'} badge-outline shrink-0`}>
                    <MonitorSmartphone className="mr-1 h-3 w-3" />
                    {f.runner}
                  </span>
                </div>
                <p className="text-[13px] leading-relaxed opacity-70">{f.description}</p>

                {f.command ? (
                  <div className="rounded-lg bg-base-200/70 px-3 py-2 font-mono text-xs">
                    <div className="flex items-center justify-between gap-2">
                      <span className="truncate">{f.command}</span>
                      <CopyButton text={f.command} />
                    </div>
                  </div>
                ) : (
                  <div className="rounded-lg bg-accent/10 px-3 py-2 text-xs font-medium text-accent-content">
                    {f.ui_hint}
                  </div>
                )}

                {f.command && (
                  <div className="flex items-center justify-between text-xs">
                    <span className="opacity-50">产出：{f.outputs.join(' · ')}</span>
                    {f.prompt_file && (
                      <button className="btn btn-ghost btn-xs" onClick={() => togglePrompt(f)}>
                        {expanded === f.id ? '收起 prompt ▲' : '其他工具？展开通用 prompt ▼'}
                      </button>
                    )}
                  </div>
                )}

                {expanded === f.id && f.prompt_file && (
                  <div className="relative mt-1 max-h-72 overflow-auto rounded-lg bg-neutral p-3 text-neutral-content">
                    <div className="sticky right-0 float-right">
                      <CopyButton text={promptText || f.prompt_file} label="复制全文" />
                    </div>
                    <pre className="pr-16 text-[11px] leading-relaxed whitespace-pre-wrap">
                      {promptText || `（读取 ${f.prompt_file} 失败 —— 在仓库根目录运行服务后重试）`}
                    </pre>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>

      <div className="card mt-8 bg-base-200/60">
        <div className="card-body p-5 text-[13px] leading-relaxed opacity-75">
          <div className="font-semibold opacity-90">四家 harness 怎么接？</div>
          <ul className="mt-1 list-disc space-y-1 pl-5">
            <li><b>ZCode / 通用 agent</b>：技能与命令在 <code>.agents/</code>（SKILL.md 通用格式），直接用上面的 <code>/study:*</code> 命令。</li>
            <li><b>WorkBuddy</b>：完整适配 —— CODEBUDDY.md + .codebuddy/ 下的技能与 /study:* 命令开箱即用。</li>
            <li><b>Kimi CLI</b>：原生读 AGENTS.md，复制「通用 prompt」发给它即可。</li>
            <li><b>Trae</b>：设置 → Rules → 勾选导入 AGENTS.md；.trae/rules/deep-study.mdc 已就位，之后同样用通用 prompt。</li>
            <li>所有 agent 共享同一份 data/ 文件契约 —— 换工具不换数据。</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
