package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/internal/testsupport"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

type fixedCourse struct {
	calls  int
	skills []domain.LearningSkill
}

func (c *fixedCourse) GetCourseContext(context.Context, int64, int64, *int64) (ports.CourseContext, error) {
	c.calls++
	return ports.CourseContext{Skills: c.skills, CourseID: 42, CourseTitle: "Python", TopicTitle: "Input practice", Materials: []ports.CourseMaterial{{PageID: 10, Title: "Input", Content: "input returns a string"}}, AllowedTopics: []ports.CourseTopic{{PageID: 10, Title: "Input"}}}, nil
}

type recordingGenerator struct {
	requests   []ports.GenerateRoundRequest
	failSecond bool
}

func (l *recordingGenerator) GenerateRound(_ context.Context, req ports.GenerateRoundRequest) (ports.GeneratedRound, ports.GenerationMetadata, error) {
	l.requests = append(l.requests, req)
	if l.failSecond && req.RoundNumber == 2 {
		return ports.GeneratedRound{}, ports.GenerationMetadata{}, fmt.Errorf("test generation failure")
	}
	result := ports.GeneratedRound{Title: "Input"}
	topic := int64(10)
	for i := 0; i < req.QuestionCount; i++ {
		result.Questions = append(result.Questions, ports.GeneratedQuestion{TopicPageID: &topic, TopicTitle: "Input", Question: fmt.Sprintf("Question %d/%d: input returns?", len(l.requests), i), Difficulty: 2, Options: []ports.GeneratedOption{{Key: "A", Text: "string"}, {Key: "B", Text: "integer"}}, CorrectOptionKey: "A", Explanation: "input returns a string", KnowledgeBasis: "course"})
	}
	if len(req.LearningPlan) > 0 {
		for i, target := range req.LearningPlan {
			result.Questions[i].SkillCode = target.SkillCode
			id := target.TopicPageID
			result.Questions[i].TopicPageID = &id
			result.Questions[i].Difficulty = target.Difficulty
		}
	}
	return result, ports.GenerationMetadata{Model: "fake", PromptVersion: "integration"}, nil
}
func (*recordingGenerator) GenerateFeedback(context.Context, ports.GenerateFeedbackRequest) (ports.GeneratedFeedback, error) {
	return ports.GeneratedFeedback{Summary: "Keep practicing input"}, nil
}

func TestLearningEvidenceSurvivesPersistenceAndPersonalizesNextSession(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	courses := &fixedCourse{}
	llm := &recordingGenerator{}
	service := application.NewAdaptiveTestService(repo, courses, llm, repo)
	ctx := context.Background()
	practice := int64(100)
	session, err := service.CreateSession(ctx, 7, 42, &practice, 3)
	if err != nil {
		t.Fatal(err)
	}
	var stored []byte
	if err := conn.Get(&stored, `SELECT context_snapshot FROM adaptive_sessions WHERE id=$1`, session.ID); err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Version int
		Course  ports.CourseContext
	}
	if err := json.Unmarshal(stored, &snapshot); err != nil || snapshot.Version != 1 || snapshot.Course.TopicTitle != "Input practice" {
		t.Fatalf("snapshot=%s err=%v", stored, err)
	}
	answers := map[int64]string{}
	for i, q := range session.Rounds[0].Questions {
		answers[q.ID] = "A"
		if i == 0 {
			answers[q.ID] = "B"
		}
	}
	session, err = service.SubmitAnswers(ctx, 7, session.ID, session.Rounds[0].ID, answers)
	if err != nil {
		t.Fatal(err)
	}
	req := llm.requests[1]
	if req.Strategy != "targeted_practice" || len(req.PreviousResults) != 1 {
		t.Fatalf("no targeted evidence: %+v", req)
	}
	evidence := req.PreviousResults[0]
	if evidence.Source != "current_adaptive" || evidence.TopicPageID == nil || *evidence.TopicPageID != 10 || evidence.CorrectCount != 2 || evidence.TotalCount != 3 || len(evidence.Questions) != 3 || evidence.Questions[0].SelectedAnswer != "integer" || evidence.Questions[0].CorrectAnswer != "string" || evidence.Questions[0].IsCorrect {
		t.Fatalf("lost error details: %+v", evidence)
	}
	answers = map[int64]string{}
	for _, q := range session.Rounds[1].Questions {
		answers[q.ID] = "A"
	}
	session, err = service.SubmitAnswers(ctx, 7, session.ID, session.Rounds[1].ID, answers)
	if err != nil {
		t.Fatal(err)
	}
	if session.Status != domain.SessionCompleted {
		t.Fatal(session.Status)
	}
	if courses.calls != 1 {
		t.Fatalf("course snapshot not reused: calls=%d", courses.calls)
	}
	if _, err := service.CreateSession(ctx, 7, 42, &practice, 3); err != nil {
		t.Fatal(err)
	}
	if len(llm.requests[2].PreviousResults) != 2 || llm.requests[2].PreviousResults[0].Source != "past_adaptive" || llm.requests[2].PreviousResults[0].Questions[0].SelectedAnswer != "integer" {
		t.Fatalf("history not reused: %+v", llm.requests[2].PreviousResults)
	}
	for _, tc := range []struct{ user, course, scope int64 }{{8, 42, 100}, {7, 43, 100}, {7, 42, 101}} {
		sessions, err := repo.ListCompletedSessions(ctx, tc.user, tc.course, &tc.scope, time.Now().Add(-time.Hour), 3)
		if err != nil || len(sessions) != 0 {
			t.Fatalf("history leaked for %+v: %+v, %v", tc, sessions, err)
		}
	}
	old, err := repo.ListCompletedSessions(ctx, 7, 42, &practice, time.Now().Add(time.Hour), 3)
	if err != nil || len(old) != 0 {
		t.Fatalf("cutoff ignored: %+v %v", old, err)
	}
	// Insert extra completed session headers to verify deterministic recency and the requested history limit.
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("00000000-0000-4000-8000-%012d", i)
		_, err := conn.Exec(`INSERT INTO adaptive_sessions(id,user_id,course_id,topic_page_id,status,question_count,completed_at) VALUES($1,7,42,100,'completed',3,$2)`, id, time.Now().Add(time.Duration(i)*time.Minute))
		if err != nil {
			t.Fatal(err)
		}
	}
	recent, err := repo.ListCompletedSessions(ctx, 7, 42, &practice, time.Now().Add(-time.Hour), 3)
	if err != nil || len(recent) != 3 || recent[0].ID != "00000000-0000-4000-8000-000000000005" {
		t.Fatalf("recency/cap broken: %+v %v", recent, err)
	}
}

func TestPersistentSkillProfilePlansAndAtomicEvidence(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	ctx := context.Background()
	scope := int64(100)
	course := &fixedCourse{skills: []domain.LearningSkill{{Code: "math.linear.balance", Title: "Linear equations", PageIDs: []int64{10}}}}
	generator := &recordingGenerator{}
	service := application.NewAdaptiveTestService(repo, course, generator, repo)
	profile, err := service.GetLearningProfile(ctx, 7, 42, &scope)
	if err != nil || len(profile.Skills) != 1 || profile.Skills[0].Status != "insufficient_data" {
		t.Fatalf("initial profile=%+v err=%v", profile, err)
	}
	session, err := service.CreateSession(ctx, 7, 42, &scope, 3)
	if err != nil {
		t.Fatal(err)
	}
	originalRound := session.Rounds[0]
	answersFor := func(round domain.AdaptiveRound, key string) map[int64]string {
		result := map[int64]string{}
		for _, q := range round.Questions {
			result[q.ID] = key
		}
		return result
	}
	session, err = service.SubmitAnswers(ctx, 7, session.ID, originalRound.ID, answersFor(originalRound, "B"))
	if err != nil {
		t.Fatal(err)
	}
	if session.Rounds[1].LearningPlan[0].Reason != "remediate" || session.Rounds[1].Questions[0].Difficulty != 1 {
		t.Fatalf("not remedial: %+v", session.Rounds[1].LearningPlan)
	}
	session, err = service.SubmitAnswers(ctx, 7, session.ID, session.Rounds[1].ID, answersFor(session.Rounds[1], "A"))
	if err != nil {
		t.Fatal(err)
	}
	profile, err = service.GetLearningProfile(ctx, 7, 42, &scope)
	if err != nil || profile.Skills[0].State.NovelAnswers != 6 || profile.Skills[0].State.Sessions != 1 || profile.Skills[0].Status == "provisionally_mastered" {
		t.Fatalf("premature mastery: %+v %v", profile, err)
	}
	second, err := service.CreateSession(ctx, 7, 42, &scope, 3)
	if err != nil {
		t.Fatal(err)
	}
	first := second.Rounds[0]
	second, err = service.SubmitAnswers(ctx, 7, second.ID, first.ID, answersFor(first, "A"))
	if err != nil {
		t.Fatal(err)
	}
	if second.Rounds[1].LearningPlan[0].Reason != "challenge" || second.Rounds[1].Questions[0].Difficulty != 3 {
		t.Fatalf("not challenging: %+v", second.Rounds[1].LearningPlan)
	}
	profile, err = service.GetLearningProfile(ctx, 7, 42, &scope)
	state := profile.Skills[0].State
	if err != nil || state.NovelAnswers != 9 || state.CorrectAnswers != 6 || state.Sessions != 2 || profile.Skills[0].Status != "provisionally_mastered" {
		t.Fatalf("lost evidence: %+v %v", profile, err)
	}
	// Retry is idempotent at both the application boundary and the persistence boundary.
	calls := len(generator.requests)
	if _, err := service.SubmitAnswers(ctx, 7, second.ID, first.ID, answersFor(first, "A")); err != nil {
		t.Fatal(err)
	}
	var retried []domain.AdaptiveAnswer
	for _, q := range first.Questions {
		retried = append(retried, domain.AdaptiveAnswer{QuestionID: q.ID, SelectedOptionKey: "A", IsCorrect: true})
	}
	if _, err := repo.SubmitLearningRound(ctx, 7, second.ID, first.ID, retried); err != nil {
		t.Fatal(err)
	}
	if len(generator.requests) != calls {
		t.Fatal("retry regenerated a round")
	}
	// Replaying exactly the same problem in a different session creates an audit record,
	// but does not increase independent evidence or postpone review.
	duplicateID := "00000000-0000-4000-9000-000000000001"
	if err := repo.CreateSession(ctx, domain.AdaptiveSession{ID: duplicateID, UserID: 7, CourseID: 42, Status: domain.SessionInProgress, CurrentRound: 1, QuestionCount: 3, CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	copyRound := originalRound
	copyRound.SessionID = duplicateID
	copyRound.Status = domain.RoundGenerated
	copyRound.ID = 0
	saved, err := repo.SaveRound(ctx, copyRound)
	if err != nil {
		t.Fatal(err)
	}
	var duplicates []domain.AdaptiveAnswer
	for _, q := range saved.Questions {
		duplicates = append(duplicates, domain.AdaptiveAnswer{QuestionID: q.ID, SelectedOptionKey: "A", IsCorrect: true})
	}
	if _, err := repo.SubmitLearningRound(ctx, 7, duplicateID, saved.ID, duplicates); err != nil {
		t.Fatal(err)
	}
	profile, err = service.GetLearningProfile(ctx, 7, 42, &scope)
	if err != nil || profile.Skills[0].State.NovelAnswers != 9 || profile.Skills[0].State.Sessions != 2 {
		t.Fatalf("duplicate evidence counted: %+v %v", profile, err)
	}
	var count int
	if err := conn.Get(&count, `SELECT count(*) FROM learning_evidence WHERE session_id=$1 AND NOT contributes`, duplicateID); err != nil || count != 3 {
		t.Fatalf("duplicate audit count=%d err=%v", count, err)
	}
	other, err := service.GetLearningProfile(ctx, 8, 42, &scope)
	if err != nil || other.Skills[0].State.NovelAnswers != 0 {
		t.Fatal("another user's profile leaked", other, err)
	}
	// Failure on the second insert must roll back the first answer and its profile update.
	invalid := second.Rounds[1]
	invalidAnswers := []domain.AdaptiveAnswer{}
	for i, q := range invalid.Questions {
		key := "A"
		if i == 1 {
			key = "Z"
		}
		invalidAnswers = append(invalidAnswers, domain.AdaptiveAnswer{QuestionID: q.ID, SelectedOptionKey: key, IsCorrect: key == "A"})
	}
	if _, err := repo.SubmitLearningRound(ctx, 7, second.ID, invalid.ID, invalidAnswers); err == nil {
		t.Fatal("invalid selected option accepted")
	}
	if err := conn.Get(&count, `SELECT count(*) FROM adaptive_answers WHERE round_id=$1`, invalid.ID); err != nil || count != 0 {
		t.Fatalf("partial answers persisted: %d %v", count, err)
	}
	profile, err = service.GetLearningProfile(ctx, 7, 42, &scope)
	if err != nil || profile.Skills[0].State.NovelAnswers != 9 {
		t.Fatal("failed transaction changed profile", profile, err)
	}
	// Two simultaneous retries can commit only once and both receive a consistent result.
	type completion struct {
		applied bool
		err     error
	}
	results := make(chan completion, 2)
	var valid []domain.AdaptiveAnswer
	for _, q := range invalid.Questions {
		valid = append(valid, domain.AdaptiveAnswer{QuestionID: q.ID, SelectedOptionKey: "A", IsCorrect: true})
	}
	for i := 0; i < 2; i++ {
		go func() {
			applied, err := repo.SubmitLearningRound(ctx, 7, second.ID, invalid.ID, valid)
			results <- completion{applied, err}
		}()
	}
	commits := 0
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err != nil {
			t.Fatal(r.err)
		}
		if r.applied {
			commits++
		}
	}
	if commits != 1 {
		t.Fatalf("concurrent completions=%d", commits)
	}
	profile, err = service.GetLearningProfile(ctx, 7, 42, &scope)
	if err != nil || profile.Skills[0].State.NovelAnswers != 12 {
		t.Fatalf("concurrent evidence lost or duplicated: %+v %v", profile, err)
	}

	// Persisted plans allow inspecting why each question was assigned.
	restored, err := repo.GetSession(ctx, second.ID, 7)
	if err != nil || len(restored.Rounds[1].LearningPlan) != 3 || restored.Rounds[1].Questions[0].SkillCode != "math.linear.balance" {
		t.Fatalf("plan attribution lost: %+v %v", restored, err)
	}
}

func TestSkillEvidenceSurvivesNextRoundGenerationFailure(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	ctx := context.Background()
	scope := int64(100)
	course := &fixedCourse{skills: []domain.LearningSkill{{Code: "math.linear.balance", PageIDs: []int64{10}}}}
	generator := &recordingGenerator{failSecond: true}
	service := application.NewAdaptiveTestService(repo, course, generator, repo)
	session, err := service.CreateSession(ctx, 7, 42, &scope, 3)
	if err != nil {
		t.Fatal(err)
	}
	answers := map[int64]string{}
	for _, q := range session.Rounds[0].Questions {
		answers[q.ID] = "A"
	}
	if _, err := service.SubmitAnswers(ctx, 7, session.ID, session.Rounds[0].ID, answers); err == nil {
		t.Fatal("expected test generator failure")
	}
	profile, err := service.GetLearningProfile(ctx, 7, 42, &scope)
	if err != nil || profile.Skills[0].State.NovelAnswers != 3 || profile.Skills[0].State.CorrectAnswers != 3 {
		t.Fatalf("evidence lost after generator failure: %+v %v", profile, err)
	}
}
