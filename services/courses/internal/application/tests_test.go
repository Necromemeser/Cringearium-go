package application

import (
	"context"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/mocks"
	"go.uber.org/mock/gomock"
)

func testFixture() *domain.Test {
	return &domain.Test{
		ID: 10,
		PageID: 20,
		PassingScore: 70,
		Questions: []domain.TestQuestion{
			{ID: 1, TestID: 10, Question: "2 + 2", Answers: []domain.TestAnswer{{ID: 11, QuestionID: 1, Text: "4", IsCorrect: true}, {ID: 12, QuestionID: 1, Text: "5"}}},
			{ID: 2, TestID: 10, Question: "3 + 3", Answers: []domain.TestAnswer{{ID: 21, QuestionID: 2, Text: "6", IsCorrect: true}, {ID: 22, QuestionID: 2, Text: "7"}}},
		},
	}
}

func TestSubmitTestPassesAndSavesResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repository)
	testData := testFixture()

	repository.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(10)).Return(testData, nil)
	repository.EXPECT().SubmitTest(gomock.Any(), int64(42), int64(10), []domain.TestAttemptAnswer{{QuestionID: 1, AnswerID: 11}, {QuestionID: 2, AnswerID: 21}}, 100, true).
		Return(&domain.TestResult{AttemptID: 99, Score: 100, Passed: true}, nil)

	result, err := service.SubmitTest(context.Background(), 42, 10, []domain.TestAttemptAnswer{{QuestionID: 1, AnswerID: 11}, {QuestionID: 2, AnswerID: 21}})
	if err != nil { t.Fatalf("SubmitTest returned error: %v", err) }
	if result.AttemptID != 99 || result.Score != 100 || !result.Passed || result.PassingScore != 70 { t.Fatalf("unexpected result: %+v", result) }
}

func TestSubmitTestFailsWithInvalidAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repository)
	testData := testFixture()

	repository.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(10)).Return(testData, nil)

	_, err := service.SubmitTest(context.Background(), 42, 10, []domain.TestAttemptAnswer{{QuestionID: 1, AnswerID: 999}, {QuestionID: 2, AnswerID: 21}})
	if err != ErrInvalidTestAnswers { t.Fatalf("expected ErrInvalidTestAnswers, got %v", err) }
}

func TestSubmitTestFailsWhenQuestionIsMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockCourseRepository(ctrl)
	service := NewCourses(repository)
	testData := testFixture()

	repository.EXPECT().FindTestByID(gomock.Any(), int64(42), int64(10)).Return(testData, nil)

	_, err := service.SubmitTest(context.Background(), 42, 10, []domain.TestAttemptAnswer{{QuestionID: 1, AnswerID: 11}})
	if err != ErrInvalidTestAnswers { t.Fatalf("expected ErrInvalidTestAnswers, got %v", err) }
}
