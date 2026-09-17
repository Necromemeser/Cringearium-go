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

export async function getCourses(): Promise<Course[]> {
  const response = await fetch('/api/courses')

  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || 'Не удалось загрузить курсы')
  }

  return (await response.json()) as Course[]
}

export function formatCoursePrice(price: number): string {
  if (price === 0) {
    return 'Бесплатно'
  }

  return `${new Intl.NumberFormat('ru-RU').format(price / 100)} ₽`
}
