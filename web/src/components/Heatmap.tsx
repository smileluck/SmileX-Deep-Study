import type { Stats } from '../api'

// 90 天复习热力图：按周分列的小方块网格。
export default function Heatmap({ data }: { data: Stats['heatmap'] }) {
  if (!data?.length) return null
  const max = Math.max(1, ...data.map((d) => d.count))
  const level = (n: number) => {
    if (!n) return 'bg-base-300/60'
    const r = n / max
    if (r > 0.66) return 'bg-primary'
    if (r > 0.33) return 'bg-primary/70'
    return 'bg-primary/40'
  }
  // 补齐首周前的空白，使每列为一个自然周（周一起）
  const firstDate = new Date(data[0].date + 'T00:00:00')
  const pad = (firstDate.getDay() + 6) % 7
  const cells: (Stats['heatmap'][number] | null)[] = [
    ...Array.from({ length: pad }, () => null),
    ...data,
  ]
  const weeks: (Stats['heatmap'][number] | null)[][] = []
  for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7))
  const months: string[] = []
  let last = ''
  for (const w of weeks) {
    const c = w.find(Boolean)
    if (!c) {
      months.push('')
      continue
    }
    const m = c.date.slice(0, 7)
    months.push(m !== last ? `${parseInt(c.date.slice(5, 7))}月` : '')
    last = m
  }
  return (
    <div className="overflow-x-auto">
      <div className="inline-block">
        <div className="mb-1 flex gap-[3px] pl-[26px] text-[10px] opacity-55">
          {months.map((m, i) => (
            <div key={i} className="w-[13px] whitespace-nowrap">
              {m}
            </div>
          ))}
        </div>
        <div className="flex gap-[3px]">
          <div className="mr-1 flex w-[18px] flex-col gap-[3px] text-[10px] opacity-55">
            <span className="leading-[13px]">一</span>
            <span className="leading-[13px]">四</span>
            <span className="leading-[13px]">日</span>
          </div>
          {weeks.map((w, wi) => (
            <div key={wi} className="flex flex-col gap-[3px]">
              {w.map((c, ci) => (
                <div
                  key={ci}
                  title={c ? `${c.date}：${c.count} 次复习` : ''}
                  className={`h-[13px] w-[13px] rounded-[3px] ${c ? level(c.count) : 'bg-transparent'}`}
                />
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
