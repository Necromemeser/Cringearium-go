package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

const LearningPolicyVersion = "skills-v1"

type LearningSkill struct {
	Code             string   `json:"code"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	PageIDs          []int64  `json:"page_ids"`
	Prerequisites    []string `json:"prerequisites"`
	PrerequisiteOnly bool     `json:"prerequisite_only"`
}
type SkillState struct {
	Code            string     `json:"code"`
	NovelAnswers    int        `json:"novel_answers"`
	CorrectAnswers  int        `json:"correct_answers"`
	Sessions        int        `json:"sessions"`
	Recent          []bool     `json:"recent"`
	LastPracticedAt *time.Time `json:"last_practiced_at,omitempty"`
}

// Status is a transparent evidence rule, not a calibrated probability of knowledge.
func (s SkillState) Status(now time.Time) string {
	n := len(s.Recent)
	if n >= 2 && !s.Recent[n-1] && !s.Recent[n-2] {
		return "needs_practice"
	}
	if s.NovelAnswers < 3 {
		return "insufficient_data"
	}
	correct := 0
	for _, v := range s.Recent {
		if v {
			correct++
		}
	}
	if s.NovelAnswers >= 6 && s.Sessions >= 2 && n >= 3 && s.Recent[n-1] && s.Recent[n-2] && s.Recent[n-3] && correct*5 >= n*4 {
		if s.LastPracticedAt != nil && now.Sub(*s.LastPracticedAt) >= 14*24*time.Hour {
			return "review_due"
		}
		return "provisionally_mastered"
	}
	return "practicing"
}
func (s SkillState) WithEvidence(correct bool, sessions int, now time.Time) SkillState {
	s.NovelAnswers++
	if correct {
		s.CorrectAnswers++
	}
	s.Sessions = sessions
	s.Recent = append(append([]bool(nil), s.Recent...), correct)
	if len(s.Recent) > 5 {
		s.Recent = s.Recent[len(s.Recent)-5:]
	}
	s.LastPracticedAt = &now
	return s
}

type QuestionTarget struct {
	SkillCode   string `json:"skill_code"`
	SkillTitle  string `json:"skill_title"`
	TopicPageID int64  `json:"topic_page_id"`
	Difficulty  int    `json:"difficulty"`
	Reason      string `json:"reason"`
}
type SkillView struct {
	LearningSkill
	State  SkillState `json:"state"`
	Status string     `json:"status"`
}
type LearningProfile struct {
	PolicyVersion string           `json:"policy_version"`
	Skills        []SkillView      `json:"skills"`
	NextPractice  []QuestionTarget `json:"next_practice"`
}

func BuildProfile(skills []LearningSkill, states []SkillState, count int, now time.Time) LearningProfile {
	byCode := map[string]SkillState{}
	for _, s := range states {
		byCode[s.Code] = s
	}
	result := LearningProfile{PolicyVersion: LearningPolicyVersion, Skills: make([]SkillView, 0, len(skills)), NextPractice: []QuestionTarget{}}
	type candidate struct {
		target   QuestionTarget
		priority int
	}
	var candidates []candidate
	for _, skill := range skills {
		state := byCode[skill.Code]
		state.Code = skill.Code
		status := state.Status(now)
		result.Skills = append(result.Skills, SkillView{LearningSkill: skill, State: state, Status: status})
		if len(skill.PageIDs) == 0 {
			continue
		}
		reason, priority, difficulty := "practice", 60, 2
		switch status {
		case "insufficient_data":
			reason, priority = "diagnose", 70
		case "needs_practice":
			reason, priority, difficulty = "remediate", 100, 1
		case "provisionally_mastered":
			reason, priority, difficulty = "challenge", 20, 3
		case "review_due":
			reason, priority = "review", 90
		}
		if skill.PrerequisiteOnly {
			if status != "needs_practice" && status != "review_due" {
				continue
			}
			reason, priority = "prerequisite", 120
		}
		for _, code := range skill.Prerequisites {
			p := byCode[code]
			if p.Status(now) == "needs_practice" {
				priority -= 30
				break
			}
		}
		candidates = append(candidates, candidate{QuestionTarget{SkillCode: skill.Code, SkillTitle: skill.Title, TopicPageID: skill.PageIDs[0], Difficulty: difficulty, Reason: reason}, priority})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority > candidates[j].priority
		}
		return candidates[i].target.SkillCode < candidates[j].target.SkillCode
	})
	if count <= 0 || len(candidates) == 0 {
		return result
	}
	// A majority of the block targets the greatest need; remaining slots sample other skills.
	focus := (count + 1) / 2
	for i := 0; i < count; i++ {
		index := 0
		if i >= focus && len(candidates) > 1 {
			index = 1 + (i-focus)%(len(candidates)-1)
		}
		result.NextPractice = append(result.NextPractice, candidates[index].target)
	}
	return result
}

func QuestionFingerprint(q AdaptiveQuestion) string {
	normalize := func(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
	options := make([]string, 0, len(q.Options))
	for _, o := range q.Options {
		options = append(options, normalize(o.Text))
	}
	sort.Strings(options)
	sum := sha256.Sum256([]byte(normalize(q.Question) + "\x00" + strings.Join(options, "\x00")))
	return hex.EncodeToString(sum[:])
}
