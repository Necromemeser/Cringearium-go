package domain

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func near(t *testing.T, got *float64, want, tolerance float64) {
	t.Helper()
	if got == nil || math.Abs(*got-want) > tolerance {
		t.Fatalf("got %v want %.9f", got, want)
	}
}
func TestResearchStatisticsUseSampleVariationAndWelchDifference(t *testing.T) {
	var rows []ResearchParticipant
	for _, arm := range []string{"control", "adaptive"} {
		for i := 0; i < 3; i++ {
			pre, post := float64(20+i*10), float64(20+i*20)
			if arm == "adaptive" {
				post += 20
			}
			gain := post - pre
			rows = append(rows, ResearchParticipant{Arm: arm, PreScore: &pre, PostScore: &post, Gain: &gain, Eligible: true})
		}
	}
	groups := []ResearchGroupMetrics{groupMetrics("control", rows), groupMetrics("adaptive", rows)}
	near(t, groups[0].Gain.Mean, 10, 1e-9)
	near(t, groups[0].Gain.SD, 10, 1e-9)
	near(t, groups[0].Gain.Median, 10, 1e-9)
	comparison := compareGroups(groups)
	near(t, comparison.GainDifference, 20, 1e-9)
	near(t, comparison.HedgesG, 1.6, 1e-9)
	near(t, comparison.CILow, 20-2.776*math.Sqrt(200.0/3), 1e-6)
	if Summarize(nil).Mean != nil || Summarize([]float64{0}).SD != nil {
		t.Fatal("missing uncertainty replaced with zero")
	}
	near(t, Summarize([]float64{0}).Mean, 0, 0)
}
func TestANCOVARecoversKnownCoefficientsAndHandlesSingularInputs(t *testing.T) {
	var rows []ResearchParticipant
	for g := 0; g < 2; g++ {
		for i := 0; i < 8; i++ {
			pre := float64(10 + i*8)
			post := 15 + .6*pre + 12*float64(g)
			gain := post - pre
			arm := "control"
			if g == 1 {
				arm = "adaptive"
			}
			rows = append(rows, ResearchParticipant{Arm: arm, PreScore: &pre, PostScore: &post, Gain: &gain, Eligible: true})
		}
	}
	result := FitANCOVA(rows)
	if !result.Available {
		t.Fatal(result)
	}
	near(t, result.Intercept, 15, 1e-8)
	near(t, result.PreCoefficient, .6, 1e-8)
	near(t, result.GroupCoefficient, 12, 1e-8)
	near(t, result.R2, 1, 1e-8)
	if FitANCOVA(rows[:6]).Available {
		t.Fatal("model fit on an insufficient sample")
	}
	for i := range rows {
		constant := 50.0
		rows[i].PreScore = &constant
	}
	if result := FitANCOVA(rows); result.Available || result.Reason != "singular_design" {
		t.Fatalf("singularity ignored: %+v", result)
	}
}
func TestResearchPairingVersionsMissingnessAndAssignedDenominators(t *testing.T) {
	now := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	hashPre, hashPost := "pre", "post"
	study := ResearchStudy{ID: 1, Design: "observational", Scope: "course", ControlCourseID: 1, AdaptiveCourseID: 2, PreHash: hashPre, PostHash: hashPost, ExportSalt: "test-only-salt"}
	data := ResearchDataset{Study: study, Members: []ResearchMember{{UserID: 1, Arm: "control"}, {UserID: 2, Arm: "adaptive"}, {UserID: 3, Arm: "adaptive"}, {UserID: 4, Arm: "control"}}, AIAvailable: false}
	add := func(id, user, course int64, kind string, score int, at time.Time, hash *string) {
		data.Assessments = append(data.Assessments, AssessmentResult{ID: id, UserID: user, CourseID: course, Type: kind, Scope: "course", Score: score, CompletedAt: at, FormHash: hash})
	}
	add(1, 1, 1, "pretest", 0, now, &hashPre)
	add(2, 1, 1, "pretest", 80, now.Add(time.Minute), &hashPre)
	add(3, 1, 1, "posttest", 40, now.Add(time.Hour), &hashPost)
	add(4, 1, 1, "posttest", 95, now.Add(2*time.Hour), &hashPost)
	add(5, 2, 2, "pretest", 20, now, nil)
	add(6, 2, 2, "posttest", 50, now.Add(time.Hour), nil)
	add(7, 3, 2, "pretest", 30, now, &hashPre)
	bad := "changed"
	add(8, 4, 1, "pretest", 10, now, &bad)
	add(9, 4, 1, "posttest", 60, now.Add(time.Hour), &hashPost)
	strict := BuildResearchReport(data, false, now.Add(3*time.Hour))
	if strict.Groups[0].Assigned != 2 || strict.Groups[1].Assigned != 2 || strict.Groups[0].Paired != 1 || strict.Groups[1].Paired != 0 {
		t.Fatal(strict.Groups)
	}
	row := strict.Participants[0]
	near(t, row.PreScore, 0, 0)
	near(t, row.PostScore, 40, 0)
	if row.PreID == nil || *row.PreID != 1 || row.PostID == nil || *row.PostID != 3 {
		t.Fatal("best/latest attempt selected", row)
	}
	if row.AISessions != nil || strict.Participants[2].PostScore != nil || strict.Participants[2].Gain != nil {
		t.Fatal("unavailable or missing data replaced with zero")
	}
	exploratory := BuildResearchReport(data, true, now)
	if exploratory.Groups[1].Paired != 1 || exploratory.Participants[3].Eligible {
		t.Fatal("unversioned option admits known incompatible forms", exploratory.Participants)
	}
	if strict.Participants[0].ParticipantID != exploratory.Participants[0].ParticipantID || strict.Participants[0].ParticipantID == ParticipantKey(ResearchStudy{ID: 2, ExportSalt: "different"}, 1) {
		t.Fatal("pseudonym not stable/scoped")
	}
	raw, _ := json.Marshal(study)
	if string(raw) == "" || jsonContains(raw, "export_salt") {
		t.Fatal("export salt leaked")
	}
}
func jsonContains(raw []byte, key string) bool {
	var fields map[string]any
	_ = json.Unmarshal(raw, &fields)
	_, ok := fields[key]
	return ok
}
func TestRandomizedSelectionExcludesPreAssignmentAndReversedTests(t *testing.T) {
	now := time.Now()
	hash := "same"
	data := ResearchDataset{Study: ResearchStudy{ID: 1, Design: "randomized", Scope: "course", ControlCourseID: 1, AdaptiveCourseID: 2, PreHash: hash, PostHash: hash, ExportSalt: "salt"}, Members: []ResearchMember{{UserID: 1, Arm: "adaptive", AssignedAt: now}}}
	for i, kind := range []string{"pretest", "posttest", "pretest"} {
		data.Assessments = append(data.Assessments, AssessmentResult{ID: int64(i + 1), UserID: 1, CourseID: 2, Type: kind, Scope: "course", Score: 50, FormHash: &hash, CompletedAt: now.Add(time.Duration(i-1) * time.Hour)})
	}
	report := BuildResearchReport(data, false, now.Add(3*time.Hour))
	if report.Participants[0].Eligible || report.QualityCounts["post_before_pretest"] != 1 {
		t.Fatal(report.Participants)
	}
}
func TestAllocationBalancesWithoutLookingAtOutcomes(t *testing.T) {
	if AllocateResearchArm(2, 1, false) != "adaptive" || AllocateResearchArm(1, 2, true) != "control" || AllocateResearchArm(1, 1, true) != "adaptive" || AllocateResearchArm(1, 1, false) != "control" {
		t.Fatal("allocation policy")
	}
}
func TestEquivalentInstrumentVersionsIgnoreIDsButDetectChangedAnswers(t *testing.T) {
	a := Test{ID: 1, Questions: []TestQuestion{{ID: 1, Question: "2x = 4", Answers: []TestAnswer{{ID: 1, Text: "2", IsCorrect: true}, {ID: 2, Text: "3"}}}}}
	b := Test{ID: 2, Questions: []TestQuestion{{ID: 99, Question: " 2x = 4 ", Answers: []TestAnswer{{ID: 9, Text: "3"}, {ID: 8, Text: "2", IsCorrect: true}}}}}
	ha, _ := InstrumentSnapshot(&a)
	hb, _ := InstrumentSnapshot(&b)
	if ha != hb {
		t.Fatal("IDs/order changed equivalent form hash")
	}
	b.Questions[0].Answers[0].Text = "4"
	hc, _ := InstrumentSnapshot(&b)
	if hc == ha {
		t.Fatal("changed form not detected")
	}
}

func TestANCOVAHC3MatchesIndependentNumpyOracle(t *testing.T) {
	noise := []float64{-2, 1, 3, -1, 2, -3, 1, -1, 1, -2, 2, 0, 3, -1, -2, 2}
	var rows []ResearchParticipant
	for i, n := range noise {
		pre := float64(10 + (i%8)*8)
		group := i / 8
		post := 15 + .6*pre + 12*float64(group) + n
		gain := post - pre
		arm := "control"
		if group == 1 {
			arm = "adaptive"
		}
		rows = append(rows, ResearchParticipant{Arm: arm, PreScore: &pre, PostScore: &post, Gain: &gain, Eligible: true})
	}
	got := FitANCOVA(rows)
	if !got.Available {
		t.Fatal(got)
	}
	near(t, got.Intercept, 15.197916666666698, 1e-8)
	near(t, got.PreCoefficient, .5947916666666664, 1e-8)
	near(t, got.GroupCoefficient, 12.375, 1e-8)
	near(t, got.SE, 1.1393438188800538, 1e-8)
}
