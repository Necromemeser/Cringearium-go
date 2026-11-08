import { useState } from 'react'
import type { User } from '../../api/auth'
import type { ResearchExclusion, ResearchParticipant } from '../../api/research'

const reasons: Record<string, string> = { already_assigned: 'Уже участвует', prior_course_access: 'Уже получен доступ к одному из курсов', prior_learning: 'Есть результаты предыдущего обучения' }
const roles: Record<string, string> = { student: 'Ученик', teacher: 'Преподаватель', admin: 'Администратор' }

export default function ParticipantPicker({ users, exclusions, participants, selected, onChange, disabled }: {
  users: User[]; exclusions: ResearchExclusion[]; participants: ResearchParticipant[]; selected: number[]; onChange: (ids: number[]) => void; disabled: boolean
}) {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('available')
  const [page, setPage] = useState(0)
  const blocked = new Map(exclusions.map(item => [item.user_id, reasons[item.reason] ?? 'Недоступен для назначения']))
  for (const person of participants) blocked.set(person.account_id, 'Уже участвует')
  const query = search.trim().toLocaleLowerCase('ru')
  const rows = users.filter(user => `${user.username} ${user.email}`.toLocaleLowerCase('ru').includes(query) && (filter === 'all' || (filter === 'selected' ? selected.includes(user.id) : !blocked.has(user.id))))
  const lastPage = Math.max(0, Math.ceil(rows.length / 20) - 1)
  const currentPage = Math.min(page, lastPage)
  const visible = rows.slice(currentPage * 20, (currentPage + 1) * 20)
  const toggle = (id: number) => onChange(selected.includes(id) ? selected.filter(value => value !== id) : [...selected, id])
  return <fieldset className="research-picker" disabled={disabled}>
    <legend>Выбери участников</legend>
    <div className="research-picker-controls"><label>Поиск пользователей<input type="search" value={search} onChange={e => { setSearch(e.target.value); setPage(0) }} placeholder="Имя или email" /></label><label>Показывать<select value={filter} onChange={e => { setFilter(e.target.value); setPage(0) }}><option value="available">Доступных</option><option value="all">Всех пользователей</option><option value="selected">Выбранных</option></select></label></div>
    <div className="research-picker-actions"><strong role="status">Выбрано: {selected.length} из 500</strong><button type="button" className="small-button" onClick={() => onChange([...new Set([...selected, ...visible.filter(user => !blocked.has(user.id)).map(user => user.id)])].slice(0, 500))}>Выбрать на странице</button><button type="button" className="small-button" onClick={() => onChange([])} disabled={!selected.length}>Снять выбор</button></div>
    <div className="research-picker-list">{visible.map(user => <label key={user.id} className={`research-picker-user${blocked.has(user.id) ? ' unavailable' : ''}`}><input type="checkbox" checked={selected.includes(user.id)} disabled={blocked.has(user.id) || (!selected.includes(user.id) && selected.length >= 500)} onChange={() => toggle(user.id)} /><span><strong>{user.username}</strong><small>{user.email}</small></span><span className="research-picker-reason">{blocked.get(user.id) ?? roles[user.role] ?? user.role}</span></label>)}{!visible.length && <p className="research-help">{users.length ? 'Нет пользователей с такими условиями. Попробуй изменить поиск или фильтр.' : 'Список пользователей пуст или не загружен.'}</p>}</div>
    <div className="research-pagination"><button type="button" className="small-button" onClick={() => setPage(currentPage - 1)} disabled={currentPage === 0}>Предыдущие пользователи</button><span>{rows.length ? currentPage * 20 + 1 : 0}–{Math.min(rows.length, (currentPage + 1) * 20)} из {rows.length}</span><button type="button" className="small-button" onClick={() => setPage(currentPage + 1)} disabled={currentPage >= lastPage}>Следующие пользователи</button></div>
    <p className="research-help">Выбор сохраняется при поиске и переключении страниц. Уже назначенные и неподходящие участники недоступны; окончательную проверку выполняет сервер.</p>
  </fieldset>
}
