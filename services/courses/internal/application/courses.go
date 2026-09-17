package application

import (
	"context"
	"errors"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/ports"
)

var (
	ErrCourseNotFound     = errors.New("course not found")
	ErrCourseNotPublished = errors.New("course is not published")
	ErrCourseNotFree      = errors.New("course is not free")
	ErrAlreadyHasAccess   = errors.New("user already has access to course")
	ErrPageNotFound       = errors.New("page not found")
	ErrPageNotInCourse    = errors.New("page does not belong to course")
	ErrTestNotFound       = errors.New("test not found")
	ErrTestAlreadyPassed  = errors.New("test already passed")
	ErrInvalidTestAnswers = errors.New("invalid test answers")
)

type Courses struct {
	repository ports.CourseRepository
}

func NewCourses(repository ports.CourseRepository) *Courses {
	return &Courses{repository: repository}
}

func (c *Courses) GetAll(ctx context.Context) ([]*domain.Course, error) {
	return c.repository.FindAll(ctx)
}

func (c *Courses) GetByID(ctx context.Context, id int64) (*domain.CourseDetails, error) {
	course, err := c.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}
	return course, nil
}

func (c *Courses) FindEnrolled(ctx context.Context, userID int64) ([]*domain.Course, error) {
	return c.repository.FindEnrolled(ctx, userID)
}

func (c *Courses) Enroll(ctx context.Context, userID, courseID int64) error {
	course, err := c.repository.FindByID(ctx, courseID)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrCourseNotFound
	}
	if course.Course.Status != domain.CourseStatusPublished {
		return ErrCourseNotPublished
	}
	if course.Course.Price != 0 {
		return ErrCourseNotFree
	}
	hasAccess, err := c.repository.HasAccess(ctx, userID, courseID)
	if err != nil {
		return err
	}
	if hasAccess {
		return ErrAlreadyHasAccess
	}
	return c.repository.GrantAccess(ctx, userID, courseID)
}

func (c *Courses) HasAccess(ctx context.Context, userID, courseID int64) (bool, error) {
	if _, err := c.GetByID(ctx, courseID); err != nil {
		return false, err
	}
	return c.repository.HasAccess(ctx, userID, courseID)
}

func (c *Courses) GetCompletedPages(ctx context.Context, userID, courseID int64) ([]int64, error) {
	if _, err := c.GetByID(ctx, courseID); err != nil {
		return nil, err
	}
	return c.repository.GetCompletedPages(ctx, userID, courseID)
}

func (c *Courses) CompletePage(ctx context.Context, userID, pageID int64) error {
	return c.repository.CompletePage(ctx, userID, pageID)
}

func (c *Courses) GetTest(ctx context.Context, userID, pageID int64) (*domain.Test, *domain.TestAttempt, error) {
	test, err := c.repository.FindTestByPageID(ctx, userID, pageID)
	if err != nil {
		return nil, nil, err
	}
	if test == nil {
		return nil, nil, ErrTestNotFound
	}

	attempt, err := c.repository.FindLatestTestAttempt(ctx, userID, test.ID)
	if err != nil {
		return nil, nil, err
	}
	if attempt != nil {
		enrichAttemptAnswers(test, attempt.Answers)
	}

	return test, attempt, nil
}

func (c *Courses) SubmitTest(ctx context.Context, userID, testID int64, answers []domain.TestAttemptAnswer) (*domain.TestResult, error) {
	if len(answers) == 0 {
		return nil, ErrInvalidTestAnswers
	}

	test, err := c.repository.FindTestByID(ctx, userID, testID)
	if err != nil {
		return nil, err
	}
	if test == nil {
		return nil, ErrTestNotFound
	}

	attempt, err := c.repository.FindLatestTestAttempt(ctx, userID, testID)
	if err != nil {
		return nil, err
	}
	if attempt != nil && attempt.Passed {
		return nil, ErrTestAlreadyPassed
	}

	if len(test.Questions) == 0 || len(answers) != len(test.Questions) {
		return nil, ErrInvalidTestAnswers
	}

	correctByQuestion := make(map[int64]int64, len(test.Questions))
	questionIDs := make(map[int64]struct{}, len(test.Questions))
	answerIDsByQuestion := make(map[int64]map[int64]struct{}, len(test.Questions))

	for _, question := range test.Questions {
		questionIDs[question.ID] = struct{}{}
		answerIDsByQuestion[question.ID] = make(map[int64]struct{}, len(question.Answers))

		for _, answer := range question.Answers {
			answerIDsByQuestion[question.ID][answer.ID] = struct{}{}
			if answer.IsCorrect {
				correctByQuestion[question.ID] = answer.ID
			}
		}

		if _, ok := correctByQuestion[question.ID]; !ok {
			return nil, ErrInvalidTestAnswers
		}
	}

	seen := make(map[int64]struct{}, len(answers))
	correct := 0
	for i := range answers {
		submitted := &answers[i]

		if _, ok := questionIDs[submitted.QuestionID]; !ok {
			return nil, ErrInvalidTestAnswers
		}
		if _, ok := seen[submitted.QuestionID]; ok {
			return nil, ErrInvalidTestAnswers
		}
		if _, ok := answerIDsByQuestion[submitted.QuestionID][submitted.AnswerID]; !ok {
			return nil, ErrInvalidTestAnswers
		}

		seen[submitted.QuestionID] = struct{}{}
		submitted.CorrectAnswerID = correctByQuestion[submitted.QuestionID]
		submitted.IsCorrect = submitted.AnswerID == submitted.CorrectAnswerID
		if submitted.IsCorrect {
			correct++
		}
	}

	if len(seen) != len(test.Questions) {
		return nil, ErrInvalidTestAnswers
	}

	score := correct * 100 / len(test.Questions)
	passed := score >= test.PassingScore

	result, err := c.repository.SubmitTest(ctx, userID, testID, answers, score, passed)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("test result is nil")
	}

	result.PassingScore = test.PassingScore
	result.Answers = answers
	return result, nil
}

func enrichAttemptAnswers(test *domain.Test, answers []domain.TestAttemptAnswer) {
	correctByQuestion := make(map[int64]int64, len(test.Questions))

	for _, question := range test.Questions {
		for _, answer := range question.Answers {
			if answer.IsCorrect {
				correctByQuestion[question.ID] = answer.ID
				break
			}
		}
	}

	for i := range answers {
		correctAnswerID := correctByQuestion[answers[i].QuestionID]

		answers[i].CorrectAnswerID = correctAnswerID
		answers[i].IsCorrect = answers[i].AnswerID == correctAnswerID
	}
}
