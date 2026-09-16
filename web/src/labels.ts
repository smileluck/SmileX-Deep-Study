// 会话类型与掌握度证据 kind 的展示标签 —— 仪表盘/会话/掌握度三处共用，新增类型只改这里。
export const sessionTypeLabel: Record<string, string> = {
  tutor: '导师',
  feynman: '费曼',
  quiz: '自测',
  diagnose: '诊断',
  import: '导入',
  plan: '规划',
  merge: '合并',
}

export const sessionTypeBadge: Record<string, string> = {
  tutor: 'badge-primary',
  feynman: 'badge-secondary',
  quiz: 'badge-accent',
  diagnose: 'badge-warning',
  import: 'badge-info',
  plan: 'badge-success',
  merge: 'badge-neutral',
}

export const evidenceKindLabel: Record<string, string> = {
  quiz: '自测',
  review: '复习',
  feynman: '费曼',
  diagnose: '诊断',
  import: '导入',
  merge: '合并',
}
