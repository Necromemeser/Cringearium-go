package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func RoundPurpose(round AdaptiveRound) string {
	review, core := false, false
	for _, target := range round.LearningPlan {
		if target.Reason == "prerequisite" {
			review = true
		} else {
			core = true
		}
	}
	if review && core {
		return "mixed"
	} // Previously saved sessions remain honestly labelled.
	if review {
		return "prerequisite_review"
	}
	return "practice"
}
func RoundScore(round AdaptiveRound) (correct, answered int) {
	questions := map[int64]bool{}
	for _, q := range round.Questions {
		questions[q.ID] = true
	}
	seen := map[int64]bool{}
	for _, a := range round.Answers {
		if !questions[a.QuestionID] || seen[a.QuestionID] {
			continue
		}
		seen[a.QuestionID] = true
		answered++
		if a.IsCorrect {
			correct++
		}
	}
	return
}
func RussianText(text string) bool {
	for _, r := range text {
		if unicode.In(r, unicode.Cyrillic) {
			return true
		}
	}
	return false
}
func validStudentFeedback(f AdaptiveFeedback) bool {
	if !RussianText(f.Summary) {
		return false
	}
	for _, group := range [][]TopicFeedback{f.MasteredTopics, f.TopicsToReview} {
		for _, t := range group {
			if !RussianText(t.Title) || !RussianText(t.Reason) {
				return false
			}
		}
	}
	for _, step := range f.NextSteps {
		if !RussianText(step.Title) || !RussianText(step.Description) {
			return false
		}
	}
	return true
}

// LocalizedFeedback also repairs presentation of historical English or incomplete results.
// It never translates invented model claims: the fallback uses only recorded answers.
func LocalizedFeedback(session AdaptiveSession) AdaptiveFeedback {
	if session.Feedback != nil && validStudentFeedback(*session.Feedback) {
		feedback := *session.Feedback
		if feedback.MasteredTopics == nil {
			feedback.MasteredTopics = []TopicFeedback{}
		}
		if feedback.TopicsToReview == nil {
			feedback.TopicsToReview = []TopicFeedback{}
		}
		if feedback.NextSteps == nil {
			feedback.NextSteps = []NextStep{}
		}
		return feedback
	}
	result := AdaptiveFeedback{SessionID: session.ID, Summary: "Практика завершена.", CreatedAt: time.Now(), MasteredTopics: []TopicFeedback{}, TopicsToReview: []TopicFeedback{}, NextSteps: []NextStep{}}
	type topicScore struct {
		title          string
		id             *int64
		correct, total int
	}
	var topics []topicScore
	positions := map[string]int{}
	total, correct := 0, 0
	for _, round := range session.Rounds {
		c, n := RoundScore(round)
		total += n
		correct += c
		answers := map[int64]AdaptiveAnswer{}
		for _, a := range round.Answers {
			answers[a.QuestionID] = a
		}
		for _, q := range round.Questions {
			a, ok := answers[q.ID]
			if !ok {
				continue
			}
			title := strings.TrimSpace(q.TopicTitle)
			if !RussianText(title) {
				title = "Тема практики"
			}
			key := title
			if q.TopicPageID != nil {
				key = fmt.Sprint(*q.TopicPageID)
			}
			index, ok := positions[key]
			if !ok {
				index = len(topics)
				positions[key] = index
				topics = append(topics, topicScore{title: title, id: q.TopicPageID})
			}
			topics[index].total++
			if a.IsCorrect {
				topics[index].correct++
			}
		}
	}
	if total > 0 {
		result.Summary = fmt.Sprintf("Практика завершена. Верных ответов: %d из %d. Результаты каждого раунда приведены ниже.", correct, total)
	}
	for _, t := range topics {
		reason := fmt.Sprintf("Верных ответов: %d из %d.", t.correct, t.total)
		item := TopicFeedback{Title: t.title, TopicPageID: t.id, Reason: reason}
		if t.correct == t.total {
			result.MasteredTopics = append(result.MasteredTopics, item)
		} else {
			item.Reason += " Стоит ещё потренироваться."
			result.TopicsToReview = append(result.TopicsToReview, item)
			result.NextSteps = append(result.NextSteps, NextStep{Title: "Повторить: " + t.title, Description: "Вернись к материалу по этой теме, разбери пример и попробуй решить новое задание самостоятельно.", TopicPageID: t.id})
		}
	}
	return result
}
