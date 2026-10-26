package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Necromemeser/Cringearium-go/internal/testsupport"
	coursehttp "github.com/Necromemeser/Cringearium-go/services/courses/internal/adapters/http"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func TestAIContextUsesPracticeSectionAndStoredAnswers(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	var courseID, sectionID, otherSectionID, theoryID, practiceID, testPageID, testID int64
	insert := func(target *int64, query string, args ...any) {
		t.Helper()
		if err := conn.Get(target, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	insert(&courseID, `INSERT INTO courses(title,status) VALUES('Context regression','published') RETURNING id`)
	insert(&sectionID, `INSERT INTO course_sections(course_id,title,position) VALUES($1,'Variables',0) RETURNING id`, courseID)
	insert(&otherSectionID, `INSERT INTO course_sections(course_id,title,position) VALUES($1,'Unrelated',1) RETURNING id`, courseID)
	insert(&theoryID, `INSERT INTO course_pages(section_id,title,type,content,position) VALUES($1,'Input conversion','theory','input returns a string',0) RETURNING id`, sectionID)
	insert(&practiceID, `INSERT INTO course_pages(section_id,title,type,position) VALUES($1,'AI input practice','ai_test',1) RETURNING id`, sectionID)
	insert(&testPageID, `INSERT INTO course_pages(section_id,title,type,position) VALUES($1,'Input test','test',2) RETURNING id`, sectionID)
	insert(&testID, `INSERT INTO tests(page_id,passing_score) VALUES($1,50) RETURNING id`, testPageID)
	if _, err := conn.Exec(`INSERT INTO course_pages(section_id,title,type,content,position) VALUES($1,'Hidden topic','theory','not relevant',0)`, otherSectionID); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`INSERT INTO course_access(user_id,course_id) VALUES(7,$1)`, courseID); err != nil {
		t.Fatal(err)
	}
	var answers []domain.TestAttemptAnswer
	for i := 0; i < 2; i++ {
		var questionID, correctID, incorrectID int64
		insert(&questionID, `INSERT INTO test_questions(test_id,question,position) VALUES($1,'What does input return?', $2) RETURNING id`, testID, i)
		insert(&correctID, `INSERT INTO test_answers(question_id,text,is_correct,position) VALUES($1,'string',true,0) RETURNING id`, questionID)
		insert(&incorrectID, `INSERT INTO test_answers(question_id,text,is_correct,position) VALUES($1,'integer',false,1) RETURNING id`, questionID)
		selected := correctID
		if i == 1 {
			selected = incorrectID
		}
		answers = append(answers, domain.TestAttemptAnswer{QuestionID: questionID, AnswerID: selected})
	}
	service := application.NewCourses(db, db)
	if _, err := service.SubmitTest(context.Background(), 7, testID, answers); err != nil {
		t.Fatal(err)
	}
	handler := coursehttp.NewHandler(service)
	req := httptest.NewRequest(http.MethodGet, "/internal/courses/"+strconv.FormatInt(courseID, 10)+"/ai-context?topic_page_id="+strconv.FormatInt(practiceID, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(courseID, 10))
	req.Header.Set("X-User-ID", "7")
	response := httptest.NewRecorder()
	handler.GetAIContext(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("HTTP %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		TopicTitle string `json:"topic_title"`
		Materials  []struct {
			PageID  int64  `json:"page_id"`
			Content string `json:"content"`
		} `json:"materials"`
		AllowedTopics []struct {
			PageID int64 `json:"page_id"`
		} `json:"allowed_topics"`
		TestResults []struct {
			PageID       int64  `json:"page_id"`
			Title        string `json:"title"`
			CorrectCount int    `json:"correct_count"`
			TotalCount   int    `json:"total_count"`
		} `json:"test_results"`
		OrdinaryTests []struct {
			Questions []struct {
				Answered       bool   `json:"answered"`
				IsCorrect      bool   `json:"is_correct"`
				SelectedAnswer string `json:"selected_answer"`
				CorrectAnswer  string `json:"correct_answer"`
			} `json:"questions"`
		} `json:"ordinary_tests"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.TopicTitle != "AI input practice" || len(payload.Materials) != 1 || payload.Materials[0].PageID != theoryID || payload.Materials[0].Content != "input returns a string" {
		t.Fatalf("wrong scope: %+v", payload)
	}
	if len(payload.AllowedTopics) != 1 || payload.AllowedTopics[0].PageID != theoryID {
		t.Fatalf("wrong allowed topics: %+v", payload.AllowedTopics)
	}
	if len(payload.TestResults) != 1 || payload.TestResults[0].CorrectCount != 1 || payload.TestResults[0].TotalCount != 2 || payload.TestResults[0].PageID != testPageID || payload.TestResults[0].Title != "Input test" {
		t.Fatalf("lost result attribution/correctness: %+v", payload.TestResults)
	}
	qs := payload.OrdinaryTests[0].Questions
	if len(qs) != 2 || !qs[0].Answered || !qs[0].IsCorrect || qs[1].IsCorrect || qs[1].SelectedAnswer != "integer" || qs[1].CorrectAnswer != "string" {
		t.Fatalf("lost answer evidence: %+v", qs)
	}
	for _, id := range []int64{999999, testPageID} {
		if _, err := service.GetAIContext(context.Background(), 7, courseID, &id); !errors.Is(err, application.ErrPageNotInCourse) {
			t.Fatalf("invalid page %d error=%v", id, err)
		}
	}
	if _, err := service.GetAIContext(context.Background(), 8, courseID, &practiceID); !errors.Is(err, application.ErrCourseNotFound) {
		t.Fatalf("unauthorized context error=%v", err)
	}
}

func TestMathSkillCatalogueAndPrerequisitesUseCourseLocalPages(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	var courseID, otherID, practiceID int64
	if err := conn.Get(&courseID, `SELECT id FROM courses WHERE title='Математика: курс с ИИ'`); err != nil {
		t.Fatal(err)
	}
	if err := conn.Get(&otherID, `SELECT id FROM courses WHERE title='Математика: обычные тесты'`); err != nil {
		t.Fatal(err)
	}
	if err := conn.Get(&practiceID, `SELECT p.id FROM course_pages p JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=$1 AND s.title='Квадратные уравнения' AND p.type='ai_test'`, courseID); err != nil {
		t.Fatal(err)
	}
	a, err := db.FindLearningSkills(context.Background(), courseID)
	if err != nil {
		t.Fatal(err)
	}
	b, err := db.FindLearningSkills(context.Background(), otherID)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 9 || len(b) != 9 {
		t.Fatalf("catalogue sizes %d/%d", len(a), len(b))
	}
	for i := range a {
		if a[i].Code != b[i].Code || a[i].PageIDs[0] == b[i].PageIDs[0] {
			t.Fatalf("unstable code or foreign page mapping: %+v %+v", a[i], b[i])
		}
	}
	service := application.NewCourses(db, db)
	if err := service.Enroll(context.Background(), 7, courseID); err != nil {
		t.Fatal(err)
	}
	result, err := service.GetAIContext(context.Background(), 7, courseID, &practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skills) != 6 || len(result.Materials) != 6 || len(result.AllowedTopics) != 6 {
		t.Fatalf("missing core/prerequisite scope: %+v", result)
	}
	deps := 0
	for _, s := range result.Skills {
		if s.PrerequisiteOnly {
			deps++
		}
		if s.Code == "math.quadratic.roots" && (len(s.Prerequisites) != 1 || s.Prerequisites[0] != "math.quadratic.discriminant") {
			t.Fatalf("missing root dependency: %+v", s)
		}
	}
	if deps != 2 {
		t.Fatalf("prerequisite count %d", deps)
	}
}
