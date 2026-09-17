import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import './App.css'

type Course = {
  id: number
  title: string
  theme: string
  description: string
  price: string
  icon: string
}

const courses: Course[] = [
  {
    id: 1,
    title: 'Математика для СДВГ-шников',
    theme: 'Квадратные уравнения',
    description: 'Короткий практический курс с понятной теорией, примерами и небольшими тестами.',
    price: 'Бесплатно',
    icon: '∑',
  },
  {
    id: 2,
    title: 'Основы программирования',
    theme: 'Для начинающих',
    description: 'Переменные, условия, циклы и первые алгоритмы без лишней теории.',
    price: '990 ₽',
    icon: '{ }',
  },
  {
    id: 3,
    title: 'Алгоритмы и структуры данных',
    theme: 'Практика',
    description: 'Разбираемся с основными структурами данных и учимся выбирать подходящий алгоритм.',
    price: '1 490 ₽',
    icon: 'λ',
  },
]

const routes = ['/', '/courses', '/about', '/login', '/profile']

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

function Link({ href, children, className = '' }: { href: string; children: React.ReactNode; className?: string }) {
  return (
    <a
      href={href}
      className={className}
      onClick={(event) => {
        if (href.startsWith('/')) {
          event.preventDefault()
          navigate(href)
        }
      }}
    >
      {children}
    </a>
  )
}

function Header() {
  return (
    <header className="header">
      <div className="container nav-container">
        <Link href="/" className="brand">
          <span className="brand-icon">⌂</span>
          <span>Cringearium</span>
        </Link>

        <nav className="nav-links">
          <Link href="/courses">📚 Каталог курсов</Link>
          <Link href="/about">ℹ О нас</Link>
        </nav>

        <nav className="nav-actions">
          <Link href="/profile" className="profile-link">◉ Личный кабинет</Link>
          <Link href="/login" className="login-link">Войти</Link>
        </nav>
      </div>
    </header>
  )
}

function Footer() {
  return (
    <footer className="footer">
      <div className="container footer-grid">
        <div>
          <div className="footer-brand">Cringearium</div>
          <p>Образовательная платформа, где учиться немного проще.</p>
        </div>
        <div>
          <h3>Навигация</h3>
          <Link href="/courses">Каталог курсов</Link>
          <Link href="/about">О нас</Link>
          <Link href="/login">Вход</Link>
        </div>
        <div>
          <h3>Проект</h3>
          <a href="https://github.com/Necromemeser/Cringearium-go" target="_blank" rel="noreferrer">GitHub</a>
          <span>© 2026 Cringearium</span>
        </div>
      </div>
    </footer>
  )
}

function HomePage() {
  return (
    <>
      <section className="hero">
        <div className="container hero-content">
          <div className="hero-copy">
            <span className="eyebrow">ОБРАЗОВАТЕЛЬНАЯ ПЛАТФОРМА</span>
            <h1>Учиться можно<br /><span>без лишнего кринжа.</span></h1>
            <p>Курсы, практика и персональная помощь в одном месте. Разбираемся со сложным нормальным человеческим языком.</p>
            <div className="hero-actions">
              <Link href="/courses" className="button button-primary">Смотреть курсы</Link>
              <Link href="/about" className="button button-secondary">Узнать больше</Link>
            </div>
          </div>
          <div className="hero-card">
            <div className="hero-card-icon">✦</div>
            <h2>Курс дня</h2>
            <p>Математика для СДВГ-шников</p>
            <span>Квадратные уравнения</span>
            <Link href="/courses" className="text-link">Перейти к курсу →</Link>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="section-heading">
            <div>
              <span className="eyebrow">ВОЗМОЖНОСТИ</span>
              <h2>Всё необходимое для обучения</h2>
            </div>
          </div>
          <div className="feature-grid">
            <article className="feature-card"><div className="feature-icon">📖</div><h3>Понятные курсы</h3><p>Материалы разбиты на небольшие последовательные шаги.</p></article>
            <article className="feature-card"><div className="feature-icon">✓</div><h3>Практика</h3><p>Тесты помогают сразу проверить, что материал действительно усвоен.</p></article>
            <article className="feature-card"><div className="feature-icon">✦</div><h3>Умный помощник</h3><p>Персональная помощь появится в следующих этапах развития платформы.</p></article>
          </div>
        </div>
      </section>

      <section className="section section-muted">
        <div className="container">
          <div className="section-heading inline-heading">
            <div><span className="eyebrow">КУРСЫ</span><h2>Популярные направления</h2></div>
            <Link href="/courses" className="text-link">Весь каталог →</Link>
          </div>
          <CourseGrid items={courses.slice(0, 3)} />
        </div>
      </section>
    </>
  )
}

function CourseGrid({ items }: { items: Course[] }) {
  return <div className="course-grid">{items.map((course) => <CourseCard key={course.id} course={course} />)}</div>
}

function CourseCard({ course }: { course: Course }) {
  return (
    <article className="course-card">
      <div className="course-cover"><span>{course.icon}</span></div>
      <div className="course-body">
        <span className="course-theme">{course.theme}</span>
        <h3>{course.title}</h3>
        <p>{course.description}</p>
        <div className="course-footer"><strong>{course.price}</strong><Link href="/login" className="small-button">Подробнее</Link></div>
      </div>
    </article>
  )
}

function CoursesPage() {
  return (
    <main className="page">
      <div className="container">
        <div className="page-heading"><span className="eyebrow">ОБУЧЕНИЕ</span><h1>Каталог курсов</h1><p>Выбирай направление и двигайся вперёд в удобном темпе.</p></div>
        <div className="catalog-toolbar"><input aria-label="Поиск курсов" placeholder="Поиск по названию курса..." /><button type="button">Найти</button></div>
        <CourseGrid items={courses} />
      </div>
    </main>
  )
}

function LoginPage() {
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => event.preventDefault()

  return (
    <main className="auth-page">
      <div className="auth-card">
        <div className="auth-icon">◉</div>
        <span className="eyebrow">С возвращением</span>
        <h1>Вход в Cringearium</h1>
        <p className="auth-description">Войди, чтобы продолжить обучение.</p>
        <form onSubmit={handleSubmit} className="auth-form">
          <label>Email<input type="email" placeholder="you@example.com" required /></label>
          <label>Пароль<input type="password" placeholder="••••••••" required /></label>
          <button className="button button-primary" type="submit">Войти</button>
        </form>
        <p className="auth-bottom">Нет аккаунта? <span>Регистрация появится вместе с API авторизации.</span></p>
      </div>
    </main>
  )
}

function ProfilePage() {
  return (
    <main className="page">
      <div className="container">
        <div className="page-heading"><span className="eyebrow">ПРОФИЛЬ</span><h1>Личный кабинет</h1><p>Здесь будет твой прогресс и список доступных курсов.</p></div>
        <div className="profile-layout">
          <section className="profile-card profile-summary"><div className="avatar">N</div><div><h2>Новый пользователь</h2><p>Ученик</p></div><button type="button" className="small-button">Редактировать</button></section>
          <section className="profile-card"><div className="card-heading"><div><span className="eyebrow">ОБУЧЕНИЕ</span><h2>Мои курсы</h2></div></div><div className="empty-state"><div className="empty-icon">📚</div><h3>Пока нет курсов</h3><p>Загляни в каталог и выбери что-нибудь для обучения.</p><Link href="/courses" className="button button-primary">Открыть каталог</Link></div></section>
        </div>
      </div>
    </main>
  )
}

function AboutPage() {
  return (
    <main className="page">
      <div className="container narrow">
        <div className="page-heading"><span className="eyebrow">CRINGEARium</span><h1>О проекте</h1><p>Образовательная платформа, которую мы собираем с нуля на современном стеке.</p></div>
        <div className="about-card"><h2>Зачем это всё?</h2><p>Cringearium задуман как место, где учебный материал не приходится пробивать головой через бесконечные стены текста. Курсы состоят из небольших страниц, практических заданий и тестов.</p><p>Платформа развивается как отдельный проект: Go-бэкенд, React-фронтенд и набор независимых сервисов. В будущем здесь появятся AI-помощник и персонализированные тесты.</p></div>
        <div className="about-stats"><div><strong>Go</strong><span>backend</span></div><div><strong>React</strong><span>frontend</span></div><div><strong>5+</strong><span>сервисов</span></div></div>
      </div>
    </main>
  )
}

function NotFound() {
  return <main className="page"><div className="container empty-page"><span className="eyebrow">404</span><h1>Страница не найдена</h1><p>Похоже, такой страницы пока нет.</p><Link href="/" className="button button-primary">На главную</Link></div></main>
}

function App() {
  const path = usePath()
  const page = routes.includes(path) ? path : '/404'

  return (
    <div className="app-shell">
      <Header />
      {page === '/' && <HomePage />}
      {page === '/courses' && <CoursesPage />}
      {page === '/login' && <LoginPage />}
      {page === '/profile' && <ProfilePage />}
      {page === '/about' && <AboutPage />}
      {page === '/404' && <NotFound />}
      <Footer />
    </div>
  )
}

export default App
