import { useEffect, useState } from 'react'
import type { FormEvent, ReactNode } from 'react'
import { getMe, login, type User } from './api/auth'
import { formatCoursePrice, getCourses, type Course } from './api/courses'
import './App.css'

const routes = ['/', '/courses', '/about', '/login', '/profile', '/register']
const TOKEN_KEY = 'cringearium_token'

function navigate(path: string) {
  window.history.pushState({}, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

function usePath() {
  const [path, setPath] = useState(window.location.pathname)

  useEffect(() => {
    const handlePopState = () => setPath(window.location.pathname)
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  return path
}

function Link({ href, children, className = '' }: { href: string; children: ReactNode; className?: string }) {
  return (
    <a href={href} className={className} onClick={(event) => {
      if (href.startsWith('/')) {
        event.preventDefault()
        navigate(href)
      }
    }}>
      {children}
    </a>
  )
}

function Header({ user, onLogout }: { user: User | null; onLogout: () => void }) {
  return (
    <header className="header">
      <div className="container nav-container">
        <Link href="/" className="brand"><span className="brand-icon">⌂</span><span>Cringearium</span></Link>
        <nav className="nav-links"><Link href="/courses">📚 Каталог курсов</Link><Link href="/about">ℹ О нас</Link></nav>
        <nav className="nav-actions">
          {user ? <><Link href="/profile" className="profile-link">◉ {user.username}</Link><button type="button" className="nav-button" onClick={onLogout}>Выйти</button></> : <Link href="/login" className="login-link">Войти</Link>}
        </nav>
      </div>
    </header>
  )
}

function Footer() {
  return <footer className="footer"><div className="container footer-grid"><div><div className="footer-brand">Cringearium</div><p>Образовательная платформа, где учиться немного проще.</p></div><div><h3>Навигация</h3><Link href="/courses">Каталог курсов</Link><Link href="/about">О нас</Link><Link href="/login">Вход</Link></div><div><h3>Проект</h3><a href="https://github.com/Necromemeser/Cringearium-go" target="_blank" rel="noreferrer">GitHub</a><span>© 2026 Cringearium</span></div></div></footer>
}

function CourseGrid({ items }: { items: Course[] }) {
  return <div className="course-grid">{items.map((course) => <CourseCard key={course.id} course={course} />)}</div>
}

function CourseCard({ course }: { course: Course }) {
  return <article className="course-card"><div className="course-cover"><span>{course.theme ? '∑' : '📚'}</span></div><div className="course-body"><span className="course-theme">{course.theme || 'Курс'}</span><h3>{course.title}</h3><p>{course.description || 'Описание курса пока не добавлено.'}</p><div className="course-footer"><strong>{formatCoursePrice(course.price)}</strong><Link href={`/courses/${course.id}`} className="small-button">Подробнее</Link></div></div></article>
}

function HomePage({ courses }: { courses: Course[] }) {
  return <><section className="hero"><div className="container hero-content"><div className="hero-copy"><span className="eyebrow">ОБРАЗОВАТЕЛЬНАЯ ПЛАТФОРМА</span><h1>Учиться можно<br /><span>без лишнего кринжа.</span></h1><p>Курсы, практика и персональная помощь в одном месте. Разбираемся со сложным нормальным человеческим языком.</p><div className="hero-actions"><Link href="/courses" className="button button-primary">Смотреть курсы</Link><Link href="/about" className="button button-secondary">Узнать больше</Link></div></div><div className="hero-card"><div className="hero-card-icon">✦</div><h2>Курс дня</h2><p>{courses[0]?.title || 'Курсы скоро появятся'}</p><span>{courses[0]?.theme || 'Новые материалы'}</span><Link href="/courses" className="text-link">Перейти к курсу →</Link></div></div></section><section className="section"><div className="container"><div className="section-heading"><div><span className="eyebrow">ВОЗМОЖНОСТИ</span><h2>Всё необходимое для обучения</h2></div></div><div className="feature-grid"><article className="feature-card"><div className="feature-icon">📖</div><h3>Понятные курсы</h3><p>Материалы разбиты на небольшие последовательные шаги.</p></article><article className="feature-card"><div className="feature-icon">✓</div><h3>Практика</h3><p>Тесты помогают сразу проверить, что материал действительно усвоен.</p></article><article className="feature-card"><div className="feature-icon">✦</div><h3>Умный помощник</h3><p>Персональная помощь появится в следующих этапах развития платформы.</p></article></div></div></section><section className="section section-muted"><div className="container"><div className="section-heading inline-heading"><div><span className="eyebrow">КУРСЫ</span><h2>Популярные направления</h2></div><Link href="/courses" className="text-link">Весь каталог →</Link></div><CourseGrid items={courses} /></div></section></>
}

function CoursesPage() {
  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [query, setQuery] = useState('')

  useEffect(() => {
    getCourses()
      .then((data) => setCourses(data.filter((course) => course.status === 'published')))
      .catch((err) => setError(err instanceof Error ? err.message : 'Не удалось загрузить курсы'))
      .finally(() => setLoading(false))
  }, [])

  const filteredCourses = courses.filter((course) => {
    const value = query.trim().toLowerCase()
    if (!value) return true
    return `${course.title} ${course.theme} ${course.description}`.toLowerCase().includes(value)
  })

  return <main className="page"><div className="container"><div className="page-heading"><span className="eyebrow">ОБУЧЕНИЕ</span><h1>Каталог курсов</h1><p>Выбирай направление и двигайся вперёд в удобном темпе.</p></div><form className="catalog-toolbar" onSubmit={(event) => { event.preventDefault(); setQuery(search) }}><input aria-label="Поиск курсов" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Поиск по названию курса..." /><button type="submit">Найти</button></form>{loading && <div className="loading-screen">Загружаем курсы...</div>}{error && <div className="form-error">{error}</div>}{!loading && !error && filteredCourses.length === 0 && <div className="empty-state"><div className="empty-icon">📚</div><h3>{query ? 'Ничего не найдено' : 'Курсов пока нет'}</h3><p>{query ? 'Попробуй изменить поисковый запрос.' : 'Опубликованные курсы появятся здесь.'}</p></div>}{!loading && !error && filteredCourses.length > 0 && <CourseGrid items={filteredCourses} />}</div></main>
}

function LoginPage({ onLogin }: { onLogin: (email: string, password: string) => Promise<void> }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); setError(''); setLoading(true)
    try { await onLogin(email, password) } catch (err) { setError(err instanceof Error ? err.message : 'Не удалось войти') } finally { setLoading(false) }
  }

  return <main className="auth-page"><div className="auth-card"><div className="auth-icon">◉</div><span className="eyebrow">С возвращением</span><h1>Вход в Cringearium</h1><p className="auth-description">Войди, чтобы продолжить обучение.</p><form onSubmit={handleSubmit} className="auth-form"><label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder="you@example.com" required /></label><label>Пароль<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="••••••••" required /></label>{error && <div className="form-error">{error}</div>}<button className="button button-primary" type="submit" disabled={loading}>{loading ? 'Входим...' : 'Войти'}</button></form><p className="auth-bottom">Нет аккаунта? <Link href="/register">Зарегистрироваться</Link></p></div></main>
}

function RegisterPage({ onRegister }: { onRegister: (username: string, email: string, password: string) => Promise<void> }) {
  const [username, setUsername] = useState(''); const [email, setEmail] = useState(''); const [password, setPassword] = useState(''); const [error, setError] = useState(''); const [loading, setLoading] = useState(false)
  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); setError(''); setLoading(true); try { await onRegister(username, email, password) } catch (err) { setError(err instanceof Error ? err.message : 'Не удалось зарегистрироваться') } finally { setLoading(false) } }
  return <main className="auth-page"><div className="auth-card"><div className="auth-icon">✦</div><span className="eyebrow">Новый аккаунт</span><h1>Регистрация</h1><p className="auth-description">Создай аккаунт и начни обучение.</p><form onSubmit={handleSubmit} className="auth-form"><label>Имя пользователя<input value={username} onChange={(event) => setUsername(event.target.value)} placeholder="username" required /></label><label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder="you@example.com" required /></label><label>Пароль<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="Минимум 8 символов" minLength={8} required /></label>{error && <div className="form-error">{error}</div>}<button className="button button-primary" type="submit" disabled={loading}>{loading ? 'Создаём...' : 'Зарегистрироваться'}</button></form><p className="auth-bottom">Уже есть аккаунт? <Link href="/login">Войти</Link></p></div></main>
}

function ProfilePage({ user }: { user: User | null }) {
  if (!user) return <main className="page"><div className="container empty-page"><h1>Нужна авторизация</h1><p>Войди в аккаунт, чтобы открыть личный кабинет.</p><Link href="/login" className="button button-primary">Войти</Link></div></main>
  return <main className="page"><div className="container"><div className="page-heading"><span className="eyebrow">ПРОФИЛЬ</span><h1>Личный кабинет</h1><p>Твоя учётная запись и учебный прогресс.</p></div><div className="profile-layout"><section className="profile-card profile-summary"><div className="avatar">{user.username.charAt(0).toUpperCase()}</div><div><h2>{user.username}</h2><p>{user.email}</p><p>Роль: {user.role}</p></div></section><section className="profile-card"><div className="card-heading"><div><span className="eyebrow">ОБУЧЕНИЕ</span><h2>Мои курсы</h2></div></div><div className="empty-state"><div className="empty-icon">📚</div><h3>Пока нет курсов</h3><p>Загляни в каталог и выбери что-нибудь для обучения.</p><Link href="/courses" className="button button-primary">Открыть каталог</Link></div></section></div></div></main>
}

function AboutPage() { return <main className="page"><div className="container narrow"><div className="page-heading"><span className="eyebrow">CRINGEARium</span><h1>О проекте</h1><p>Образовательная платформа, которую мы собираем с нуля на современном стеке.</p></div><div className="about-card"><h2>Зачем это всё?</h2><p>Cringearium задуман как место, где учебный материал не приходится пробивать головой через бесконечные стены текста. Курсы состоят из небольших страниц, практических заданий и тестов.</p><p>Платформа развивается как отдельный проект: Go-бэкенд, React-фронтенд и набор независимых сервисов. В будущем здесь появятся AI-помощник и персонализированные тесты.</p></div><div className="about-stats"><div><strong>Go</strong><span>backend</span></div><div><strong>React</strong><span>frontend</span></div><div><strong>5+</strong><span>сервисов</span></div></div></div></main> }
function NotFound() { return <main className="page"><div className="container empty-page"><span className="eyebrow">404</span><h1>Страница не найдена</h1><p>Похоже, такой страницы пока нет.</p><Link href="/" className="button button-primary">На главную</Link></div></main> }

function App() {
  const path = usePath(); const [user, setUser] = useState<User | null>(null); const [authLoading, setAuthLoading] = useState(true); const [homeCourses, setHomeCourses] = useState<Course[]>([])
  useEffect(() => { const token = localStorage.getItem(TOKEN_KEY); if (!token) { setAuthLoading(false); return } getMe(token).then(setUser).catch(() => localStorage.removeItem(TOKEN_KEY)).finally(() => setAuthLoading(false)) }, [])
  useEffect(() => { getCourses().then((data) => setHomeCourses(data.filter((course) => course.status === 'published'))).catch(() => setHomeCourses([])) }, [])
  const handleLogin = async (email: string, password: string) => { const token = await login(email, password); localStorage.setItem(TOKEN_KEY, token); const currentUser = await getMe(token); setUser(currentUser); navigate('/profile') }
  const handleRegister = async (username: string, email: string, password: string) => { await import('./api/auth').then(({ register }) => register(username, email, password)); await handleLogin(email, password) }
  const handleLogout = () => { localStorage.removeItem(TOKEN_KEY); setUser(null); navigate('/') }
  if (authLoading) return <div className="loading-screen">Загрузка...</div>
  const page = routes.includes(path) || path.startsWith('/courses/') ? path : '/404'
  return <div className="app-shell"><Header user={user} onLogout={handleLogout} />{page === '/' && <HomePage courses={homeCourses} />}{page === '/courses' && <CoursesPage />}{page === '/login' && <LoginPage onLogin={handleLogin} />}{page === '/register' && <RegisterPage onRegister={handleRegister} />}{page === '/profile' && <ProfilePage user={user} />}{page === '/about' && <AboutPage />}{page === '/404' && <NotFound />}<Footer /></div>
}

export default App
