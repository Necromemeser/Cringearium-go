import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { getMe, login, type User } from './api/auth'
import { TOKEN_KEY } from './constants/auth'
import Footer from './components/common/Footer'
import Header from './components/common/Header'
import Link from './components/common/Link'
import AboutPage from './pages/AboutPage'
import CoursePage from './pages/CoursePage'
import CoursesPage from './pages/CoursesPage'
import HomePage from './pages/HomePage'
import LoginPage from './pages/LoginPage'
import ProfilePage from './pages/ProfilePage'
import RegisterPage from './pages/RegisterPage'
import { navigate, usePath } from './utils/navigation'
import './App.css'

function App() {
  const path = usePath()
  const [user, setUser] = useState<User | null>(null)
  const [authLoading, setAuthLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) { setAuthLoading(false); return }
    getMe(token)
      .then(setUser)
      .catch(() => { localStorage.removeItem(TOKEN_KEY); setUser(null) })
      .finally(() => setAuthLoading(false))
  }, [])

  const handleLogin = async (email: string, password: string) => {
    const token = await login(email, password)
    localStorage.setItem(TOKEN_KEY, token)
    const user = await getMe(token)
    setUser(user)
    navigate('/profile')
  }

  const handleLogout = () => {
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
    navigate('/')
  }

  if (authLoading) return <div className="loading-screen">Загружаем...</div>

  let page: ReactNode
  if (path === '/') page = <HomePage courses={[]} />
  else if (path === '/courses') page = <CoursesPage />
  else if (path === '/about') page = <AboutPage />
  else if (path === '/login') page = <LoginPage onLogin={handleLogin} />
  else if (path === '/register') page = <RegisterPage />
  else if (path === '/profile') page = <ProfilePage user={user} />
  else {
    const match = path.match(/^\/courses\/(\d+)$/)
    page = match ? <CoursePage courseId={Number(match[1])} user={user} /> : (
      <main className="page"><div className="container empty-page"><h1>404</h1><p>Страница не найдена.</p><Link href="/" className="button button-primary">На главную</Link></div></main>
    )
  }

  return <div className="app-shell"><Header user={user} onLogout={handleLogout} />{page}<Footer /></div>
}

export default App
