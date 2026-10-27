package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/ports"
)

type Research struct {
	repo ports.ResearchRepository
	ai   ports.ResearchAIClient
}

func NewResearch(repo ports.ResearchRepository, ai ports.ResearchAIClient) *Research {
	return &Research{repo, ai}
}
func (s *Research) List(ctx context.Context) ([]domain.ResearchStudy, error) {
	return s.repo.ListResearchStudies(ctx)
}
func (s *Research) Create(ctx context.Context, study domain.ResearchStudy) (domain.ResearchStudy, error) {
	study.Title = strings.TrimSpace(study.Title)
	if study.Title == "" || len(study.Title) > 300 || len(study.ProtocolNotes) > 10000 || study.ControlCourseID <= 0 || study.AdaptiveCourseID <= 0 || study.ControlCourseID == study.AdaptiveCourseID || (study.Design != "randomized" && study.Design != "observational") || (study.Scope != "course" && study.Scope != "trigonometry") {
		return domain.ResearchStudy{}, ports.ErrResearchInvalid
	}
	hashes := map[string]string{}
	for _, courseID := range []int64{study.ControlCourseID, study.AdaptiveCourseID} {
		aiPages, err := s.repo.ResearchAIPageCount(ctx, courseID)
		if err != nil {
			return domain.ResearchStudy{}, err
		}
		if (courseID == study.ControlCourseID && aiPages != 0) || (courseID == study.AdaptiveCourseID && aiPages == 0) {
			return domain.ResearchStudy{}, fmt.Errorf("%w: контрольный курс должен быть без AI-страниц, экспериментальный — с ними", ports.ErrResearchInvalid)
		}
		instruments, err := s.repo.ResearchInstruments(ctx, courseID, study.Scope)
		if err != nil {
			return domain.ResearchStudy{}, err
		}
		if len(instruments) != 2 {
			return domain.ResearchStudy{}, fmt.Errorf("%w: нужны ровно один входной и один итоговый тест выбранного объёма", ports.ErrResearchInvalid)
		}
		found := map[string]bool{}
		for _, instrument := range instruments {
			if found[instrument.Kind] || len(instrument.Test.Questions) == 0 {
				return domain.ResearchStudy{}, ports.ErrResearchInvalid
			}
			found[instrument.Kind] = true
			hash, _ := domain.InstrumentSnapshot(&instrument.Test)
			if previous, exists := hashes[instrument.Kind]; exists && previous != hash {
				return domain.ResearchStudy{}, fmt.Errorf("%w: контрольные тесты групп должны совпадать", ports.ErrResearchInvalid)
			}
			hashes[instrument.Kind] = hash
		}
		if !found["pretest"] || !found["posttest"] {
			return domain.ResearchStudy{}, ports.ErrResearchInvalid
		}
		content, err := s.repo.ResearchContentHash(ctx, courseID)
		if err != nil {
			return domain.ResearchStudy{}, err
		}
		if study.ContentHash != "" && study.ContentHash != content {
			return domain.ResearchStudy{}, fmt.Errorf("%w: теоретические материалы групп различаются", ports.ErrResearchInvalid)
		}
		study.ContentHash = content
	}
	study.PreHash, study.PostHash = hashes["pretest"], hashes["posttest"]
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return domain.ResearchStudy{}, err
	}
	study.ExportSalt = hex.EncodeToString(salt)
	return s.repo.CreateResearchStudy(ctx, study)
}
func (s *Research) Assign(ctx context.Context, id int64, userIDs []int64, arm string) error {
	study, err := s.repo.GetResearchStudy(ctx, id)
	if err != nil {
		return err
	}
	if len(userIDs) == 0 || len(userIDs) > 500 || (study.Design == "observational" && arm != "control" && arm != "adaptive") {
		return ports.ErrResearchInvalid
	}
	seen := map[int64]bool{}
	allocations := make([]ports.ResearchAllocation, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 || seen[userID] {
			return ports.ErrResearchInvalid
		}
		seen[userID] = true
		var bit [1]byte
		if _, err := rand.Read(bit[:]); err != nil {
			return err
		}
		allocations = append(allocations, ports.ResearchAllocation{UserID: userID, Arm: arm, TieBit: bit[0]&1 == 1})
	}
	return s.repo.AssignResearchParticipants(ctx, id, allocations)
}
func (s *Research) Dataset(ctx context.Context, id int64) (domain.ResearchDataset, error) {
	data, err := s.repo.ResearchDataset(ctx, id)
	if err != nil {
		return data, err
	}
	// Form versions are frozen per submission; content drift is reported, never hidden.
	for _, courseID := range []int64{data.Study.ControlCourseID, data.Study.AdaptiveCourseID} {
		current, err := s.repo.ResearchContentHash(ctx, courseID)
		if err != nil {
			return data, err
		}
		if current != data.Study.ContentHash {
			data.Warnings = append(data.Warnings, "content_changed_since_protocol")
		}
	}
	if s.ai != nil {
		events, err := s.ai.Events(ctx, []int64{data.Study.ControlCourseID, data.Study.AdaptiveCourseID})
		if err == nil {
			data.AIAvailable = true
			members := map[int64]bool{}
			for _, m := range data.Members {
				members[m.UserID] = true
			}
			for _, event := range events {
				if members[event.UserID] {
					data.AIEvents = append(data.AIEvents, event)
				}
			}
		}
	}
	return data, nil
}
func (s *Research) Report(ctx context.Context, id int64, includeUnversioned bool) (domain.ResearchReport, error) {
	data, err := s.Dataset(ctx, id)
	if err != nil {
		return domain.ResearchReport{}, err
	}
	return domain.BuildResearchReport(data, includeUnversioned, time.Now().UTC()), nil
}

var _ ports.ResearchUseCases = (*Research)(nil)
