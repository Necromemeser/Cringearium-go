package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application/mocks"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"go.uber.org/mock/gomock"
)

func TestCoursesGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	courses := []*domain.Course{{ID: 1, Title: "Course 1"}}
	repo.EXPECT().FindAll(gomock.Any()).Return(courses, nil)

	got, err := service.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(got) != 1 || got[0] != courses[0] {
		t.Fatalf("GetAll() = %#v, want %#v", got, courses)
	}
}

func TestCoursesGetAllRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	wantErr := errors.New("repository error")
	repo.EXPECT().FindAll(gomock.Any()).Return(nil, wantErr)

	_, err := service.GetAll(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetAll() error = %v, want %v", err, wantErr)
	}
}

func TestCoursesGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	course := &domain.CourseDetails{Course: domain.Course{ID: 1, Title: "Course 1"}}
	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(course, nil)

	got, err := service.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != course {
		t.Fatalf("GetByID() = %#v, want %#v", got, course)
	}
}

func TestCoursesGetByIDNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, nil)

	_, err := service.GetByID(context.Background(), 1)
	if !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, ErrCourseNotFound)
	}
}

func TestCoursesGetByIDRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	wantErr := errors.New("repository error")
	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, wantErr)

	_, err := service.GetByID(context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetByID() error = %v, want %v", err, wantErr)
	}
}

func TestCoursesEnroll(t *testing.T) {
	tests := []struct {
		name    string
		course  *domain.CourseDetails
		hasAccess bool
		setup   func(*mocks.MockCourseRepository)
		wantErr error
	}{
		{
			name:   "course not found",
			course: nil,
			setup: func(repo *mocks.MockCourseRepository) {
				repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, nil)
			},
			wantErr: ErrCourseNotFound,
		},
		{
			name: "course not published",
			course: &domain.CourseDetails{Course: domain.Course{
				ID: 1, Status: domain.CourseStatusDraft,
			}},
			setup: func(repo *mocks.MockCourseRepository) {
				repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusDraft}}, nil)
			},
			wantErr: ErrCourseNotPublished,
		},
		{
			name: "course is paid",
			setup: func(repo *mocks.MockCourseRepository) {
				repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished, Price: 99000}}, nil)
			},
			wantErr: ErrCourseNotFree,
		},
		{
			name: "already has access",
			setup: func(repo *mocks.MockCourseRepository) {
				repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished}}, nil)
				repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(true, nil)
			},
			wantErr: ErrAlreadyHasAccess,
		},
		{
			name: "has access check error",
			setup: func(repo *mocks.MockCourseRepository) {
				repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished}}, nil)
				repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(false, errors.New("repository error"))
			},
			wantErr: errors.New("repository error"),
		},
		{
			name: "grant access error",
			setup: func(repo *mocks.MockCourseRepository) {
				repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished}}, nil)
				repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(false, nil)
				repo.EXPECT().GrantAccess(gomock.Any(), int64(42), int64(1)).Return(errors.New("grant error"))
			},
			wantErr: errors.New("grant error"),
		},
		{
			name: "success",
			setup: func(repo *mocks.MockCourseRepository) {
				repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished}}, nil)
				repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(false, nil)
				repo.EXPECT().GrantAccess(gomock.Any(), int64(42), int64(1)).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockCourseRepository(ctrl)
			service := NewCourses(repo)
			tt.setup(repo)

			err := service.Enroll(context.Background(), 42, 1)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Enroll() error = %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr.Error() {
				t.Fatalf("Enroll() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCoursesEnrollFindByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	wantErr := errors.New("repository error")
	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, wantErr)

	err := service.Enroll(context.Background(), 42, 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Enroll() error = %v, want %v", err, wantErr)
	}
}

func TestCoursesHasAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1}}, nil)
	repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(true, nil)

	got, err := service.HasAccess(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("HasAccess() error = %v", err)
	}
	if !got {
		t.Fatal("HasAccess() = false, want true")
	}
}

func TestCoursesHasAccessCourseNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, nil)

	got, err := service.HasAccess(context.Background(), 42, 1)
	if !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf("HasAccess() error = %v, want %v", err, ErrCourseNotFound)
	}
	if got {
		t.Fatal("HasAccess() = true, want false")
	}
}

func TestCoursesHasAccessRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	wantErr := errors.New("repository error")
	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1}}, nil)
	repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(false, wantErr)

	_, err := service.HasAccess(context.Background(), 42, 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("HasAccess() error = %v, want %v", err, wantErr)
	}
}
