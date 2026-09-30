import type { User } from '../../api/auth'
import Link from './Link'

export default function Header({ user, onLogout }: {
  user: User | null
  onLogout: () => void
}) {
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
