package domain

// Codes are stable across course variants; page IDs remain local to the course.
type LearningSkill struct {
	Code             string   `json:"code"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	PageIDs          []int64  `json:"page_ids"`
	Prerequisites    []string `json:"prerequisites"`
	PrerequisiteOnly bool     `json:"prerequisite_only"`
}
