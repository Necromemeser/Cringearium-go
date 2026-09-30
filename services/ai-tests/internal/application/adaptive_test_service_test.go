package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

type sessionRepositoryMock struct {
	sessions    map[string]domain.AdaptiveSession
	nextRoundID int64
}

func newSessionRepositoryMock() *sessionRepositoryMock {
	return &sessionRepositoryMock{
		sessions:    make(map[string]domain.AdaptiveSession),
		nextRoundID: 1,
	}
}

func (r *sessionRepositoryMock) CreateSession(_ context.Context, session domain.AdaptiveSession) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *sessionRepositoryMock) GetSession(_ context.Context, sessionID string, userID int64) (domain.AdaptiveSession, error) {
	session, ok := r.sessions[sessionID]
	if !ok || session.UserID != userID {
		return domain.AdaptiveSession{}, errors.New("not found")
	}
	return cloneSession(session), nil
}

func (r *sessionRepositoryMock) SaveRound(_ context.Context, round domain.AdaptiveRound) (domain.AdaptiveRound, error) {
	round.ID = r.nextRoundID
	r.nextRoundID++

	for i := range round.Questions {
		round.Questions[i].ID = int64(i + 1)
		round.Questions[i].RoundID = round.ID
		for j := range round.Questions[i].Options {
			round.Questions[i].Options[j].ID = int64(j + 1)
			round.Questions[i].Options[j].QuestionID = round.Questions[i].ID
		}
	}

	session := r.sessions[round.SessionID]
	session.Rounds = append(session.Rounds, round)
	r.sessions[round.SessionID] = session
	return round, nil
}

func (r *sessionRepositoryMock) SaveAnswers(_ context.Context, sessionID string, roundID int64, answers []domain.AdaptiveAnswer) error {
	session := r.sessions[sessionID]
	for i := range session.Rounds {
		if session.Rounds[i].ID == roundID {
			session.Rounds[i].Answers = append([]domain.AdaptiveAnswer(nil), answers...)
			r.sessions[sessionID] = session
			return nil
		}
	}
	return errors.New("round not found")
}

func (r *sessionRepositoryMock) CompleteRound(_ context.Context, sessionID string, roundID int64) error {
	session := r.sessions[sessionID]
	for i := range session.Rounds {
		if session.Rounds[i].ID == roundID {
			session.Rounds[i].Status = domain.RoundCompleted
			now := time.Now()
			session.Rounds[i].CompletedAt = &now
			r.sessions[sessionID] = session
			return nil
		}
	}
	return errors.New("round not found")
}

func (r *sessionRepositoryMock) FailSession(_ context.Context, sessionID string) error {
	session := r.sessions[sessionID]
	session.Status = domain.SessionFailed
	now := time.Now()
	session.CompletedAt = &now
	r.sessions[sessionID] = session
	return nil
}

func (r *sessionRepositoryMock) CompleteSession(_ context.Context, sessionID string, feedback domain.AdaptiveFeedback) error {
	session := r.sessions[sessionID]
	session.Status = domain.SessionCompleted
	now := time.Now()
	session.CompletedAt = &now
	session.Feedback = &feedback
	r.sessions[sessionID] = session
	return nil
}

type courseClientMock struct {
	context ports.CourseContext
	calls   int
	err     error
}

func (c *courseClientMock) GetCourseContext(_ context.Context, _ int64, _ int64, _ *int64) (ports.CourseContext, error) {
	c.calls++
	if c.err != nil {
		return ports.CourseContext{}, c.err
	}
	return c.context, nil
}

type llmClientMock struct {
	rounds        []ports.GeneratedRound
	feedback      ports.GeneratedFeedback
	roundCalls    []ports.GenerateRoundRequest
	feedbackCalls []ports.GenerateFeedbackRequest
	roundErr      error
	feedbackErr   error
}

func (l *llmClientMock) GenerateRound(_ context.Context, request ports.GenerateRoundRequest) (ports.GeneratedRound, ports.GenerationMetadata, error) {
	l.roundCalls = append(l.roundCalls, request)
	if l.roundErr != nil {
		return ports.GeneratedRound{}, ports.GenerationMetadata{}, l.roundErr
	}
	if len(l.rounds) == 0 {
		return ports.GeneratedRound{}, ports.GenerationMetadata{}, errors.New("no generated round")
	}
	result := l.rounds[0]
	l.rounds = l.rounds[1:]
	return result, ports.GenerationMetadata{Model: "test-model", PromptVersion: "test-v1"}, nil
}

func (l *llmClientMock) GenerateFeedback(_ context.Context, request ports.GenerateFeedbackRequest) (ports.GeneratedFeedback, error) {
	l.feedbackCalls = append(l.feedbackCalls, request)
	if l.feedbackErr != nil {
		return ports.GeneratedFeedback{}, l.feedbackErr
	}
	return l.feedback, nil
}

var _ ports.SessionRepository = (*sessionRepositoryMock)(nil)
var _ ports.CourseClient = (*courseClientMock)(nil)
var _ ports.LLMClient = (*llmClientMock)(nil)

func testCourseContext() ports.CourseContext {
	return ports.CourseContext{
		CourseID:    42,
		CourseTitle: "Go",
		Materials: []ports.CourseMaterial{
			{PageID: 10, Title: "Variables", Content: "Variables store values."},
		},
		OrdinaryTests: []ports.OrdinaryTest{
			{
				ID:    100,
				PageID: 10,
				Title: "Variables test",
				Questions: []ports.OrdinaryTestQuestion{
					{Question: "What stores a value?", Options: []string{"variable", "loop"}},
				},
			},
		},
		TestResults: []ports.OrdinaryTestResult{
			{TestID: 100, TopicPageID: int64Ptr(10), TopicTitle: "Variables", CorrectCount: 1, TotalCount: 2},
		},
		AllowedTopics: []ports.CourseTopic{
			{PageID: 10, Title: "Variables"},
			{PageID: 11, Title: "Loops"},
		},
	}
}

func generatedRound(questionCount int, topicPageID int64) ports.GeneratedRound {
	questions := make([]ports.GeneratedQuestion, 0, questionCount)
	for i := 0; i < questionCount; i++ {
		questions = append(questions, ports.GeneratedQuestion{
			TopicPageID: &topicPageID,
			TopicTitle:  "Variables",
			Question:    "Question",
			Difficulty:  2,
			Options: []ports.GeneratedOption{
				{Key: "A", Text: "Correct"},
				{Key: "B", Text: "Wrong"},
			},
			CorrectOptionKey: "A",
			Explanation:      "Explanation",
			KnowledgeBasis:   string(domain.BasisCourse),
		})
	}
	return ports.GeneratedRound{
		Title:        "Round",
		Instructions: "Answer all questions",
		Questions:    questions,
	}
}

func cloneSession(session domain.AdaptiveSession) domain.AdaptiveSession {
	result := session
	result.Rounds = append([]domain.AdaptiveRound(nil), session.Rounds...)
	for i := range result.Rounds {
		result.Rounds[i].Questions = append([]domain.AdaptiveQuestion(nil), session.Rounds[i].Questions...)
		result.Rounds[i].Answers = append([]domain.AdaptiveAnswer(nil), session.Rounds[i].Answers...)
		for j := range result.Rounds[i].Questions {
			result.Rounds[i].Questions[j].Options = append([]domain.AdaptiveOption(nil), result.Rounds[i].Questions[j].Options...)
		}
	}
	return result
}

func TestCreateSession(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{rounds: []ports.GeneratedRound{generatedRound(3, 10)}}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, int64Ptr(10), 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if session.Status != domain.SessionInProgress {
		t.Fatalf("status = %q, want %q", session.Status, domain.SessionInProgress)
	}
	if session.CurrentRound != 1 {
		t.Fatalf("current round = %d, want 1", session.CurrentRound)
	}
	if len(session.Rounds) != 1 || len(session.Rounds[0].Questions) != 3 {
		t.Fatalf("round/questions = %d/%d, want 1/3", len(session.Rounds), len(session.Rounds[0].Questions))
	}
	if session.Rounds[0].ID == 0 {
		t.Fatal("saved round ID was not assigned")
	}
	if len(llm.roundCalls) != 1 {
		t.Fatalf("GenerateRound calls = %d, want 1", len(llm.roundCalls))
	}
	if llm.roundCalls[0].RoundNumber != 1 || llm.roundCalls[0].Strategy != string(domain.StrategyInitial) {
		t.Fatalf("first request = round %d, strategy %q", llm.roundCalls[0].RoundNumber, llm.roundCalls[0].Strategy)
	}
}

func TestCreateSessionRejectsInvalidQuestionCount(t *testing.T) {
	service := NewAdaptiveTestService(
		newSessionRepositoryMock(),
		&courseClientMock{context: testCourseContext()},
		&llmClientMock{},
	)

	for _, count := range []int{0, 2, 11} {
		_, err := service.CreateSession(context.Background(), 7, 42, nil, count)
		if !errors.Is(err, ErrInvalidQuestionCount) {
			t.Errorf("question count %d: error = %v, want ErrInvalidQuestionCount", count, err)
		}
	}
}

func TestSubmitAnswersUsesIncreaseDifficultyWhenAllCorrect(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{
		rounds: []ports.GeneratedRound{
			generatedRound(3, 10),
			generatedRound(3, 11),
		},
	}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	answers := map[int64]string{}
	for _, question := range session.Rounds[0].Questions {
		answers[question.ID] = "A"
	}

	updated, err := service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[0].ID, answers)
	if err != nil {
		t.Fatalf("SubmitAnswers() error = %v", err)
	}

	if updated.CurrentRound != 2 || len(updated.Rounds) != 2 {
		t.Fatalf("round state = current %d, rounds %d, want 2/2", updated.CurrentRound, len(updated.Rounds))
	}
	if updated.Rounds[1].Strategy != domain.StrategyIncreaseDifficulty {
		t.Fatalf("strategy = %q, want %q", updated.Rounds[1].Strategy, domain.StrategyIncreaseDifficulty)
	}
	if llm.roundCalls[1].Strategy != string(domain.StrategyIncreaseDifficulty) {
		t.Fatalf("LLM strategy = %q, want %q", llm.roundCalls[1].Strategy, domain.StrategyIncreaseDifficulty)
	}
}

func TestSubmitAnswersUsesTargetedPracticeWhenNotAllCorrect(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{
		rounds: []ports.GeneratedRound{
			generatedRound(3, 10),
			generatedRound(3, 10),
		},
	}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	answers := map[int64]string{}
	for i, question := range session.Rounds[0].Questions {
		answers[question.ID] = "A"
		if i == 1 {
			answers[question.ID] = "B"
		}
	}

	updated, err := service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[0].ID, answers)
	if err != nil {
		t.Fatalf("SubmitAnswers() error = %v", err)
	}

	if updated.Rounds[1].Strategy != domain.StrategyTargetedPractice {
		t.Fatalf("strategy = %q, want %q", updated.Rounds[1].Strategy, domain.StrategyTargetedPractice)
	}
	if llm.roundCalls[1].Strategy != string(domain.StrategyTargetedPractice) {
		t.Fatalf("LLM strategy = %q, want %q", llm.roundCalls[1].Strategy, domain.StrategyTargetedPractice)
	}
}

func TestSubmitAnswersCompletesSessionAndGeneratesFeedback(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{
		rounds: []ports.GeneratedRound{
			generatedRound(3, 10),
			generatedRound(3, 11),
		},
		feedback: ports.GeneratedFeedback{
			Summary: "Good work",
			MasteredTopics: []ports.GeneratedTopicFeedback{
				{Title: "Variables", TopicPageID: int64Ptr(10), Reason: "Strong result"},
			},
			TopicsToReview: []ports.GeneratedTopicFeedback{
				{Title: "Loops", TopicPageID: int64Ptr(11), Reason: "Review needed"},
			},
			NextSteps: []ports.GeneratedNextStep{
				{Title: "Practice", Description: "Solve more tasks", TopicPageID: int64Ptr(11)},
			},
		},
	}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	firstAnswers := map[int64]string{}
	for _, question := range session.Rounds[0].Questions {
		firstAnswers[question.ID] = "A"
	}

	session, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[0].ID, firstAnswers)
	if err != nil {
		t.Fatal(err)
	}

	secondAnswers := map[int64]string{}
	for _, question := range session.Rounds[1].Questions {
		secondAnswers[question.ID] = "B"
	}

	completed, err := service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[1].ID, secondAnswers)
	if err != nil {
		t.Fatalf("second SubmitAnswers() error = %v", err)
	}

	if completed.Status != domain.SessionCompleted {
		t.Fatalf("status = %q, want %q", completed.Status, domain.SessionCompleted)
	}
	if completed.Feedback == nil || completed.Feedback.Summary != "Good work" {
		t.Fatalf("feedback = %#v, want generated feedback", completed.Feedback)
	}
	if len(llm.feedbackCalls) != 1 {
		t.Fatalf("GenerateFeedback calls = %d, want 1", len(llm.feedbackCalls))
	}
	if len(llm.feedbackCalls[0].RoundResults) != 2 {
		t.Fatalf("feedback round results = %d, want 2", len(llm.feedbackCalls[0].RoundResults))
	}
	if llm.feedbackCalls[0].RoundResults[1].CorrectCount != 0 {
		t.Fatalf("second round correct count = %d, want 0", llm.feedbackCalls[0].RoundResults[1].CorrectCount)
	}
}

func TestSubmitAnswersRejectsInvalidAnswerCount(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{rounds: []ports.GeneratedRound{generatedRound(3, 10)}}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[0].ID, map[int64]string{
		session.Rounds[0].Questions[0].ID: "A",
	})
	if !errors.Is(err, ErrInvalidAnswer) {
		t.Fatalf("error = %v, want ErrInvalidAnswer", err)
	}
}

func TestSubmitAnswersRejectsUnknownRound(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{rounds: []ports.GeneratedRound{generatedRound(3, 10)}}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.SubmitAnswers(context.Background(), 7, session.ID, 999, map[int64]string{})
	if !errors.Is(err, ErrRoundNotFound) {
		t.Fatalf("error = %v, want ErrRoundNotFound", err)
	}
}

func TestSubmitAnswersRejectsCompletedSession(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{
		rounds: []ports.GeneratedRound{
			generatedRound(3, 10),
			generatedRound(3, 11),
		},
		feedback: ports.GeneratedFeedback{Summary: "Done"},
	}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	firstAnswers := map[int64]string{}
	for _, question := range session.Rounds[0].Questions {
		firstAnswers[question.ID] = "A"
	}
	session, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[0].ID, firstAnswers)
	if err != nil {
		t.Fatal(err)
	}

	secondAnswers := map[int64]string{}
	for _, question := range session.Rounds[1].Questions {
		secondAnswers[question.ID] = "A"
	}
	session, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[1].ID, secondAnswers)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[1].ID, secondAnswers)
	if !errors.Is(err, ErrSessionNotInProgress) {
		t.Fatalf("error = %v, want ErrSessionNotInProgress", err)
	}
}

func TestCreateSessionRejectsInvalidGeneratedRound(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	llm := &llmClientMock{rounds: []ports.GeneratedRound{generatedRound(3, 999)}}
	service := NewAdaptiveTestService(repository, courses, llm)

	_, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if !errors.Is(err, ErrInvalidGeneratedRound) {
		t.Fatalf("error = %v, want ErrInvalidGeneratedRound", err)
	}
	if len(repository.sessions) != 0 {
		t.Fatalf("sessions = %d, want 0", len(repository.sessions))
	}
}

func TestCreateSessionPropagatesLLMError(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	expected := errors.New("llm unavailable")
	llm := &llmClientMock{rounds: []ports.GeneratedRound{{}}, roundErr: expected}
	service := NewAdaptiveTestService(repository, courses, llm)

	_, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if !errors.Is(err, expected) {
		t.Fatalf("error = %v, want wrapped LLM error", err)
	}
	if len(repository.sessions) != 0 {
		t.Fatalf("sessions = %d, want 0", len(repository.sessions))
	}
}

func TestSubmitAnswersFailsSessionWhenSecondRoundGenerationFails(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	expected := errors.New("llm unavailable")
	llm := &llmClientMock{
		rounds:   []ports.GeneratedRound{generatedRound(3, 10)},
		roundErr: expected,
	}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	answers := map[int64]string{}
	for _, question := range session.Rounds[0].Questions {
		answers[question.ID] = "A"
	}

	_, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[0].ID, answers)
	if !errors.Is(err, expected) {
		t.Fatalf("error = %v, want wrapped LLM error", err)
	}

	failed, err := repository.GetSession(context.Background(), session.ID, 7)
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != domain.SessionFailed {
		t.Fatalf("status = %q, want %q", failed.Status, domain.SessionFailed)
	}
}

func TestSubmitAnswersFailsSessionWhenFeedbackGenerationFails(t *testing.T) {
	repository := newSessionRepositoryMock()
	courses := &courseClientMock{context: testCourseContext()}
	expected := errors.New("feedback unavailable")
	llm := &llmClientMock{
		rounds:      []ports.GeneratedRound{generatedRound(3, 10), generatedRound(3, 11)},
		feedbackErr: expected,
	}
	service := NewAdaptiveTestService(repository, courses, llm)

	session, err := service.CreateSession(context.Background(), 7, 42, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	firstAnswers := map[int64]string{}
	for _, question := range session.Rounds[0].Questions {
		firstAnswers[question.ID] = "A"
	}
	session, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[0].ID, firstAnswers)
	if err != nil {
		t.Fatal(err)
	}

	secondAnswers := map[int64]string{}
	for _, question := range session.Rounds[1].Questions {
		secondAnswers[question.ID] = "A"
	}

	_, err = service.SubmitAnswers(context.Background(), 7, session.ID, session.Rounds[1].ID, secondAnswers)
	if !errors.Is(err, expected) {
		t.Fatalf("error = %v, want wrapped feedback error", err)
	}

	failed, err := repository.GetSession(context.Background(), session.ID, 7)
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != domain.SessionFailed {
		t.Fatalf("status = %q, want %q", failed.Status, domain.SessionFailed)
	}
}

func TestBuildRoundRequestIncludesCourseContext(t *testing.T) {
	context := testCourseContext()
	request := buildRoundRequest(context, 5, 2, domain.StrategyTargetedPractice, nil)

	if request.CourseTitle != "Go" {
		t.Fatalf("course title = %q, want Go", request.CourseTitle)
	}
	if len(request.Materials) != 1 || request.Materials[0].PageID != 10 {
		t.Fatalf("materials = %#v, want one material with page 10", request.Materials)
	}
	if len(request.OrdinaryTests) != 1 || len(request.OrdinaryTests[0].Questions) != 1 {
		t.Fatalf("ordinary tests = %#v, want one test with one question", request.OrdinaryTests)
	}
	if len(request.PreviousResults) != 1 || request.PreviousResults[0].CorrectCount != 1 {
		t.Fatalf("previous results = %#v, want course result", request.PreviousResults)
	}
	if len(request.AllowedTopics) != 2 {
		t.Fatalf("allowed topics = %d, want 2", len(request.AllowedTopics))
	}
	if request.RoundNumber != 2 || request.Strategy != string(domain.StrategyTargetedPractice) {
		t.Fatalf("round metadata = %d/%q", request.RoundNumber, request.Strategy)
	}
}
