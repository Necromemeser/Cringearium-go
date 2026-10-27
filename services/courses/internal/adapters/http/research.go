package http

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/ports"
)

type ResearchHandler struct{ service ports.ResearchUseCases }

func NewResearchHandler(service ports.ResearchUseCases) *ResearchHandler {
	return &ResearchHandler{service}
}
func (h *ResearchHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /internal/admin/research/studies", h.list)
	mux.HandleFunc("POST /internal/admin/research/studies", h.create)
	mux.HandleFunc("POST /internal/admin/research/studies/{studyId}/participants", h.assign)
	mux.HandleFunc("GET /internal/admin/research/studies/{studyId}/report", h.report)
	mux.HandleFunc("GET /internal/admin/research/studies/{studyId}/export", h.export)
}
func researchError(w http.ResponseWriter, err error) {
	if errors.Is(err, ports.ErrResearchNotFound) {
		http.Error(w, "Исследование не найдено", 404)
		return
	}
	if errors.Is(err, ports.ErrResearchInvalid) {
		text := strings.TrimPrefix(err.Error(), ports.ErrResearchInvalid.Error()+": ")
		if text == err.Error() {
			text = "Проверь параметры протокола и список участников"
		}
		http.Error(w, text, 400)
		return
	}
	http.Error(w, "Не удалось обработать данные исследования", 500)
}
func studyID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("studyId"), 10, 64)
	if err != nil || id <= 0 {
		return 0, ports.ErrResearchInvalid
	}
	return id, nil
}
func decodeResearch(r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func (h *ResearchHandler) list(w http.ResponseWriter, r *http.Request) {
	studies, err := h.service.List(r.Context())
	if err != nil {
		researchError(w, err)
		return
	}
	writeJSON(w, 200, studies)
}
func (h *ResearchHandler) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string `json:"title"`
		Design   string `json:"design"`
		Scope    string `json:"scope"`
		Control  int64  `json:"control_course_id"`
		Adaptive int64  `json:"adaptive_course_id"`
		Notes    string `json:"protocol_notes"`
	}
	if err := decodeResearch(r, &input); err != nil {
		researchError(w, ports.ErrResearchInvalid)
		return
	}
	study, err := h.service.Create(r.Context(), domain.ResearchStudy{Title: input.Title, Design: input.Design, Scope: input.Scope, ControlCourseID: input.Control, AdaptiveCourseID: input.Adaptive, ProtocolNotes: input.Notes})
	if err != nil {
		researchError(w, err)
		return
	}
	writeJSON(w, 201, study)
}
func (h *ResearchHandler) assign(w http.ResponseWriter, r *http.Request) {
	id, err := studyID(r)
	if err != nil {
		researchError(w, err)
		return
	}
	var input struct {
		UserIDs []int64 `json:"user_ids"`
		Arm     string  `json:"arm"`
	}
	if err := decodeResearch(r, &input); err != nil {
		researchError(w, ports.ErrResearchInvalid)
		return
	}
	if err := h.service.Assign(r.Context(), id, input.UserIDs, input.Arm); err != nil {
		researchError(w, err)
		return
	}
	writeJSON(w, 200, map[string]int{"assigned": len(input.UserIDs)})
}
func (h *ResearchHandler) report(w http.ResponseWriter, r *http.Request) {
	id, err := studyID(r)
	if err != nil {
		researchError(w, err)
		return
	}
	report, err := h.service.Report(r.Context(), id, r.URL.Query().Get("include_unversioned") == "true")
	if err != nil {
		researchError(w, err)
		return
	}
	writeJSON(w, 200, report)
}
func csvValue(v any) string {
	switch value := v.(type) {
	case nil:
		return ""
	case *float64:
		if value == nil {
			return ""
		}
		return strconv.FormatFloat(*value, 'g', -1, 64)
	case *int:
		if value == nil {
			return ""
		}
		return strconv.Itoa(*value)
	case *int64:
		if value == nil {
			return ""
		}
		return strconv.FormatInt(*value, 10)
	case *bool:
		if value == nil {
			return ""
		}
		return strconv.FormatBool(*value)
	case *string:
		if value == nil {
			return ""
		}
		return *value
	case time.Time:
		return value.UTC().Format(time.RFC3339Nano)
	case *time.Time:
		if value == nil {
			return ""
		}
		return value.UTC().Format(time.RFC3339Nano)
	case json.RawMessage:
		return string(value)
	}
	return fmt.Sprint(v)
}
func safeCSV(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed != "" && strings.ContainsAny(trimmed[:1], "=+-@") {
		if number, err := strconv.ParseFloat(trimmed, 64); err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return "'" + value
		}
	}
	return value
}
func (h *ResearchHandler) export(w http.ResponseWriter, r *http.Request) {
	id, err := studyID(r)
	if err != nil {
		researchError(w, err)
		return
	}
	dataset := r.URL.Query().Get("dataset")
	allowed := map[string]bool{"participants": true, "assessments": true, "answers": true, "ai_events": true, "metrics": true, "protocol": true, "mapping": true}
	if !allowed[dataset] {
		researchError(w, ports.ErrResearchInvalid)
		return
	}
	data, err := h.service.Dataset(r.Context(), id)
	if err != nil {
		researchError(w, err)
		return
	}
	if dataset == "ai_events" && !data.AIAvailable {
		http.Error(w, "Детализация ИИ недоступна. Экспорт не будет подменён пустыми данными.", 503)
		return
	}
	now := time.Now().UTC()
	report := domain.BuildResearchReport(data, r.URL.Query().Get("include_unversioned") == "true", now)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="study-%d-%s-%s.csv"`, id, dataset, now.Format("20060102T150405Z")))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte("\xef\xbb\xbf"))
	writer := csv.NewWriter(w)
	defer writer.Flush()
	emit := func(values ...any) {
		row := make([]string, len(values))
		for i, v := range values {
			row[i] = safeCSV(csvValue(v))
		}
		_ = writer.Write(row)
	}
	members := map[int64]domain.ResearchMember{}
	for _, m := range data.Members {
		members[m.UserID] = m
	}
	switch dataset {
	case "participants":
		emit("study_id", "participant_id", "assigned_arm", "course_id", "assigned_at_utc", "pre_assessment_id", "post_assessment_id", "pre_score", "post_score", "gain_pp", "pre_at_utc", "post_at_utc", "eligible", "quality_flags", "completed_pages", "total_pages", "ai_sessions", "ai_answers", "ai_correct", "include_unversioned", "exported_at_utc")
		for _, p := range report.Participants {
			emit(id, p.ParticipantID, p.Arm, p.CourseID, p.AssignedAt, p.PreID, p.PostID, p.PreScore, p.PostScore, p.Gain, p.PreAt, p.PostAt, p.Eligible, strings.Join(p.Flags, ";"), p.CompletedPages, p.TotalPages, p.AISessions, p.AIAnswers, p.AICorrect, report.IncludeUnversioned, now)
		}
	case "assessments":
		emit("study_id", "participant_id", "assigned_arm", "observed_course_id", "assessment_id", "type", "scope", "score", "completed_at_utc", "form_hash", "answers_json", "instrument_snapshot_json", "exported_at_utc")
		for _, a := range data.Assessments {
			member := members[a.UserID]
			emit(id, domain.ParticipantKey(data.Study, a.UserID), member.Arm, a.CourseID, a.ID, a.Type, a.Scope, a.Score, a.CompletedAt, a.FormHash, a.Answers, a.InstrumentSnapshot, now)
		}
	case "answers":
		emit("study_id", "participant_id", "assigned_arm", "course_id", "assessment_id", "type", "scope", "form_hash", "completed_at_utc", "question_id", "answer_id", "correct_answer_id", "is_correct", "question_text", "selected_answer_text", "options_json", "exported_at_utc")
		for _, a := range data.Assessments {
			var answers []domain.TestAttemptAnswer
			var snapshot struct{ Test domain.Test }
			if err := json.Unmarshal(a.Answers, &answers); err != nil {
				continue
			}
			_ = json.Unmarshal(a.InstrumentSnapshot, &snapshot)
			questions := map[int64]domain.TestQuestion{}
			for _, q := range snapshot.Test.Questions {
				questions[q.ID] = q
			}
			for _, answer := range answers {
				q := questions[answer.QuestionID]
				selected := ""
				for _, option := range q.Answers {
					if option.ID == answer.AnswerID {
						selected = option.Text
					}
				}
				options, _ := json.Marshal(q.Answers)
				emit(id, domain.ParticipantKey(data.Study, a.UserID), members[a.UserID].Arm, a.CourseID, a.ID, a.Type, a.Scope, a.FormHash, a.CompletedAt, answer.QuestionID, answer.AnswerID, answer.CorrectAnswerID, answer.IsCorrect, q.Question, selected, string(options), now)
			}
		}
	case "ai_events":
		emit("study_id", "participant_id", "assigned_arm", "observed_course_id", "session_id", "session_status", "session_created_at_utc", "session_completed_at_utc", "round_id", "round_number", "question_id", "question_text", "skill_code", "difficulty", "reason", "selected_option", "correct_option", "is_correct", "answered_at_utc", "model", "prompt_version", "policy_version", "contributes_to_profile", "exported_at_utc")
		for _, event := range data.AIEvents {
			emit(id, domain.ParticipantKey(data.Study, event.UserID), members[event.UserID].Arm, event.CourseID, event.SessionID, event.SessionStatus, event.CreatedAt, event.CompletedAt, event.RoundID, event.RoundNumber, event.QuestionID, event.QuestionText, event.SkillCode, event.Difficulty, event.Reason, event.SelectedOption, event.CorrectOption, event.IsCorrect, event.AnsweredAt, event.Model, event.PromptVersion, event.PolicyVersion, event.Contributes, now)
		}
	case "metrics":
		emit("study_id", "record_type", "arm", "metric", "n", "value", "ci_low", "ci_high", "method", "include_unversioned", "exported_at_utc")
		for _, group := range report.Groups {
			for name, value := range map[string]int{"assigned": group.Assigned, "with_pre": group.WithPre, "with_post": group.WithPost, "paired": group.Paired, "finished_course": group.FinishedCourse} {
				emit(id, "count", group.Arm, name, value, value, nil, nil, "count", report.IncludeUnversioned, now)
			}
			for name, stat := range map[string]domain.SummaryStats{"pre": group.Pre, "post": group.Post, "gain_pp": group.Gain} {
				emit(id, "summary", group.Arm, name, stat.N, stat.Mean, stat.CILow, stat.CIHigh, "mean_t95", report.IncludeUnversioned, now)
				emit(id, "summary", group.Arm, name+"_sd", stat.N, stat.SD, nil, nil, "sample_sd", report.IncludeUnversioned, now)
				emit(id, "summary", group.Arm, name+"_median", stat.N, stat.Median, nil, nil, "median", report.IncludeUnversioned, now)
			}
		}
		c := report.Comparison
		emit(id, "comparison", "adaptive-control", "gain_difference_pp", report.Groups[0].Paired+report.Groups[1].Paired, c.GainDifference, c.CILow, c.CIHigh, "Welch_t95", report.IncludeUnversioned, now)
		emit(id, "comparison", "adaptive-control", "hedges_g_gain", report.Groups[0].Paired+report.Groups[1].Paired, c.HedgesG, nil, nil, "Hedges_g_pooled_gain_sd", report.IncludeUnversioned, now)
		m := report.Model
		emit(id, "model", "adaptive-control", "adjusted_post_difference_pp", m.N, m.GroupCoefficient, m.CILow, m.CIHigh, "OLS_post_pre_group_HC3_t95", report.IncludeUnversioned, now)
	case "mapping":
		emit("study_id", "participant_id", "account_id", "assigned_arm", "assigned_at_utc")
		for _, member := range data.Members {
			emit(id, domain.ParticipantKey(data.Study, member.UserID), member.UserID, member.Arm, member.AssignedAt)
		}
	case "protocol":
		emit("study_id", "title", "design", "scope", "control_course_id", "adaptive_course_id", "expected_pre_hash", "expected_post_hash", "theory_content_hash", "protocol_notes", "pairing_policy", "analysis_population", "missing_values", "include_unversioned", "ai_source_available", "exported_at_utc")
		emit(id, data.Study.Title, data.Study.Design, data.Study.Scope, data.Study.ControlCourseID, data.Study.AdaptiveCourseID, data.Study.PreHash, data.Study.PostHash, data.Study.ContentHash, data.Study.ProtocolNotes, "first_pre_then_first_post_by_time_and_id", "complete_pairs_in_assigned_groups", "empty_not_zero", report.IncludeUnversioned, data.AIAvailable, now)
	}
}
