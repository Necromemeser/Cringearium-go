package domain

type AdminCourse struct {
	ID             int64
	Title          string
	Theme          string
	Description    string
	Price          int
	AuthorID       *int64
	Status         CourseStatus
	EnrolledUsers  int
	SectionCount   int
	PageCount      int
}
