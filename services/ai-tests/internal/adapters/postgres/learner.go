package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

func (r *Repository) GetSkillStates(ctx context.Context, userID int64, codes []string) ([]domain.SkillState, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	var rows [][]byte
	if err := r.db.SelectContext(ctx, &rows, `SELECT state FROM learner_skill_states WHERE user_id=$1 AND skill_code=ANY($2::text[]) ORDER BY skill_code`, userID, codes); err != nil {
		return nil, err
	}
	result := make([]domain.SkillState, 0, len(rows))
	for _, row := range rows {
		var state domain.SkillState
		if err := json.Unmarshal(row, &state); err != nil {
			return nil, err
		}
		result = append(result, state)
	}
	return result, nil
}

func (r *Repository) SubmitLearningRound(ctx context.Context, userID int64, sessionID string, roundID int64, answers []domain.AdaptiveAnswer) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	// Serialize profile changes for one learner, including previously unseen skill rows.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('learning-profile:' || $1::bigint::text,0))`, userID); err != nil {
		return false, err
	}
	var status string
	if err := tx.GetContext(ctx, &status, `SELECT r.status FROM adaptive_rounds r JOIN adaptive_sessions s ON s.id=r.session_id
 WHERE s.id=$1 AND s.user_id=$2 AND r.id=$3 FOR UPDATE OF s,r`, sessionID, userID, roundID); err != nil {
		return false, err
	}
	var stored []struct {
		ID         int64
		Skill      string `db:"skill_code"`
		Question   string
		Correct    string `db:"correct_option_key"`
		Difficulty int
		Options    []byte
	}
	if err := tx.SelectContext(ctx, &stored, `SELECT q.id,q.skill_code,q.question,q.correct_option_key,q.difficulty,
 json_agg(json_build_object('Key',o.option_key,'Text',o.text) ORDER BY o.position) AS options
 FROM adaptive_questions q JOIN adaptive_options o ON o.question_id=q.id
 WHERE q.round_id=$1 GROUP BY q.id ORDER BY q.position`, roundID); err != nil {
		return false, err
	}
	byID := map[int64]domain.AdaptiveAnswer{}
	for _, a := range answers {
		byID[a.QuestionID] = a
	}
	if len(stored) != len(answers) || len(byID) != len(answers) {
		return false, errors.New("incomplete or duplicate answers")
	}
	if status == string(domain.RoundCompleted) {
		var previous []struct {
			QuestionID int64  `db:"question_id"`
			Key        string `db:"selected_option_key"`
		}
		if err := tx.SelectContext(ctx, &previous, `SELECT question_id,selected_option_key FROM adaptive_answers WHERE round_id=$1`, roundID); err != nil {
			return false, err
		}
		if len(previous) != len(answers) {
			return false, errors.New("round already submitted")
		}
		for _, a := range previous {
			if byID[a.QuestionID].SelectedOptionKey != a.Key {
				return false, errors.New("round already submitted with different answers")
			}
		}
		return false, tx.Commit()
	}
	if status != string(domain.RoundGenerated) {
		return false, errors.New("round is not answerable")
	}
	now := time.Now().UTC()
	for _, q := range stored {
		a, ok := byID[q.ID]
		if !ok || a.IsCorrect != (a.SelectedOptionKey == q.Correct) {
			return false, errors.New("invalid graded answer")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO adaptive_answers(session_id,round_id,question_id,selected_option_key,is_correct) VALUES($1,$2,$3,$4,$5)`, sessionID, roundID, q.ID, a.SelectedOptionKey, a.IsCorrect); err != nil {
			return false, err
		}
		if q.Skill == "" {
			continue
		} // Legacy and non-pilot questions do not create invented skill evidence.
		var options []domain.AdaptiveOption
		if err := json.Unmarshal(q.Options, &options); err != nil {
			return false, err
		}
		fingerprint := domain.QuestionFingerprint(domain.AdaptiveQuestion{Question: q.Question, Options: options})
		var seen bool
		if err := tx.GetContext(ctx, &seen, `SELECT EXISTS(SELECT 1 FROM learning_evidence WHERE user_id=$1 AND skill_code=$2 AND fingerprint=$3 AND contributes)`, userID, q.Skill, fingerprint); err != nil {
			return false, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO learning_evidence(question_id,session_id,user_id,skill_code,fingerprint,is_correct,difficulty,contributes,policy_version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, q.ID, sessionID, userID, q.Skill, fingerprint, a.IsCorrect, q.Difficulty, !seen, domain.LearningPolicyVersion); err != nil {
			return false, err
		}
		if seen {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO learner_skill_states(user_id,skill_code,state) VALUES($1,$2,'{}') ON CONFLICT DO NOTHING`, userID, q.Skill); err != nil {
			return false, err
		}
		var raw []byte
		if err := tx.GetContext(ctx, &raw, `SELECT state FROM learner_skill_states WHERE user_id=$1 AND skill_code=$2 FOR UPDATE`, userID, q.Skill); err != nil {
			return false, err
		}
		var state domain.SkillState
		if err := json.Unmarshal(raw, &state); err != nil {
			return false, err
		}
		state.Code = q.Skill
		var sessions int
		if err := tx.GetContext(ctx, &sessions, `SELECT count(DISTINCT session_id) FROM learning_evidence WHERE user_id=$1 AND skill_code=$2 AND contributes`, userID, q.Skill); err != nil {
			return false, err
		}
		state = state.WithEvidence(a.IsCorrect, sessions, now)
		encoded, err := json.Marshal(state)
		if err != nil {
			return false, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE learner_skill_states SET state=$3 WHERE user_id=$1 AND skill_code=$2`, userID, q.Skill, string(encoded)); err != nil {
			return false, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE adaptive_rounds SET status='completed',completed_at=$2 WHERE id=$1`, roundID, now); err != nil {
		return false, fmt.Errorf("complete learning round: %w", err)
	}
	return true, tx.Commit()
}

var _ ports.LearnerRepository = (*Repository)(nil)
