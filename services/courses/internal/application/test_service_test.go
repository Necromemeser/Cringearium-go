package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/mocks"
	"go.uber.org/mock/gomock"
)

func testFixture() *domain.Test {
	return &domain.Test{
		ID: 1, PageID: 10, PassingScore: 70,
		Questions: []domain.TestQuestion{
			{ID: 101, TestID: 1, Question: "2 + 2 = ?", Position: 0, Answers: []domain.TestAnswer{
				{ID: 1001, QuestionID: 101, Text: "3", Position: 0},
				{ID: 1002, QuestionID: 101, Text: "4", Position: 1, IsCorrect: true},
			}},
			{ID: 102, TestID: 1, Question: "3 + 3 = ?", Position: 1, Answers: []domain.TestAnswer{
				{ID: 1003, QuestionID: 102, Text: "5", Position: 0},
				{ID: 1004, QuestionID: 102, Text: "6", Position: 1, IsCorrect: true},
			}},
			{ID: 103, TestID: 1, Question: "4 + 4 = ?", Position: 2, Answers: []domain.TestAnswer{
				{ID: 1005, QuestionID: 103, Text: "7", Position: 0},
				{ID: 1006, QuestionID: 103, Text: "8", Position: 1, IsCorrect: true},
			}},
		},
	}
}

func TestCoursesGetTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()

	repo.EXPECT().FindTestByPageID(gomock.Any(), int64(42), int64(10)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, nil)

	got, attempt, err := service.GetTest(context.Background(), 42, 10)
	if err != nil { t.Fatalf("GetTest() error = %v", err) }
	if got != test { t.Fatalf("GetTest() = %#v, want %#v", got, test) }
	if attempt != nil { t.Fatalf("GetTest() attempt = %#v, want nil", attempt) }
}

func TestCoursesGetTestWithAttempt(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()
	attempt := &domain.TestAttempt{ID: 15, Score: 33, Answers: []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1001, IsCorrect: false},
		{QuestionID: 102, AnswerID: 1004, IsCorrect: true},
		{QuestionID: 103, AnswerID: 1005, IsCorrect: false},
	}}

	repo.EXPECT().FindTestByPageID(gomock.Any(), int64(42), int64(10)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(attempt, nil)

	_, got, err := service.GetTest(context.Background(), 42, 10)
	if err != nil { t.Fatalf("GetTest() error = %v", err) }
	if got != attempt { t.Fatalf("GetTest() attempt = %#v, want %#v", got, attempt) }
	if got.Answers[0].CorrectAnswerID != 1002 || got.Answers[1].CorrectAnswerID != 1004 || got.Answers[2].CorrectAnswerID != 1006 {
		t.Fatalf("GetTest() correct answer ids = %#v", got.Answers)
	}
}

func TestCoursesGetTestNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	repo.EXPECT().FindTestByPageID(gomock.Any(), int64(42), int64(10)).Return(nil, nil)

	_, _, err := service.GetTest(context.Background(), 42, 10)
	if !errors.Is(err, ErrTestNotFound) { t.Fatalf("GetTest() error = %v, want %v", err, ErrTestNotFound) }
}

func TestCoursesGetTestAttemptRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()
	wantErr := errors.New("repository error")

	repo.EXPECT().FindTestByPageID(gomock.Any(), int64(42), int64(10)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, wantErr)

	_, _, err := service.GetTest(context.Background(), 42, 10)
	if !errors.Is(err, wantErr) { t.Fatalf("GetTest() error = %v, want %v", err, wantErr) }
}

func TestCoursesSubmitTestPassed(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()
	answers := []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}, {QuestionID: 102, AnswerID: 1004}, {QuestionID: 103, AnswerID: 1006}}
	expectedResult := &domain.TestResult{AttemptID: 15, Score: 100, Passed: true}

	repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, nil)
	repo.EXPECT().SubmitTest(gomock.Any(), int64(42), int64(1), answers, 100, true).Return(expectedResult, nil)

	got, err := service.SubmitTest(context.Background(), 42, 1, answers)
	if err != nil { t.Fatalf("SubmitTest() error = %v", err) }
	if got.Score != 100 || !got.Passed || got.PassingScore != 70 { t.Fatalf("SubmitTest() result = %#v", got) }
	if got.Answers[0].CorrectAnswerID != 1002 || got.Answers[1].CorrectAnswerID != 1004 || got.Answers[2].CorrectAnswerID != 1006 { t.Fatalf("SubmitTest() answers = %#v", got.Answers) }
}

func TestCoursesSubmitTestFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()
	answers := []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}, {QuestionID: 102, AnswerID: 1003}, {QuestionID: 103, AnswerID: 1005}}
	expectedResult := &domain.TestResult{AttemptID: 16, Score: 33, Passed: false}

	repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, nil)
	repo.EXPECT().SubmitTest(gomock.Any(), int64(42), int64(1), answers, 33, false).Return(expectedResult, nil)

	got, err := service.SubmitTest(context.Background(), 42, 1, answers)
	if err != nil { t.Fatalf("SubmitTest() error = %v", err) }
	if got.Score != 33 || got.Passed || got.PassingScore != 70 { t.Fatalf("SubmitTest() result = %#v", got) }
	if got.Answers[0].IsCorrect || !got.Answers[1].IsCorrect || got.Answers[2].IsCorrect { t.Fatalf("SubmitTest() grading = %#v", got.Answers) }
}

func TestCoursesSubmitTestAlreadyPassed(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()
	attempt := &domain.TestAttempt{ID: 15, Score: 100, Passed: true}

	repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(attempt, nil)

	_, err := service.SubmitTest(context.Background(), 42, 1, []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}, {QuestionID: 102, AnswerID: 1004}, {QuestionID: 103, AnswerID: 1006}})
	if !errors.Is(err, ErrTestAlreadyPassed) { t.Fatalf("SubmitTest() error = %v, want %v", err, ErrTestAlreadyPassed) }
}

func TestCoursesSubmitTestNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(nil, nil)

	_, err := service.SubmitTest(context.Background(), 42, 1, []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}})
	if !errors.Is(err, ErrTestNotFound) { t.Fatalf("SubmitTest() error = %v, want %v", err, ErrTestNotFound) }
}

func TestCoursesSubmitTestRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()
	wantErr := errors.New("repository error")
	answers := []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}, {QuestionID: 102, AnswerID: 1004}, {QuestionID: 103, AnswerID: 1006}}

	repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, nil)
	repo.EXPECT().SubmitTest(gomock.Any(), int64(42), int64(1), answers, 100, true).Return(nil, wantErr)

	_, err := service.SubmitTest(context.Background(), 42, 1, answers)
	if !errors.Is(err, wantErr) { t.Fatalf("SubmitTest() error = %v, want %v", err, wantErr) }
}

func TestCoursesSubmitTestEmptyAnswers(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	_, err := service.SubmitTest(context.Background(), 42, 1, nil)
	if !errors.Is(err, ErrInvalidTestAnswers) { t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers) }
}

func TestCoursesSubmitTestInvalidAnswers(t *testing.T) {
	tests := []struct {
		name string
		answers []domain.TestAttemptAnswer
	}{
		{"incomplete", []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}, {QuestionID: 102, AnswerID: 1004}}},
		{"duplicate question", []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}, {QuestionID: 101, AnswerID: 1001}, {QuestionID: 102, AnswerID: 1004}}},
		{"unknown question", []domain.TestAttemptAnswer{{QuestionID: 999, AnswerID: 1002}, {QuestionID: 102, AnswerID: 1004}, {QuestionID: 103, AnswerID: 1006}}},
		{"unknown answer", []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 999}, {QuestionID: 102, AnswerID: 1004}, {QuestionID: 103, AnswerID: 1006}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockCourseRepository(ctrl)
			service := NewCourses(repo)
			test := testFixture()
			repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(test, nil)
			repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, nil)

			_, err := service.SubmitTest(context.Background(), 42, 1, tt.answers)
			if !errors.Is(err, ErrInvalidTestAnswers) { t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers) }
		})
	}
}

func TestCoursesSubmitTestQuestionWithoutCorrectAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := testFixture()
	test.Questions[0].Answers[1].IsCorrect = false

	repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, nil)

	_, err := service.SubmitTest(context.Background(), 42, 1, []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1001}, {QuestionID: 102, AnswerID: 1004}, {QuestionID: 103, AnswerID: 1006}})
	if !errors.Is(err, ErrInvalidTestAnswers) { t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers) }
}

func TestCoursesSubmitTestEmptyTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)
	test := &domain.Test{ID: 1, PageID: 10, PassingScore: 70}

	repo.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(1)).Return(test, nil)
	repo.EXPECT().FindLatestTestAttempt(gomock.Any(), int64(42), int64(1)).Return(nil, nil)

	_, err := service.SubmitTest(context.Background(), 42, 1, []domain.TestAttemptAnswer{{QuestionID: 101, AnswerID: 1002}})
	if !errors.Is(err, ErrInvalidTestAnswers) { t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers) }
}
