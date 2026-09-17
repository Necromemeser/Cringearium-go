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
  return (await response.json()) as CourseDetails
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
  return data.completed_page_ids
}

export async function completePage(id: number, token: string): Promise<void> {
  const response = await fetch(`/api/pages/${id}/complete`, { method: 'POST', headers: { Authorization: `Bearer ${token}` } })
  if (!response.ok) return parseError(response, 'Не удалось сохранить прогресс')
}

export function formatCoursePrice(price: number): string {
  if (price === 0) return 'Бесплатно'
  return `${new Intl.NumberFormat('ru-RU').format(price)} ₽`
}
