const team = [
  { name: 'Нейро Мемесер', role: 'Основатель и разработчик', quote: 'Не кринжует лишь тот, кто не учится.', image: 'team1.jpg', width: 484, height: 527 },
  { name: 'Котик из мема', role: 'Буквально я', quote: 'ЭЭааЫЫВуаааа', image: 'team2.jpg', width: 752, height: 485 },
  { name: 'Райан Гослинг', role: 'Буквально я 2', quote: '...', image: 'team3.jpg', width: 640, height: 360 },
]

export default function AboutPage() {
  return <main className="about-page">
    <section className="about-intro"><div className="container"><span className="eyebrow">ЗНАКОМИМСЯ ПОБЛИЖЕ</span><h1>О нашей<br />образовательной платформе</h1><p>Мы создаем будущее образования, делая обучение доступным, интересным и эффективным</p></div></section>
    <section className="about-mission"><div className="container about-mission-grid"><div><span className="eyebrow">ЗАЧЕМ МЫ ЗДЕСЬ</span><h2>Наша миссия</h2><p className="about-lead">Cringearium был основан в 2025 году с целью изменить подход к онлайн-образованию.</p><p>Мы верим, что качественное образование должно быть доступно каждому, независимо от места проживания, возраста или финансовых возможностей.</p></div><img src="/images/about/mission.jpg" alt="Мем о преподавании: «Где ты дримкору учился? Я преподавал»" width="1280" height="960" /></div></section>
    <section className="about-team"><div className="container"><div className="public-section-heading"><span className="eyebrow">ЛИЦА ПРОЕКТА</span><h2>Наша команда</h2><p>Люди, которые делают образование лучше</p></div><div className="about-team-grid">{team.map(member => <article className="about-member" key={member.name}><img src={`/images/about/${member.image}`} alt={member.name} width={member.width} height={member.height} loading="lazy" /><div><h3>{member.name}</h3><p>{member.role}</p><blockquote>«{member.quote}»</blockquote></div></article>)}</div></div></section>
    <section className="about-achievements"><div className="container"><div className="public-section-heading"><span className="eyebrow">НЕМНОГО САМОИРОНИИ</span><h2>Мы гордимся</h2><p>Нашими результатами работы</p></div><dl>{[['0', 'Дофамина'], ['0', 'Серотонина'], ['0', 'Смысла жизни'], ['146%', 'Выгорания']].map(([value, label]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl></div></section>
  </main>
}
