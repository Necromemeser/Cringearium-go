package postgres

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/Necromemeser/Cringearium-go/internal/testsupport"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/ports"
)

type emptyResearchAI struct{}

func (emptyResearchAI) Events(context.Context, []int64) ([]domain.AIResearchEvent, error) {
	return []domain.AIResearchEvent{}, nil
}
func TestResearchProtocolSnapshotsAllocationAndVersionedOutcomes(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	service := application.NewResearch(db, emptyResearchAI{})
	var control, adaptive int64
	if err := conn.Get(&control, `SELECT id FROM courses WHERE title='Математика: обычные тесты'`); err != nil {
		t.Fatal(err)
	}
	if err := conn.Get(&adaptive, `SELECT id FROM courses WHERE title='Математика: курс с ИИ'`); err != nil {
		t.Fatal(err)
	}
	protocol := domain.ResearchStudy{Title: "Исследование", Design: "observational", Scope: "course", ControlCourseID: control, AdaptiveCourseID: adaptive}
	study, err := service.Create(ctx, protocol)
	if err != nil {
		t.Fatal(err)
	}
	if len(study.PreHash) != 64 || len(study.ExportSalt) != 64 {
		t.Fatal("protocol not fingerprinted")
	}
	if err := service.Assign(ctx, study.ID, []int64{101}, "control"); err != nil {
		t.Fatal(err)
	}
	if err := service.Assign(ctx, study.ID, []int64{102}, "adaptive"); err != nil {
		t.Fatal(err)
	}
	native := application.NewCourses(db, db)
	take := func(user, course int64, kind string, correct int) {
		t.Helper()
		instruments, err := db.ResearchInstruments(ctx, course, "course")
		if err != nil {
			t.Fatal(err)
		}
		for _, instrument := range instruments {
			if instrument.Kind != kind {
				continue
			}
			var answers []domain.TestAttemptAnswer
			for i, q := range instrument.Test.Questions {
				var selected int64
				for _, option := range q.Answers {
					if option.IsCorrect == (i < correct) {
						selected = option.ID
						break
					}
				}
				answers = append(answers, domain.TestAttemptAnswer{QuestionID: q.ID, AnswerID: selected})
			}
			if _, err := native.SubmitTest(ctx, user, instrument.Test.ID, answers); err != nil {
				t.Fatal(err)
			}
			return
		}
		t.Fatal("missing test")
	}
	take(101, control, "pretest", 0)
	take(101, control, "posttest", 12)
	take(102, adaptive, "pretest", 24)
	take(102, adaptive, "posttest", 24)
	report, err := service.Report(ctx, study.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Groups[0].Paired != 1 || report.Groups[1].Paired != 1 || report.Participants[0].PreScore == nil || *report.Participants[0].PreScore != 0 || report.Participants[0].Gain == nil || *report.Participants[0].Gain != 50 {
		t.Fatalf("wrong measured pairs: %+v", report)
	}
	if !report.AIAvailable || report.Participants[0].AIAnswers == nil || *report.Participants[0].AIAnswers != 0 {
		t.Fatal("empty source confused with failure")
	}
	dataset, err := service.Dataset(ctx, study.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(dataset.Assessments) != 4 {
		t.Fatal(len(dataset.Assessments))
	}
	for _, a := range dataset.Assessments {
		if a.FormHash == nil || len(a.InstrumentSnapshot) == 0 {
			t.Fatal("submission did not capture instrument", a)
		}
	}
	// Metadata, rather than a mutable page title, determines assessment purpose.
	if _, err := conn.Exec(`UPDATE course_pages p SET title='Новое название диагностики' FROM tests t JOIN assessment_definitions d ON d.test_id=t.id WHERE p.id=t.page_id AND d.kind='pretest' AND p.section_id IN(SELECT id FROM course_sections WHERE course_id=$1)`, control); err != nil {
		t.Fatal(err)
	}
	if err := service.Assign(ctx, study.ID, []int64{103}, "control"); err != nil {
		t.Fatal(err)
	}
	take(103, control, "pretest", 12)
	report, err = service.Report(ctx, study.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Groups[0].WithPre != 2 {
		t.Fatal("title rename lost diagnostic purpose")
	}
	// A prospective trial rejects prior access, balances fresh participants, and never reassigns.
	protocol.Design = "randomized"
	trial, err := service.Create(ctx, protocol)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Assign(ctx, trial.ID, []int64{101}, ""); !errors.Is(err, ports.ErrResearchInvalid) {
		t.Fatalf("experienced participant randomized: %v", err)
	}
	if err := service.Assign(ctx, trial.ID, []int64{201, 202, 203, 204, 205, 206, 207, 208}, ""); err != nil {
		t.Fatal(err)
	}
	assigned, err := service.Report(ctx, trial.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.Groups[0].Assigned != 4 || assigned.Groups[1].Assigned != 4 || assigned.Groups[0].Paired != 0 {
		t.Fatalf("bad allocation: %+v", assigned.Groups)
	}
	if err := service.Assign(ctx, trial.ID, []int64{201, 209}, ""); !errors.Is(err, ports.ErrResearchInvalid) {
		t.Fatal("assignment mutable")
	}
	var count int
	if err := conn.Get(&count, `SELECT count(*) FROM research_participants WHERE study_id=$1 AND user_id=209`, trial.ID); err != nil || count != 0 {
		t.Fatal("partial failed batch persisted")
	}
	// A changed outcome form cannot silently enter a new paired experiment.
	if _, err := conn.Exec(`UPDATE test_questions q SET question=q.question||' изменено' FROM tests t JOIN course_pages p ON p.id=t.page_id JOIN course_sections s ON s.id=p.section_id WHERE q.test_id=t.id AND s.course_id=$1`, adaptive); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, protocol); !errors.Is(err, ports.ErrResearchInvalid) {
		t.Fatalf("incompatible forms accepted: %v", err)
	}
}
