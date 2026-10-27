package postgres

import (
	"context"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

func (r *Repository) ResearchEvents(ctx context.Context, courses []int64) ([]domain.ResearchEvent, bool, error) {
	result := []domain.ResearchEvent{}
	err := r.db.SelectContext(ctx, &result, `SELECT s.id AS session_id,s.user_id,s.course_id,s.status AS session_status,s.created_at,s.completed_at,
 r.id AS round_id,r.round_number,q.id AS question_id,q.question AS question_text,q.correct_option_key AS correct_option,
 r.learning_plan -> (q.position-1) ->> 'reason' AS reason,q.skill_code,q.difficulty,
 a.selected_option_key AS selected_option,a.is_correct,a.answered_at,r.prompt_version,r.model,e.policy_version,e.contributes
 FROM adaptive_sessions s LEFT JOIN adaptive_rounds r ON r.session_id=s.id
 LEFT JOIN adaptive_questions q ON q.round_id=r.id LEFT JOIN adaptive_answers a ON a.question_id=q.id
 LEFT JOIN learning_evidence e ON e.question_id=q.id
 WHERE s.course_id=ANY($1::bigint[]) ORDER BY s.created_at,s.id,r.round_number,q.position LIMIT 50001`, courses)
	truncated := len(result) > 50000
	if truncated {
		result = result[:50000]
	}
	return result, truncated, err
}
