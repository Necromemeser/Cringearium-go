import type { ResearchGroup, ResearchParticipant } from '../../api/research'
export const armLabel = (arm: string) => arm === 'adaptive' ? 'ИИ-практика' : 'Обычные тесты'
const color = (arm: string) => arm === 'adaptive' ? '#7c3aed' : '#16729b'
export const number = (value: number | null | undefined, digits = 1) => value == null ? '—' : new Intl.NumberFormat('ru-RU', { maximumFractionDigits: digits }).format(value)

export function GainChart({ groups }: { groups: ResearchGroup[] }) {
  const available = groups.filter(group => group.gain.mean != null)
  if (!available.length) return <p className="research-empty">Для графика прироста нужны полные сопоставимые пары.</p>
  const values = available.flatMap(group => [group.gain.ci_low ?? group.gain.mean!, group.gain.ci_high ?? group.gain.mean!])
  const low = Math.min(-10, ...values); const high = Math.max(10, ...values)
  const x = (value: number) => 155 + (value - low) / (high - low) * 360
  return <svg viewBox="0 0 600 205" role="img" aria-label="Средний прирост баллов по группам с 95-процентным доверительным интервалом">
    <line x1={x(0)} x2={x(0)} y1="20" y2="150" stroke="#a4afc0" strokeDasharray="4" />
    {available.map((group, i) => {
      const y = 48 + i * 64
      return <g key={group.arm}>
        <text x="4" y={y} fontSize="13">{armLabel(group.arm)}</text>
        <text x="4" y={y + 18} fontSize="11" fill="#64748b">Полных пар: {group.paired}</text>
        {group.gain.ci_low != null && group.gain.ci_high != null && <><line x1={x(group.gain.ci_low)} x2={x(group.gain.ci_high)} y1={y} y2={y} stroke={color(group.arm)} strokeWidth="3" /><line x1={x(group.gain.ci_low)} x2={x(group.gain.ci_low)} y1={y - 6} y2={y + 6} stroke={color(group.arm)} /><line x1={x(group.gain.ci_high)} x2={x(group.gain.ci_high)} y1={y - 6} y2={y + 6} stroke={color(group.arm)} /></>}
        <circle cx={x(group.gain.mean!)} cy={y} r="6" fill={color(group.arm)}><title>{`${armLabel(group.arm)}: ${number(group.gain.mean)} п. п.; 95% ДИ ${number(group.gain.ci_low)}…${number(group.gain.ci_high)}`}</title></circle>
        <text x="530" y={y + 4} fontSize="13">{number(group.gain.mean)} п. п.</text>
      </g>
    })}
    <text x="155" y="174" fontSize="11">{number(low)}</text><text x="492" y="174" fontSize="11">{number(high)}</text>
    <text x="185" y="195" fontSize="12">Прирост: итоговый − входной, п. п.</text>
  </svg>
}
export function ScoreScatter({ participants }: { participants: ResearchParticipant[] }) {
  const paired = participants.filter(row => row.eligible && row.pre_score != null && row.post_score != null)
  if (!paired.length) return <p className="research-empty">Нет полных пар для сопоставления входных и итоговых баллов.</p>
  const x = (v: number) => 58 + v * 4.9; const y = (v: number) => 245 - v * 2
  return <svg viewBox="0 0 600 295" role="img" aria-label="Входной и итоговый балл каждого участника: выше диагонали результат улучшился">
    {[0, 25, 50, 75, 100].map(tick => <g key={tick}><line x1={x(tick)} x2={x(tick)} y1="45" y2="245" stroke="#edf0f4" /><line x1="58" x2="548" y1={y(tick)} y2={y(tick)} stroke="#edf0f4" /><text x={x(tick)} y="262" fontSize="11" textAnchor="middle">{tick}</text><text x="43" y={y(tick) + 4} fontSize="11" textAnchor="end">{tick}</text></g>)}
    <line x1={x(0)} y1={y(0)} x2={x(100)} y2={y(100)} stroke="#8795aa" strokeDasharray="4" />
    {paired.map(row => <circle key={row.participant_id} cx={x(row.pre_score!)} cy={y(row.post_score!)} r="4.5" fill={color(row.arm)} opacity="0.6"><title>{`${row.participant_id}: ${armLabel(row.arm)}, входной ${row.pre_score}%, итоговый ${row.post_score}%, прирост ${row.gain} п. п.`}</title></circle>)}
    <text x="185" y="286" fontSize="12">Входной тест, % верных ответов</text>
    <text x="12" y="152" fontSize="12" transform="rotate(-90 12 152)">Итоговый тест, %</text>
    <text x="58" y="22" fontSize="12" fill="#16729b">● Обычные тесты</text><text x="225" y="22" fontSize="12" fill="#7c3aed">● ИИ-практика</text>
  </svg>
}
export function ParticipationChart({ groups }: { groups: ResearchGroup[] }) {
  const max = Math.max(1, ...groups.map(group => group.assigned))
  return <div className="research-funnel">{groups.map(group => <div key={group.arm}><h4>{armLabel(group.arm)}</h4>{[['Назначено', group.assigned], ['Есть входной', group.with_pre], ['Есть итоговый', group.with_post], ['Полные сопоставимые пары', group.paired]].map(([label, value]) => <div className="research-funnel-row" key={label}><span>{label}</span><div className="research-funnel-track"><div style={{ width: `${Number(value) / max * 100}%`, background: color(group.arm) }} /></div><strong>{value}</strong></div>)}</div>)}</div>
}
