import type { AdaptiveSession } from '../../api/adaptiveTests'

const russian = (value: string | undefined) => /[А-Яа-яЁё]/.test(value ?? '')

export default function AdaptiveResults({ session }: { session: AdaptiveSession }) {
  const feedback = session.feedback
  const strong = (feedback?.mastered_topics ?? []).filter(item => russian(item.title) && russian(item.reason)) ?? []
  const review = (feedback?.topics_to_review ?? []).filter(item => russian(item.title) && russian(item.reason)) ?? []
  const next = (feedback?.next_steps ?? []).filter(item => russian(item.title) && russian(item.description)) ?? []
  return (
    <div className="adaptive-feedback">
      <span className="adaptive-icon" aria-hidden="true">✓</span>
      <h3>Практика завершена</h3>
      <p>{russian(feedback?.summary) ? feedback!.summary : 'Ты завершил практику. Ниже — результаты раундов.'}</p>
      <div className="adaptive-result-grid">
        {session.rounds.filter(round => round.status === 'completed' && (round.answered_count ?? 0) > 0).map(round => {
          const count = round.answered_count!
          const correct = round.correct_count ?? 0
          const percent = Math.round(correct / count * 100)
          const label = round.purpose === 'prerequisite_review' ? 'Повторение основ' : round.purpose === 'mixed' ? 'Практика с повторением' : 'Основная практика'
          return <section className="adaptive-result-card" key={round.id} aria-label={label}>
            <span>{label}</span>
            <strong>{percent}%</strong>
            <p>Верных ответов: {correct} из {count}</p>
            <progress max={count} value={correct} aria-label={`Верных ответов: ${correct} из ${count}`} />
          </section>
        })}
      </div>
      {strong.length > 0 && <section><h4>Что получилось</h4><ul>{strong.map((item, i) => <li key={i}><strong>{item.title}</strong><span>{item.reason}</span></li>)}</ul></section>}
      {review.length > 0 && <section><h4>Что стоит повторить</h4><ul>{review.map((item, i) => <li key={i}><strong>{item.title}</strong><span>{item.reason}</span></li>)}</ul></section>}
      {next.length > 0 && <section><h4>Следующие шаги</h4><ul>{next.map((item, i) => <li key={i}><strong>{item.title}</strong><span>{item.description}</span></li>)}</ul></section>}
    </div>
  )
}
