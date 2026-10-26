package domain

import (
	"testing"
	"time"
)

func TestSkillStatusRequiresIndependentEvidenceAcrossSessions(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	state := SkillState{Code: "a"}
	for i := 0; i < 6; i++ {
		state = state.WithEvidence(true, 1, now)
	}
	if state.Status(now) != "practicing" {
		t.Fatal("one session was treated as independent confirmation")
	}
	state = state.WithEvidence(true, 2, now)
	if state.Status(now) != "provisionally_mastered" {
		t.Fatal(state)
	}
	if state.Status(now.Add(15*24*time.Hour)) != "review_due" {
		t.Fatal("old evidence was treated as fresh")
	}
	state = state.WithEvidence(false, 2, now)
	state = state.WithEvidence(false, 2, now)
	if state.Status(now) != "needs_practice" {
		t.Fatal(state)
	}
	if (SkillState{}).Status(now) != "insufficient_data" {
		t.Fatal("no evidence interpreted as failure")
	}
}
func TestPlannerDistinguishesGapsUnknownAndStrongSkills(t *testing.T) {
	now := time.Now()
	skills := []LearningSkill{{Code: "a", Title: "A", PageIDs: []int64{1}}, {Code: "b", Title: "B", PageIDs: []int64{2}}, {Code: "c", Title: "C", PageIDs: []int64{3}}}
	cases := []struct {
		name         string
		states       []SkillState
		code, reason string
		difficulty   int
	}{
		{"unknown", nil, "a", "diagnose", 2},
		{"gap", []SkillState{{Code: "b", NovelAnswers: 4, Recent: []bool{false, false}}}, "b", "remediate", 1},
		{"strong", []SkillState{{Code: "a", NovelAnswers: 6, Sessions: 2, Recent: []bool{true, true, true}, LastPracticedAt: &now}}, "b", "diagnose", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile := BuildProfile(skills, tc.states, 5, now)
			target := profile.NextPractice[0]
			if target.SkillCode != tc.code || target.Reason != tc.reason || target.Difficulty != tc.difficulty {
				t.Fatalf("plan=%+v", profile.NextPractice)
			}
			if len(profile.NextPractice) != 5 || profile.NextPractice[3].SkillCode == target.SkillCode {
				t.Fatal("no coverage beyond focus skill")
			}
		})
	}
}
func TestPlannerRevisitsKnownWeakPrerequisite(t *testing.T) {
	skills := []LearningSkill{{Code: "roots", PageIDs: []int64{2}, Prerequisites: []string{"discriminant"}}, {Code: "discriminant", PageIDs: []int64{1}, PrerequisiteOnly: true}}
	states := []SkillState{{Code: "discriminant", NovelAnswers: 2, Recent: []bool{false, false}}}
	plan := BuildProfile(skills, states, 3, time.Now()).NextPractice
	if plan[0].SkillCode != "discriminant" || plan[0].Reason != "prerequisite" {
		t.Fatal(plan)
	}
	plan = BuildProfile(skills, nil, 3, time.Now()).NextPractice
	if plan[0].SkillCode != "roots" || plan[0].Reason != "diagnose" {
		t.Fatal("unknown prerequisite blocks all progress", plan)
	}
}
func TestQuestionFingerprintIgnoresWhitespaceCaseAndOptionOrder(t *testing.T) {
	a := AdaptiveQuestion{Question: " Solve  2x=4 ", Options: []AdaptiveOption{{Key: "A", Text: "2"}, {Key: "B", Text: "3"}}}
	b := AdaptiveQuestion{Question: "solve 2x=4", Options: []AdaptiveOption{{Key: "A", Text: "3"}, {Key: "B", Text: "2"}}}
	if QuestionFingerprint(a) != QuestionFingerprint(b) {
		t.Fatal("same problem counted twice")
	}
	b.Question = "solve 3x=6"
	if QuestionFingerprint(a) == QuestionFingerprint(b) {
		t.Fatal("different problem not distinguishable")
	}
}
