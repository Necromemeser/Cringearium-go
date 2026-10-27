package application

import (
	"context"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func (c *Courses) addLearningSkills(ctx context.Context, course *domain.CourseDetails, result *domain.AIContext) error {
	skills, err := c.repository.FindLearningSkills(ctx, course.Course.ID)
	if err != nil {
		return err
	}
	selectedPages := map[int64]bool{}
	for _, m := range result.Materials {
		selectedPages[m.PageID] = true
	}
	all := map[string]domain.LearningSkill{}
	selected := map[string]bool{}
	for _, skill := range skills {
		all[skill.Code] = skill
		for _, id := range skill.PageIDs {
			if selectedPages[id] {
				selected[skill.Code] = true
			}
		}
	}
	// A visited set also makes malformed cyclic catalogues terminate safely.
	needed := map[string]bool{}
	var visit func(string)
	visit = func(code string) {
		if needed[code] {
			return
		}
		skill, exists := all[code]
		if !exists {
			return
		}
		needed[code] = true
		for _, p := range skill.Prerequisites {
			visit(p)
		}
	}
	for code := range selected {
		visit(code)
	}
	pageIDs := map[int64]bool{}
	for _, skill := range skills {
		if needed[skill.Code] {
			skill.PrerequisiteOnly = !selected[skill.Code]
			result.Skills = append(result.Skills, skill)
			for _, id := range skill.PageIDs {
				pageIDs[id] = true
			}
		}
	}
	// Supply explicitly linked prerequisite material, without widening to unrelated sections.
	for _, section := range course.Sections {
		for _, page := range section.Pages {
			if page.Type == domain.PageTypeTheory && pageIDs[page.ID] && !selectedPages[page.ID] {
				result.Materials = append(result.Materials, domain.AIContextMaterial{PageID: page.ID, Title: page.Title, Content: page.Content})
				result.AllowedTopics = append(result.AllowedTopics, domain.AIContextTopic{PageID: page.ID, Title: page.Title})
				selectedPages[page.ID] = true
			}
		}
	}
	return nil
}
