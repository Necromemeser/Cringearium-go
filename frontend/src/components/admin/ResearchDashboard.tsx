import { useEffect, useState, type FormEvent } from 'react'
import type { User } from '../../api/auth'
import type { AdminCourse } from '../../api/admin'
import { assignParticipants, createStudy, downloadResearchCSV, getResearchReport, getStudies, type ResearchReport, type ResearchStudy } from '../../api/research'
import { armLabel, GainChart, number, ParticipationChart, ScoreScatter } from './ResearchCharts'

const quality: Record<string, string> = {
  missing_pretest: 'Нет входного теста', missing_posttest: 'Нет подходящего итогового теста', multiple_pretests: 'Несколько входных попыток', multiple_posttests: 'Несколько итоговых попыток',
  post_before_pretest: 'Итоговый раньше входного', unversioned_pretest: 'Не сохранена версия входного', unversioned_posttest: 'Не сохранена версия итогового', incompatible_pretest: 'Другая версия входного', incompatible_posttest: 'Другая версия итогового',
  other_arm_access: 'Есть доступ к другому варианту', ai_in_other_arm: 'Сессии ИИ в другом варианте',
}
const warnings: Record<string, string> = {
  same_pre_post_form: 'Входной и итоговый тест используют одну форму. Возможен эффект повторного знакомства с заданиями.',
  uncalibrated_instruments: 'Одинаковость форм не доказывает их педагогическую валидность. Пока нет психометрической калибровки контрольных заданий.',
  observational_design: 'Группы назначены вручную. Самоотбор и исходные различия могут объяснять результат; сравнение не доказывает причинный эффект ИИ.',
  complete_cases_only: 'Оценки прироста и модель используют только полные сопоставимые пары. Пропуски не заменяются нулями; при различиях в выбывании возможна систематическая ошибка.',
  active_time_not_measured: 'Активное время обучения пока не измеряется. Интервал между тестами и длительность сессии включают паузы и не равны времени учёбы.',
  no_delayed_retention_test: 'Отложенной проверки сохранения знаний пока нет. Эти данные оценивают ближайший результат, а не долговременное освоение.',
  small_sample: 'В одной из групп меньше 10 полных пар. Оценки нестабильны; отсутствие заметной разницы не доказывает равенство методов.',
  ai_source_unavailable: 'Детализация ИИ недоступна или превышает лимит выгрузки. Основные результаты доступны; отсутствующие данные ИИ не заменяются нулями.',
  content_changed_since_protocol: 'Теория изменилась после фиксации протокола. Сопоставимость обучения требует отдельной проверки.',
}
const date = (value: string) => new Intl.DateTimeFormat('ru-RU', { dateStyle: 'short', timeStyle: 'short', timeZone: 'Europe/Moscow' }).format(new Date(value))

export default function ResearchDashboard({ token, courses, users }: { token: string; courses: AdminCourse[]; users: User[] }) {
  const [studies, setStudies] = useState<ResearchStudy[]>([])
  const [selected, setSelected] = useState<number | null>(null)
  const [report, setReport] = useState<ResearchReport | null>(null)
  const [includeOld, setIncludeOld] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [refresh, setRefresh] = useState(0)
  const [busy, setBusy] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [title, setTitle] = useState('')
  const [design, setDesign] = useState('randomized')
  const [scope, setScope] = useState('course')
  const [control, setControl] = useState('')
  const [adaptive, setAdaptive] = useState('')
  const [notes, setNotes] = useState('')
  const [userIds, setUserIds] = useState('')
  const [arm, setArm] = useState('control')
  const [page, setPage] = useState(0)
  const [onlyFlags, setOnlyFlags] = useState(false)

  useEffect(() => {
    let cancelled = false
    void getStudies(token).then(data => { if (!cancelled) { setStudies(data); setSelected(current => current ?? data[0]?.id ?? null) } }).catch(() => { if (!cancelled) setError('Не удалось загрузить список исследований.') })
    return () => { cancelled = true }
  }, [token])
  useEffect(() => {
    if (!selected) return
    let cancelled = false
    setLoading(true); setReport(null); setError(''); setPage(0)
    void getResearchReport(token, selected, includeOld).then(data => { if (!cancelled) setReport(data) }).catch(e => { if (!cancelled) setError(e instanceof Error ? e.message : 'Не удалось загрузить отчёт') }).finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [token, selected, includeOld, refresh])
  async function saveStudy(event: FormEvent) {
    event.preventDefault(); setBusy(true); setError(''); setMessage('')
    try {
      const study = await createStudy(token, { title, design, scope, control_course_id: Number(control), adaptive_course_id: Number(adaptive), protocol_notes: notes })
      setStudies(current => [study, ...current]); setSelected(study.id); setShowCreate(false); setTitle(''); setMessage('Протокол сохранён. Теперь можно добавить участников.')
    } catch (e) { setError(e instanceof Error ? e.message : 'Протокол не сохранён') } finally { setBusy(false) }
  }
  async function addParticipants(event: FormEvent) {
    event.preventDefault(); if (!selected) return
    const ids = userIds.split(/[\s,;]+/).filter(Boolean).map(Number)
    const known = new Set(users.map(user => user.id))
    if (!ids.length || ids.some(id => !Number.isSafeInteger(id) || !known.has(id)) || new Set(ids).size !== ids.length) { setError('Укажи неповторяющиеся ID существующих аккаунтов из таблицы пользователей.'); return }
    setBusy(true); setError(''); setMessage('')
    try { const result = await assignParticipants(token, selected, ids, arm); setUserIds(''); setMessage(`Участников добавлено: ${result.assigned}. Доступ к назначенному курсу выдан.`); setRefresh(value => value + 1) } catch (e) { setError(e instanceof Error ? e.message : 'Не удалось распределить участников') } finally { setBusy(false) }
  }
  async function download(dataset: string) {
    if (!selected) return
    setBusy(true); setError('')
    try { await downloadResearchCSV(token, selected, dataset, includeOld) } catch (e) { setError(e instanceof Error ? e.message : 'Экспорт не выполнен') } finally { setBusy(false) }
  }
  const study = studies.find(item => item.id === selected)
  const rows = report?.participants.filter(row => !onlyFlags || row.flags.length > 0) ?? []
  const courseOptions = courses.map(course => <option key={course.id} value={course.id}>{course.title} · ID {course.id}</option>)
  return <section className="admin-card research-dashboard">
    <div className="admin-card-heading"><div><span className="eyebrow">ИССЛЕДОВАНИЕ ОБУЧЕНИЯ</span><h2>Эффект адаптивной практики</h2><p className="admin-card-description">Сравнение групп по фиксированным входным и итоговым тестам. Верные ответы в ИИ-практике сами по себе не измеряют эффект обучения.</p></div><button className="button button-secondary" type="button" onClick={() => setShowCreate(value => !value)}>Новое исследование</button></div>
    <div className="research-body">
      {error && <p className="form-error" role="alert">{error}</p>}{message && <p className="research-success" role="status">{message}</p>}
      {showCreate && <form className="research-form" onSubmit={saveStudy}>
        <h3>Зафиксировать протокол</h3>
        <label>Название<input required maxLength={300} value={title} onChange={e => setTitle(e.target.value)} placeholder="Математика · пилотная группа" /></label>
        <label>Дизайн<select value={design} onChange={e => setDesign(e.target.value)}><option value="randomized">Случайное распределение новых участников 1:1</option><option value="observational">Наблюдательное: назначить группы вручную</option></select></label>
        <label>Контрольная группа<select required value={control} onChange={e => setControl(e.target.value)}><option value="">Курс с обычными тестами</option>{courseOptions}</select></label>
        <label>Экспериментальная группа<select required value={adaptive} onChange={e => setAdaptive(e.target.value)}><option value="">Курс с ИИ-практикой</option>{courseOptions}</select></label>
        <label>Объём проверки<select value={scope} onChange={e => setScope(e.target.value)}><option value="course">Весь курс</option><option value="trigonometry">Только тригонометрия</option></select></label>
        <label className="research-wide">Протокол и условия<textarea value={notes} onChange={e => setNotes(e.target.value)} maxLength={10000} placeholder="Гипотеза, планируемый размер выборки, одинаковый бюджет времени, сроки, критерии включения. Лучше записать до сбора результатов." /></label>
        <p className="research-wide research-help">Система проверит совпадение теории и контрольных тестов. Группы участников и версии тестов фиксируются; задним числом улучшить исходный балл выбором лучшей попытки нельзя.</p>
        <button className="button button-primary" disabled={busy || !courses.length}>Сохранить протокол</button>
      </form>}
      <div className="research-toolbar"><label>Исследование<select value={selected ?? ''} onChange={e => { setSelected(Number(e.target.value)); setIncludeOld(false); setMessage('') }}><option value="" disabled>Выбери исследование</option>{studies.map(item => <option key={item.id} value={item.id}>{item.title} · №{item.id}</option>)}</select></label><button className="button button-secondary" onClick={() => setRefresh(value => value + 1)} disabled={!selected || loading}>Обновить отчёт</button></div>
      {!studies.length && <p className="research-empty">Пока нет исследований. Создай протокол для пары курсов и добавь участников — в том числе уже обучавшихся, если выбран наблюдательный дизайн.</p>}
      {study && <details className="research-protocol"><summary>Протокол и добавление участников</summary><p><strong>{study.design === 'randomized' ? 'Случайное блочное распределение 1:1' : 'Наблюдательное сравнение'}</strong> · {study.scope === 'course' ? 'Весь курс' : 'Тригонометрия'}. Контроль: курс {study.control_course_id}; ИИ: курс {study.adaptive_course_id}.</p>{study.protocol_notes && <p className="research-notes">{study.protocol_notes}</p>}<p className="research-help">Пары выбираются одинаково: первый входной и первый последующий итоговый, в том же курсе и объёме. При случайном распределении учитываются только результаты после назначения. Участник сохраняет назначенную группу независимо от фактического использования ИИ.</p>
        <form className="research-assign" onSubmit={addParticipants}><label>ID аккаунтов участников<textarea value={userIds} onChange={e => setUserIds(e.target.value)} required placeholder="Например: 12, 13, 14. ID доступны в таблице пользователей ниже." /></label>{study.design === 'observational' && <label>Назначить в группу<select value={arm} onChange={e => setArm(e.target.value)}><option value="control">Обычные тесты</option><option value="adaptive">ИИ-практика</option></select></label>}<p className="research-help">{study.design === 'randomized' ? 'Для случайного распределения нужны новые участники без прежнего доступа к обоим вариантам. Повторное назначение запрещено.' : 'Группы выбирает исследователь. Ранее полученные результаты доступны с явной отметкой отсутствующих версий.'}</p><button className="button button-primary" disabled={busy || !users.length}>Добавить и выдать доступ</button></form>
      </details>}
      {selected && <label className="research-toggle"><input type="checkbox" checked={includeOld} onChange={e => setIncludeOld(e.target.checked)} />Включить исторические результаты без сохранённой версии теста (только разведочный анализ)</label>}
      {loading && <p role="status">Собираем данные исследования...</p>}
      {report && <>
        <p className="research-stamp">Срез: {date(report.generated_at)} МСК. Экспорты содержат время UTC и псевдонимы участников без имён и email.</p>
        {includeOld && <p className="research-warning">Включены неверсионированные попытки. Сопоставимость их заданий проверить нельзя; выводы считаются разведочными.</p>}
        <div className="research-groups">{report.groups.map(group => <article key={group.arm}><h3>{armLabel(group.arm)}</h3><strong>{group.paired}<small> полных сопоставимых пар</small></strong><dl><div><dt>Назначено</dt><dd>{group.assigned}</dd></div><div><dt>Нет выбранного итогового</dt><dd>{group.assigned - group.with_post}</dd></div><div><dt>Входной → итоговый, средние пар</dt><dd>{number(group.pre.mean)} → {number(group.post.mean)}%</dd></div><div><dt>Средний прирост</dt><dd>{number(group.gain.mean)} п. п.</dd></div><div><dt>Медиана / SD прироста</dt><dd>{number(group.gain.median)} / {number(group.gain.sd)}</dd></div><div><dt>95% ДИ среднего прироста</dt><dd>{number(group.gain.ci_low)} … {number(group.gain.ci_high)}</dd></div><div><dt>Дошли до итогового</dt><dd>{number(group.post_completion)}%</dd></div><div><dt>Завершили все страницы</dt><dd>{group.finished_course}</dd></div></dl></article>)}</div>
        <div className="research-comparison"><h3>ИИ − обычные тесты: сравнение прироста</h3><strong>{number(report.comparison.gain_difference)} п. п.</strong><p>95% ДИ: {number(report.comparison.ci_low)} … {number(report.comparison.ci_high)} п. п. · Hedges g по приросту: {number(report.comparison.hedges_g, 3)}</p><p className="research-help">Это разница изменений баллов, а не процент эффективности ИИ. Пустое значение означает, что оценку рассчитать нельзя. Если интервал включает ноль, данные совместимы и с отсутствием разницы.</p></div>
        <div className="research-charts"><article><h3>Прирост и неопределённость</h3><GainChart groups={report.groups} /></article><article><h3>Исходный уровень и результат</h3><ScoreScatter participants={report.participants} /></article><article className="research-wide"><h3>Полнота наблюдений</h3><ParticipationChart groups={report.groups} /></article></div>
        <section className="research-model"><h3>Модель с поправкой на входной балл</h3><p>Итоговый балл = константа + входной балл + назначенная группа. OLS, устойчивые стандартные ошибки HC3, 95% t-интервал.</p>{report.model.available ? <><strong>Скорректированная разница: {number(report.model.group_coefficient)} п. п.</strong><p>95% ДИ {number(report.model.ci_low)} … {number(report.model.ci_high)} · n={report.model.n}, df={report.model.df} · R²={number(report.model.r2, 3)}</p><p>Коэффициент входного балла: {number(report.model.pre_coefficient, 3)}; константа: {number(report.model.intercept, 3)}.</p></> : <p className="research-empty">Модель недоступна: {report.model.reason === 'singular_design' ? 'входные баллы не позволяют отделить параметры модели' : report.model.reason === 'high_leverage' ? 'отдельные наблюдения полностью определяют оценку' : 'нужно минимум 10 полных пар, не менее трёх в каждой группе и достаточное разнообразие входных баллов'}.</p>}<p className="research-help">Интервалы предполагают независимость участников. Для классов и общих учебных групп нужны кластерные или многоуровневые модели. Поправка на входной балл не устраняет скрытые различия между людьми. Модель не доказывает причинный эффект при самоотборе и не восстанавливает пропущенные итоговые тесты.</p></section>
        <details className="research-quality" open><summary>Качество данных и ограничения</summary><ul>{Array.from(new Set(report.warnings)).map(flag => <li key={flag}>{warnings[flag] ?? flag}</li>)}</ul>{Object.keys(report.quality_counts).length > 0 && <div className="research-flag-grid">{Object.entries(report.quality_counts).map(([flag, count]) => <span key={flag}>{quality[flag] ?? flag}: <strong>{count}</strong></span>)}</div>}<p className="research-help">Доступ к другому варианту не доказывает его прохождение. Эти признаки сохраняются в выгрузке; участники не переводятся автоматически в другую группу.</p></details>
        <div className="research-exports"><h3>Данные для Python / R / Excel</h3><p>UTF-8 CSV. Пустая ячейка — отсутствующее измерение, не ноль. Все попытки и ответы доступны отдельно от отфильтрованных пар.</p><div>{[['participants', 'Участники и пары'], ['assessments', 'Все контрольные попытки'], ['answers', 'Ответы по заданиям'], ['ai_events', 'Детализация ИИ'], ['metrics', 'Метрики и модель'], ['protocol', 'Протокол и версии']].map(([dataset, label]) => <button key={dataset} type="button" className="button button-secondary" disabled={busy || (dataset === 'ai_events' && !report.ai_available)} onClick={() => download(dataset)}>{label} · CSV</button>)}</div><p className="research-help">Выгрузки живые: фиксируй файлы вместе с протоколом после завершения сбора. Для сложных, многоуровневых моделей и анализа пропусков используй внешние инструменты; в ответах ИИ есть версии модели, промпта и политики выбора заданий.</p></div>
        <details className="research-protocol"><summary>Ключ для сопоставления с внешними данными</summary><p>Отдельный служебный файл связывает псевдоним участника с ID аккаунта. Он нужен для добавления своих ковариат; храни его отдельно от основной выгрузки.</p><button className="button button-secondary" disabled={busy} onClick={() => download('mapping')}>Ключ ID · CSV</button></details>
        <div className="research-toolbar"><h3>Участники исследования</h3><label><input type="checkbox" checked={onlyFlags} onChange={e => { setOnlyFlags(e.target.checked); setPage(0) }} />Только с замечаниями</label></div>
        <div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>Участник</th><th>Группа</th><th>Входной</th><th>Итоговый</th><th>Прирост, п. п.</th><th>ИИ-сессий / ответов</th><th>В расчёте</th><th>Качество</th></tr></thead><tbody>{rows.slice(page * 25, page * 25 + 25).map(row => <tr key={row.participant_id}><td className="admin-mono">{row.participant_id}<small>Аккаунт ID: {row.account_id}</small></td><td>{armLabel(row.arm)}</td><td>{number(row.pre_score)}</td><td>{number(row.post_score)}</td><td>{number(row.gain)}</td><td>{number(row.ai_sessions, 0)} / {number(row.ai_answers, 0)}</td><td>{row.eligible ? 'Да' : 'Нет'}</td><td>{row.flags.map(flag => quality[flag] ?? flag).join('; ') || '—'}</td></tr>)}{!rows.length && <tr><td colSpan={8}>Участников с выбранными условиями пока нет.</td></tr>}</tbody></table></div>
        <div className="research-pagination"><button className="button button-secondary" disabled={page === 0} onClick={() => setPage(value => value - 1)}>Назад</button><span>{rows.length ? page * 25 + 1 : 0}–{Math.min(rows.length, page * 25 + 25)} из {rows.length}</span><button className="button button-secondary" disabled={(page + 1) * 25 >= rows.length} onClick={() => setPage(value => value + 1)}>Далее</button></div>
      </>}
    </div>
  </section>
}
