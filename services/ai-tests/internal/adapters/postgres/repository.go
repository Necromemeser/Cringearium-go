package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

var ErrNotFound = sql.ErrNoRows

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *DB) *Repository {
	return &Repository{db: db.conn}
}

func (r *Repository) CreateSession(ctx context.Context, session domain.AdaptiveSession) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO adaptive_sessions (
			id, user_id, course_id, topic_page_id, status,
			current_round, question_count, context_snapshot,
			created_at, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		session.ID,
		session.UserID,
		session.CourseID,
		session.TopicPageID,
		session.Status,
		session.CurrentRound,
		session.QuestionCount,
		emptyJSON(session.ContextSnapshot),
		session.CreatedAt,
		session.CompletedAt,
	)
	return err
}

func (r *Repository) GetSession(ctx context.Context, sessionID string, userID int64) (domain.AdaptiveSession, error) {
	var row sessionRow

	err := r.db.GetContext(ctx, &row, `
		SELECT
			id, user_id, course_id, topic_page_id, status,
			current_round, question_count, context_snapshot,
			created_at, completed_at
		FROM adaptive_sessions
		WHERE id = $1 AND user_id = $2
	`, sessionID, userID)
	if err != nil {
		return domain.AdaptiveSession{}, err
	}

	session := domain.AdaptiveSession{
		ID:              row.ID,
		UserID:          row.UserID,
		CourseID:        row.CourseID,
		TopicPageID:     row.TopicPageID,
		Status:          domain.SessionStatus(row.Status),
		CurrentRound:    row.CurrentRound,
		QuestionCount:   row.QuestionCount,
		ContextSnapshot: row.ContextSnapshot,
		CreatedAt:       row.CreatedAt,
		CompletedAt:     row.CompletedAt,
	}

	rounds, err := r.getRounds(ctx, sessionID)
	if err != nil {
		return domain.AdaptiveSession{}, err
	}

	session.Rounds = rounds
	return session, nil
}

func (r *Repository) SaveRound(ctx context.Context, round domain.AdaptiveRound) (domain.AdaptiveRound, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return domain.AdaptiveRound{}, err
	}
	defer tx.Rollback()

	var roundID int64

	err = tx.GetContext(ctx, &roundID, `
		INSERT INTO adaptive_rounds (
			session_id, round_number, strategy, status,
			title, instructions, model, prompt_version, prompt_hash,
			input_tokens, output_tokens, generation_error,
			created_at, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`,
		round.SessionID,
		round.RoundNumber,
		round.Strategy,
		round.Status,
		round.Title,
		round.Instructions,
		round.Model,
		round.PromptVersion,
		round.PromptHash,
		round.InputTokens,
		round.OutputTokens,
		round.GenerationError,
		round.CreatedAt,
		round.CompletedAt,
	)
	if err != nil {
		return domain.AdaptiveRound{}, err
	}

	round.ID = roundID

	if _, err := tx.ExecContext(ctx, `
		UPDATE adaptive_sessions
		SET current_round = $2
		WHERE id = $1 AND status = 'in_progress'
	`, round.SessionID, round.RoundNumber); err != nil {
		return domain.AdaptiveRound{}, err
	}

	for i := range round.Questions {
		question := &round.Questions[i]

		sources, err := json.Marshal(question.Sources)
		if err != nil {
			return domain.AdaptiveRound{}, err
		}

		var questionID int64
		err = tx.GetContext(ctx, &questionID, `
			INSERT INTO adaptive_questions (
				round_id, position, topic_page_id, topic_title,
				question, difficulty, correct_option_key,
				explanation, knowledge_basis, sources
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`,
			roundID,
			question.Position,
			question.TopicPageID,
			question.TopicTitle,
			question.Question,
			question.Difficulty,
			question.CorrectOptionKey,
			question.Explanation,
			question.KnowledgeBasis,
			sources,
		)
		if err != nil {
			return domain.AdaptiveRound{}, err
		}

		question.ID = questionID
		question.RoundID = roundID

		for i := range question.Options {
			option := &question.Options[i]
			var optionID int64

			err = tx.GetContext(ctx, &optionID, `
				INSERT INTO adaptive_options (
					question_id, option_key, text, position
				)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`,
				questionID,
				option.Key,
				option.Text,
				option.Position,
			)
			if err != nil {
				return domain.AdaptiveRound{}, err
			}

			option.ID = optionID
			option.QuestionID = questionID
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.AdaptiveRound{}, err
	}

	return round, nil
}

func (r *Repository) SaveAnswers(ctx context.Context, sessionID string, roundID int64, answers []domain.AdaptiveAnswer) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, answer := range answers {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO adaptive_answers (
				session_id, round_id, question_id,
				selected_option_key, is_correct
			)
			VALUES ($1, $2, $3, $4, $5)
		`,
			sessionID,
			roundID,
			answer.QuestionID,
			answer.SelectedOptionKey,
			answer.IsCorrect,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) CompleteRound(ctx context.Context, sessionID string, roundID int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE adaptive_rounds
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1 AND session_id = $2 AND status = 'generated'
	`, roundID, sessionID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("round cannot be completed")
	}

	return nil
}

func (r *Repository) FailSession(ctx context.Context, sessionID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE adaptive_sessions
		SET status = 'failed', completed_at = NOW()
		WHERE id = $1 AND status = 'in_progress'
	`, sessionID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("session cannot be failed")
	}

	return nil
}

func (r *Repository) CompleteSession(ctx context.Context, sessionID string, feedback domain.AdaptiveFeedback) error {
	mastered, err := json.Marshal(feedback.MasteredTopics)
	if err != nil {
		return err
	}
	review, err := json.Marshal(feedback.TopicsToReview)
	if err != nil {
		return err
	}
	nextSteps, err := json.Marshal(feedback.NextSteps)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO adaptive_feedback (
			session_id, summary, mastered_topics, topics_to_review, next_steps
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		sessionID, feedback.Summary, mastered, review, nextSteps,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE adaptive_sessions
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1 AND status = 'in_progress'
	`, sessionID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type sessionRow struct {
	ID              string     `db:"id"`
	UserID          int64      `db:"user_id"`
	CourseID        int64      `db:"course_id"`
	TopicPageID     *int64     `db:"topic_page_id"`
	Status          string     `db:"status"`
	CurrentRound    int        `db:"current_round"`
	QuestionCount   int        `db:"question_count"`
	ContextSnapshot []byte     `db:"context_snapshot"`
	CreatedAt       time.Time  `db:"created_at"`
	CompletedAt     *time.Time `db:"completed_at"`
}

type roundRow struct {
	ID               int64          `db:"id"`
	SessionID        string         `db:"session_id"`
	RoundNumber      int            `db:"round_number"`
	Strategy         string         `db:"strategy"`
	Status           string         `db:"status"`
	Title            string         `db:"title"`
	Instructions     string         `db:"instructions"`
	Model            sql.NullString `db:"model"`
	PromptVersion    sql.NullString `db:"prompt_version"`
	PromptHash       sql.NullString `db:"prompt_hash"`
	InputTokens      sql.NullInt64  `db:"input_tokens"`
	OutputTokens     sql.NullInt64  `db:"output_tokens"`
	GenerationError  sql.NullString `db:"generation_error"`
	CreatedAt        time.Time      `db:"created_at"`
	CompletedAt      *time.Time     `db:"completed_at"`
}

func (r *Repository) getRounds(ctx context.Context, sessionID string) ([]domain.AdaptiveRound, error) {
	var rows []roundRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, session_id, round_number, strategy, status,
		       title, instructions, model, prompt_version, prompt_hash,
		       input_tokens, output_tokens, generation_error,
		       created_at, completed_at
		FROM adaptive_rounds
		WHERE session_id = $1
		ORDER BY round_number
	`, sessionID); err != nil {
		return nil, err
	}

	result := make([]domain.AdaptiveRound, 0, len(rows))
	for _, row := range rows {
		round := domain.AdaptiveRound{
			ID:             row.ID,
			SessionID:      row.SessionID,
			RoundNumber:    row.RoundNumber,
			Strategy:       domain.RoundStrategy(row.Strategy),
			Status:         domain.RoundStatus(row.Status),
			Title:          row.Title,
			Instructions:   row.Instructions,
			Model:          row.Model.String,
			PromptVersion:  row.PromptVersion.String,
			PromptHash:     row.PromptHash.String,
			CreatedAt:      row.CreatedAt,
			CompletedAt:    row.CompletedAt,
		}

		questions, err := r.getQuestions(ctx, row.ID)
		if err != nil {
			return nil, err
		}

		round.Questions = questions
		if err := r.loadAnswers(ctx, sessionID, round.ID, &round.Answers); err != nil {
			return nil, err
		}
		result = append(result, round)
	}

	return result, nil
}


func (r *Repository) loadAnswers(
	ctx context.Context,
	sessionID string,
	roundID int64,
	target *[]domain.AdaptiveAnswer,
) error {
	var rows []answerRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT session_id, round_id, question_id, selected_option_key, is_correct, answered_at
		FROM adaptive_answers
		WHERE session_id = $1 AND round_id = $2
		ORDER BY question_id
	`, sessionID, roundID); err != nil {
		return err
	}

	answers := make([]domain.AdaptiveAnswer, 0, len(rows))
	for _, row := range rows {
		answers = append(answers, domain.AdaptiveAnswer{
			SessionID:         row.SessionID,
			RoundID:           row.RoundID,
			QuestionID:        row.QuestionID,
			SelectedOptionKey: row.SelectedOptionKey,
			IsCorrect:         row.IsCorrect,
			AnsweredAt:        row.AnsweredAt,
		})
	}

	*target = answers
}

type optionRow struct {
	ID         int64  `db:"id"`
	QuestionID int64  `db:"question_id"`
	Key        string `db:"key"`
	Text       string `db:"text"`
	Position   int    `db:"position"`
}

type answerRow struct {
	SessionID         string    `db:"session_id"`
	RoundID           int64     `db:"round_id"`
	QuestionID        int64     `db:"question_id"`
	SelectedOptionKey string    `db:"selected_option_key"`
	IsCorrect         bool      `db:"is_correct"`
	AnsweredAt        time.Time `db:"answered_at"`
}

type questionRow struct {
	ID               int64          `db:"id"`
	RoundID          int64          `db:"round_id"`
	Position         int            `db:"position"`
	TopicPageID      *int64         `db:"topic_page_id"`
	TopicTitle       sql.NullString `db:"topic_title"`
	Question         string         `db:"question"`
	Difficulty       int            `db:"difficulty"`
	CorrectOptionKey string         `db:"correct_option_key"`
	Explanation      string         `db:"explanation"`
	KnowledgeBasis   string         `db:"knowledge_basis"`
	Sources          []byte         `db:"sources"`
}

func (r *Repository) getQuestions(ctx context.Context, roundID int64) ([]domain.AdaptiveQuestion, error) {
	var rows []questionRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, round_id, position, topic_page_id, topic_title,
		       question, difficulty, correct_option_key, explanation,
		       knowledge_basis, sources
		FROM adaptive_questions
		WHERE round_id = $1
		ORDER BY position
	`, roundID); err != nil {
		return nil, err
	}

	result := make([]domain.AdaptiveQuestion, 0, len(rows))
	for _, row := range rows {
		var sources []domain.QuestionSource
		if len(row.Sources) > 0 {
			if err := json.Unmarshal(row.Sources, &sources); err != nil {
				return nil, err
			}
		}

		question := domain.AdaptiveQuestion{
			ID:               row.ID,
			RoundID:          row.RoundID,
			Position:         row.Position,
			TopicPageID:      row.TopicPageID,
			TopicTitle:       row.TopicTitle.String,
			Question:         row.Question,
			Difficulty:       row.Difficulty,
			CorrectOptionKey: row.CorrectOptionKey,
			Explanation:      row.Explanation,
			KnowledgeBasis:   domain.KnowledgeBasis(row.KnowledgeBasis),
			Sources:          sources,
		}

		var options []optionRow
		if err := r.db.SelectContext(ctx, &options, `
			SELECT id, question_id, option_key AS key, text, position
			FROM adaptive_options
			WHERE question_id = $1
			ORDER BY position
		`, row.ID); err != nil {
			return nil, err
		}

		question.Options = make([]domain.AdaptiveOption, 0, len(options))
		for _, option := range options {
			question.Options = append(question.Options, domain.AdaptiveOption{
				ID:         option.ID,
				QuestionID: option.QuestionID,
				Key:        option.Key,
				Text:       option.Text,
				Position:   option.Position,
			})
		}
		result = append(result, question)
	}

	return result, nil
}

func emptyJSON(value []byte) []byte {
	if len(value) == 0 {
		return []byte("{}")
	}
	return value
}

var _ ports.SessionRepository = (*Repository)(nil)

