package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

type ResearchStudy struct {
	ID               int64     `json:"id" db:"id"`
	Title            string    `json:"title" db:"title"`
	Design           string    `json:"design" db:"design"`
	Scope            string    `json:"scope" db:"scope"`
	ControlCourseID  int64     `json:"control_course_id" db:"control_course_id"`
	AdaptiveCourseID int64     `json:"adaptive_course_id" db:"adaptive_course_id"`
	PreHash          string    `json:"pre_hash" db:"pre_hash"`
	PostHash         string    `json:"post_hash" db:"post_hash"`
	ContentHash      string    `json:"content_hash" db:"content_hash"`
	ProtocolNotes    string    `json:"protocol_notes" db:"protocol_notes"`
	ExportSalt       string    `json:"-" db:"export_salt"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
type ResearchMember struct {
	UserID         int64
	Arm            string
	AssignedAt     time.Time
	CompletedPages int
	TotalPages     int
	OtherArmAccess bool
}
type ResearchInstrument struct {
	Kind  string
	Scope string
	Test  Test
}
type AIResearchEvent struct {
	QuestionText   *string    `json:"question_text"`
	CorrectOption  *string    `json:"correct_option"`
	Reason         *string    `json:"reason"`
	SessionID      string     `json:"session_id"`
	UserID         int64      `json:"user_id"`
	CourseID       int64      `json:"course_id"`
	SessionStatus  string     `json:"session_status"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	RoundID        *int64     `json:"round_id"`
	RoundNumber    *int       `json:"round_number"`
	QuestionID     *int64     `json:"question_id"`
	SkillCode      *string    `json:"skill_code"`
	Difficulty     *int       `json:"difficulty"`
	SelectedOption *string    `json:"selected_option"`
	IsCorrect      *bool      `json:"is_correct"`
	AnsweredAt     *time.Time `json:"answered_at"`
	PromptVersion  *string    `json:"prompt_version"`
	Model          *string    `json:"model"`
	PolicyVersion  *string    `json:"policy_version"`
	Contributes    *bool      `json:"contributes"`
}
type ResearchDataset struct {
	Study       ResearchStudy
	Members     []ResearchMember
	Assessments []AssessmentResult
	AIEvents    []AIResearchEvent
	AIAvailable bool
	Warnings    []string
}
type ResearchParticipant struct {
	ParticipantID  string     `json:"participant_id"`
	UserID         int64      `json:"account_id"`
	Arm            string     `json:"arm"`
	CourseID       int64      `json:"course_id"`
	AssignedAt     time.Time  `json:"assigned_at"`
	PreScore       *float64   `json:"pre_score"`
	PostScore      *float64   `json:"post_score"`
	Gain           *float64   `json:"gain"`
	PreID          *int64     `json:"pre_id"`
	PostID         *int64     `json:"post_id"`
	PreAt          *time.Time `json:"pre_at"`
	PostAt         *time.Time `json:"post_at"`
	Eligible       bool       `json:"eligible"`
	Flags          []string   `json:"flags"`
	CompletedPages int        `json:"completed_pages"`
	TotalPages     int        `json:"total_pages"`
	AISessions     *int       `json:"ai_sessions"`
	AIAnswers      *int       `json:"ai_answers"`
	AICorrect      *int       `json:"ai_correct"`
}
type ResearchReport struct {
	Study              ResearchStudy          `json:"study"`
	GeneratedAt        time.Time              `json:"generated_at"`
	IncludeUnversioned bool                   `json:"include_unversioned"`
	Participants       []ResearchParticipant  `json:"participants"`
	Groups             []ResearchGroupMetrics `json:"groups"`
	Comparison         ResearchComparison     `json:"comparison"`
	Model              RegressionResult       `json:"model"`
	Warnings           []string               `json:"warnings"`
	AIAvailable        bool                   `json:"ai_available"`
	QualityCounts      map[string]int         `json:"quality_counts"`
}

func ParticipantKey(study ResearchStudy, userID int64) string {
	mac := hmac.New(sha256.New, []byte(study.ExportSalt))
	fmt.Fprintf(mac, "study:%d:user:%d", study.ID, userID)
	return hex.EncodeToString(mac.Sum(nil))[:16]
}
func before(a, b AssessmentResult) bool {
	if !a.CompletedAt.Equal(b.CompletedAt) {
		return a.CompletedAt.Before(b.CompletedAt)
	}
	return a.ID < b.ID
}
func BuildResearchReport(data ResearchDataset, includeUnversioned bool, now time.Time) ResearchReport {
	report := ResearchReport{Study: data.Study, GeneratedAt: now, IncludeUnversioned: includeUnversioned, Participants: []ResearchParticipant{}, Warnings: append([]string{}, data.Warnings...), AIAvailable: data.AIAvailable, QualityCounts: map[string]int{}}
	if data.Study.Design == "observational" {
		report.Warnings = append(report.Warnings, "observational_design")
	}
	if !data.AIAvailable {
		report.Warnings = append(report.Warnings, "ai_source_unavailable")
	}
	if data.Study.PreHash == data.Study.PostHash {
		report.Warnings = append(report.Warnings, "same_pre_post_form")
	}
	report.Warnings = append(report.Warnings, "uncalibrated_instruments", "complete_cases_only", "active_time_not_measured", "no_delayed_retention_test")
	attempts := append([]AssessmentResult{}, data.Assessments...)
	sort.Slice(attempts, func(i, j int) bool { return before(attempts[i], attempts[j]) })
	for _, member := range data.Members {
		courseID := data.Study.ControlCourseID
		if member.Arm == "adaptive" {
			courseID = data.Study.AdaptiveCourseID
		}
		row := ResearchParticipant{ParticipantID: ParticipantKey(data.Study, member.UserID), UserID: member.UserID, Arm: member.Arm, CourseID: courseID, AssignedAt: member.AssignedAt, Flags: []string{}, CompletedPages: member.CompletedPages, TotalPages: member.TotalPages}
		if member.OtherArmAccess {
			row.Flags = append(row.Flags, "other_arm_access")
		}
		var pres, posts []AssessmentResult
		for _, attempt := range attempts {
			if attempt.UserID != member.UserID || attempt.CourseID != courseID || attempt.Scope != data.Study.Scope {
				continue
			}
			if data.Study.Design == "randomized" && attempt.CompletedAt.Before(member.AssignedAt) {
				continue
			}
			if attempt.Type == "pretest" {
				pres = append(pres, attempt)
			} else if attempt.Type == "posttest" {
				posts = append(posts, attempt)
			}
		}
		if len(pres) > 1 {
			row.Flags = append(row.Flags, "multiple_pretests")
		}
		if len(posts) > 1 {
			row.Flags = append(row.Flags, "multiple_posttests")
		}
		valid := true
		if len(pres) == 0 {
			row.Flags = append(row.Flags, "missing_pretest")
			valid = false
		} else {
			pre := pres[0]
			value := float64(pre.Score)
			row.PreScore = &value
			id := pre.ID
			row.PreID = &id
			date := pre.CompletedAt
			row.PreAt = &date
			if pre.FormHash == nil || *pre.FormHash == "" {
				row.Flags = append(row.Flags, "unversioned_pretest")
				if !includeUnversioned {
					valid = false
				}
			} else if *pre.FormHash != data.Study.PreHash {
				row.Flags = append(row.Flags, "incompatible_pretest")
				valid = false
			}
		}
		var post *AssessmentResult
		for i := range posts {
			if len(pres) > 0 && !before(pres[0], posts[i]) {
				row.Flags = appendUnique(row.Flags, "post_before_pretest")
				valid = false
				continue
			}
			if post == nil {
				post = &posts[i]
			}
		}
		if post == nil {
			row.Flags = append(row.Flags, "missing_posttest")
			valid = false
		} else {
			value := float64(post.Score)
			row.PostScore = &value
			id := post.ID
			row.PostID = &id
			date := post.CompletedAt
			row.PostAt = &date
			if post.FormHash == nil || *post.FormHash == "" {
				row.Flags = append(row.Flags, "unversioned_posttest")
				if !includeUnversioned {
					valid = false
				}
			} else if *post.FormHash != data.Study.PostHash {
				row.Flags = append(row.Flags, "incompatible_posttest")
				valid = false
			}
		}
		if row.PreScore != nil && row.PostScore != nil {
			gain := *row.PostScore - *row.PreScore
			row.Gain = &gain
		}
		row.Eligible = valid && row.Gain != nil
		if data.AIAvailable {
			sessions := map[string]bool{}
			answers, correct := 0, 0
			for _, event := range data.AIEvents {
				if event.UserID != member.UserID {
					continue
				}
				if data.Study.Design == "randomized" && event.CreatedAt.Before(member.AssignedAt) {
					continue
				}
				if event.CourseID != courseID {
					if member.Arm == "control" && event.CourseID == data.Study.AdaptiveCourseID {
						row.Flags = appendUnique(row.Flags, "ai_in_other_arm")
					}
					continue
				}
				sessions[event.SessionID] = true
				if event.IsCorrect != nil {
					answers++
					if *event.IsCorrect {
						correct++
					}
				}
			}
			n := len(sessions)
			row.AISessions = &n
			row.AIAnswers = &answers
			row.AICorrect = &correct
		}
		for _, flag := range row.Flags {
			report.QualityCounts[flag]++
		}
		report.Participants = append(report.Participants, row)
	}
	for _, arm := range []string{"control", "adaptive"} {
		report.Groups = append(report.Groups, groupMetrics(arm, report.Participants))
	}
	report.Comparison = compareGroups(report.Groups)
	report.Model = FitANCOVA(report.Participants)
	if report.Groups[0].Paired < 10 || report.Groups[1].Paired < 10 {
		report.Warnings = append(report.Warnings, "small_sample")
	}
	return report
}
func appendUnique(values []string, value string) []string {
	for _, v := range values {
		if v == value {
			return values
		}
	}
	return append(values, value)
}

func AllocateResearchArm(control, adaptive int, tieBit bool) string {
	if adaptive < control || (adaptive == control && tieBit) {
		return "adaptive"
	}
	return "control"
}
