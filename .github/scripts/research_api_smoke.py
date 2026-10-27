"""Research workflow against a disposable Compose stack, no LLM calls.
Creates unique synthetic accounts and a study; removes its own fixtures in finally.
Requires Docker access to the stack's PostgreSQL for test-only role setup and answer keys.
"""
import csv
import io
import json
import subprocess
import sys
import urllib.request
import uuid

sys.dont_write_bytecode = True
from smoke_backend import request, sql, BASE_URL


def auth_sql(query):
    return subprocess.run([
        'docker', 'exec', '-i', 'cringearium-postgres', 'sh', '-c',
        'exec psql -X -qAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d cringearium_auth',
    ], input=query, text=True, capture_output=True, check=True).stdout.strip()


def main():
    accounts = []
    study_id = None
    try:
        for i in range(3):
            name = 'research_' + uuid.uuid4().hex
            credentials = {'email': name + '@example.test', 'password': uuid.uuid4().hex}
            user = request('POST', '/api/auth/register', {'username': name, **credentials}, expected=201)
            uid = int(user['id'])
            accounts.append(uid)
            if i == 0:
                auth_sql(f"UPDATE users SET role='admin' WHERE id={uid};")
                admin = request('POST', '/api/auth/login', credentials)['token']
            else:
                user['token'] = request('POST', '/api/auth/login', credentials)['token']
                if i == 1:
                    learner1 = user
                else:
                    learner2 = user
        path = '/api/admin/research/studies'
        request('GET', path, expected=401)
        request('GET', path, token=learner1['token'], expected=403)
        courses = request('GET', '/api/courses')
        control = next(int(c['id']) for c in courses if c['title'] == 'Математика: обычные тесты')
        adaptive = next(int(c['id']) for c in courses if c['title'] == 'Математика: курс с ИИ')
        study = request('POST', path, {'title': '[SMOKE] Synthetic research', 'design': 'randomized',
                        'scope': 'course', 'control_course_id': control, 'adaptive_course_id': adaptive,
                        'protocol_notes': 'Synthetic integration fixtures, removed after verification.'}, admin, 201)
        study_id = int(study['id'])
        root = path + '/' + str(study_id)
        request('POST', root + '/participants', {'user_ids': accounts[1:]}, admin)
        report = request('GET', root + '/report', token=admin)
        assert [g['assigned'] for g in report['groups']] == [1, 1]
        assert all(p['pre_score'] is None and p['gain'] is None for p in report['participants'])
        request('POST', root + '/participants', {'user_ids': accounts[1:]}, admin, 400)
        learners = {int(u['id']): u for u in [learner1, learner2]}
        for participant in report['participants']:
            uid, cid = int(participant['account_id']), int(participant['course_id'])
            token = learners[uid]['token']
            instruments = json.loads(sql(f"""
                SELECT json_agg(x ORDER BY kind) FROM (
                    SELECT t.id, d.kind FROM tests t JOIN assessment_definitions d ON d.test_id=t.id
                    JOIN course_pages p ON p.id=t.page_id JOIN course_sections s ON s.id=p.section_id
                    WHERE s.course_id={cid} AND d.scope='course'
                ) x;
            """))
            for kind in ['pretest', 'posttest']:
                test_id = int(next(t['id'] for t in instruments if t['kind'] == kind))
                options = json.loads(sql(f"""
                    SELECT json_agg(x ORDER BY id) FROM (
                        SELECT q.id, min(a.id) FILTER (WHERE a.is_correct) AS correct,
                            min(a.id) FILTER (WHERE NOT a.is_correct) AS incorrect
                        FROM test_questions q JOIN test_answers a ON a.question_id=q.id
                        WHERE q.test_id={test_id} GROUP BY q.id
                    ) x;
                """))
                count = 0 if kind == 'pretest' else len(options) // 2
                answers = [{'question_id': option['id'], 'answer_id': option['correct' if i < count else 'incorrect']} for i, option in enumerate(options)]
                result = request('POST', f'/api/tests/{test_id}/attempts', {'answers': answers}, token)
                assert result['score'] == (0 if kind == 'pretest' else 50)
        report = request('GET', root + '/report', token=admin)
        assert all(p['eligible'] and p['pre_score'] == 0 and p['gain'] == 50 for p in report['participants'])
        assert report['comparison']['gain_difference'] == 0
        assert report['ai_available'] and all(p['ai_answers'] == 0 for p in report['participants'])
        for dataset in ['participants', 'assessments', 'answers', 'ai_events', 'metrics', 'protocol', 'mapping']:
            url = root + '/export?dataset=' + dataset
            request('GET', url, token=learner1['token'], expected=403)
            req = urllib.request.Request(BASE_URL + url, headers={'Authorization': 'Bearer ' + admin})
            with urllib.request.urlopen(req, timeout=15) as response:
                payload = response.read()
                assert payload.startswith(b'\xef\xbb\xbf')
                rows = list(csv.DictReader(io.StringIO(payload.decode('utf-8-sig'))))
                assert dataset == 'ai_events' or rows, dataset
                if dataset == 'participants':
                    assert len(rows) == 2 and all(r['pre_score'] == '0' and r['gain_pp'] == '50' for r in rows)
                    assert 'account_id' not in rows[0] and 'email' not in rows[0]
                if dataset == 'assessments':
                    assert len(rows) == 4 and all(len(r['form_hash']) == 64 for r in rows)
                if dataset == 'answers':
                    assert all(r['question_text'] and r['selected_answer_text'] for r in rows)
        print('PASS research API: admin authorization, protocol, randomized grants, immutable assignment, versioned pairs, missing versus zero, seven CSV datasets')
    finally:
        if study_id is not None:
            sql(f'DELETE FROM research_participants WHERE study_id={study_id}; DELETE FROM research_studies WHERE id={study_id};')
        if accounts:
            ids = ','.join(str(uid) for uid in accounts)
            sql(f'DELETE FROM test_attempts WHERE user_id IN ({ids}); DELETE FROM page_progress WHERE user_id IN ({ids}); DELETE FROM course_access WHERE user_id IN ({ids});')
            auth_sql(f'DELETE FROM users WHERE id IN ({ids});')


if __name__ == '__main__':
    main()
