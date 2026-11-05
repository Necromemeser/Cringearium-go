package domain

import (
	"math"
	"testing"
	"time"
)

func TestResearchEngagementUsesExplicitDenominatorsAndOutcomeWindow(t *testing.T) {
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	hash := "version"
	data := ResearchDataset{Study: ResearchStudy{ID: 1, Design: "observational", Scope: "course", ControlCourseID: 1, AdaptiveCourseID: 2, PreHash: hash, PostHash: hash}, AIAvailable: true}
	for i := 1; i <= 4; i++ {
		total, completed := 10, (3-i)*5
		if i == 4 {
			total, completed = 0, 0
		}
		data.Members = append(data.Members, ResearchMember{UserID: int64(i), Arm: "adaptive", TotalPages: total, CompletedPages: completed})
		if i != 4 {
			data.Assessments = append(data.Assessments, AssessmentResult{ID: int64(i * 10), UserID: int64(i), CourseID: 2, Scope: "course", Type: "pretest", Score: 50, FormHash: &hash, CompletedAt: start})
		}
		data.Assessments = append(data.Assessments, AssessmentResult{ID: int64(i*10 + 1), UserID: int64(i), CourseID: 2, Scope: "course", Type: "posttest", Score: 70 - i*10, FormHash: &hash, CompletedAt: start.Add(48 * time.Hour)})
	}
	data.Assessments = append(data.Assessments, AssessmentResult{ID: 12, UserID: 1, CourseID: 2, Scope: "course", Type: "pretest", Score: 100, FormHash: &hash, CompletedAt: start.Add(time.Hour)})
	for i, offset := range []time.Duration{-time.Hour, time.Hour, 72 * time.Hour} {
		answered := start.Add(offset)
		correct := i != 1
		data.AIEvents = append(data.AIEvents, AIResearchEvent{UserID: 1, CourseID: 2, SessionID: "s", AnsweredAt: &answered, IsCorrect: &correct})
	}
	report := BuildResearchReport(data, false, start.Add(100*time.Hour))
	e := report.Groups[1].Engagement
	if e.Improved != 1 || e.Unchanged != 1 || e.Declined != 1 || e.ImprovementRate == nil || math.Abs(*e.ImprovementRate-100.0/3) > 1e-9 {
		t.Fatal("wrong complete-pair denominator", e)
	}
	if e.Progress.N != 3 || *e.Progress.Mean != 50 || e.CalendarDays.N != 3 || *e.CalendarDays.Median != 2 {
		t.Fatal("missing progress or calendar treated as zero", e)
	}
	if e.PreAttempts.N != 4 || *e.PreAttempts.Mean != 1 || e.AIAnswers.N != 4 || *e.AIAnswers.Mean != 0.75 || e.AITotalAnswers != 3 || math.Abs(*e.AIAccuracy-200.0/3) > 1e-9 {
		t.Fatal("wrong observed counts", e)
	}
	if report.Participants[0].PreAttempts != 2 || *report.Participants[0].AIWindowAnswers != 1 || report.Participants[3].CalendarDays != nil || report.Participants[3].AIWindowAnswers != nil {
		t.Fatal("outcome-window filter or missing pairs lost", report.Participants)
	}
	data.AIAvailable = false
	report = BuildResearchReport(data, false, start)
	if report.Groups[1].Engagement.AIAnswers.N != 0 || report.Groups[1].Engagement.AIAccuracy != nil || report.Participants[0].AIWindowAnswers != nil {
		t.Fatal("unavailable source became zero")
	}
	empty := BuildResearchReport(ResearchDataset{}, false, start)
	if empty.Groups[0].Engagement.ImprovementRate != nil || empty.Groups[0].Engagement.Progress.Mean != nil {
		t.Fatal("empty study invented measurements")
	}
}
