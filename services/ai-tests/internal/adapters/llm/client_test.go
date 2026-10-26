package llm

import (
	"strings"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

func TestRoundPromptContainsConcreteMistakesAndUnknownAnswers(t *testing.T) {
	topic := int64(10)
	prompt := buildRoundPrompt(ports.GenerateRoundRequest{TopicTitle: "Input practice", Strategy: "targeted_practice", OrdinaryTests: []ports.OrdinaryTestContext{{Questions: []ports.OrdinaryQuestionContext{{Question: "input returns?", Answered: true, SelectedAnswer: "integer", CorrectAnswer: "string", IsCorrect: false}, {Question: "Unanswered"}}}}, PreviousResults: []ports.PreviousResultContext{{Source: "current_adaptive", TopicPageID: &topic, TopicTitle: "Input", CorrectCount: 0, TotalCount: 1, Questions: []ports.AnsweredQuestionContext{{Question: "input returns?", SelectedAnswer: "integer", CorrectAnswer: "string", Difficulty: 2}}}}})
	for _, want := range []string{"CURRENT TOPIC: Input practice", "Learner selected: integer; correct answer: string; correct: false", "Not answered; do not treat as an error.", `"Source": "current_adaptive"`, `"SelectedAnswer": "integer"`, `"Difficulty": 2`, "avoid repeating identical questions", "Missing results mean unknown knowledge"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestRoundPromptMakesSkillPlanAuthoritative(t *testing.T) {
	prompt := buildRoundPrompt(ports.GenerateRoundRequest{LearningPlan: []domain.QuestionTarget{{SkillCode: "math.linear.balance", TopicPageID: 10, Difficulty: 1, Reason: "remediate"}}, Skills: []domain.LearningSkill{{Code: "math.linear.balance", Title: "Линейные уравнения"}}})
	for _, want := range []string{"BACKEND LEARNING PLAN", `"skill_code":"math.linear.balance"`, `"difficulty":1`, "Do not replace the plan", "skillCode"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("missing %q", want)
		}
	}
}
