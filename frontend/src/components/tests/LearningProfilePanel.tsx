import type { LearningProfile } from '../../api/adaptiveTests'

const statuses = {
  insufficient_data: 'Недостаточно данных',
  needs_practice: 'Нужна практика',
  practicing: 'В процессе освоения',
  provisionally_mastered: 'Предварительно освоено',
  review_due: 'Пора повторить',
}
const reasons = {
  diagnose: 'проверим знания',
  remediate: 'проработаем ошибки',
  practice: 'закрепим навык',
  challenge: 'попробуем более сложное задание',
  review: 'повторим после перерыва',
  prerequisite: 'вернёмся к предварительному навыку',
}

export default function LearningProfilePanel({ profile }: { profile: LearningProfile | null }) {
  if (!profile?.skills.length) return null
  const next = profile.next_practice[0]
  return (
    <section className="learning-profile" aria-label="Профиль знаний">
      <h4>Твои навыки</h4>
      <p>Оценка уточняется по разным заданиям. Завершение теста само по себе не означает освоение темы.</p>
      <ul>
        {profile.skills.map(skill => (
          <li key={skill.code}>
            <strong>{skill.title}</strong>
            <span>{statuses[skill.status]} · разных заданий: {skill.state.novel_answers}{skill.prerequisite_only ? ' · предварительный навык' : ''}</span>
          </li>
        ))}
      </ul>
      {next && <p><strong>Следующий фокус: {next.skill_title}</strong> — {reasons[next.reason]}.</p>}
    </section>
  )
}
