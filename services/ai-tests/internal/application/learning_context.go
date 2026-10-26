package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

const historySessionLimit = 3
const historyMaxAge = 30 * 24 * time.Hour

// A snapshot keeps subsequent rounds tied to the materials and evidence used at creation.
type learningSnapshot struct {
	InitialProfile domain.LearningProfile
	Version        int
	Course         ports.CourseContext
	History        []ports.PreviousResultContext
}

func (s *AdaptiveTestService) sessionContext(ctx context.Context, session domain.AdaptiveSession) (learningSnapshot, error) {
	var snapshot learningSnapshot
	if len(session.ContextSnapshot) > 0 && string(session.ContextSnapshot) != "{}" {
		if err := json.Unmarshal(session.ContextSnapshot, &snapshot); err != nil {
			return snapshot, fmt.Errorf("decode learning snapshot: %w", err)
		}
		if snapshot.Version != 1 {
			return snapshot, fmt.Errorf("unsupported learning snapshot version %d", snapshot.Version)
		}
		return snapshot, nil
	}
	// Sessions created before snapshots were introduced remain resumable.
	course, err := s.courses.GetCourseContext(ctx, session.UserID, session.CourseID, session.TopicPageID)
	return learningSnapshot{Version: 1, Course: course}, err
}

// Group actual answered questions by topic; never infer weaknesses from only a total score.
func roundEvidence(session domain.AdaptiveSession, source string) []ports.PreviousResultContext {
	var results []ports.PreviousResultContext
	for _, round := range session.Rounds {
		if round.Status != domain.RoundCompleted {
			continue
		}
		answers := make(map[int64]domain.AdaptiveAnswer, len(round.Answers))
		for _, a := range round.Answers {
			answers[a.QuestionID] = a
		}
		groups := make(map[int64]int)
		for _, q := range round.Questions {
			answer, ok := answers[q.ID]
			if !ok {
				continue
			}
			topicID := int64(0)
			if q.TopicPageID != nil {
				topicID = *q.TopicPageID
			}
			index, exists := groups[topicID]
			if !exists {
				item := ports.PreviousResultContext{Source: source, SessionID: session.ID, RoundNumber: round.RoundNumber, TopicPageID: q.TopicPageID, TopicTitle: q.TopicTitle}
				if round.CompletedAt != nil {
					item.CompletedAt = round.CompletedAt.UTC().Format(time.RFC3339)
				}
				results = append(results, item)
				index = len(results) - 1
				groups[topicID] = index
			}
			result := &results[index]
			selected, correct := "", ""
			for _, option := range q.Options {
				if option.Key == answer.SelectedOptionKey {
					selected = option.Text
				}
				if option.Key == q.CorrectOptionKey {
					correct = option.Text
				}
			}
			result.TotalCount++
			if answer.IsCorrect {
				result.CorrectCount++
			}
			result.Questions = append(result.Questions, ports.AnsweredQuestionContext{Question: q.Question, SelectedAnswer: selected, CorrectAnswer: correct, IsCorrect: answer.IsCorrect, Difficulty: q.Difficulty})
		}
	}
	return results
}

// Old sessions could generate questions anywhere in the course. Only relevant evidence
// belongs in a newly scoped practice, even when the practice page ID is unchanged.
func scopedEvidence(evidence []ports.PreviousResultContext, topics []ports.CourseTopic) []ports.PreviousResultContext {
	allowed := make(map[int64]bool, len(topics))
	for _, topic := range topics {
		allowed[topic.PageID] = true
	}
	var result []ports.PreviousResultContext
	for _, item := range evidence {
		if item.TopicPageID != nil && allowed[*item.TopicPageID] {
			result = append(result, item)
		}
	}
	return result
}
