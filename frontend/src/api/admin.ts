import type { User } from './auth'

export type AdminCourse = {
  id: number
  title: string
  theme: string
  description: string
  price: number
  author_id?: number
  status: string
  enrolled_users: number
  section_count: number
  page_count: number
}

export type AssessmentAnswer = {
  question_id: number
  answer_id: number
  correct_answer_id: number
  is_correct: boolean
}

export type AdminAssessmentResult = {
  id: number
  user_id: number
  course_id: number
  course_title: string
  type: 'pretest' | 'posttest'
  score: number
  answers: AssessmentAnswer[]
  completed_at: string
}

export type AdminAISession = {
  id: string
  user_id: number
  course_id: number
  topic_page_id?: number
  status: string
  current_round: number
  question_count: number
  created_at: string
  completed_at?: string
  summary?: string
}

export type AdminChatStats = {
  conversations: number
  messages: number
  user_messages: number
  ai_messages: number
  active_users: number
}

export type AdminAIStats = {
  total_sessions: number
  completed_sessions: number
  failed_sessions: number
  questions_generated: number
  answers_submitted: number
  accuracy_percent: number
  recent_sessions: AdminAISession[]
}

const API_BASE = '/api'

async function request<T>(path: string, token: string): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
  })

  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || `Request failed with status ${response.status}`)
  }

  return (await response.json()) as T
}

export function getAdminUsers(token: string): Promise<User[]> {
  return request<User[]>('/admin/users', token)
}

export function getAdminCourses(token: string): Promise<AdminCourse[]> {
  return request<AdminCourse[]>('/admin/courses', token)
}

export function getAdminAIStats(token: string): Promise<AdminAIStats> {
  return request<AdminAIStats>('/admin/ai-tests', token)
}


export function getAdminChatStats(token: string): Promise<AdminChatStats> {
  return request<AdminChatStats>('/admin/ai-chat', token)
}

export function getAdminAssessmentResults(token: string): Promise<AdminAssessmentResult[]> {
  return request<AdminAssessmentResult[]>('/admin/assessment-results', token)
}
