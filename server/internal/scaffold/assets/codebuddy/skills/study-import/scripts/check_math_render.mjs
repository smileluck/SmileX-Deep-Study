// 笔记公式渲染体检（复刻 Web 端真实渲染链路，用于导入后验收）
//
//   node check_math_render.mjs <topic> [notesDir] [repoRoot]
//
// 复刻链路：normalizeDisplayMath（与 web/src/components/MarkdownInner.tsx 逐字一致）
//          → remark-parse + remark-gfm + remark-math → mdast
//          → 对每个 inlineMath / math 节点跑 katex.renderToString（参数与 rehype-katex 一致）
//
// 报告三类问题：
//   ① KaTeX 渲染失败（页面上会显示成红字 katex-error）——如 \hfill 不存在、一个公式里出现两个 \tag
//   ② 同段 ≥2 个未转义字面 $（会被 remark-math 配对误吞成公式 → 内容损坏）。
//      单个不成对的 $（如货币符号，正确写法 \$）为良性：渲染为字面 $，仅计数不判失败
//   ③ 表格行「格子数多于表头」（GFM 会直接丢弃多出的格子 → 内容丢失）
//
// 依赖：项目 web/ 已 pnpm install（脚本自动从 .pnpm 里挑版本，无需写死版本号）
import fs from 'node:fs'
import path from 'node:path'
import { createRequire } from 'node:module'
import { pathToFileURL } from 'node:url'

const TOPIC = process.argv[2] || 'signals-and-systems'
const NOTES = path.resolve(process.argv[3] || 'data/notes')
// 默认认为脚本位于 <root>/.workbuddy/ 或 <root>/.agents/skills/study-import/scripts/
const REPO = path.resolve(process.argv[4] || process.cwd())
const PNPM = path.join(REPO, 'web/node_modules/.pnpm')

if (!fs.existsSync(PNPM)) {
  console.error('找不到 ' + PNPM + '，请先在 web/ 执行 pnpm install')
  process.exit(2)
}

const require = createRequire(import.meta.url)

function cmpVer(a, b) {
  const pa = a.split(/[.\-+]/), pb = b.split(/[.\-+]/)
  for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
    const x = pa[i] ?? '0', y = pb[i] ?? '0'
    const nx = /^\d+$/.test(x), ny = /^\d+$/.test(y)
    if (nx && ny) { const d = Number(x) - Number(y); if (d) return d }
    else if (x !== y) return x < y ? -1 : 1
  }
  return 0
}

/** 在 .pnpm 里挑 <name>@<最高版本> */
function pick(name) {
  const dirs = fs.readdirSync(PNPM).filter((d) => d.startsWith(name + '@'))
  if (!dirs.length) throw new Error('.pnpm 里找不到 ' + name)
  dirs.sort((a, b) => cmpVer(a.slice(name.length + 1), b.slice(name.length + 1)))
  return path.join(PNPM, dirs[dirs.length - 1], 'node_modules', name)
}

async function loadEsm(pkg) {
  const dir = pick(pkg)
  for (const entry of ['index.js', 'index.mjs']) {
    const p = path.join(dir, entry)
    if (fs.existsSync(p)) return (await import(pathToFileURL(p).href)).default
  }
  throw new Error(pkg + ' 入口未找到')
}

/** katex：优先用 rehype-katex 实际解析到的那一份，保证与页面一致 */
function loadKatex() {
  const cands = fs.readdirSync(PNPM).filter((d) => d.startsWith('rehype-katex@'))
  for (const c of cands) {
    const link = path.join(PNPM, c, 'node_modules', 'katex')
    if (fs.existsSync(link)) {
      const real = fs.realpathSync(link)
      const js = path.join(real, 'dist/katex.js')
      if (fs.existsSync(js)) return require(js)
    }
  }
  const kl = fs.readdirSync(PNPM).filter((d) => d.startsWith('katex@'))
  kl.sort((a, b) => cmpVer(a.slice(6), b.slice(6)))
  return require(path.join(PNPM, kl[kl.length - 1], 'node_modules', 'katex', 'dist/katex.js'))
}

const { unified } = await import(pathToFileURL(path.join(pick('unified'), 'index.js')).href)
const remarkParse = await loadEsm('remark-parse')
const remarkGfm = await loadEsm('remark-gfm')
const remarkMath = await loadEsm('remark-math')
const katex = loadKatex()

// —— 与 web/src/components/MarkdownInner.tsx 中的实现逐字一致 ——
function normalizeDisplayMath(src) {
  let inFence = false
  let inMath = false
  const out = []
  for (const line of src.split('\n')) {
    if (/^\s*```/.test(line)) { inFence = !inFence; out.push(line); continue }
    if (inFence) { out.push(line); continue }
    if (inMath) {
      const close = line.match(/^(.*?)\$\$\s*$/)
      if (close) {
        if (close[1].trim()) out.push(close[1])
        out.push('$$')
        inMath = false
      } else out.push(line)
      continue
    }
    const single = line.match(/^\s*\$\$([^$]+)\$\$\s*$/)
    if (single) { out.push('$$', single[1].trim(), '$$'); continue }
    const open = line.match(/^\s*\$\$(.+)$/)
    if (open && !open[1].includes('$$')) { out.push('$$', open[1]); inMath = true; continue }
    out.push(line)
  }
  return out.join('\n')
}

const FM = /^---\r?\n[\s\S]*?\r?\n---\r?\n?/
const processor = unified().use(remarkParse).use(remarkGfm).use(remarkMath)

function walk(node, fn, parent = null) {
  fn(node, parent)
  if (node.children) for (const c of node.children) walk(c, fn, node)
}

const only = process.argv.includes('--verbose')
let total = 0, errTotal = 0, riskyTotal = 0, benignTotal = 0, badCellTotal = 0
const report = {}

for (const fn of fs.readdirSync(NOTES).filter((f) => f.endsWith('.md')).sort()) {
  const raw = fs.readFileSync(path.join(NOTES, fn), 'utf8')
  const m = FM.exec(raw)
  // topic 值可能带引号也可能不带（不同批次写法不一），两种都要匹配
  const topicRe = new RegExp('^topic:\\s*("?' + TOPIC.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '"?)\\s*$', 'm')
  if (!m || !topicRe.test(m[0])) continue
  const nid = fn.replace(/\.md$/, '')
  total++
  let tree
  try {
    tree = processor.parse(normalizeDisplayMath(raw.slice(m[0].length)))
  } catch (e) {
    report[nid] = { parseError: String(e) }
    continue
  }
  const errors = []
  let badCells = 0
  const dollars = [] // { parent, value, cnt } —— 按父节点（段落级）聚合判定配对风险
  walk(tree, (node, parent) => {
    if (node.type === 'table') {
      const ncol = node.children[0] ? node.children[0].children.length : 0
      for (const row of node.children.slice(1)) if (row.children.length > ncol) badCells++
    }
    if (node.type === 'inlineMath' || node.type === 'math') {
      const displayMode = node.type === 'math'
      try {
        katex.renderToString(node.value, { strict: false, displayMode, throwOnError: true })
      } catch (e) {
        errors.push({ displayMode, src: node.value.slice(0, 160), msg: String(e.message || e) })
      }
    }
    if (node.type === 'text' && node.value.includes('$')) {
      const cnt = (node.value.match(/\$/g) || []).length
      dollars.push({ parent, value: node.value.replace(/\n/g, ' ⏎ ').slice(0, 110), cnt })
    }
  })
  // 同一父节点（段落/标题/表格单元格）下 ≥2 个 $ 会互相配对成 math → 风险；单个为良性
  const byParent = new Map()
  for (const d of dollars) byParent.set(d.parent, (byParent.get(d.parent) || 0) + d.cnt)
  let leftover = 0, benign = 0
  const samples = []
  for (const [p, c] of byParent) {
    if (c >= 2) {
      leftover += c
      for (const d of dollars) if (d.parent === p && samples.length < 3) samples.push(d.value)
    } else benign += c
  }
  errTotal += errors.length
  riskyTotal += leftover
  benignTotal += benign
  badCellTotal += badCells
  if (errors.length || leftover || badCells) report[nid] = { errors, leftover, benign, badCells, samples }
}

console.log('主题:', TOPIC, ' 笔记数:', total)
console.log('KaTeX 渲染失败公式数:', errTotal)
console.log('同段多$配对风险个数:', riskyTotal, '（良性单个$如货币符号:', benignTotal, '）')
console.log('表格「格子多于表头」行数:', badCellTotal)
console.log('受影响笔记数:', Object.keys(report).length)
for (const [nid, r] of Object.entries(report)) {
  if (!only && Object.keys(report).length > 15) break
  console.log('--- ' + nid + '  leftover$=' + r.leftover + ' 表格坏行=' + r.badCells)
  if (r.parseError) console.log('    PARSE ERROR: ' + r.parseError)
  for (const s of r.samples || []) console.log('    残留$: ' + s)
  for (const e of (r.errors || []).slice(0, 4)) {
    console.log('    [' + (e.displayMode ? 'block' : 'inline') + '] ' + e.msg)
    console.log('        src: ' + e.src.replace(/\n/g, ' ⏎ '))
  }
}
process.exit(errTotal || riskyTotal || badCellTotal ? 1 : 0)
