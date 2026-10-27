import Link from '../components/common/Link'

const steps = [
  { title: 'Разберись в теме', text: 'Короткая теория, наглядные примеры и разбор типичных ошибок.', label: 'Теория' },
  { title: 'Попробуй сам', text: 'Проверь понимание на задачах, а не только на знакомых формулировках.', label: 'Практика' },
  { title: 'Закрепи результат', text: 'Посмотри обратную связь и вернись к тому, что пока вызывает вопросы.', label: 'Обратная связь' },
]

export default function HomePage() {
  return <main className="home-page">
    <section className="home-intro">
      <div className="container home-intro-grid">
        <div className="home-intro-copy">
          <span className="eyebrow">CRINGEARIUM · УЧИМСЯ С ПОНИМАНИЕМ</span>
          <h1>Сложные темы.<br /><span>Понятные шаги.</span></h1>
          <p>От первого «ничего не понятно» до самостоятельного решения. Теория, практика и помощь ИИ — в одном месте и в твоём темпе.</p>
          <div className="home-actions"><Link href="/courses" className="button button-primary">Выбрать курс <span aria-hidden="true">→</span></Link><Link href="/about" className="home-about-link">Познакомиться с проектом</Link></div>
          <p className="home-intro-note">Начни с интересной темы. Остальное разберём по ходу.</p>
        </div>
        <div className="home-learning-path" aria-label="Как устроено обучение">
          <div className="home-path-heading"><span>ТВОЙ ПУТЬ В ТЕМЕ</span><span aria-hidden="true">01 — 03</span></div>
          <ol>{steps.map((step, index) => <li key={step.title}><span className="home-step-number" aria-hidden="true">0{index + 1}</span><div><span className="home-step-label">{step.label}</span><h2>{step.title}</h2><p>{step.text}</p></div></li>)}</ol>
          <div className="home-path-note">Ошибаться можно. Разбираться — полезно.</div>
        </div>
      </div>
    </section>
    <section className="home-subjects" aria-label="Направления обучения"><div className="container"><span>Есть из чего выбрать</span><ul>{['Математика', 'Python', 'Русский язык', 'Психология', 'Философия', 'Мемология'].map(subject => <li key={subject}>{subject}</li>)}</ul></div></section>
    <section className="home-approach">
      <div className="container">
        <div className="home-section-heading"><span className="eyebrow">ОБУЧЕНИЕ, КОТОРОЕ МОЖНО ПОПРОБОВАТЬ</span><h2>Не просто прочитать.<br />Понять, как это работает.</h2><p>У каждого курса есть структура. А у тебя — возможность выбрать подходящий способ практики.</p></div>
        <div className="home-feature-grid">
          <article><span className="home-feature-index">01 / ОСНОВА</span><h3>Материал по делу</h3><p>Небольшие уроки, примеры и самопроверка. Сложная тема становится последовательностью посильных шагов.</p></article>
          <article><span className="home-feature-index">02 / ТВОЙ ВЫБОР</span><h3>Два формата практики</h3><p>Обычные тесты или адаптивные задания с ИИ. Во втором формате следующий раунд учитывает твои ответы и помогает вернуться к пробелам.</p></article>
          <article><span className="home-feature-index">03 / ПОДДЕРЖКА</span><h3>Кринжик рядом</h3><p>Если хочется уточнить понятие или разобрать пример, продолжи разговор с ИИ-помощником. Вопросы здесь приветствуются.</p><Link href="/chat" className="text-link">Познакомиться с Кринжиком →</Link></article>
        </div>
        <div className="home-closing"><div><h2>Любопытство — уже начало.</h2><p>Выбери тему, которую давно хотелось понять.</p></div><Link href="/courses" className="button button-primary">Открыть каталог →</Link></div>
      </div>
    </section>
  </main>
}
