package domain

import "time"

type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "draft"
	CourseStatusPublished CourseStatus = "published"
	CourseStatusArchived  CourseStatus = "archived"
)

type PageType string

const (
	PageTypeTheory PageType = "theory"
	PageTypeTest   PageType = "test"
	PageTypeAITest PageType = "ai_test"
)

type Course struct {
	ID          int64
	Title       string
	Theme       string
	Description string
	Price       int
	ImageID     *string
	AuthorID    *int64
	Status      CourseStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Section struct {
	ID          int64
	CourseID    int64
	Title       string
	Description string
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Page struct {
	ID        int64
	SectionID int64
	Title     string
	Type      PageType
	Content   string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CourseDetails struct {
	Course   Course
	Sections []SectionDetails
}

type SectionDetails struct {
	Section Section
	Pages   []Page
}
