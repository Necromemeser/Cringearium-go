package domain

import "time"

type RoundStrategy string

const (
	StrategyInitial            RoundStrategy = "initial"
	StrategyIncreaseDifficulty RoundStrategy = "increase_difficulty"
	StrategyTargetedPractice   RoundStrategy = "targeted_practice"
)

type RoundStatus string

const (
	RoundGenerating RoundStatus = "generating"
	RoundGenerated  RoundStatus = "generated"
	RoundCompleted  RoundStatus = "completed"
	RoundFailed     RoundStatus = "failed"
)

type AdaptiveRound struct {
	ID           int64
	SessionID    string
	RoundNumber  int
	Strategy     RoundStrategy
	Status       RoundStatus
	Title        string
	Instructions string

	Model         string
	PromptVersion string
	PromptHash    string

	InputTokens  *int
	OutputTokens *int

	GenerationError *string
	CreatedAt       time.Time
	CompletedAt     *time.Time

	Questions []AdaptiveQuestion
}
