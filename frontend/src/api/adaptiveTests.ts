export type AdaptiveOption = {
  key: string
  text: string
  position: number
}

export type AdaptiveQuestion = {
  id: number
  position: number
  topic_page_id?: number
  topic_title?: string
  question: string
  difficulty: number
  knowledge_basis: 'course' | 'external_knowledge' | 'mixed'
  explanation?: string
  options: AdaptiveOption[]
}

export type AdaptiveRound = {
  id: number
  round_number: number
  strategy: 'initial' | 'increase_difficulty' | 'targeted_practice'
  status: 'generating' | 'generated' | 'completed' | 'failed'
  title: string
  instructions?: string
  questions: AdaptiveQuestion[]
}

export type AdaptiveTopicFeedback = {
  title: string
  topic_page_id?: number
  reason: string
}

export type AdaptiveNextStep = {
  title: string
  description: string
  topic_page_id?: number
}

export type AdaptiveFeedback = {
  summary: string
  mastered_topics: AdaptiveTopicFeedback[]
  topics_to_review: AdaptiveTopicFeedback[]
  next_steps: AdaptiveNextStep[]
}

export type AdaptiveSession = {
  id: string
  course_id: number
  topic_page_id?: number
  status: 'in_progress' | 'completed' | 'failed'
  current_round: number
  question_count: number
  rounds: AdaptiveRound[]
  feedback?: AdaptiveFeedback
}

async function parseError(response: Response, fallback: string): Promise<never> {
  const message = (await response.text()).trim()
  throw new Error(message || fallback)
}

export async function createAdaptiveSession(courseId: number, topicPageId: number, questionCount: number, token: string): Promise<AdaptiveSession> {
  const response = await fetch('/api/adaptive-tests', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ course_id: courseId, topic_page_id: topicPageId, question_count: questionCount }),
  })
  if (!response.ok) return parseError(response, 'Не удалось создать AI-тест')
  return (await response.json()) as AdaptiveSession
}

export async function submitAdaptiveAnswers(sessionId: string, roundId: number, answers: Record<number, string>, token: string): Promise<AdaptiveSession> {
  const response = await fetch(`/api/adaptive-tests/${sessionId}/answers`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ round_id: roundId, answers: Object.fromEntries(Object.entries(answers).map(([id, key]) => [id, key])) }),
  })
  if (!response.ok) return parseError(response, 'Не удалось отправить ответы')
  return (await response.json()) as AdaptiveSession
}

export async function getAdaptiveSession(sessionId: string, token: string): Promise<AdaptiveSession> {
  const response = await fetch(`/api/adaptive-tests/${sessionId}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!response.ok) return parseError(response, 'Не удалось загрузить AI-тест')
  return (await response.json()) as AdaptiveSession
}
