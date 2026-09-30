package application

import (
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

func TestEvaluateAnswers(t *testing.T) {
	round := domain.AdaptiveRound{
		ID: 10,
		Questions: []domain.AdaptiveQuestion{
			{ID: 1, CorrectOptionKey: "a"},
			{ID: 2, CorrectOptionKey: "b"},
		},
	}

	answers, result := evaluateAnswers("session", round, map[int64]string{
		1: "a",
		2: "x",
	})

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	if result.CorrectCount != 1 || result.TotalCount != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestValidateGeneratedRound(t *testing.T) {
	topics := []ports.AllowedTopic{{PageID: 10, Title: "Algebra"}}

	generated := ports.GeneratedRound{
		Questions: []ports.GeneratedQuestion{
			{
				TopicPageID:      int64Ptr(10),
				Question:         "2+2?",
				Difficulty:       1,
				Options: []ports.GeneratedOption{
					{Key: "a", Text: "4"},
					{Key: "b", Text: "5"},
				},
				CorrectOptionKey: "a",
				Explanation:      "Basic arithmetic.",
				KnowledgeBasis:   "course",
			},
		},
	}

	if err := validateGeneratedRound(generated, topics, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateGeneratedRoundRejectsUnknownTopic(t *testing.T) {
	generated := ports.GeneratedRound{
		Questions: []ports.GeneratedQuestion{
			{
				TopicPageID:      int64Ptr(999),
				Question:         "2+2?",
				Difficulty:       1,
				Options: []ports.GeneratedOption{
					{Key: "a", Text: "4"},
					{Key: "b", Text: "5"},
				},
				CorrectOptionKey: "a",
				Explanation:      "Basic arithmetic.",
				KnowledgeBasis:   "course",
			},
		},
	}

	if err := validateGeneratedRound(generated, nil, 1); err == nil {
		t.Fatal("expected unknown topic error")
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
