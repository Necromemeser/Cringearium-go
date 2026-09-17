import { useEffect, useState } from 'react'
import type { FormEvent, ReactNode } from 'react'
import { getMe, login, type User } from './api/auth'
import { completePage, enrollCourse, formatCoursePrice, getCourse, getCourseProgress, getCourses, getEnrolledCourses, type Course, type CourseDetails } from './api/courses'
import './App.css'

const routes = ['/', '/courses', '/about', '/login', '/profile', '/register']
const TOKEN_KEY = 'cringearium_token'

function navigate(path: string) { window.history.pushState({}, '', path); window.dispatchEvent(new PopStateEvent('popstate')) }
function usePath() { const [path, setPath] = useState(window.location.pathname); useEffect(() => { const handle = () => setPath(window.location.pathname); window.addEventListener('popstate', handle); return () => window.removeEventListener('popstate', handle) }, []); return path }
function Link({ href, children, className = '' }: { href: string; children: ReactNode; className?: string }) { return <a href={href} className={className} onClick={(e) => { if (href.startsWith('/')) { e.preventDefault(); navigate(href) } }}>{children}</a> }

function renderInlineMarkdown(text: string): ReactNode[] {
  const pattern = /(\*\*([^*]+)\*\*|__([^_]+)__|`([^`]+)`|\[([^\]]+)\]\(([^)]+)\)|(\*|_)([^*_]+)\7)/g
  const result: ReactNode[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null
  let key = 0

  while ((match = pattern.exec(text)) !== null) {
    if (match.index > lastIndex) result.push(text.slice(lastIndex, match.index))

    if (match[2] || match[3]) result.push(<strong key={key++}>{match[2] || match[3]}</strong>)
    else if (match[4]) result.push(<code key={key++}>{match[4]}</code>)
    else if (match[5] && match[6]) result.push(<a key={key++} href={match[6]} target="_blank" rel="noreferrer">{match[5]}</a>)
    else if (match[8]) result.push(<em key={key++}>{match[8]}</em>)

    lastIndex = match.index + match[0].length
  }

  if (lastIndex < text.length) result.push(text.slice(lastIndex))
  return result
}

function Markdown({ content }: { content: string }) {
  const lines = content.replace(/\r\n?/g, '\n').split('\n')
  const blocks: ReactNode[] = []
  let paragraph: string[] = []
  let list: string[] = []
  let orderedList = false

  const flushParagraph = () => {
    if (!paragraph.length) return
    blocks.push(<p key={`p-${blocks.length}`}>{renderInlineMarkdown(paragraph.join(' '))}</p>)
    paragraph = []
  }

  const flushList = () => {
    if (!list.length) return
    const items = list.map((item, index) => <li key={index}>{renderInlineMarkdown(item)}</li>)
    blocks.push(orderedList ? <ol key={`ol-${blocks.length}`}>{items}</ol> : <ul key={`ul-${blocks.length}`}>{items}</ul>)
    list = []
  }

  lines.forEach((line) => {
    const trimmed = line.trim()
    if (!trimmed) {
      flushParagraph()
      flushList()
      return
    }

    const heading = /^(#{1,6})\s+(.+)$/.exec(trimmed)
    if (heading) {
      flushParagraph()
      flushList()
      const level = Math.min(heading[1].length, 4)
      const text = renderInlineMarkdown(heading[2])
      if (level === 1) blocks.push(<h3 key={`h-${blocks.length}`}>{text}</h3>)
      else if (level === 2) blocks.push(<h4 key={`h-${blocks.length}`}>{text}</h4>)
      else blocks.push(<h5 key={`h-${blocks.length}`}>{text}</h5>)
      return
    }

    const unordered = /^[-*+]\s+(.+)$/.exec(trimmed)
    const ordered = /^\d+[.)]\s+(.+)$/.exec(trimmed)
    if (unordered || ordered) {
      flushParagraph()
      if (list.length && orderedList !== Boolean(ordered)) flushList()
      orderedList = Boolean(ordered)
      list.push((unordered || ordered)![1])
      return
    }

    if (trimmed.startsWith('> ')) {
      flushParagraph()
      flushList()
      blocks.push(<blockquote key={`q-${blocks.length}`}>{renderInlineMarkdown(trimmed.slice(2))}</blockquote>)
      return
    }

    if (/^---+$/.test(trimmed)) {
      flushParagraph()
      flushList()
      blocks.push(<hr key={`hr-${blocks.length}`} />)
      return
    }

    paragraph.push(trimmed)
  })

  flushParagraph()
  flushList()

  return <div className="markdown-content">{blocks.length ? blocks : <p>Материал этой страницы пока не добавлен.</p>}</div>
}

function Header({ user, onLogout }: { user: User | null; onLogout: () => void }) { return <header className="header"><div className="container nav-container"><Link href="/" className="brand"><span className="brand-icon">⌂</span><span>Cringearium</span></Link><nav className="nav-links"><Link href="/courses">📚 Каталог курсов</Link><Link href="/about">ℹ О нас</Link></nav><nav className="nav-actions">{user ? <><Link href="/profile" className="profile-link">◉ {user.username}</Link><button type="button" className="nav-button" onClick={onLogout}>Выйти</button></> : <Link href="/login" className="login-link">Войти</Link>}</nav></div></header> }
function Footer() { return <footer className="footer"><div className="container footer-grid"><div><div className="footer-brand">Cringearium</div><p>Образовательная платформа, где учиться немного проще.</p></div><div><h3>Навигация</h3><Link href="/courses">Каталог курсов</Link><Link href="/about">О нас</Link><Link href="/login">Вход</Link></div><div><h3>Проект</h3><a href="https://github.com/Necromemeser/Cringearium-go" target="_blank" rel="noreferrer">GitHub</a><span>© 2026 Cringearium</span></div></div></footer> }
function CourseGrid({ items }: { items: Course[] }) { return <div className="course-grid">{items.map((course) => <CourseCard key={course.id} course={course} />)}</div> }
function CourseCard({ course }: { course: Course }) { return <article className="course-card"><div className="course-cover"><span>{course.theme ? '∑' : '📚'}</span></div><div className="course-body"><span className="course-theme">{course.theme || 'Курс'}</span><h3>{course.title}</h3><p>{course.description || 'Описание курса пока не добавлено.'}</p><div className="course-footer"><strong>{formatCoursePrice(course.price)}</strong><Link href={`/courses/${course.id}`} className="small-button">Подробнее</Link></div></div></article> }
function HomePage({ courses }: { courses: Course[] }) { return <><section className="hero"><div className="container hero-content"><div className="hero-copy"><span className="eyebrow">ОБРАЗОВАТЕЛЬНАЯ ПЛАТФОРМА</span><h1>Учиться можно<br /><span>без лишнего кринжа.</span></h1><p>Курсы, практика и персональная помощь в одном месте. Разбираемся со сложным нормальным человеческим языком.</p><div className="hero-actions"><Link href="/courses" className="button button-primary">Смотреть курсы</Link><Link href="/about" className="button button-secondary">Узнать больше</Link></div></div><div className="hero-card"><div className="hero-card-icon">✦</div><h2>Курс дня</h2><p>{courses[0]?.title || 'Курсы скоро появятся'}</p><span>{courses[0]?.theme || 'Новые материалы'}</span><Link href={courses[0] ? `/courses/${courses[0].id}` : '/courses'} className="text-link">Перейти к курсу →</Link></div></div></section><section className="section"><div className="container"><div className="section-heading"><div><span className="eyebrow">ВОЗМОЖНОСТИ</span><h2>Всё необходимое для обучения</h2></div></div><div className="feature-grid"><article className="feature-card"><div className="feature-icon">📖</div><h3>Понятные курсы</h3><p>Материалы разбиты на небольшие последовательные шаги.</p></article><article className="feature-card"><div className="feature-icon">✓</div><h3>Практика</h3><p>Тесты помогают сразу проверить, что материал действительно усвоен.</p></article><article className="feature-card"><div className="feature-icon">✦</div><h3>Умный помощник</h3><p>Персональная помощь появится в следующих этапах развития платформы.</p></article></div></div></section><section className="section section-muted"><div className="container"><div className="section-heading inline-heading"><div><span className="eyebrow">КУРСЫ</span><h2>Популярные направления</h2></div><Link href="/courses" className="text-link">Весь каталог →</Link></div><CourseGrid items={courses} /></div></section></> }

function CoursesPage() { const [courses, setCourses] = useState<Course[]>([]); const [loading, setLoading] = useState(true); const [error, setError] = useState(''); const [search, setSearch] = useState(''); const [query, setQuery] = useState(''); useEffect(() => { getCourses().then((data) => setCourses(data.filter((c) => c.status === 'published'))).catch((e) => setError(e instanceof Error ? e.message : 'Не удалось загрузить курсы')).finally(() => setLoading(false)) }, []); const filtered = courses.filter((c) => { const q = query.trim().toLowerCase(); return !q || `${c.title} ${c.theme} ${c.description}`.toLowerCase().includes(q) }); return <main className="page"><div className="container"><div className="page-heading"><span className="eyebrow">ОБУЧЕНИЕ</span><h1>Каталог курсов</h1><p>Выбирай направление и двигайся вперёд в удобном темпе.</p></div><form className="catalog-toolbar" onSubmit={(e) => { e.preventDefault(); setQuery(search) }}><input aria-label="Поиск курсов" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Поиск по названию курса..." /><button type="submit">Найти</button></form>{loading && <div className="loading-screen">Загружаем курсы...</div>}{error && <div className="form-error">{error}</div>}{!loading && !error && filtered.length === 0 && <div className="empty-state"><div className="empty-icon">📚</div><h3>{query ? 'Ничего не найдено' : 'Курсов пока нет'}</h3><p>{query ? 'Попробуй изменить поисковый запрос.' : 'Опубликованные курсы появятся здесь.'}</p></div>}{!loading && !error && filtered.length > 0 && <CourseGrid items={filtered} />}</div></main> }

function CoursePage({ courseId, user }: { courseId: number; user: User | null }) { const [course, setCourse] = useState<CourseDetails | null>(null); const [loading, setLoading] = useState(true); const [error, setError] = useState(''); const [enrolling, setEnrolling] = useState(false); const [enrolled, setEnrolled] = useState(false); const [progress, setProgress] = useState<number[]>([]); const [selectedPage, setSelectedPage] = useState<number | null>(null); const [progressLoading, setProgressLoading] = useState(false); const [actionError, setActionError] = useState('')
  const token = localStorage.getItem(TOKEN_KEY)
  useEffect(() => { getCourse(courseId).then(setCourse).catch((e) => setError(e instanceof Error ? e.message : 'Не удалось загрузить курс')).finally(() => setLoading(false)) }, [courseId])
  useEffect(() => { if (!user || !token) return; getCourseProgress(courseId, token).then(setProgress).catch(() => setProgress([])) }, [courseId, user, token])
  useEffect(() => { if (!course || selectedPage !== null) return; const first = course.sections.flatMap((s) => s.pages)[0]; if (first) setSelectedPage(first.id) }, [course, selectedPage])
  const handleEnroll = async () => { if (!user || !token) { navigate('/login'); return }; setEnrolling(true); setActionError(''); try { await enrollCourse(courseId, token); setEnrolled(true) } catch (e) { setActionError(e instanceof Error ? e.message : 'Не удалось записаться на курс') } finally { setEnrolling(false) } }
  const markComplete = async (pageId: number) => { if (!user || !token) { navigate('/login'); return }; setProgressLoading(true); setActionError(''); try { await completePage(pageId, token); setProgress((old) => old.includes(pageId) ? old : [...old, pageId]) } catch (e) { setActionError(e instanceof Error ? e.message : 'Не удалось сохранить прогресс') } finally { setProgressLoading(false) } }
  if (loading) return <main className="page"><div className="container"><div className="loading-screen">Загружаем курс...</div></div></main>
  if (error || !course) return <main className="page"><div className="container empty-page"><span className="eyebrow">ОШИБКА</span><h1>Не удалось открыть курс</h1><p>{error || 'Курс не найден.'}</p><Link href="/courses" className="button button-primary">Вернуться в каталог</Link></div></main>
  const pages = course.sections.flatMap((s) => s.pages); const currentPage = pages.find((p) => p.id === selectedPage) || pages[0]; const progressPercent = pages.length ? Math.round(progress.filter((id) => pages.some((p) => p.id === id)).length / pages.length * 100) : 0; const isFree = course.price === 0
  return <main className="page"><div className="container"><div className="course-detail-hero"><div><span className="eyebrow">{course.theme || 'ОБУЧЕНИЕ'}</span><h1>{course.title}</h1><p>{course.description || 'Описание курса пока не добавлено.'}</p></div><div className="course-detail-action"><strong>{formatCoursePrice(course.price)}</strong>{enrolled ? <span className="enroll-success">✓ Ты записан на курс</span> : isFree ? user ? <button type="button" className="button button-primary" onClick={handleEnroll} disabled={enrolling}>{enrolling ? 'Записываем...' : 'Записаться на курс'}</button> : <Link href="/login" className="button button-primary">Войти и записаться</Link> : <span className="course-coming-soon">Оплата появится позже</span>}{actionError && <div className="form-error">{actionError}</div>}</div></div>
    <section className="course-learning"><aside className="course-sidebar"><div className="course-progress"><span>Прогресс</span><strong>{progressPercent}%</strong><div className="progress-bar"><span style={{ width: `${progressPercent}%` }} /></div></div>{course.sections.map((section) => <div className="course-section-nav" key={section.id}><span>Раздел {section.position + 1}</span><h3>{section.title}</h3>{section.pages.map((page) => <button type="button" key={page.id} className={`course-page-nav ${selectedPage === page.id ? 'active' : ''}`} onClick={() => setSelectedPage(page.id)}><span>{progress.includes(page.id) ? '✓' : page.type === 'theory' ? '📖' : page.type === 'test' ? '✓' : '✦'}</span>{page.title}</button>)}</div>)}</aside>
      <div className="course-lesson">{currentPage ? <><div className="lesson-header"><span className="course-page-type">{currentPage.type === 'theory' ? 'ТЕОРИЯ' : currentPage.type === 'test' ? 'ТЕСТ' : 'AI-ТЕСТ'}</span><h2>{currentPage.title}</h2></div>{currentPage.type === 'theory' ? <div className="lesson-content"><Markdown content={currentPage.content || ''} /></div> : <div className="lesson-placeholder"><div className="empty-icon">{currentPage.type === 'test' ? '✓' : '✦'}</div><h3>{currentPage.type === 'test' ? 'Тест' : 'AI-тест'}</h3><p>{currentPage.type === 'test' ? 'Интерактивный тест подключим следующим этапом.' : 'Персонализированный AI-тест пока находится в разработке.'}</p></div>}{user && <div className="lesson-actions">{progress.includes(currentPage.id) ? <span className="enroll-success">✓ Страница пройдена</span> : <button type="button" className="button button-primary" onClick={() => markComplete(currentPage.id)} disabled={progressLoading}>{progressLoading ? 'Сохраняем...' : currentPage.type === 'theory' ? 'Отметить как прочитанное' : 'Отметить как пройденное'}</button>}</div>}<div className="lesson-navigation">{(() => { const index = pages.findIndex((p) => p.id === currentPage.id); const previous = pages[index - 1]; const next = pages[index + 1]; return <>{previous ? <button type="button" className="button button-secondary" onClick={() => setSelectedPage(previous.id)}>← Назад</button> : <span />}{next ? <button type="button" className="button button-primary" onClick={() => setSelectedPage(next.id)}>Следующая →</button> : <span className="enroll-success">Курс завершён</span>}</> })()}</div></> : <div className="empty-state"><h3>В курсе пока нет страниц</h3></div>}</div></section><Link href="/courses" className="text-link">← Вернуться в каталог</Link></div></main> }

function LoginPage({ onLogin }: { onLogin: (email: string, password: string) => Promise<void> }) { const [email, setEmail] = useState(''); const [password, setPassword] = useState(''); const [error, setError] = useState(''); const [loading, setLoading] = useState(false); const submit = async (e: FormEvent<HTMLFormElement>) => { e.preventDefault(); setError(''); setLoading(true); try { await onLogin(email, password) } catch (err) { setError(err instanceof Error ? err.message : 'Не удалось войти') } finally { setLoading(false) } }; return <main className="auth-page"><div className="auth-card"><div className="auth-icon">◉</div><span className="eyebrow">С возвращением</span><h1>Вход в Cringearium</h1><p className="auth-description">Войди, чтобы продолжить обучение.</p><form onSubmit={submit} className="auth-form"><label>Email<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" required /></label><label>Пароль<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" required /></label>{error && <div className="form-error">{error}</div>}<button type="submit" className="button button-primary" disabled={loading}>{loading ? 'Входим...' : 'Войти'}</button></form><p className="auth-bottom">Нет аккаунта? <Link href="/register" className="text-link">Зарегистрироваться</Link></p></div></main> }

function RegisterPage() { return <main className="auth-page"><div className="auth-card"><div className="auth-icon">✦</div><span className="eyebrow">НОВЫЙ АККАУНТ</span><h1>Регистрация</h1><p className="auth-description">Создай аккаунт и начни обучение.</p><div className="empty-state"><h3>Регистрация подключается</h3><p>Форма регистрации будет добавлена следующим этапом.</p></div></div></main> }

function ProfilePage({ user }: { user: User | null }) { const [courses, setCourses] = useState<Course[]>([]); const [loading, setLoading] = useState(true); const [error, setError] = useState(''); const token = localStorage.getItem(TOKEN_KEY); useEffect(() => { if (!token) { setLoading(false); return } getEnrolledCourses(token).then(setCourses).catch((e) => setError(e instanceof Error ? e.message : 'Не удалось загрузить мои курсы')).finally(() => setLoading(false)) }, [token]); if (!user) return <main className="page"><div className="container empty-page"><span className="eyebrow">ПРОФИЛЬ</span><h1>Войдите в аккаунт</h1><p>Здесь будут отображаться ваши курсы и прогресс.</p><Link href="/login" className="button button-primary">Войти</Link></div></main>; return <main className="page"><div className="container"><div className="page-heading"><span className="eyebrow">ЛИЧНЫЙ КАБИНЕТ</span><h1>Профиль</h1></div><div className="profile-layout"><section className="profile-card profile-summary"><div className="avatar">{user.username.slice(0, 1).toUpperCase()}</div><div><h2>{user.username}</h2><p>{user.email}</p></div></section><section className="profile-card"><div className="card-heading"><span className="eyebrow">ОБУЧЕНИЕ</span><h2>Мои курсы</h2></div>{loading && <div className="loading-screen">Загружаем курсы...</div>}{error && <div className="form-error">{error}</div>}{!loading && !error && courses.length === 0 && <div className="empty-state"><div className="empty-icon">📚</div><h3>Пока нет курсов</h3><p>Запишись на курс из каталога, и он появится здесь.</p><Link href="/courses" className="button button-primary">Перейти в каталог</Link></div>}{!loading && !error && courses.length > 0 && <CourseGrid items={courses} />}</section></div></div></main> }

function AboutPage() { return <main className="page"><div className="container narrow"><div className="page-heading"><span className="eyebrow">ПРОЕКТ</span><h1>О Cringearium</h1><p>Небольшая образовательная платформа для курсов, практики и персонализированного обучения.</p></div><section className="about-card"><h2>Зачем это всё?</h2><p>Cringearium создаётся как современная учебная платформа с понятной структурой курсов, практическими заданиями и персональной помощью.</p><div className="about-stats"><div><strong>Go</strong><span>Backend</span></div><div><strong>React</strong><span>Frontend</span></div><div><strong>AI</strong><span>Следующий этап</span></div></div></section></div></main> }

function App() { const path = usePath(); const [user, setUser] = useState<User | null>(null); const [authLoading, setAuthLoading] = useState(true); useEffect(() => { const token = localStorage.getItem(TOKEN_KEY); if (!token) { setAuthLoading(false); return } getMe(token).then(setUser).catch(() => { localStorage.removeItem(TOKEN_KEY); setUser(null) }).finally(() => setAuthLoading(false)) }, []); const handleLogin = async (email: string, password: string) => { const result = await login(email, password); localStorage.setItem(TOKEN_KEY, result.token); setUser(result.user); navigate('/profile') }; const handleLogout = () => { localStorage.removeItem(TOKEN_KEY); setUser(null); navigate('/') }; if (authLoading) return <div className="loading-screen">Загружаем...</div>; let page: ReactNode; if (path === '/') page = <HomePage courses={[]} />; else if (path === '/courses') page = <CoursesPage />; else if (path === '/about') page = <AboutPage />; else if (path === '/login') page = <LoginPage onLogin={handleLogin} />; else if (path === '/register') page = <RegisterPage />; else if (path === '/profile') page = <ProfilePage user={user} />; else { const match = path.match(/^\/courses\/(\d+)$/); page = match ? <CoursePage courseId={Number(match[1])} user={user} /> : <main className="page"><div className="container empty-page"><h1>404</h1><p>Страница не найдена.</p><Link href="/" className="button button-primary">На главную</Link></div></main> } return <div className="app-shell"><Header user={user} onLogout={handleLogout} />{page}<Footer /></div> }

export default App
