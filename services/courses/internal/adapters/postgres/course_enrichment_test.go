package postgres

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"testing"

	"github.com/Necromemeser/Cringearium-go/internal/testsupport"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func TestEnrichedCoursePairsHaveComparableInstrumentsAndScopedSkills(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	pairs := [][3]string{
		{"python", "Python: основы программирования", "Python с ИИ: основы программирования"},
		{"russian", "Русский язык: подготовка к ЕГЭ", "Русский язык с ИИ: подготовка к ЕГЭ"},
		{"psychology", "Основы психологии", "Основы психологии с ИИ"},
		{"philosophy", "Философия без занудства", "Философия без занудства с ИИ"},
		{"memology", "Мемология: история и теория мемов", "Мемология с ИИ: история и теория мемов"},
	}
	ctx := context.Background()
	research := application.NewResearch(db, emptyResearchAI{})
	service := application.NewCourses(db, db)
	for _, pair := range pairs {
		t.Run(pair[0], func(t *testing.T) {
			var control, adaptive int64
			if err := conn.Get(&control, `SELECT id FROM courses WHERE title=$1`, pair[1]); err != nil {
				t.Fatal(err)
			}
			if err := conn.Get(&adaptive, `SELECT id FROM courses WHERE title=$1`, pair[2]); err != nil {
				t.Fatal(err)
			}
			a, err := db.FindLearningSkills(ctx, control)
			if err != nil {
				t.Fatal(err)
			}
			b, err := db.FindLearningSkills(ctx, adaptive)
			if err != nil {
				t.Fatal(err)
			}
			if len(a) != 6 || len(b) != 6 {
				t.Fatalf("skills missing: %d/%d", len(a), len(b))
			}
			for i := range a {
				if a[i].Code != b[i].Code || !strings.HasPrefix(a[i].Code, pair[0]+".") || len(a[i].PageIDs) != 1 || len(b[i].PageIDs) != 1 || a[i].PageIDs[0] == b[i].PageIDs[0] {
					t.Fatal("unstable or cross-course mapping", a[i], b[i])
				}
			}
			// Creating a protocol exercises identical theory and independent assessment versions.
			if _, err := research.Create(ctx, domain.ResearchStudy{Title: "Pilot " + pair[0], Design: "randomized", Scope: "course", ControlCourseID: control, AdaptiveCourseID: adaptive}); err != nil {
				t.Fatal(err)
			}
			var counts struct {
				Pages     int
				Questions int
				Correct   int
			}
			if err := conn.Get(&counts, `SELECT count(DISTINCT p.id) AS pages,count(DISTINCT q.id) AS questions,count(*) FILTER(WHERE a.is_correct) AS correct FROM course_upgrade_v10_added_pages u JOIN course_pages p ON p.id=u.page_id JOIN course_sections s ON s.id=p.section_id JOIN tests t ON t.page_id=p.id JOIN test_questions q ON q.test_id=t.id JOIN test_answers a ON a.question_id=q.id WHERE s.course_id=$1`, control); err != nil {
				t.Fatal(err)
			}
			if counts.Pages != 3 || counts.Questions != 18 || counts.Correct != 18 {
				t.Fatal("fixed practice incomplete", counts)
			}
			if err := service.Enroll(ctx, 888, control); err != nil {
				t.Fatal(err)
			}
			var practicePage int64
			if err := conn.Get(&practicePage, `SELECT min(p.id) FROM course_upgrade_v10_added_pages u JOIN course_pages p ON p.id=u.page_id JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=$1`, control); err != nil {
				t.Fatal(err)
			}
			fixed, err := db.FindTestByPageID(ctx, 888, practicePage)
			if err != nil || fixed == nil {
				t.Fatal("fixed practice inaccessible", err)
			}
			for _, correctCount := range []int{3, 4} {
				var selected []domain.TestAttemptAnswer
				for i, q := range fixed.Questions {
					for _, a := range q.Answers {
						if a.IsCorrect == (i < correctCount) {
							selected = append(selected, domain.TestAttemptAnswer{QuestionID: q.ID, AnswerID: a.ID})
							break
						}
					}
				}
				attempt, err := service.SubmitTest(ctx, 888, fixed.ID, selected)
				if err != nil || attempt.Passed != (correctCount == 4) || attempt.Score != correctCount*100/6 {
					t.Fatal("practice grading/threshold changed", attempt, err)
				}
				var completed, assessments int
				if err := conn.Get(&completed, `SELECT count(*) FROM page_progress WHERE user_id=888 AND page_id=$1`, practicePage); err != nil {
					t.Fatal(err)
				}
				if err := conn.Get(&assessments, `SELECT count(*) FROM assessment_results WHERE test_attempt_id=$1`, attempt.AttemptID); err != nil {
					t.Fatal(err)
				}
				if completed != correctCount-3 || assessments != 0 {
					t.Fatal("practice mixed with objective assessments", completed, assessments)
				}
			}
			if err := service.Enroll(ctx, 777, adaptive); err != nil {
				t.Fatal(err)
			}
			var practiceIDs []int64
			if err := conn.Select(&practiceIDs, `SELECT p.id FROM course_pages p JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=$1 AND p.type='ai_test' ORDER BY s.position,p.position`, adaptive); err != nil {
				t.Fatal(err)
			}
			if len(practiceIDs) != 3 {
				t.Fatal("missing AI module", practiceIDs)
			}
			for _, id := range practiceIDs {
				result, err := service.GetAIContext(ctx, 777, adaptive, &id)
				if err != nil {
					t.Fatal(err)
				}
				core := 0
				allowed := map[int64]bool{}
				for _, skill := range result.Skills {
					if !skill.PrerequisiteOnly {
						core++
					}
					if !strings.HasPrefix(skill.Code, pair[0]+".") {
						t.Fatal("foreign skill", skill)
					}
					for _, p := range skill.PageIDs {
						allowed[p] = true
					}
				}
				if core != 2 || len(result.Materials) != len(allowed) {
					t.Fatalf("scope/prerequisites invalid: %+v", result)
				}
				for _, m := range result.Materials {
					if !allowed[m.PageID] || !strings.Contains(m.Content, "Самопроверка") {
						t.Fatal("missing lesson or foreign material", m.PageID)
					}
				}
			}
			var final bool
			if err := conn.Get(&final, `SELECT EXISTS(SELECT 1 FROM course_pages p JOIN course_sections s ON s.id=p.section_id JOIN tests t ON t.page_id=p.id JOIN assessment_definitions d ON d.test_id=t.id WHERE s.course_id=$1 AND d.kind='posttest' AND p.id=(SELECT p2.id FROM course_pages p2 JOIN course_sections s2 ON s2.id=p2.section_id WHERE s2.course_id=$1 ORDER BY s2.position DESC,p2.position DESC LIMIT 1))`, control); err != nil || !final {
				t.Fatal("fixed practice displaced the final assessment", err)
			}
		})
	}
}

func TestCourseEnrichmentUpgradeAndRollbackPreserveExistingAttempts(t *testing.T) {
	conn := testsupport.Postgres(t)
	files, err := fs.Glob(migrationFiles, "migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	for _, name := range files {
		if name >= "migrations/000010_" {
			continue
		}
		raw, err := migrationFiles.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Exec(string(raw)); err != nil {
			t.Fatal(name, err)
		}
	}
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	ctx := context.Background()
	service := application.NewCourses(db, db)
	var courseID int64
	if err := conn.Get(&courseID, `SELECT id FROM courses WHERE title='Основы психологии'`); err != nil {
		t.Fatal(err)
	}
	if err := service.Enroll(ctx, 123, courseID); err != nil {
		t.Fatal(err)
	}
	instruments, err := db.ResearchInstruments(ctx, courseID, "course")
	if err != nil {
		t.Fatal(err)
	}
	var test domain.Test
	for _, v := range instruments {
		if v.Kind == "pretest" {
			test = v.Test
		}
	}
	oldHash, _ := domain.InstrumentSnapshot(&test)
	var answers []domain.TestAttemptAnswer
	for _, q := range test.Questions {
		for _, a := range q.Answers {
			if !a.IsCorrect {
				answers = append(answers, domain.TestAttemptAnswer{QuestionID: q.ID, AnswerID: a.ID})
				break
			}
		}
	}
	result, err := service.SubmitTest(ctx, 123, test.ID, answers)
	if err != nil {
		t.Fatal(err)
	}
	var theoryID int64
	if err := conn.Get(&theoryID, `SELECT min(p.id) FROM course_pages p JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=$1 AND p.type='theory'`, courseID); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`INSERT INTO page_progress(user_id,page_id) VALUES(123,$1)`, theoryID); err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		var attempts, progress int
		var hash string
		if err := conn.Get(&attempts, `SELECT count(*) FROM assessment_results WHERE test_attempt_id=$1 AND user_id=123 AND score=0 AND form_hash=$2`, result.AttemptID, oldHash); err != nil {
			t.Fatal(err)
		}
		if err := conn.Get(&progress, `SELECT count(*) FROM page_progress WHERE user_id=123 AND page_id=$1`, theoryID); err != nil {
			t.Fatal(err)
		}
		current, err := db.ResearchInstruments(ctx, courseID, "course")
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range current {
			if v.Kind == "pretest" {
				hash, _ = domain.InstrumentSnapshot(&v.Test)
			}
		}
		if attempts != 1 || progress != 1 || hash != oldHash {
			t.Fatal("history or objective test changed", attempts, progress, hash)
		}
	}
	for _, direction := range []string{"up", "down", "up"} {
		raw, err := migrationFiles.ReadFile("migrations/000010_course_enrichment." + direction + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Exec(string(raw)); err != nil {
			t.Fatal(direction, err)
		}
		check()
	}
}

func TestMemologyLessonsRemoveOutboundFootnotesButKeepImages(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	var pages []struct {
		Content  string
		Original string
	}
	if err := conn.Select(&pages, `SELECT p.content,b.content AS original FROM memology_v11_original_pages b JOIN course_pages p ON p.id=b.page_id`); err != nil {
		t.Fatal(err)
	}
	if len(pages) != 12 {
		t.Fatal("missing paired lessons", len(pages))
	}
	images, changed := 0, 0
	for _, p := range pages {
		if strings.Contains(p.Content, "](http") {
			t.Fatal("outbound learner link remains", p.Content)
		}
		if !strings.Contains(p.Content, "Самопроверка") {
			t.Fatal("lesson lost learning material")
		}
		if p.Content != p.Original {
			changed++
		}
		before, after := strings.Count(p.Original, "](/images/memes/"), strings.Count(p.Content, "](/images/memes/")
		if before != after {
			t.Fatal("original illustration lost")
		}
		images += after
	}
	if images != 10 || changed != 10 {
		t.Fatal("unexpected cleanup scope", images, changed)
	}
	down, err := migrationFiles.ReadFile("migrations/000011_memology_safe_lessons.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(string(down)); err != nil {
		t.Fatal(err)
	}
	var restored int
	if err := conn.Get(&restored, `SELECT count(*) FROM course_pages p JOIN course_sections s ON s.id=p.section_id JOIN courses c ON c.id=s.course_id WHERE c.title IN('Мемология: история и теория мемов','Мемология с ИИ: история и теория мемов') AND p.content LIKE '%](http%'`); err != nil || restored != 10 {
		t.Fatal("rollback lost original paragraphs", restored, err)
	}
}
