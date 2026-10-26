package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

type ResearchClient struct {
	baseURL string
	client  *http.Client
}

func NewResearchClient(baseURL string) *ResearchClient {
	return &ResearchClient{strings.TrimRight(baseURL, "/"), &http.Client{Timeout: 8 * time.Second}}
}
func (c *ResearchClient) Events(ctx context.Context, courses []int64) ([]domain.AIResearchEvent, error) {
	values := []string{}
	for _, id := range courses {
		values = append(values, strconv.FormatInt(id, 10))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/admin/research-events?courses="+strings.Join(values, ","), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AI research status %d", resp.StatusCode)
	}
	var result struct {
		Events    []domain.AIResearchEvent `json:"events"`
		Truncated bool                     `json:"truncated"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&result); err != nil {
		return nil, err
	}
	if result.Truncated {
		return nil, fmt.Errorf("AI research export exceeds 50000 rows")
	}
	return result.Events, nil
}
