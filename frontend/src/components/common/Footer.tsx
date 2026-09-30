import Link from './Link'

export default function Footer() {
  return (
    <footer className="footer">
      <div className="container footer-grid">
        <div><div className="footer-brand">Cringearium</div><p>Образовательная платформа, где учиться немного проще.</p></div>
        <div><h3>Навигация</h3><Link href="/courses">Каталог курсов</Link><Link href="/about">О нас</Link><Link href="/login">Вход</Link></div>
        <div><h3>Проект</h3><a href="https://github.com/Necromemeser/Cringearium-go" target="_blank" rel="noreferrer">GitHub</a><span>© 2026 Cringearium</span></div>
      </div>
    </footer>
  )
}
