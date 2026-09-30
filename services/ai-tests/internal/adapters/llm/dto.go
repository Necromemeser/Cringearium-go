package llm

type GenerateRoundRequest struct {
	CourseTitle     string
	TopicTitle      string
	Materials       []MaterialContext
	OrdinaryTests   []OrdinaryTestContext
	PreviousResults []PreviousResultContext

	QuestionCount int
	RoundNumber   int
	Strategy      string
	AllowedTopics []AllowedTopic
}

type MaterialContext struct {
	PageID  int64
	Title   string
	Content string
}

type OrdinaryTestContext struct {
	TestTitle string
	Questions []OrdinaryQuestionContext
}

type OrdinaryQuestionContext struct {
	Question       string
	Options        []string
	Correct        bool
	SelectedAnswer string
}

type PreviousResultContext struct {
	TopicTitle   string
	CorrectCount int
	TotalCount   int
}

type AllowedTopic struct {
	PageID int64
	Title  string
}

type GeneratedRound struct {
	Title        string              `json:"title"`
	Instructions string              `json:"instructions"`
	Questions    []GeneratedQuestion `json:"questions"`
}

type GeneratedQuestion struct {
	TopicPageID      *int64            `json:"topic_page_id"`
	TopicTitle       string            `json:"topic_title"`
	Question         string            `json:"question"`
	Difficulty       int               `json:"difficulty"`
	Options          []GeneratedOption `json:"options"`
	CorrectOptionKey string            `json:"correct_option_key"`
	Explanation      string            `json:"explanation"`
	KnowledgeBasis   string            `json:"knowledge_basis"`
	Sources          []GeneratedSource `json:"sources"`
}

type GeneratedOption struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

type GeneratedSource struct {
	Title string `json:"title"`
	URL   string `json:"url,omitempty"`
}
