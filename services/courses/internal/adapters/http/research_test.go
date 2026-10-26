package http

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/ports"
)

type researchDatasetStub struct {
	ports.ResearchUseCases
	data domain.ResearchDataset
}

func (s researchDatasetStub) Dataset(context.Context, int64) (domain.ResearchDataset, error) {
	return s.data, nil
}
func TestResearchCSVIsMachineReadableScopedAndFormulaSafe(t *testing.T) {
	now := time.Now().UTC()
	hash := "hash"
	test := &domain.Test{ID: 9, Questions: []domain.TestQuestion{{ID: 31, Question: `=HYPERLINK("https://example.test")`, Answers: []domain.TestAnswer{{ID: 41, Text: "=1+1", IsCorrect: true}}}}}
	_, snapshot := domain.InstrumentSnapshot(test)
	answers, _ := json.Marshal([]domain.TestAttemptAnswer{{QuestionID: 31, AnswerID: 41, CorrectAnswerID: 41, IsCorrect: true}})
	data := domain.ResearchDataset{Study: domain.ResearchStudy{ID: 1, Title: "=SUM(1,2)", Design: "observational", Scope: "course", ControlCourseID: 1, AdaptiveCourseID: 2, PreHash: hash, PostHash: hash, ExportSalt: "private-export-salt"}, Members: []domain.ResearchMember{{UserID: 90909090, Arm: "control", AssignedAt: now}}, Assessments: []domain.AssessmentResult{{ID: 11, UserID: 90909090, CourseID: 1, Type: "pretest", Scope: "course", Score: 0, FormHash: &hash, InstrumentSnapshot: snapshot, Answers: answers, CompletedAt: now}}}
	handler := NewResearchHandler(researchDatasetStub{data: data})
	mux := http.NewServeMux()
	handler.Register(mux)
	download := func(dataset string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/internal/admin/research/studies/1/export?dataset="+dataset, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		return response
	}
	parse := func(response *httptest.ResponseRecorder) [][]string {
		t.Helper()
		if response.Code != 200 {
			t.Fatal(response.Code, response.Body.String())
		}
		if !strings.Contains(response.Header().Get("Content-Type"), "text/csv") {
			t.Fatal(response.Header())
		}
		rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(response.Body.String(), "\ufeff"))).ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	response := download("answers")
	rows := parse(response)
	columns := map[string]int{}
	for i, h := range rows[0] {
		columns[h] = i
	}
	if len(rows) != 2 || rows[1][columns["question_text"]] != `'=HYPERLINK("https://example.test")` || rows[1][columns["selected_answer_text"]] != "'=1+1" || rows[1][columns["question_id"]] != "31" || rows[1][columns["is_correct"]] != "true" {
		t.Fatalf("answer export loses data or protection: %+v", rows)
	}
	if strings.Contains(response.Body.String(), "90909090") || strings.Contains(response.Body.String(), "private-export-salt") {
		t.Fatal("identity/secret leaked in default research export")
	}
	protocol := parse(download("protocol"))
	if protocol[1][1] != "'=SUM(1,2)" {
		t.Fatal(protocol)
	}
	participants := parse(download("participants"))
	if participants[1][6] != "" || participants[1][8] != "" || participants[1][9] != "" {
		t.Fatal("missing result became zero", participants)
	}
	key := parse(download("mapping"))
	if key[1][2] != "90909090" || key[1][1] != rows[1][columns["participant_id"]] {
		t.Fatal("explicit mapping is inconsistent", key)
	}
	if download("ai_events").Code != 503 {
		t.Fatal("failed AI source exported as empty data")
	}
	if download("unknown").Code != 400 {
		t.Fatal("unknown dataset accepted")
	}
	if safeCSV("-20") != "-20" || safeCSV(" @SUM(A1:A2)") != "' @SUM(A1:A2)" {
		t.Fatal("numeric or formula handling")
	}
}
