package ports

import (
	"context"
	"errors"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

var ErrResearchInvalid = errors.New("invalid research protocol or participants")
var ErrResearchNotFound = errors.New("research study not found")

type ResearchAllocation struct {
	UserID int64
	Arm    string
	TieBit bool
}
type ResearchRepository interface {
	ListResearchStudies(context.Context) ([]domain.ResearchStudy, error)
	CreateResearchStudy(context.Context, domain.ResearchStudy) (domain.ResearchStudy, error)
	GetResearchStudy(context.Context, int64) (domain.ResearchStudy, error)
	ResearchInstruments(context.Context, int64, string) ([]domain.ResearchInstrument, error)
	ResearchContentHash(context.Context, int64) (string, error)
	ResearchAIPageCount(context.Context, int64) (int, error)
	AssignResearchParticipants(context.Context, int64, []ResearchAllocation) error
	ResearchDataset(context.Context, int64) (domain.ResearchDataset, error)
}
type ResearchAIClient interface {
	Events(context.Context, []int64) ([]domain.AIResearchEvent, error)
}
type ResearchUseCases interface {
	List(context.Context) ([]domain.ResearchStudy, error)
	Create(context.Context, domain.ResearchStudy) (domain.ResearchStudy, error)
	Assign(context.Context, int64, []int64, string) error
	Dataset(context.Context, int64) (domain.ResearchDataset, error)
	Report(context.Context, int64, bool) (domain.ResearchReport, error)
}
