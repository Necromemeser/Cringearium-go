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
		ID:           1,
		PageID:       10,
		PassingScore: 70,
		Questions: []domain.TestQuestion{
			{
				ID:       101,
				TestID:   1,
				Question: "2 + 2 = ?",
				Position: 0,
				Answers: []domain.TestAnswer{
					{ID: 1001, QuestionID: 101, Text: "3", Position: 0},
					{ID: 1002, QuestionID: 101, Text: "4", Position: 1, IsCorrect: true},
				},
			},
			{
				ID:       102,
				TestID:   1,
				Question: "3 + 3 = ?",
				Position: 1,
				Answers: []domain.TestAnswer{
					{ID: 1003, QuestionID: 102, Text: "5", Position: 0},
					{ID: 1004, QuestionID: 102, Text: "6", Position: 1, IsCorrect: true},
				},
			},
			{
				ID:       103,
				TestID:   1,
				Question: "4 + 4 = ?",
				Position: 2,
				Answers: []domain.TestAnswer{
					{ID: 1005, QuestionID: 103, Text: "7", Position: 0},
					{ID: 1006, QuestionID: 103, Text: "8", Position: 1, IsCorrect: true},
				},
			},
		},
	}
}

func TestCoursesGetTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()

	repo.EXPECT().
		FindTestByPageID(gomock.Any(), int64(42), int64(10)).
		Return(test, nil)

	got, err := service.GetTest(context.Background(), 42, 10)
	if err != nil {
		t.Fatalf("GetTest() error = %v", err)
	}

	if got != test {
		t.Fatalf("GetTest() = %#v, want %#v", got, test)
	}
}

func TestCoursesGetTestNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	repo.EXPECT().
		FindTestByPageID(gomock.Any(), int64(42), int64(10)).
		Return(nil, nil)

	_, err := service.GetTest(context.Background(), 42, 10)
	if !errors.Is(err, ErrTestNotFound) {
		t.Fatalf("GetTest() error = %v, want %v", err, ErrTestNotFound)
	}
}

func TestCoursesGetTestRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	wantErr := errors.New("repository error")

	repo.EXPECT().
		FindTestByPageID(gomock.Any(), int64(42), int64(10)).
		Return(nil, wantErr)

	_, err := service.GetTest(context.Background(), 42, 10)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetTest() error = %v, want %v", err, wantErr)
	}
}

func TestCoursesSubmitTestPassed(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1002},
		{QuestionID: 102, AnswerID: 1004},
		{QuestionID: 103, AnswerID: 1006},
	}

	expectedResult := &domain.TestResult{
		AttemptID: 15,
		Score:     100,
		Passed:    true,
	}

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	repo.EXPECT().
		SubmitTest(
			gomock.Any(),
			int64(42),
			int64(1),
			answers,
			100,
			true,
		).
		Return(expectedResult, nil)

	got, err := service.SubmitTest(context.Background(), 42, 1, answers)
	if err != nil {
		t.Fatalf("SubmitTest() error = %v", err)
	}

	if got.Score != 100 {
		t.Fatalf("SubmitTest() score = %d, want 100", got.Score)
	}

	if !got.Passed {
		t.Fatal("SubmitTest() passed = false, want true")
	}

	if got.PassingScore != 70 {
		t.Fatalf("SubmitTest() passing score = %d, want 70", got.PassingScore)
	}
}

func TestCoursesSubmitTestFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1002},
		{QuestionID: 102, AnswerID: 1003},
		{QuestionID: 103, AnswerID: 1005},
	}

	expectedResult := &domain.TestResult{
		AttemptID: 16,
		Score:     33,
		Passed:    false,
	}

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	repo.EXPECT().
		SubmitTest(
			gomock.Any(),
			int64(42),
			int64(1),
			answers,
			33,
			false,
		).
		Return(expectedResult, nil)

	got, err := service.SubmitTest(context.Background(), 42, 1, answers)
	if err != nil {
		t.Fatalf("SubmitTest() error = %v", err)
	}

	if got.Score != 33 {
		t.Fatalf("SubmitTest() score = %d, want 33", got.Score)
	}

	if got.Passed {
		t.Fatal("SubmitTest() passed = true, want false")
	}

	if got.PassingScore != 70 {
		t.Fatalf("SubmitTest() passing score = %d, want 70", got.PassingScore)
	}
}

func TestCoursesSubmitTestNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(nil, nil)

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		[]domain.TestAttemptAnswer{
			{QuestionID: 101, AnswerID: 1002},
		},
	)

	if !errors.Is(err, ErrTestNotFound) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrTestNotFound)
	}
}

func TestCoursesSubmitTestRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()
	wantErr := errors.New("repository error")

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1002},
		{QuestionID: 102, AnswerID: 1004},
		{QuestionID: 103, AnswerID: 1006},
	}

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	repo.EXPECT().
		SubmitTest(
			gomock.Any(),
			int64(42),
			int64(1),
			answers,
			100,
			true,
		).
		Return(nil, wantErr)

	_, err := service.SubmitTest(context.Background(), 42, 1, answers)
	if !errors.Is(err, wantErr) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, wantErr)
	}
}

func TestCoursesSubmitTestEmptyAnswers(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		nil,
	)

	if !errors.Is(err, ErrInvalidTestAnswers) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers)
	}
}

func TestCoursesSubmitTestIncompleteAnswers(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1002},
		{QuestionID: 102, AnswerID: 1004},
	}

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		answers,
	)

	if !errors.Is(err, ErrInvalidTestAnswers) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers)
	}
}

func TestCoursesSubmitTestDuplicateQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1002},
		{QuestionID: 101, AnswerID: 1001},
		{QuestionID: 102, AnswerID: 1004},
	}

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		answers,
	)

	if !errors.Is(err, ErrInvalidTestAnswers) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers)
	}
}

func TestCoursesSubmitTestUnknownQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 999, AnswerID: 1002},
		{QuestionID: 102, AnswerID: 1004},
		{QuestionID: 103, AnswerID: 1006},
	}

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		answers,
	)

	if !errors.Is(err, ErrInvalidTestAnswers) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers)
	}
}

func TestCoursesSubmitTestUnknownAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 999},
		{QuestionID: 102, AnswerID: 1004},
		{QuestionID: 103, AnswerID: 1006},
	}

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		answers,
	)

	if !errors.Is(err, ErrInvalidTestAnswers) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers)
	}
}

func TestCoursesSubmitTestQuestionWithoutCorrectAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := testFixture()
	test.Questions[0].Answers[1].IsCorrect = false

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1001},
		{QuestionID: 102, AnswerID: 1004},
		{QuestionID: 103, AnswerID: 1006},
	}

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		answers,
	)

	if !errors.Is(err, ErrInvalidTestAnswers) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers)
	}
}

func TestCoursesSubmitTestEmptyTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repo)

	test := &domain.Test{
		ID:           1,
		PageID:       10,
		PassingScore: 70,
	}

	repo.EXPECT().
		FindTestByID(gomock.Any(), int64(42), int64(1)).
		Return(test, nil)

	answers := []domain.TestAttemptAnswer{
		{QuestionID: 101, AnswerID: 1002},
	}

	_, err := service.SubmitTest(
		context.Background(),
		42,
		1,
		answers,
	)

	if !errors.Is(err, ErrInvalidTestAnswers) {
		t.Fatalf("SubmitTest() error = %v, want %v", err, ErrInvalidTestAnswers)
	}
}
