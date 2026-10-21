"""Exercise course submission against the disposable CI Compose stack (no LLM calls)."""
import json
import subprocess
import time
import urllib.error
import urllib.request
import uuid

BASE_URL = "http://localhost:8080"


def request(method, path, body=None, token=None, expected=200):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(BASE_URL + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            assert response.status == expected, (path, response.status)
            payload = response.read()
            return json.loads(payload) if payload else None
    except urllib.error.HTTPError as error:
        if error.code != expected:
            raise RuntimeError(f"{method} {path}: expected {expected}, got {error.code}") from None


def sql(query):
    result = subprocess.run(
        ["docker", "exec", "-i", "cringearium-postgres", "sh", "-c",
         'exec psql -X -qAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d cringearium_courses'],
        input=query, text=True, capture_output=True, check=True,
    )
    return result.stdout.strip()


def main():
    # Gateway health alone does not establish that auth and course migrations finished.
    for attempt in range(30):
        try:
            request("GET", "/api/courses")
            request("GET", "/api/auth/me", expected=401)
            break
        except (OSError, RuntimeError):
            if attempt == 29:
                raise
            time.sleep(1)

    username = "smoke_" + uuid.uuid4().hex
    credentials = {"email": username + "@example.test", "password": uuid.uuid4().hex}
    user = request("POST", "/api/auth/register", {"username": username, **credentials}, expected=201)
    token = request("POST", "/api/auth/login", credentials)["token"]
    user_id = int(user["id"])

    # Own fixtures keep this check independent of changes to educational seed content.
    fixtures = json.loads(sql(f"""
        WITH course AS (
            INSERT INTO courses (title, status, price) VALUES ('{username}', 'published', 0) RETURNING id
        ), section AS (
            INSERT INTO course_sections (course_id, title, position)
            SELECT id, 'Smoke tests', 0 FROM course RETURNING id, course_id
        ), pages AS (
            INSERT INTO course_pages (section_id, title, type, position)
            SELECT section.id, fixture.title, 'test', fixture.position FROM section
            CROSS JOIN (VALUES ('Входной тест: smoke', 0), ('Итоговый тест: smoke', 1), ('Практика', 2)) fixture(title, position)
            RETURNING id, section_id, title
        ), tests AS (
            INSERT INTO tests (page_id, passing_score) SELECT id, 70 FROM pages RETURNING id, page_id
        ), questions AS (
            INSERT INTO test_questions (test_id, question, position)
            SELECT id, '2 + 2 = ?', 0 FROM tests RETURNING id
        ), answers AS (
            INSERT INTO test_answers (question_id, text, is_correct, position)
            SELECT questions.id, option.text, option.correct, option.position FROM questions
            CROSS JOIN (VALUES ('4', true, 0), ('5', false, 1)) option(text, correct, position)
            RETURNING id
        )
        SELECT json_agg(fixture) FROM (
            SELECT tests.id, pages.id AS page_id, section.course_id,
                CASE WHEN pages.title LIKE 'Входной тест:%' THEN 'pretest'
                     WHEN pages.title LIKE 'Итоговый тест:%' THEN 'posttest'
                     ELSE 'ordinary' END AS kind
            FROM tests JOIN pages ON pages.id=tests.page_id
            JOIN section ON section.id=pages.section_id
        ) fixture;
    """))
    assert {item["kind"] for item in fixtures} == {"pretest", "posttest", "ordinary"}
    enrolled = set()
    for fixture in fixtures:
        test_id, page_id, course_id = (int(fixture[key]) for key in ("id", "page_id", "course_id"))
        if course_id not in enrolled:
            request("POST", f"/api/courses/{course_id}/enroll", token=token, expected=201)
            enrolled.add(course_id)
        test = request("GET", f"/api/pages/{page_id}/test", token=token)
        assert test["id"] == test_id and test["questions"]
        options = json.loads(sql(f"""
            SELECT json_agg(options) FROM (
                SELECT q.id AS question_id,
                    min(a.id) FILTER (WHERE a.is_correct) AS correct,
                    min(a.id) FILTER (WHERE NOT a.is_correct) AS incorrect
                FROM test_questions q JOIN test_answers a ON a.question_id=q.id
                WHERE q.test_id={test_id} GROUP BY q.id ORDER BY q.id
            ) options;
        """))
        for passed in (False, True):
            answers = [{"question_id": option["question_id"],
                        "answer_id": option["correct" if passed else "incorrect"]} for option in options]
            result = request("POST", f"/api/tests/{test_id}/attempts", {"answers": answers}, token, expected=200)
            assert result["passed"] == passed and result["score"] == (100 if passed else 0), result
            assert all(answer["is_correct"] == passed for answer in result["answers"])
            attempt_id = int(result["attempt_id"])
            stored = sql(f"SELECT bool_and(is_correct={'true' if passed else 'false'}) FROM test_attempt_answers WHERE attempt_id={attempt_id};")
            assert stored == "t", "persisted grading differs from application grading"
            progress = sql(f"SELECT count(*) FROM page_progress WHERE user_id={user_id} AND page_id={page_id};")
            assert int(progress) == int(passed), "page completion differs from pass status"
            assessment = sql(f"SELECT assessment_type || ':' || score FROM assessment_results WHERE test_attempt_id={attempt_id};")
            expected = "" if fixture["kind"] == "ordinary" else fixture["kind"] + (":100" if passed else ":0")
            assert assessment == expected, (assessment, expected)
        request("POST", f"/api/tests/{test_id}/attempts", {"answers": answers}, token, expected=409)
        print(f"PASS {fixture['kind']}: failed/passed grading, assessment, progress, duplicate prevention")
    request("GET", "/api/adaptive-tests/" + str(uuid.uuid4()), token=token, expected=404)
    print("PASS missing adaptive session: HTTP 404")


if __name__ == "__main__":
    main()
