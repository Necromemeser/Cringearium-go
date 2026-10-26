package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPrerequisitePlanIsAnExclusiveReviewRound(t *testing.T) {
	skills := []LearningSkill{{Code: "quadratic", PageIDs: []int64{10}}, {Code: "linear", PageIDs: []int64{20}, PrerequisiteOnly: true}}
	profile := BuildProfile(skills, []SkillState{{Code: "linear", Recent: []bool{false, false}, NovelAnswers: 2}}, 5, time.Now())
	if len(profile.NextPractice) != 5 {
		t.Fatal(profile)
	}
	for _, slot := range profile.NextPractice {
		if slot.SkillCode != "linear" || slot.Reason != "prerequisite" {
			t.Fatalf("mixed core question into review: %+v", slot)
		}
	}
	if RoundPurpose(AdaptiveRound{LearningPlan: profile.NextPractice}) != "prerequisite_review" {
		t.Fatal("review round not identified")
	}
	if RoundPurpose(AdaptiveRound{LearningPlan: []QuestionTarget{{Reason: "prerequisite"}, {Reason: "diagnose"}}}) != "mixed" {
		t.Fatal("legacy mixed round mislabelled")
	}
}
func TestLegacyFeedbackFieldsAreReadableAndMarshalForFrontend(t *testing.T) {
	for _, source := range []string{`{"Title":"Тема","Reason":"Верно","TopicPageID":10}`, `{"title":"Тема","reason":"Верно","topicPageID":10}`, `{"title":"Тема","reason":"Верно","topic_page_id":10}`} {
		var topic TopicFeedback
		if err := json.Unmarshal([]byte(source), &topic); err != nil {
			t.Fatal(err)
		}
		if topic.TopicPageID == nil || *topic.TopicPageID != 10 {
			t.Fatal(topic)
		}
		raw, err := json.Marshal(topic)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), `"title"`) || strings.Contains(string(raw), `"Title"`) {
			t.Fatalf("frontend receives wrong field names: %s", raw)
		}
	}
}
func TestEnglishAndIncompleteFeedbackFallsBackToRecordedRussianResult(t *testing.T) {
	session := AdaptiveSession{ID: "test", Rounds: []AdaptiveRound{{Status: RoundCompleted, Questions: []AdaptiveQuestion{{ID: 1, TopicTitle: "Квадратные уравнения"}, {ID: 2, TopicTitle: "Квадратные уравнения"}}, Answers: []AdaptiveAnswer{{QuestionID: 1, IsCorrect: true}, {QuestionID: 2, IsCorrect: false}, {QuestionID: 999, IsCorrect: true}}}}}
	for _, feedback := range []*AdaptiveFeedback{nil, {Summary: "Round 1 was strong"}, {Summary: "Хорошая работа", MasteredTopics: []TopicFeedback{{Title: "", Reason: ""}}}} {
		session.Feedback = feedback
		result := LocalizedFeedback(session)
		if !RussianText(result.Summary) || !strings.Contains(result.Summary, "1 из 2") || len(result.TopicsToReview) != 1 || len(result.NextSteps) != 1 {
			t.Fatalf("wrong fallback: %+v", result)
		}
		if strings.Contains(result.Summary, "Round") {
			t.Fatal(result)
		}
	}
}
