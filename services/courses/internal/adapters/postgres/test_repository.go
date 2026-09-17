package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func (db *DB) FindTestByPageID(ctx context.Context, userID, pageID int64) (*domain.Test, error) {
	const query = `
		SELECT t.id, t.page_id, t.passing_score
		FROM tests t
		JOIN course_pages p ON p.id = t.page_id
		JOIN course_sections s ON s.id = p.section_id
		JOIN course_access ca ON ca.course_id = s.course_id AND ca.user_id = $1
		JOIN courses c ON c.id = s.course_id
		WHERE t.page_id = $2 AND p.type = 'test' AND c.status = 'published'
	`
	return db.findTest(ctx, query, userID, pageID)
}

func (db *DB) FindTestByID(ctx context.Context, userID, testID int64) (*domain.Test, error) {
	const query = `
		SELECT t.id, t.page_id, t.passing_score
		FROM tests t
		JOIN course_pages p ON p.id = t.page_id
		JOIN course_sections s ON s.id = p.section_id
		JOIN course_access ca ON ca.course_id = s.course_id AND ca.user_id = $1
		JOIN courses c ON c.id = s.course_id
		WHERE t.id = $2 AND p.type = 'test' AND c.status = 'published'
	`
	return db.findTest(ctx, query, userID, testID)
}

func (db *DB) FindLatestTestAttempt(ctx context.Context, userID, testID int64) (*domain.TestAttempt, error) {
	const attemptQuery = `
		SELECT id, score, passed, completed_at
		FROM test_attempts
		WHERE user_id = $1 AND test_id = $2
		ORDER BY completed_at DESC NULLS LAST, id DESC
		LIMIT 1
	`
	attempt := new(domain.TestAttempt)
	if err := db.conn.QueryRowxContext(ctx, attemptQuery, userID, testID).Scan(&attempt.ID, &attempt.Score, &attempt.Passed, &attempt.CompletedAt); err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}

	const answersQuery = `
		SELECT question_id, answer_id
		FROM test_attempt_answers
		WHERE attempt_id = $1
		ORDER BY question_id
	`
	rows, err := db.conn.QueryxContext(ctx, answersQuery, attempt.ID)
	if err != nil { return nil, err }
	defer rows.Close()
	for rows.Next() {
		answer := domain.TestAttemptAnswer{}
		if err := rows.Scan(&answer.QuestionID, &answer.AnswerID); err != nil { return nil, err }
		attempt.Answers = append(attempt.Answers, answer)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return attempt, nil
}

func (db *DB) findTest(ctx context.Context, query string, args ...any) (*domain.Test, error) {
	test := new(domain.Test)
	if err := db.conn.QueryRowxContext(ctx, query, args...).Scan(&test.ID, &test.PageID, &test.PassingScore); err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}

	const questionsQuery = `
		SELECT id, test_id, question, position
		FROM test_questions
		WHERE test_id = $1
		ORDER BY position
	`
	rows, err := db.conn.QueryxContext(ctx, questionsQuery, test.ID)
	if err != nil { return nil, err }
	defer rows.Close()

	for rows.Next() {
		question := domain.TestQuestion{}
		if err := rows.Scan(&question.ID, &question.TestID, &question.Question, &question.Position); err != nil { return nil, err }
		test.Questions = append(test.Questions, question)
	}
	if err := rows.Err(); err != nil { return nil, err }

	const answersQuery = `
		SELECT id, question_id, text, position, is_correct
		FROM test_answers
		WHERE question_id = $1
		ORDER BY position
	`
	for i := range test.Questions {
		answerRows, err := db.conn.QueryxContext(ctx, answersQuery, test.Questions[i].ID)
		if err != nil { return nil, err }
		for answerRows.Next() {
			answer := domain.TestAnswer{}
			if err := answerRows.Scan(&answer.ID, &answer.QuestionID, &answer.Text, &answer.Position, &answer.IsCorrect); err != nil {
				answerRows.Close()
				return nil, err
			}
			test.Questions[i].Answers = append(test.Questions[i].Answers, answer)
		}
		if err := answerRows.Err(); err != nil {
			answerRows.Close()
			return nil, err
		}
		answerRows.Close()
	}

	return test, nil
}

func (db *DB) SubmitTest(ctx context.Context, userID, testID int64, answers []domain.TestAttemptAnswer, score int, passed bool) (*domain.TestResult, error) {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil { return nil, err }
	defer tx.Rollback()

	const attemptQuery = `
		INSERT INTO test_attempts (test_id, user_id, score, passed, started_at, completed_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, completed_at
	`
	var attemptID int64
	var completedAt time.Time
	if err := tx.QueryRowContext(ctx, attemptQuery, testID, userID, score, passed).Scan(&attemptID, &completedAt); err != nil { return nil, err }

	const answerQuery = `
		INSERT INTO test_attempt_answers (attempt_id, question_id, answer_id, is_correct)
		SELECT $1, tq.id, $3, ($3 = ta.id)
		FROM test_questions tq
		LEFT JOIN test_answers ta ON ta.id = $3 AND ta.question_id = tq.id
		WHERE tq.id = $2 AND tq.test_id = $4
	`
	for _, answer := range answers {
		result, err := tx.ExecContext(ctx, answerQuery, attemptID, answer.QuestionID, answer.AnswerID, testID)
		if err != nil { return nil, err }
		if rows, err := result.RowsAffected(); err != nil || rows != 1 { return nil, sql.ErrNoRows }
	}

	if passed {
		const completeQuery = `
			INSERT INTO page_progress (user_id, page_id)
			SELECT $1, page_id FROM tests WHERE id = $2
			ON CONFLICT (user_id, page_id) DO NOTHING
		`
		if _, err := tx.ExecContext(ctx, completeQuery, userID, testID); err != nil { return nil, err }
	}

	if err := tx.Commit(); err != nil { return nil, err }
	return &domain.TestResult{AttemptID: attemptID, Score: score, Passed: passed, CompletedAt: completedAt, Answers: answers}, nil
}
