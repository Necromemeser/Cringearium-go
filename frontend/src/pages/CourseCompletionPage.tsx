import { useEffect, useState } from 'react'
import type { User } from '../api/auth'
import { getCourse, getCourseProgress, type CourseDetails } from '../api/courses'
import { TOKEN_KEY } from '../constants/auth'
import Link from '../components/common/Link'

export function CourseCompletionContent({ username, courseTitle, courseId, pageCount }: { username: string; courseTitle: string; courseId: number; pageCount: number }) {
  return <section className="course-completion">
    <span className="course-completion-icon" aria-hidden="true">✓</span>
    <span className="eyebrow">КУРС ЗАВЕРШЁН</span>
    <h1>Спасибо, что прошёл наш курс, {username}!</h1>
    <p>Ты завершил курс «{courseTitle}». Пройдено страниц: {pageCount}. Результаты тестов сохранены.</p>
    <p>Можно вернуться к материалам и повторить сложные темы или выбрать следующий курс.</p>
    <div className="course-completion-actions">
      <Link href="/courses" className="button button-primary">Выбрать следующий курс</Link>
      <Link href={`/courses/${courseId}`} className="button button-secondary">Вернуться к материалам</Link>
    </div>
  </section>
}

export default function CourseCompletionPage({ courseId, user }: { courseId: number; user: User | null }) {
  const [data, setData] = useState<{ course: CourseDetails; progress: number[] } | null>(null)
  const [error, setError] = useState('')
  const token = localStorage.getItem(TOKEN_KEY)
  useEffect(() => {
    let cancelled = false
    setData(null)
    setError('')
    if (!user || !token) return
    void Promise.all([getCourse(courseId), getCourseProgress(courseId, token)])
      .then(([course, progress]) => { if (!cancelled) setData({ course, progress }) })
      .catch(() => { if (!cancelled) setError('Не удалось загрузить результаты курса. Попробуй обновить страницу.') })
    return () => { cancelled = true }
  }, [courseId, user, token])
  if (!user || !token) return <main className="page"><div className="container empty-page"><h1>Войди в аккаунт</h1><p>Результаты курса доступны после входа.</p><Link href="/login" className="button button-primary">Войти</Link></div></main>
  if (error) return <main className="page"><div className="container empty-page"><p role="alert">{error}</p><Link href={`/courses/${courseId}`}>Вернуться к курсу</Link></div></main>
  if (!data) return <main className="page"><div className="container loading-screen">Загружаем результаты курса...</div></main>
  const pages = data.course.sections.flatMap(section => section.pages)
  const completed = new Set(data.progress)
  const count = pages.filter(page => completed.has(page.id)).length
  const finished = pages.length > 0 && count === pages.length
  return <main className="page"><div className="container">
    {finished ? <CourseCompletionContent username={user.username} courseTitle={data.course.title} courseId={courseId} pageCount={pages.length} /> : (
      <section className="course-completion"><h1>До встречи на финише!</h1><p>Пока пройдено {count} из {pages.length} страниц курса. Заверши оставшиеся материалы и тесты.</p><Link href={`/courses/${courseId}`} className="button button-primary">Продолжить курс</Link></section>
    )}
  </div></main>
}
