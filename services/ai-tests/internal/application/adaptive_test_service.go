package application

import (
	"context"
	"errors"
	"fmt"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

const (
	minQuestionCount = 3
	maxQuestionCount = 10
	maxRounds        = 2
)

var (
	ErrInvalidQuestionCount = errors.New("question count must be between 3 and 10")
	ErrSessionNotInProgress = errors.New("session is not in progress")
	ErrRoundNotFound        = errors.New("round not found")
	ErrInvalidAnswer        = errors.New("invalid answer")
	ErrInvalidGeneratedRound = errors.New("invalid generated round")
)

type AdaptiveTestService struct {
	repository ports.SessionRepository
	courses    ports.CourseClient
	llm        ports.LLMClient
}

func NewAdaptiveTestService(
	repository ports.SessionRepository,
	courses ports.CourseClient,
	llm ports.LLMClient,
) *AdaptiveTestService {
	return &AdaptiveTestService{
		repository: repository,
		courses:    courses,
		llm:        llm,
	}
}

func (s *AdaptiveTestService) CreateSession(
	ctx context.Context,
	userID int64,
	courseID int64,
	topicPageID *int64,
	questionCount int,
) (domain.AdaptiveSession, error) {
	if questionCount < minQuestionCount || questionCount > maxQuestionCount {
		return domain.AdaptiveSession{}, ErrInvalidQuestionCount
	}

	courseContext, err := s.courses.GetCourseContext(ctx, userID, courseID, topicPageID)
	if err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("get course context: %w", err)
	}

	request := buildRoundRequest(
		courseContext,
		questionCount,
		1,
		domain.StrategyInitial,
		nil,
	)

	generated, metadata, err := s.llm.GenerateRound(ctx, request)
	if err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("generate first round: %w", err)
	}

	if err := validateGeneratedRound(generated, topicsToAllowed(courseContext.AllowedTopics), questionCount); err != nil {
		return domain.AdaptiveSession{}, err
	}

	sessionID := newSessionID()

	now := time.Now()

	session := domain.AdaptiveSession{
		ID:              sessionID,
		UserID:          userID,
		CourseID:        courseID,
		TopicPageID:     topicPageID,
		Status:          domain.SessionInProgress,
		CurrentRound:    1,
		QuestionCount:   questionCount,
		CreatedAt:       now,
	}

	round := generatedRoundToDomain(
		generated,
		metadata,
		sessionID,
		1,
		domain.StrategyInitial,
	)

	if err := s.repository.CreateSession(ctx, session); err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("create session: %w", err)
	}

	savedRound, err := s.repository.SaveRound(ctx, round)
	if err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("save first round: %w", err)
	}

	session.Rounds = []domain.AdaptiveRound{savedRound}
	return session, nil
}

func (s *AdaptiveTestService) GetSession(ctx context.Context, userID int64, sessionID string) (domain.AdaptiveSession, error) {
	if sessionID == "" || userID <= 0 {
		return domain.AdaptiveSession{}, ErrRoundNotFound
	}
	session, err := s.repository.GetSession(ctx, sessionID, userID)
	if err != nil {
		return domain.AdaptiveSession{}, err
	}
	return session, nil
}

func (s *AdaptiveTestService) SubmitAnswers(
	ctx context.Context,
	userID int64,
	sessionID string,
	roundID int64,
	answers map[int64]string,
) (domain.AdaptiveSession, error) {
	session, err := s.repository.GetSession(ctx, sessionID, userID)
	if err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("get session: %w", err)
	}

	if session.Status != domain.SessionInProgress {
		return domain.AdaptiveSession{}, ErrSessionNotInProgress
	}

	round, ok := findRound(session, roundID)
	if !ok {
		return domain.AdaptiveSession{}, ErrRoundNotFound
	}

	if round.RoundNumber != session.CurrentRound || round.Status != domain.RoundGenerated {
		return domain.AdaptiveSession{}, ErrSessionNotInProgress
	}

	if len(answers) != len(round.Questions) {
		return domain.AdaptiveSession{}, ErrInvalidAnswer
	}

	domainAnswers, result := evaluateAnswers(sessionID, round, answers)

	if len(domainAnswers) != len(round.Questions) {
		return domain.AdaptiveSession{}, ErrInvalidAnswer
	}

	if err := s.repository.SaveAnswers(ctx, sessionID, roundID, domainAnswers); err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("save answers: %w", err)
	}

	if err := s.repository.CompleteRound(ctx, sessionID, roundID); err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("complete round: %w", err)
	}

	for i := range session.Rounds {
		if session.Rounds[i].ID == roundID {
			session.Rounds[i].Status = domain.RoundCompleted
			session.Rounds[i].Answers = domainAnswers
			round = session.Rounds[i]
			break
		}
	}

	if round.RoundNumber == maxRounds {
		courseContext, err := s.courses.GetCourseContext(ctx, userID, session.CourseID, session.TopicPageID)
		if err != nil {
			return domain.AdaptiveSession{}, s.failAfterProcessing(ctx, sessionID, fmt.Errorf("get course context for feedback: %w", err))
		}

		feedback, err := s.generateFeedback(ctx, courseContext, session)
		if err != nil {
			return domain.AdaptiveSession{}, s.failAfterProcessing(ctx, sessionID, fmt.Errorf("generate feedback: %w", err))
		}

		if err := s.repository.CompleteSession(ctx, sessionID, feedback); err != nil {
			return domain.AdaptiveSession{}, fmt.Errorf("complete session: %w", err)
		}

		session.Status = domain.SessionCompleted
		session.Feedback = &feedback
		return session, nil
	}

	strategy := domain.StrategyTargetedPractice
	if result.CorrectCount == result.TotalCount {
		strategy = domain.StrategyIncreaseDifficulty
	}

	courseContext, err := s.courses.GetCourseContext(ctx, userID, session.CourseID, session.TopicPageID)
	if err != nil {
		return domain.AdaptiveSession{}, s.failAfterProcessing(ctx, sessionID, fmt.Errorf("get course context for second round: %w", err))
	}

	request := buildRoundRequest(
		courseContext,
		session.QuestionCount,
		2,
		strategy,
		[]ports.PreviousResultContext{
			{
				CorrectCount: result.CorrectCount,
				TotalCount:   result.TotalCount,
			},
		},
	)

	generated, metadata, err := s.llm.GenerateRound(ctx, request)
	if err != nil {
		return domain.AdaptiveSession{}, fmt.Errorf("generate second round: %w", err)
	}

	if err := validateGeneratedRound(generated, topicsToAllowed(courseContext.AllowedTopics), session.QuestionCount); err != nil {
		return domain.AdaptiveSession{}, s.failAfterProcessing(ctx, sessionID, err)
	}

	round2 := generatedRoundToDomain(
		generated,
		metadata,
		sessionID,
		2,
		strategy,
	)

	savedRound2, err := s.repository.SaveRound(ctx, round2)
	if err != nil {
		return domain.AdaptiveSession{}, s.failAfterProcessing(ctx, sessionID, fmt.Errorf("save second round: %w", err))
	}

	session.CurrentRound = 2
	session.Rounds = append(session.Rounds, savedRound2)
	return session, nil
}

func (s *AdaptiveTestService) failAfterProcessing(ctx context.Context, sessionID string, err error) error {
	if failErr := s.repository.FailSession(ctx, sessionID); failErr != nil {
		return fmt.Errorf("%w; fail session: %v", err, failErr)
	}
	return err
}

type roundResult struct {
	CorrectCount int
	TotalCount   int
}

func evaluateAnswers(
	sessionID string,
	round domain.AdaptiveRound,
	answers map[int64]string,
) ([]domain.AdaptiveAnswer, roundResult) {
	result := roundResult{TotalCount: len(round.Questions)}
	domainAnswers := make([]domain.AdaptiveAnswer, 0, len(round.Questions))

	for _, question := range round.Questions {
		selected, ok := answers[question.ID]
		if !ok {
			continue
		}

		correct := selected == question.CorrectOptionKey
		if correct {
			result.CorrectCount++
		}

		domainAnswers = append(domainAnswers, domain.AdaptiveAnswer{
			SessionID:         sessionID,
			RoundID:           round.ID,
			QuestionID:        question.ID,
			SelectedOptionKey: selected,
			IsCorrect:         correct,
		})
	}

	return domainAnswers, result
}

func findRound(session domain.AdaptiveSession, roundID int64) (domain.AdaptiveRound, bool) {
	for _, round := range session.Rounds {
		if round.ID == roundID {
			return round, true
		}
	}
	return domain.AdaptiveRound{}, false
}

func buildRoundRequest(
	c ports.CourseContext,
	questionCount int,
	roundNumber int,
	strategy domain.RoundStrategy,
	previous []ports.PreviousResultContext,
) ports.GenerateRoundRequest {
	materials := make([]ports.MaterialContext, 0, len(c.Materials))
	for _, material := range c.Materials {
		materials = append(materials, ports.MaterialContext{
			PageID: material.PageID,
			Title: material.Title,
			Content: material.Content,
		})
	}

	topics := make([]ports.AllowedTopic, 0, len(c.AllowedTopics))
	for _, topic := range c.AllowedTopics {
		topics = append(topics, ports.AllowedTopic{PageID: topic.PageID, Title: topic.Title})
	}

	tests := make([]ports.OrdinaryTestContext, 0, len(c.OrdinaryTests))
	for _, test := range c.OrdinaryTests {
		item := ports.OrdinaryTestContext{
			TestID: test.ID,
			Title: test.Title,
			Questions: make([]ports.OrdinaryQuestionContext, 0, len(test.Questions)),
		}
		for _, question := range test.Questions {
			item.Questions = append(item.Questions, ports.OrdinaryQuestionContext{
				Question: question.Question,
				Options: question.Options,
			})
		}
		tests = append(tests, item)
	}

	results := make([]ports.PreviousResultContext, 0, len(c.TestResults))
	for _, result := range c.TestResults {
		results = append(results, ports.PreviousResultContext{
			CorrectCount: result.CorrectCount,
			TotalCount: result.TotalCount,
		})
	}

	return ports.GenerateRoundRequest{
		CourseTitle: c.CourseTitle,
		Materials: materials,
		OrdinaryTests: tests,
		QuestionCount: questionCount,
		RoundNumber: roundNumber,
		Strategy: string(strategy),
		AllowedTopics: topics,
		PreviousResults: append(results, previous...),
	}
}

func generatedRoundToDomain(
	g ports.GeneratedRound,
	metadata ports.GenerationMetadata,
	sessionID string,
	roundNumber int,
	strategy domain.RoundStrategy,
) domain.AdaptiveRound {
	questions := make([]domain.AdaptiveQuestion, 0, len(g.Questions))

	for i, question := range g.Questions {
		options := make([]domain.AdaptiveOption, 0, len(question.Options))
		for j, option := range question.Options {
			options = append(options, domain.AdaptiveOption{
				Key:      option.Key,
				Text:     option.Text,
				Position: j + 1,
			})
		}

		sources := make([]domain.QuestionSource, 0)

		questions = append(questions, domain.AdaptiveQuestion{
			RoundID:          0,
			Position:         i + 1,
			TopicPageID:      question.TopicPageID,
			TopicTitle:       question.TopicTitle,
			Question:         question.Question,
			Difficulty:       question.Difficulty,
			CorrectOptionKey: question.CorrectOptionKey,
			Explanation:      question.Explanation,
			KnowledgeBasis:   domain.KnowledgeBasis(question.KnowledgeBasis),
			Sources:          sources,
			Options:          options,
		})
	}

	return domain.AdaptiveRound{
		SessionID:    sessionID,
		RoundNumber:  roundNumber,
		Strategy:     strategy,
		Status:       domain.RoundGenerated,
		Title:        g.Title,
		Instructions: g.Instructions,
		Model:        metadata.Model,
		PromptVersion: metadata.PromptVersion,
		PromptHash:    metadata.PromptHash,
		InputTokens:  metadata.InputTokens,
		OutputTokens: metadata.OutputTokens,
		Questions:    questions,
	}
}

func validateGeneratedRound(
	g ports.GeneratedRound,
	allowedTopics []ports.AllowedTopic,
	expectedQuestionCount int,
) error {
	if len(g.Questions) != expectedQuestionCount {
		return fmt.Errorf("%w: expected %d questions, got %d", ErrInvalidGeneratedRound, expectedQuestionCount, len(g.Questions))
	}

	allowed := make(map[int64]struct{}, len(allowedTopics))
	for _, topic := range allowedTopics {
		allowed[topic.PageID] = struct{}{}
	}

	for _, question := range g.Questions {
		if question.Question == "" || question.Explanation == "" {
			return fmt.Errorf("%w: empty question or explanation", ErrInvalidGeneratedRound)
		}

		if question.Difficulty < 1 || question.Difficulty > 5 {
			return fmt.Errorf("%w: invalid difficulty", ErrInvalidGeneratedRound)
		}

		if len(question.Options) < 2 || question.CorrectOptionKey == "" {
			return fmt.Errorf("%w: invalid options", ErrInvalidGeneratedRound)
		}

		keys := make(map[string]struct{}, len(question.Options))
		correctExists := false

		for _, option := range question.Options {
			if option.Key == "" || option.Text == "" {
				return fmt.Errorf("%w: empty option", ErrInvalidGeneratedRound)
			}
			if _, exists := keys[option.Key]; exists {
				return fmt.Errorf("%w: duplicate option key", ErrInvalidGeneratedRound)
			}
			keys[option.Key] = struct{}{}

			if option.Key == question.CorrectOptionKey {
				correctExists = true
			}
		}

		if !correctExists {
			return fmt.Errorf("%w: correct option does not exist", ErrInvalidGeneratedRound)
		}

		if question.TopicPageID != nil {
			if _, exists := allowed[*question.TopicPageID]; !exists {
				return fmt.Errorf("%w: unknown topic page", ErrInvalidGeneratedRound)
			}
		}

		switch domain.KnowledgeBasis(question.KnowledgeBasis) {
		case domain.BasisCourse, domain.BasisExternalKnowledge, domain.BasisMixed:
		default:
			return fmt.Errorf("%w: invalid knowledge basis", ErrInvalidGeneratedRound)
		}
	}

	return nil
}

func (s *AdaptiveTestService) generateFeedback(
	ctx context.Context,
	course ports.CourseContext,
	session domain.AdaptiveSession,
) (domain.AdaptiveFeedback, error) {
	topics := make([]ports.FeedbackTopicContext, 0, len(course.AllowedTopics))
	for _, topic := range course.AllowedTopics {
		topics = append(topics, ports.FeedbackTopicContext{
			PageID: topic.PageID,
			Title:  topic.Title,
		})
	}

	results := make([]ports.RoundResultContext, 0, len(session.Rounds))
	for _, round := range session.Rounds {
		result := ports.RoundResultContext{
			RoundNumber: round.RoundNumber,
			TotalCount:  len(round.Questions),
			Topics:      make([]ports.TopicResultContext, 0),
		}
		for _, answer := range round.Answers {
			if answer.IsCorrect {
				result.CorrectCount++
			}
			for _, question := range round.Questions {
				if question.ID == answer.QuestionID {
					result.Topics = append(result.Topics, ports.TopicResultContext{
						TopicPageID: question.TopicPageID,
						TopicTitle:  question.TopicTitle,
						CorrectCount: boolToInt(answer.IsCorrect),
						TotalCount:   1,
					})
					break
				}
			}
		}
		results = append(results, result)
	}

	generated, err := s.llm.GenerateFeedback(ctx, ports.GenerateFeedbackRequest{
		CourseTitle:   course.CourseTitle,
		Topics:        topics,
		RoundResults:  results,
		AllowedTopics: topicsToAllowed(course.AllowedTopics),
	})
	if err != nil {
		return domain.AdaptiveFeedback{}, err
	}

	return domain.AdaptiveFeedback{
		SessionID:       session.ID,
		Summary:         generated.Summary,
		MasteredTopics:  generatedTopicsToDomain(generated.MasteredTopics),
		TopicsToReview:  generatedTopicsToDomain(generated.TopicsToReview),
		NextSteps:       generatedNextStepsToDomain(generated.NextSteps),
		CreatedAt:       time.Now(),
	}, nil
}

func generatedTopicsToDomain(items []ports.GeneratedTopicFeedback) []domain.TopicFeedback {
	result := make([]domain.TopicFeedback, 0, len(items))
	for _, item := range items {
		result = append(result, domain.TopicFeedback{
			Title: item.Title, TopicPageID: item.TopicPageID, Reason: item.Reason,
		})
	}
	return result
}

func generatedNextStepsToDomain(items []ports.GeneratedNextStep) []domain.NextStep {
	result := make([]domain.NextStep, 0, len(items))
	for _, item := range items {
		result = append(result, domain.NextStep{
			Title: item.Title, Description: item.Description, TopicPageID: item.TopicPageID,
		})
	}
	return result
}

func boolToInt(value bool) int {
	if value { return 1 }
	return 0
}

func topicsToAllowed(topics []ports.CourseTopic) []ports.AllowedTopic {
	result := make([]ports.AllowedTopic, 0, len(topics))
	for _, topic := range topics {
		result = append(result, ports.AllowedTopic{
			PageID: topic.PageID,
			Title:  topic.Title,
		})
	}
	return result
}

func newSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic("failed to generate session id")
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(buf[0:4]),
		hex.EncodeToString(buf[4:6]),
		hex.EncodeToString(buf[6:8]),
		hex.EncodeToString(buf[8:10]),
		hex.EncodeToString(buf[10:16]),
	)
}
