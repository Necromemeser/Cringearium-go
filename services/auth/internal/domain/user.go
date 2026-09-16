package domain

type Role string

const (
	RoleStudent Role = "student"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID             int64
	Username       string
	Email          string
	PasswordHash   string
	Role           Role
	ProfileImageID *string
}
