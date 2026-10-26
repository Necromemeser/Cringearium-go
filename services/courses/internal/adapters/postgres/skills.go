package postgres

import (
	"context"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func (db *DB) FindLearningSkills(ctx context.Context, courseID int64) ([]domain.LearningSkill, error) {
	var rows []struct {
		Code        string
		Title       string
		Description string
		PageID      int64 `db:"page_id"`
	}
	err := db.conn.SelectContext(ctx, &rows, `SELECT sk.code,sk.title,sk.description,p.id AS page_id
 FROM learning_skills sk JOIN page_learning_skills ps ON ps.skill_code=sk.code
 JOIN course_pages p ON p.id=ps.page_id JOIN course_sections s ON s.id=p.section_id
 WHERE s.course_id=$1 ORDER BY sk.code,p.id`, courseID)
	if err != nil {
		return nil, err
	}
	var result []domain.LearningSkill
	positions := map[string]int{}
	for _, row := range rows {
		i, ok := positions[row.Code]
		if !ok {
			i = len(result)
			positions[row.Code] = i
			result = append(result, domain.LearningSkill{Code: row.Code, Title: row.Title, Description: row.Description})
		}
		result[i].PageIDs = append(result[i].PageIDs, row.PageID)
	}
	// A single query avoids per-skill dependency lookups.
	var deps []struct {
		Skill        string `db:"skill_code"`
		Prerequisite string `db:"prerequisite_code"`
	}
	if err := db.conn.SelectContext(ctx, &deps, `SELECT DISTINCT d.skill_code,d.prerequisite_code
 FROM learning_skill_prerequisites d JOIN page_learning_skills ps ON ps.skill_code=d.skill_code
 JOIN course_pages p ON p.id=ps.page_id JOIN course_sections s ON s.id=p.section_id
 WHERE s.course_id=$1 ORDER BY d.skill_code,d.prerequisite_code`, courseID); err != nil {
		return nil, err
	}
	for _, d := range deps {
		if i, ok := positions[d.Skill]; ok {
			result[i].Prerequisites = append(result[i].Prerequisites, d.Prerequisite)
		}
	}
	return result, nil
}
