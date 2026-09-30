package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

type CoursesClient struct {
	baseURL string
	client  *http.Client
}

func NewCoursesClient(baseURL string) *CoursesClient {
	return &CoursesClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *CoursesClient) GetAIContext(
	ctx context.Context,
	userID int64,
	courseID int64,
	topicPageID *int64,
) (ports.CourseContext, error) {
	u, err := url.Parse(c.baseURL + "/internal/courses/" + strconv.FormatInt(courseID, 10) + "/ai-context")
	if err != nil {
		return ports.CourseContext{}, err
	}

	query := u.Query()
	if topicPageID != nil {
		query.Set("topic_page_id", strconv.FormatInt(*topicPageID, 10))
	}
	u.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ports.CourseContext{}, err
	}
	request.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))

	response, err := c.client.Do(request)
	if err != nil {
		return ports.CourseContext{}, fmt.Errorf("courses request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ports.CourseContext{}, fmt.Errorf("courses returned status %d", response.StatusCode)
	}

	var payload aiContextResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return ports.CourseContext{}, fmt.Errorf("decode courses response: %w", err)
	}

	result := ports.CourseContext{
		CourseID:      payload.CourseID,
		CourseTitle:   payload.CourseTitle,
		Materials:     make([]ports.CourseMaterial, 0, len(payload.Materials)),
		OrdinaryTests: make([]ports.OrdinaryTest, 0, len(payload.OrdinaryTests)),
		TestResults:   make([]ports.OrdinaryTestResult, 0, len(payload.TestResults)),
		AllowedTopics: make([]ports.CourseTopic, 0, len(payload.AllowedTopics)),
	}

	for _, material := range payload.Materials {
		result.Materials = append(result.Materials, ports.CourseMaterial{
			PageID: material.PageID, Title: material.Title, Content: material.Content,
		})
	}

	for _, test := range payload.OrdinaryTests {
		item := ports.OrdinaryTest{
			ID: test.ID, PageID: test.PageID, Title: test.Title,
			Questions: make([]ports.OrdinaryTestQuestion, 0, len(test.Questions)),
		}
		for _, question := range test.Questions {
			item.Questions = append(item.Questions, ports.OrdinaryTestQuestion{
				Question: question.Question, Options: question.Options,
			})
		}
		result.OrdinaryTests = append(result.OrdinaryTests, item)
	}

	for _, item := range payload.TestResults {
		result.TestResults = append(result.TestResults, ports.OrdinaryTestResult{
			TestID: item.TestID, CorrectCount: item.CorrectCount, TotalCount: item.TotalCount,
		})
	}

	for _, topic := range payload.AllowedTopics {
		result.AllowedTopics = append(result.AllowedTopics, ports.CourseTopic{
			PageID: topic.PageID, Title: topic.Title,
		})
	}

	return result, nil
}

var _ ports.CourseClient = (*CoursesClient)(nil)
