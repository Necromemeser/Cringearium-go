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

export function GainDistribution({ participants }: { participants: ResearchParticipant[] }) {
  const rows = participants.filter(row => row.eligible && row.gain != null)
  if (!rows.length) return <p className="research-empty">Для распределения нужны полные сопоставимые пары.</p>
  const arms = ['control', 'adaptive']
  const bins = arms.map(arm => Array.from({ length: 10 }, (_, index) => rows.filter(row => row.arm === arm && row.gain! >= -100 + index * 20 && (index === 9 ? row.gain! <= 100 : row.gain! < -80 + index * 20)).length))
  const max = Math.max(1, ...bins.flat())
  return <svg viewBox="0 0 600 300" role="img" aria-label="Распределение прироста по группам с интервалами 20 процентных пунктов">
    {arms.map((arm, a) => <text key={arm} x={55 + a * 245} y="22" fontSize="12">{armLabel(arm)} · n={rows.filter(row => row.arm === arm).length}</text>)}
    {bins.map((counts, a) => counts.map((n, i) => <rect key={`${a}-${i}`} x={55 + i * 49 + a * 21} y={240 - n / max * 180} width="19" height={n / max * 180} fill={color(arms[a])}><title>{`${armLabel(arms[a])}: от ${-100 + i * 20} до ${-80 + i * 20}${i === 9 ? ' включительно' : ' (не включая верхнюю границу)'} п. п.; участников: ${n}`}</title></rect>))}
    <line x1="52" x2="548" y1="240" y2="240" stroke="#94a3b8" />
    {[0, 2.5, 5, 7.5, 10].map(tick => <text key={tick} x={55 + tick * 49} y="260" fontSize="11" textAnchor="middle">{-100 + tick * 20}</text>)}
    <text x="45" y="60" textAnchor="end" fontSize="11">{max}</text><text x="45" y="244" textAnchor="end" fontSize="11">0</text>
    <text x="175" y="286" fontSize="12">Прирост, п. п. · высота: число участников</text>
  </svg>
}

export function MeanTrajectory({ groups }: { groups: ResearchGroup[] }) {
  const available = groups.filter(group => group.paired > 0 && group.pre.mean != null && group.post.mean != null)
  if (!available.length) return <p className="research-empty">Для сравнения средних нужны полные сопоставимые пары.</p>
  const y = (value: number) => 235 - value * 1.8
  return <svg viewBox="0 0 600 300" role="img" aria-label="Входной и итоговый средний балл на одной выборке полных пар">
    {[0, 25, 50, 75, 100].map(tick => <g key={tick}><line x1="70" x2="540" y1={y(tick)} y2={y(tick)} stroke="#e5eaf2" /><text x="55" y={y(tick) + 4} fontSize="11" textAnchor="end">{tick}%</text></g>)}
    {available.map((group, i) => <g key={group.arm}><text x={70 + i * 240} y="22" fontSize="12">{armLabel(group.arm)} · n={group.paired}</text><line x1="145" x2="465" y1={y(group.pre.mean!)} y2={y(group.post.mean!)} stroke={color(group.arm)} strokeWidth="3" />{[group.pre.mean!, group.post.mean!].map((value, index) => <circle key={index} cx={145 + index * 320} cy={y(value)} r="6" fill={color(group.arm)}><title>{`${armLabel(group.arm)}: ${index ? 'итоговый' : 'входной'} средний ${number(value)}%; одни и те же ${group.paired} участников`}</title></circle>)}</g>)}
    <text x="145" y="263" fontSize="12" textAnchor="middle">Входной</text><text x="465" y="263" fontSize="12" textAnchor="middle">Итоговый</text><text x="125" y="287" fontSize="11">Средние по одинаковым участникам, а не индивидуальные траектории</text>
  </svg>
}

export function PracticeScatter({ participants }: { participants: ResearchParticipant[] }) {
  const rows = participants.filter(row => row.arm === 'adaptive' && row.eligible && row.gain != null && row.ai_window_answers != null)
  if (!rows.length) return <p className="research-empty">Нужны сопоставимые пары ИИ-группы и доступные счётчики ответов.</p>
  const max = Math.max(1, ...rows.map(row => row.ai_window_answers!))
  const x = (value: number) => 70 + value / max * 460
  const y = (value: number) => 145 - value
  return <svg viewBox="0 0 600 300" role="img" aria-label="Число ответов в ИИ-практике и прирост балла каждого участника">
    <text x="70" y="22" fontSize="12">ИИ-практика · сопоставимых пар: {rows.length}</text>
    {[-100, -50, 0, 50, 100].map(tick => <g key={tick}><line x1="70" x2="530" y1={y(tick)} y2={y(tick)} stroke={tick === 0 ? '#9aa8bc' : '#edf0f4'} /><text x="55" y={y(tick) + 4} fontSize="11" textAnchor="end">{tick}</text></g>)}
    {rows.map(row => <circle key={row.participant_id} cx={x(row.ai_window_answers!)} cy={y(row.gain!)} r="5" fill={color('adaptive')} opacity=".65"><title>{`${row.participant_id}: ответов ${row.ai_window_answers}, прирост ${number(row.gain)} п. п.`}</title></circle>)}
    <text x="70" y="263" fontSize="11">0</text><text x="530" y="263" fontSize="11" textAnchor="end">{max}</text><text x="175" y="287" fontSize="12">Ответов между входным и итоговым тестом</text><text x="15" y="180" transform="rotate(-90 15 180)" fontSize="12">Прирост, п. п.</text>
  </svg>
}
