"""Optional live browser check for the ten enriched courses against disposable Compose.
Requires Playwright and Chromium. Uses one synthetic account, cleans its fixtures, no LLM calls.
"""
import json
import os
import sys
import uuid
from pathlib import Path

sys.dont_write_bytecode = True
from smoke_backend import request, sql
from research_api_smoke import auth_sql
from playwright.sync_api import sync_playwright, expect

BASE = os.environ.get('LEARNER_UI_URL', 'http://localhost:3000')
ROOT = Path(__file__).resolve().parents[2]


def main():
    catalog = json.loads((ROOT / 'content/courses/catalog.json').read_text())['subjects']
    uid = None
    errors = []
    try:
        name = 'catalog_' + uuid.uuid4().hex
        credentials = {'email': name + '@example.test', 'password': uuid.uuid4().hex}
        user = request('POST', '/api/auth/register', {'username': name, **credentials}, expected=201)
        uid = int(user['id'])
        token = request('POST', '/api/auth/login', credentials)['token']
        courses = request('GET', '/api/courses')
        with sync_playwright() as p:
            browser = p.chromium.launch(headless=True, executable_path=os.environ.get('CHROMIUM_PATH', '/usr/bin/chromium'), args=['--no-sandbox'])
            context = browser.new_context(viewport={'width': 1440, 'height': 1000})
            context.add_init_script('localStorage.setItem("cringearium_token", ' + json.dumps(token) + ')')
            page = context.new_page()
            page.on('pageerror', lambda error: errors.append(str(error)))
            page.on('response', lambda response: errors.append(response.url) if '/images/' in response.url and response.status >= 400 else None)
            lessons = practice = images = 0
            for subject in catalog:
                for variant in ['control', 'adaptive']:
                    cid = int(next(c['id'] for c in courses if c['title'] == subject[variant]))
                    request('POST', f'/api/courses/{cid}/enroll', token=token, expected=201)
                    profile = request('GET', f'/api/adaptive-tests/profile?course_id={cid}', token=token)
                    assert len(profile['skills']) == 6
                    page.goto(f'{BASE}/courses/{cid}')
                    for material in subject['pages']:
                        page.locator('.course-page-button').filter(has_text=material['title']).click()
                        lesson = page.locator('.lesson-content')
                        expect(lesson).to_contain_text('Самопроверка')
                        for img in lesson.locator('img').all():
                            img.scroll_into_view_if_needed()
                            expect(img).to_have_js_property('complete', True)
                            assert img.evaluate('e => e.naturalWidth > 0'), img.get_attribute('src')
                            images += 1
                        lessons += 1
                    if variant == 'control':
                        for section in subject['sections']:
                            page.locator('.course-page-button').filter(has_text='Самостоятельная практика: ' + section).click()
                            expect(page.locator('.lesson-panel input[type=radio]')).to_have_count(24)
                            practice += 1
                    page.set_viewport_size({'width': 390, 'height': 844})
                    page.locator('.course-page-button').filter(has_text=subject['pages'][0]['title']).click()
                    expect(page.locator('.lesson-content')).to_contain_text('Самопроверка')
                    assert page.evaluate('document.documentElement.scrollWidth <= innerWidth + 1'), subject['key']
                    if subject['key'] == 'memology' and variant == 'adaptive':
                        page.locator('.course-page-button').filter(has_text='Шаблон, ожидание и панчлайн').click()
                        expect(page.locator('.lesson-content img')).to_have_count(1)
                        page.locator('.lesson-content img').scroll_into_view_if_needed()
                        page.screenshot(path='/tmp/enriched-memology-mobile.png', full_page=True)
                    page.set_viewport_size({'width': 1440, 'height': 1000})
            assert lessons == 60 and practice == 15 and images >= 18
            assert not errors, errors
            browser.close()
        print(f'PASS live course catalog: {lessons} lessons, {practice} fixed practice tests, {images} image placements, ten six-skill profiles, mobile layout; no LLM calls')
    finally:
        if uid is not None:
            sql(f'DELETE FROM test_attempts WHERE user_id={uid}; DELETE FROM page_progress WHERE user_id={uid}; DELETE FROM course_access WHERE user_id={uid};')
            auth_sql(f'DELETE FROM users WHERE id={uid};')


if __name__ == '__main__':
    main()
