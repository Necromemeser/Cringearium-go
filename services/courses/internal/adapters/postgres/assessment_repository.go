package postgres

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func (db *DB) FindAssessmentResults(ctx context.Context) ([]domain.AssessmentResult, error) {
	const query = `
		SELECT ar.id, ar.user_id, ar.course_id, c.title, ar.assessment_type, ar.assessment_scope,
			ar.score, ar.answers, ar.completed_at
		FROM assessment_results ar
		JOIN courses c ON c.id = ar.course_id
		ORDER BY ar.completed_at DESC, ar.id DESC
	`

	rows, err := db.conn.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]domain.AssessmentResult, 0)
	for rows.Next() {
		var result domain.AssessmentResult
		if err := rows.Scan(
			&result.ID,
			&result.UserID,
			&result.CourseID,
			&result.CourseTitle,
			&result.Type,
			&result.Scope,
			&result.Score,
			&result.Answers,
			&result.CompletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
