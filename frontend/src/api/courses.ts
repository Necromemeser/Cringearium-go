export type CoursePage = {
  id: number
  title: string
  type: 'theory' | 'test' | 'ai_test'
  content: string
  position: number
}

export type CourseSection = {
  id: number
  title: string
  description: string
  position: number
  pages: CoursePage[]
}

export type Course = {
  id: number
  title: string
  theme: string
  description: string
  price: number
  image_id?: string
  author_id?: number
  status: 'draft' | 'published' | 'archived'
}

export type CourseDetails = Course & { sections: CourseSection[] }

export type TestAnswer = {
  id: number
  text: string
  position: number
}

export type TestQuestion = {
  id: number
  question: string
  position: number
  answers: TestAnswer[]
}

export type TestSubmissionAnswer = {
  question_id: number
  answer_id: number
}

export type TestAttemptAnswer = TestSubmissionAnswer & {
  is_correct: boolean
}

export type CourseTest = {
  id: number
  page_id: number
  passing_score: number
  questions: TestQuestion[]
  completed: boolean
  attempt_id?: number
  score?: number
  answers?: TestAttemptAnswer[]
}

export type TestResult = {
  attempt_id: number
  score: number
  passed: boolean
  passing_score: number
  answers: TestAttemptAnswer[]
}

type CourseDetailsResponse = CourseDetails | {
  course: Course
  sections: CourseSection[]
}

async function parseError(response: Response, fallback: string): Promise<never> {
  const message = (await response.text()).trim()
  throw new Error(message || fallback)
}

export async function getCourses(): Promise<Course[]> {
  const response = await fetch('/api/courses')
  if (!response.ok) return parseError(response, 'Не удалось загрузить курсы')
  return (await response.json()) as Course[]
}

export async function getCourse(id: number): Promise<CourseDetails> {
  const response = await fetch(`/api/courses/${id}`)
  if (!response.ok) return parseError(response, 'Не удалось загрузить курс')
  const data = (await response.json()) as CourseDetailsResponse
  if ('course' in data) return { ...data.course, sections: data.sections ?? [] }
  return { ...data, sections: data.sections ?? [] }
}

export async function getEnrolledCourses(token: string): Promise<Course[]> {
  const response = await fetch('/api/users/me/courses', { headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) return parseError(response, 'Не удалось загрузить мои курсы')
  return (await response.json()) as Course[]
}

export async function enrollCourse(id: number, token: string): Promise<void> {
  const response = await fetch(`/api/courses/${id}/enroll`, { method: 'POST', headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) return parseError(response, 'Не удалось записаться на курс')
}

export async function getCourseProgress(id: number, token: string): Promise<number[]> {
  const response = await fetch(`/api/courses/${id}/progress`, { headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) return parseError(response, 'Не удалось загрузить прогресс курса')
  const data = (await response.json()) as { completed_page_ids: number[] }
  return data.completed_page_ids ?? []
}

export async function completePage(id: number, token: string): Promise<void> {
  const response = await fetch(`/api/pages/${id}/complete`, { method: 'POST', headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) return parseError(response, 'Не удалось сохранить прогресс')
}

export async function getTest(pageId: number, token: string): Promise<CourseTest> {
  const response = await fetch(`/api/pages/${pageId}/test`, { headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) return parseError(response, 'Не удалось загрузить тест')
  return (await response.json()) as CourseTest
}

export async function submitTest(testId: number, answers: TestSubmissionAnswer[], token: string): Promise<TestResult> {
  const response = await fetch(`/api/tests/${testId}/attempts`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ answers }),
  })
  if (!response.ok) return parseError(response, 'Не удалось отправить тест')
  return (await response.json()) as TestResult
}

export function formatCoursePrice(price: number): string {
  if (price === 0) return 'Бесплатно'
  return `${new Intl.NumberFormat('ru-RU').format(price)} ₽`
}
