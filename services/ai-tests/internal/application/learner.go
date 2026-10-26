package application

import (
	"context"
	"fmt"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

func (s *AdaptiveTestService) GetLearningProfile(ctx context.Context, userID, courseID int64, topicPageID *int64) (domain.LearningProfile, error) {
	course, err := s.courses.GetCourseContext(ctx, userID, courseID, topicPageID)
	if err != nil {
		return domain.LearningProfile{}, err
	}
	return s.learningProfile(ctx, userID, course, 5)
}
func (s *AdaptiveTestService) learningProfile(ctx context.Context, userID int64, course ports.CourseContext, count int) (domain.LearningProfile, error) {
	codes := make([]string, 0, len(course.Skills))
	for _, skill := range course.Skills {
		codes = append(codes, skill.Code)
	}
	var states []domain.SkillState
	if len(codes) > 0 {
		var err error
		states, err = s.learners.GetSkillStates(ctx, userID, codes)
		if err != nil {
			return domain.LearningProfile{}, fmt.Errorf("load learner profile: %w", err)
		}
	}
	return domain.BuildProfile(course.Skills, states, count, time.Now()), nil
}

func validateLearningPlan(g ports.GeneratedRound, plan []domain.QuestionTarget) error {
	if len(plan) == 0 {
		for _, q := range g.Questions {
			if q.SkillCode != "" {
				return fmt.Errorf("%w: unplanned skill", ErrInvalidGeneratedRound)
			}
		}
		return nil
	}
	if len(g.Questions) != len(plan) {
		return fmt.Errorf("%w: question plan length mismatch", ErrInvalidGeneratedRound)
	}
	for i, q := range g.Questions {
		target := plan[i]
		if q.SkillCode != target.SkillCode || q.TopicPageID == nil || *q.TopicPageID != target.TopicPageID || q.Difficulty != target.Difficulty {
			return fmt.Errorf("%w: question %d does not match the backend skill plan", ErrInvalidGeneratedRound, i+1)
		}
	}
	return nil
}
