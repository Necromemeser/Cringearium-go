package application

import (
	"context"
	"errors"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/ports"
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrCourseNotPublished = errors.New("course is not published")
	ErrCourseNotFree = errors.New("course is not free")
	ErrAlreadyHasAccess = errors.New("user already has access to course")
	ErrPageNotFound = errors.New("page not found")
	ErrPageNotInCourse = errors.New("page does not belong to course")
)

type Courses struct { repository ports.CourseRepository }

func NewCourses(repository ports.CourseRepository) *Courses { return &Courses{repository: repository} }

func (c *Courses) GetAll(ctx context.Context) ([]*domain.Course, error) { return c.repository.FindAll(ctx) }

func (c *Courses) GetByID(ctx context.Context, id int64) (*domain.CourseDetails, error) {
	course, err := c.repository.FindByID(ctx, id)
	if err != nil { return nil, err }
	if course == nil { return nil, ErrCourseNotFound }
	return course, nil
}

func (c *Courses) FindEnrolled(ctx context.Context, userID int64) ([]*domain.Course, error) {
	return c.repository.FindEnrolled(ctx, userID)
}

func (c *Courses) Enroll(ctx context.Context, userID, courseID int64) error {
	course, err := c.repository.FindByID(ctx, courseID)
	if err != nil { return err }
	if course == nil { return ErrCourseNotFound }
	if course.Course.Status != domain.CourseStatusPublished { return ErrCourseNotPublished }
	if course.Course.Price != 0 { return ErrCourseNotFree }
	hasAccess, err := c.repository.HasAccess(ctx, userID, courseID)
	if err != nil { return err }
	if hasAccess { return ErrAlreadyHasAccess }
	return c.repository.GrantAccess(ctx, userID, courseID)
}

func (c *Courses) HasAccess(ctx context.Context, userID, courseID int64) (bool, error) {
	if _, err := c.GetByID(ctx, courseID); err != nil { return false, err }
	return c.repository.HasAccess(ctx, userID, courseID)
}

func (c *Courses) GetCompletedPages(ctx context.Context, userID, courseID int64) ([]int64, error) {
	if _, err := c.GetByID(ctx, courseID); err != nil { return nil, err }
	return c.repository.GetCompletedPages(ctx, userID, courseID)
}

func (c *Courses) CompletePage(ctx context.Context, userID, pageID int64) error {
	return c.repository.CompletePage(ctx, userID, pageID)
}
