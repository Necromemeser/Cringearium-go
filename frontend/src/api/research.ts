export type ResearchStudy = {
  id: number; title: string; design: 'observational' | 'randomized'; scope: 'course' | 'trigonometry'
  control_course_id: number; adaptive_course_id: number; pre_hash: string; post_hash: string; content_hash: string
  protocol_notes: string; created_at: string
}
export type ResearchParticipant = {
  participant_id: string; account_id: number; arm: 'control' | 'adaptive'; course_id: number; assigned_at: string
  pre_score: number | null; post_score: number | null; gain: number | null; pre_id: number | null; post_id: number | null
  pre_at: string | null; post_at: string | null; eligible: boolean; flags: string[]; completed_pages: number; total_pages: number
  ai_sessions: number | null; ai_answers: number | null; ai_correct: number | null
}
export type Summary = { n: number; mean: number | null; median: number | null; sd: number | null; ci_low: number | null; ci_high: number | null }
export type ResearchGroup = {
  arm: 'control' | 'adaptive'; assigned: number; with_pre: number; with_post: number; paired: number; finished_course: number
  post_completion: number | null; pre: Summary; post: Summary; gain: Summary
}
export type ResearchReport = {
  study: ResearchStudy; generated_at: string; include_unversioned: boolean; participants: ResearchParticipant[]; groups: ResearchGroup[]
  comparison: { gain_difference: number | null; ci_low: number | null; ci_high: number | null; hedges_g: number | null; completion_difference: number | null }
  model: { available: boolean; reason: string; n: number; df: number; group_coefficient: number | null; pre_coefficient: number | null; intercept: number | null; se: number | null; ci_low: number | null; ci_high: number | null; r2: number | null }
  warnings: string[]; ai_available: boolean; quality_counts: Record<string, number>
}
const base = '/api/admin/research/studies'
async function request<T>(token: string, path: string, body?: unknown): Promise<T> {
  const response = await fetch(base + path, { method: body === undefined ? 'GET' : 'POST', headers: { Authorization: `Bearer ${token}`, ...(body === undefined ? {} : { 'Content-Type': 'application/json' }) }, ...(body === undefined ? {} : { body: JSON.stringify(body) }) })
  if (!response.ok) throw new Error((await response.text()).trim() || 'Не удалось загрузить данные исследования')
  return await response.json() as T
}
export const getStudies = (token: string) => request<ResearchStudy[]>(token, '')
export const createStudy = (token: string, study: { title: string; design: string; scope: string; control_course_id: number; adaptive_course_id: number; protocol_notes: string }) => request<ResearchStudy>(token, '', study)
export const assignParticipants = (token: string, id: number, userIds: number[], arm: string) => request<{ assigned: number }>(token, `/${id}/participants`, { user_ids: userIds, arm })
export const getResearchReport = (token: string, id: number, includeUnversioned: boolean) => request<ResearchReport>(token, `/${id}/report?include_unversioned=${includeUnversioned}`)
export async function downloadResearchCSV(token: string, id: number, dataset: string, includeUnversioned: boolean) {
  const response = await fetch(`${base}/${id}/export?dataset=${dataset}&include_unversioned=${includeUnversioned}`, { headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) throw new Error((await response.text()).trim() || 'Экспорт не выполнен')
  const blob = await response.blob(); const url = URL.createObjectURL(blob)
  const link = document.createElement('a'); link.href = url; link.download = `study-${id}-${dataset}.csv`; link.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
