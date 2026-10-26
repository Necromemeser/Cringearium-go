"""Optional admin research UI smoke test with synthetic HTTP responses only.
Requires Vite (or LEARNER_UI_URL), Playwright and Chromium. No database writes or LLM calls.
"""
import copy
import json
import math
import os
import statistics
from datetime import datetime, timezone
from urllib.parse import urlparse, parse_qs
from playwright.sync_api import sync_playwright, expect

BASE = os.environ.get('LEARNER_UI_URL', 'http://127.0.0.1:5173')
NOW = datetime.now(timezone.utc).isoformat()
STUDY = {'id': 1, 'title': 'Синтетический пример исследования', 'design': 'observational', 'scope': 'course', 'control_course_id': 2, 'adaptive_course_id': 1, 'pre_hash': 'a' * 64, 'post_hash': 'b' * 64, 'content_hash': 'c' * 64, 'protocol_notes': 'Синтетические данные только для проверки интерфейса.', 'created_at': NOW}


def summary(values):
    n = len(values)
    mean = statistics.mean(values) if n else None
    sd = statistics.stdev(values) if n > 1 else None
    d = 2.365 * sd / math.sqrt(n) if sd is not None else None
    return {'n': n, 'mean': mean, 'median': statistics.median(values) if n else None, 'sd': sd, 'ci_low': mean - d if d is not None else None, 'ci_high': mean + d if d is not None else None}


def fixture():
    participants = []
    noise = [-2, 1, 3, -1, 2, -3, 1, -1, 1, -2, 2, 0, 3, -1, -2, 2]
    for i in range(16):
        adaptive = i >= 8
        pre = 10 + (i % 8) * 10
        post = 15 + .6 * pre + 12 * adaptive + noise[i]
        participants.append({'participant_id': f'synthetic-{i:03d}', 'account_id': i + 1, 'arm': 'adaptive' if adaptive else 'control', 'course_id': 1 if adaptive else 2, 'assigned_at': NOW, 'pre_score': pre, 'post_score': post, 'gain': post - pre, 'pre_id': 2 * i + 1, 'post_id': 2 * i + 2, 'pre_at': NOW, 'post_at': NOW, 'eligible': True, 'flags': [], 'completed_pages': 10, 'total_pages': 10, 'ai_sessions': 2 if adaptive else 0, 'ai_answers': 20 if adaptive else 0, 'ai_correct': 15 if adaptive else 0})
    groups = []
    for arm in ['control', 'adaptive']:
        rows = [p for p in participants if p['arm'] == arm]
        groups.append({'arm': arm, 'assigned': 8, 'with_pre': 8, 'with_post': 8, 'paired': 8, 'finished_course': 8, 'post_completion': 100, 'pre': summary([p['pre_score'] for p in rows]), 'post': summary([p['post_score'] for p in rows]), 'gain': summary([p['gain'] for p in rows])})
    return {'study': STUDY, 'generated_at': NOW, 'include_unversioned': False, 'participants': participants, 'groups': groups, 'comparison': {'gain_difference': groups[1]['gain']['mean'] - groups[0]['gain']['mean'], 'ci_low': -1.5, 'ci_high': 26.2, 'hedges_g': .9, 'completion_difference': 0}, 'model': {'available': True, 'reason': '', 'n': 16, 'df': 13, 'group_coefficient': 12.375, 'pre_coefficient': .595833333, 'intercept': 15.1875, 'se': 1.139343819, 'ci_low': 9.914, 'ci_high': 14.836, 'r2': .96}, 'warnings': ['observational_design', 'complete_cases_only', 'small_sample', 'active_time_not_measured'], 'ai_available': True, 'quality_counts': {}}


def main():
    state = {'studies': [STUDY], 'assigned': False, 'exports': [], 'calls': [], 'errors': []}
    courses = [{'id': i, 'title': 'Математика: курс с ИИ' if i == 1 else 'Математика: обычные тесты', 'theme': 'Математика', 'status': 'published', 'enrolled_users': 8, 'page_count': 10, 'section_count': 3, 'price': 0, 'description': ''} for i in [1, 2]]
    users = [{'id': i, 'username': f'test{i}', 'email': f'test{i}@example.test', 'role': 'student'} for i in range(1, 17)]

    def api(route):
        request = route.request
        parsed = urlparse(request.url)
        path = parsed.path
        if not path.startswith('/api/'):
            route.continue_()
            return
        state['calls'].append(path)
        if path == '/api/auth/me':
            body = {'id': 100, 'username': 'Исследователь', 'email': 'test@example.test', 'role': 'student' if request.headers.get('authorization') == 'Bearer student-test' else 'admin'}
        elif path == '/api/admin/users':
            body = users
        elif path == '/api/admin/courses':
            body = courses
        elif path == '/api/admin/ai-tests':
            route.fulfill(status=503, body='Unavailable legacy source')
            return
        elif path == '/api/admin/ai-chat':
            body = {'conversations': 0, 'messages': 0, 'user_messages': 0, 'ai_messages': 0, 'active_users': 0}
        elif path == '/api/admin/assessment-results':
            body = []
        elif path == '/api/admin/research/studies':
            if request.method == 'POST':
                payload = request.post_data_json
                assert payload['control_course_id'] == 2 and payload['adaptive_course_id'] == 1
                study = {**STUDY, **payload, 'id': 2}
                state['studies'].insert(0, study)
                body = study
            else:
                body = state['studies']
        elif path.endswith('/participants'):
            assert request.post_data_json['user_ids'] == [1, 2]
            state['assigned'] = True
            body = {'assigned': 2}
        elif path.endswith('/report'):
            body = fixture()
            sid = int(path.split('/')[-2])
            body['study'] = next(s for s in state['studies'] if s['id'] == sid)
            body['include_unversioned'] = parse_qs(parsed.query).get('include_unversioned') == ['true']
            if sid == 2:
                body['participants'] = []
                for g in body['groups']:
                    for key in ['assigned', 'with_pre', 'with_post', 'paired', 'finished_course']:
                        g[key] = 1 if key == 'assigned' and state['assigned'] else 0
                    for key in ['pre', 'post', 'gain']:
                        g[key] = summary([])
                body['model'].update(available=False, reason='insufficient_data')
        elif path.endswith('/export'):
            assert request.headers.get('authorization') == 'Bearer admin-test', 'Export lost authorization'
            state['exports'].append(parse_qs(parsed.query)['dataset'][0])
            route.fulfill(status=200, headers={'Content-Type': 'text/csv; charset=utf-8'}, body='\ufeffparticipant_id,assigned_arm,pre_score,post_score\nsynthetic-000,control,10,19\n')
            return
        else:
            state['errors'].append(f'Unexpected API: {path}')
            route.fulfill(status=500, body='Unexpected request')
            return
        route.fulfill(status=200, json=copy.deepcopy(body))

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True, executable_path=os.environ.get('CHROMIUM_PATH', '/usr/bin/chromium'), args=['--no-sandbox'])
        context = browser.new_context(viewport={'width': 1440, 'height': 1000})
        context.add_init_script("localStorage.setItem('cringearium_token', 'admin-test')")
        context.route('**/api/**', api)
        page = context.new_page()
        page.on('pageerror', lambda error: state['errors'].append(str(error)))
        page.goto(BASE + '/admin')
        expect(page.get_by_role('heading', name='Эффект адаптивной практики')).to_be_visible()
        expect(page.get_by_text('Скорректированная разница:', exact=False)).to_be_visible()
        expect(page.get_by_role('img', name='Средний прирост баллов по группам с 95-процентным доверительным интервалом')).to_be_visible()
        expect(page.get_by_role('img', name='Входной и итоговый балл каждого участника: выше диагонали результат улучшился')).to_be_visible()
        expect(page.get_by_text('Часть оперативной статистики недоступна.', exact=False)).to_be_visible()
        page.get_by_label('Включить исторические результаты', exact=False).check()
        expect(page.get_by_text('Включены неверсионированные попытки.', exact=False)).to_be_visible()
        with page.expect_download() as info:
            page.get_by_role('button', name='Участники и пары · CSV').click()
        assert 'participants' in info.value.suggested_filename
        assert state['exports'] == ['participants']
        page.set_viewport_size({'width': 390, 'height': 844})
        if not page.evaluate('document.documentElement.scrollWidth <= innerWidth + 1'):
            print('Overflow:', page.evaluate("Array.from(document.querySelectorAll('body *')).map(e=>({tag:e.tagName,cl:e.className,right:e.getBoundingClientRect().right,width:e.getBoundingClientRect().width})).filter(e=>e.right>innerWidth+2).slice(0,12)"))
            page.screenshot(path='/tmp/research-mobile.png', full_page=True)
            raise AssertionError('Research UI overflows the mobile viewport')
        page.set_viewport_size({'width': 1440, 'height': 1000})
        page.screenshot(path='/tmp/research-dashboard.png', full_page=True)
        page.get_by_role('button', name='Новое исследование').click()
        page.get_by_label('Название', exact=True).fill('Проверка интерфейса')
        page.get_by_label('Контрольная группа', exact=False).select_option('2')
        page.get_by_label('Экспериментальная группа', exact=False).select_option('1')
        page.get_by_role('button', name='Сохранить протокол').click()
        expect(page.get_by_text('Протокол сохранён.', exact=False)).to_be_visible()
        page.locator('.research-protocol').first.locator('summary').click()
        page.get_by_label('ID аккаунтов участников').fill('1, 2')
        page.get_by_role('button', name='Добавить и выдать доступ').click()
        expect(page.get_by_text('Участников добавлено: 2.', exact=False)).to_be_visible()
        assert state['assigned'] and not state['errors'], state['errors']
        student = browser.new_context()
        student.add_init_script("localStorage.setItem('cringearium_token', 'student-test')")
        student.route('**/api/**', api)
        denied = student.new_page()
        before = len(state['calls'])
        denied.goto(BASE + '/admin')
        expect(denied.get_by_role('heading', name='Доступ запрещён')).to_be_visible()
        assert not any(path.startswith('/api/admin/') for path in state['calls'][before:])
        browser.close()
    print('PASS research browser: charts, model, missing source isolation, old-version toggle, authenticated CSV, mobile, protocol, allocation, admin-only view')


if __name__ == '__main__':
    main()
