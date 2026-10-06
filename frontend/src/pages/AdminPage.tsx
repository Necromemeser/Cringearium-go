import { useEffect, useState } from 'react'
import type { User } from '../api/auth'
import { getAdminAIStats, getAdminChatStats, getAdminCourses, getAdminUsers, type AdminAIStats, type AdminChatStats, type AdminCourse } from '../api/admin'
import { TOKEN_KEY } from '../constants/auth'
import './AdminPage.css'

type Props = {
  user: User | null
}

export default function AdminPage({ user }: Props) {
  const [users, setUsers] = useState<User[]>([])
  const [courses, setCourses] = useState<AdminCourse[]>([])
  const [aiStats, setAIStats] = useState<AdminAIStats | null>(null)
  const [chatStats, setChatStats] = useState<AdminChatStats | null>(null)
  const [chatStatsError, setChatStatsError] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const token = localStorage.getItem(TOKEN_KEY)

  useEffect(() => {
    if (!token || user?.role !== 'admin') {
      setLoading(false)
      return
    }

    Promise.all([
      getAdminUsers(token),
      getAdminCourses(token),
      getAdminAIStats(token),
    ])
      .then(([loadedUsers, loadedCourses, loadedAIStats]) => {
        setUsers(loadedUsers)
        setCourses(loadedCourses)
        setAIStats(loadedAIStats)
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Не удалось загрузить данные панели'))
      .finally(() => setLoading(false))

    getAdminChatStats(token)
      .then(setChatStats)
      .catch((e) => setChatStatsError(e instanceof Error ? e.message : 'Не удалось загрузить статистику AI-чата'))
  }, [token, user?.role])

  if (!user || user.role !== 'admin') {
    return (
      <main className="page">
        <div className="container admin-denied">
          <span className="eyebrow">АДМИНИСТРАТОР</span>
          <h1>Доступ запрещён</h1>
          <p>Панель администратора доступна только пользователям с ролью admin.</p>
        </div>
      </main>
    )
  }

  if (loading) {
    return <main className="page"><div className="container loading-screen">Загружаем панель...</div></main>
  }

  if (error) {
    return (
      <main className="page">
        <div className="container">
          <div className="form-error">{error}</div>
        </div>
      </main>
    )
  }

  const publishedCourses = courses.filter((course) => course.status === 'published').length
  const totalEnrollments = courses.reduce((sum, course) => sum + course.enrolled_users, 0)

  return (
    <main className="page">
      <div className="container admin-page">
        <div className="page-heading">
          <span className="eyebrow">АДМИНИСТРАТОР</span>
          <h1>Панель управления</h1>
          <p>Пользователи, курсы и статистика нейронного контента в одном месте.</p>
        </div>

        <section className="admin-stat-grid">
          <StatCard title="Пользователи" value={users.length} caption="зарегистрировано" />
          <StatCard title="Курсы" value={courses.length} caption={`${publishedCourses} опубликовано`} />
          <StatCard title="Записи на курсы" value={totalEnrollments} caption="выдано доступов" />
          <StatCard title="AI-тесты" value={aiStats?.total_sessions ?? 0} caption="создано сессий" />
          <StatCard title="Вопросы AI" value={aiStats?.questions_generated ?? 0} caption="сгенерировано" />
          <StatCard title="Точность" value={`${(aiStats?.accuracy_percent ?? 0).toFixed(1)}%`} caption="по всем ответам" />
        </section>

        <section className="admin-card">
          <div className="admin-card-heading">
            <div>
              <span className="eyebrow">ПОЛЬЗОВАТЕЛИ</span>
              <h2>Последние аккаунты</h2>
            </div>
            <span className="admin-count">{users.length}</span>
          </div>
          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead><tr><th>ID</th><th>Имя</th><th>Email</th><th>Роль</th></tr></thead>
              <tbody>
                {users.slice(-10).reverse().map((item) => (
                  <tr key={item.id}>
                    <td>{item.id}</td>
                    <td>{item.username}</td>
                    <td>{item.email}</td>
                    <td><RoleBadge role={item.role} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <section className="admin-card">
          <div className="admin-card-heading">
            <div>
              <span className="eyebrow">КУРСЫ</span>
              <h2>Состояние учебного контента</h2>
            </div>
            <span className="admin-count">{courses.length}</span>
          </div>
          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead><tr><th>Курс</th><th>Статус</th><th>Студенты</th><th>Разделы</th><th>Страницы</th></tr></thead>
              <tbody>
                {courses.map((course) => (
                  <tr key={course.id}>
                    <td><strong>{course.title}</strong><small>{course.theme}</small></td>
                    <td><StatusBadge status={course.status} /></td>
                    <td>{course.enrolled_users}</td>
                    <td>{course.section_count}</td>
                    <td>{course.page_count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <section className="admin-card">
          <div className="admin-card-heading">
            <div>
              <span className="eyebrow">AI-ЧАТ</span>
              <h2>Разговоры с Кринжиком</h2>
            </div>
            <span className="admin-count">{chatStats?.conversations ?? 0}</span>
          </div>
          {chatStatsError ? (
            <div className="admin-section-error">Не удалось загрузить статистику AI-чата: {chatStatsError}</div>
          ) : (
            <>
              <div className="admin-ai-grid">
                <div><strong>{chatStats?.conversations ?? 0}</strong><span>разговоров</span></div>
                <div><strong>{chatStats?.active_users ?? 0}</strong><span>активных пользователей</span></div>
                <div><strong>{chatStats?.messages ?? 0}</strong><span>сообщений</span></div>
              </div>
              <div className="admin-ai-summary admin-chat-summary">
                <span>сообщений пользователей: {chatStats?.user_messages ?? 0}</span>
                <span>ответов AI: {chatStats?.ai_messages ?? 0}</span>
              </div>
            </>
          )}
        </section>

        <section className="admin-card">
          <div className="admin-card-heading">
            <div>
              <span className="eyebrow">НЕЙРОННЫЙ КОНТЕНТ</span>
              <h2>AI-тесты и результаты</h2>
            </div>
            <div className="admin-ai-summary">
              <span>завершено {aiStats?.completed_sessions ?? 0}</span>
              <span>ошибок {aiStats?.failed_sessions ?? 0}</span>
            </div>
          </div>

          <div className="admin-ai-grid">
            <div><strong>{aiStats?.questions_generated ?? 0}</strong><span>вопросов</span></div>
            <div><strong>{aiStats?.answers_submitted ?? 0}</strong><span>ответов</span></div>
            <div><strong>{(aiStats?.accuracy_percent ?? 0).toFixed(1)}%</strong><span>точность</span></div>
          </div>

          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead><tr><th>Сессия</th><th>Пользователь</th><th>Курс</th><th>Статус</th><th>Раунды</th><th>Результат AI</th></tr></thead>
              <tbody>
                {aiStats?.recent_sessions.map((session) => (
                  <tr key={session.id}>
                    <td className="admin-mono">{session.id.slice(0, 8)}</td>
                    <td>{session.user_id}</td>
                    <td>{session.course_id}</td>
                    <td><StatusBadge status={session.status} /></td>
                    <td>{session.current_round}/2</td>
                    <td>{session.summary || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </main>
  )
}

function StatCard({ title, value, caption }: { title: string; value: string | number; caption: string }) {
  return <article className="admin-stat-card"><span>{title}</span><strong>{value}</strong><small>{caption}</small></article>
}

function RoleBadge({ role }: { role: string }) {
  return <span className={`admin-badge admin-badge-${role}`}>{role}</span>
}

function StatusBadge({ status }: { status: string }) {
  return <span className={`admin-badge admin-badge-status-${status}`}>{status}</span>
}
