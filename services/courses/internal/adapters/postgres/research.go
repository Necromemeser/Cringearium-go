package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/ports"
)

const studyColumns = `id,title,design,scope,control_course_id,adaptive_course_id,pre_hash,post_hash,content_hash,protocol_notes,export_salt,created_at`

func (db *DB) ListResearchStudies(ctx context.Context) ([]domain.ResearchStudy, error) {
	result := []domain.ResearchStudy{}
	err := db.conn.SelectContext(ctx, &result, `SELECT `+studyColumns+` FROM research_studies ORDER BY created_at DESC,id DESC`)
	return result, err
}
func (db *DB) GetResearchStudy(ctx context.Context, id int64) (domain.ResearchStudy, error) {
	var result domain.ResearchStudy
	err := db.conn.GetContext(ctx, &result, `SELECT `+studyColumns+` FROM research_studies WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		err = ports.ErrResearchNotFound
	}
	return result, err
}
func (db *DB) CreateResearchStudy(ctx context.Context, study domain.ResearchStudy) (domain.ResearchStudy, error) {
	err := db.conn.GetContext(ctx, &study, `INSERT INTO research_studies(title,design,scope,control_course_id,adaptive_course_id,pre_hash,post_hash,content_hash,protocol_notes,export_salt) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING `+studyColumns, study.Title, study.Design, study.Scope, study.ControlCourseID, study.AdaptiveCourseID, study.PreHash, study.PostHash, study.ContentHash, study.ProtocolNotes, study.ExportSalt)
	return study, err
}
func (db *DB) ResearchInstruments(ctx context.Context, courseID int64, scope string) ([]domain.ResearchInstrument, error) {
	var rows []struct {
		ID   int64
		Kind string
	}
	if err := db.conn.SelectContext(ctx, &rows, `SELECT t.id,d.kind FROM assessment_definitions d JOIN tests t ON t.id=d.test_id JOIN course_pages p ON p.id=t.page_id JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=$1 AND d.scope=$2 ORDER BY d.kind,t.id`, courseID, scope); err != nil {
		return nil, err
	}
	result := []domain.ResearchInstrument{}
	for _, row := range rows {
		test, err := db.findTest(ctx, `SELECT t.id,t.page_id,t.passing_score,p.title,d.kind,d.scope FROM tests t JOIN course_pages p ON p.id=t.page_id JOIN assessment_definitions d ON d.test_id=t.id WHERE t.id=$1`, row.ID)
		if err != nil {
			return nil, err
		}
		if test != nil {
			result = append(result, domain.ResearchInstrument{Kind: row.Kind, Scope: scope, Test: *test})
		}
	}
	return result, nil
}
func (db *DB) ResearchContentHash(ctx context.Context, courseID int64) (string, error) {
	var pages []struct {
		Title   string
		Content string
	}
	if err := db.conn.SelectContext(ctx, &pages, `SELECT p.title,coalesce(p.content,'') AS content FROM course_pages p JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=$1 AND p.type='theory' ORDER BY p.title,p.content`, courseID); err != nil {
		return "", err
	}
	if len(pages) == 0 {
		return "", ports.ErrResearchInvalid
	}
	raw, _ := json.Marshal(pages)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
func (db *DB) AssignResearchParticipants(ctx context.Context, id int64, allocations []ports.ResearchAllocation) error {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var study domain.ResearchStudy
	if err := tx.GetContext(ctx, &study, `SELECT `+studyColumns+` FROM research_studies WHERE id=$1 FOR UPDATE`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.ErrResearchNotFound
		}
		return err
	}
	// Stable locking order protects parallel batches and preserves immutable assignment.
	sort.Slice(allocations, func(i, j int) bool { return allocations[i].UserID < allocations[j].UserID })
	var control, adaptive int
	if err := tx.GetContext(ctx, &control, `SELECT count(*) FROM research_participants WHERE study_id=$1 AND arm='control'`, id); err != nil {
		return err
	}
	if err := tx.GetContext(ctx, &adaptive, `SELECT count(*) FROM research_participants WHERE study_id=$1 AND arm='adaptive'`, id); err != nil {
		return err
	}
	for _, allocation := range allocations {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('research-allocation:'||$1::bigint::text,0))`, allocation.UserID); err != nil {
			return err
		}
		var exists bool
		if err := tx.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM research_participants WHERE study_id=$1 AND user_id=$2)`, id, allocation.UserID); err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: участник уже распределён; группа неизменна", ports.ErrResearchInvalid)
		}
		arm := allocation.Arm
		if study.Design == "randomized" {
			// Prospective allocation must precede access and measurement.
			if err := tx.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM course_access WHERE user_id=$1 AND course_id=ANY($2::bigint[]))`, allocation.UserID, []int64{study.ControlCourseID, study.AdaptiveCourseID}); err != nil {
				return err
			}
			if !exists {
				if err := tx.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM assessment_results WHERE user_id=$1 AND course_id=ANY($2::bigint[])) OR EXISTS(SELECT 1 FROM page_progress pp JOIN course_pages p ON p.id=pp.page_id JOIN course_sections s ON s.id=p.section_id WHERE pp.user_id=$1 AND s.course_id=ANY($2::bigint[]))`, allocation.UserID, []int64{study.ControlCourseID, study.AdaptiveCourseID}); err != nil {
					return err
				}
			}
			if exists {
				return fmt.Errorf("%w: для случайного распределения нужны участники без прежнего доступа к обоим вариантам курса", ports.ErrResearchInvalid)
			}
			arm = domain.AllocateResearchArm(control, adaptive, allocation.TieBit)
		}
		if arm != "control" && arm != "adaptive" {
			return ports.ErrResearchInvalid
		}
		courseID := study.ControlCourseID
		if arm == "adaptive" {
			courseID = study.AdaptiveCourseID
			adaptive++
		} else {
			control++
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO research_participants(study_id,user_id,arm) VALUES($1,$2,$3)`, id, allocation.UserID, arm); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO course_access(user_id,course_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, allocation.UserID, courseID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (db *DB) ResearchDataset(ctx context.Context, id int64) (domain.ResearchDataset, error) {
	study, err := db.GetResearchStudy(ctx, id)
	if err != nil {
		return domain.ResearchDataset{}, err
	}
	data := domain.ResearchDataset{Study: study, Members: []domain.ResearchMember{}, Assessments: []domain.AssessmentResult{}, AIEvents: []domain.AIResearchEvent{}}
	rows, err := db.conn.QueryxContext(ctx, `SELECT m.user_id,m.arm,m.assigned_at,
 (SELECT count(*) FROM page_progress pp JOIN course_pages p ON p.id=pp.page_id JOIN course_sections s ON s.id=p.section_id WHERE pp.user_id=m.user_id AND s.course_id=CASE WHEN m.arm='control' THEN $2::bigint ELSE $3::bigint END),
 (SELECT count(*) FROM course_pages p JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=CASE WHEN m.arm='control' THEN $2::bigint ELSE $3::bigint END),
 EXISTS(SELECT 1 FROM course_access ca WHERE ca.user_id=m.user_id AND ca.course_id=CASE WHEN m.arm='control' THEN $3::bigint ELSE $2::bigint END)
 FROM research_participants m WHERE m.study_id=$1 ORDER BY m.user_id`, id, study.ControlCourseID, study.AdaptiveCourseID)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var member domain.ResearchMember
		if err := rows.Scan(&member.UserID, &member.Arm, &member.AssignedAt, &member.CompletedPages, &member.TotalPages, &member.OtherArmAccess); err != nil {
			rows.Close()
			return data, err
		}
		data.Members = append(data.Members, member)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return data, err
	}
	rows, err = db.conn.QueryxContext(ctx, `SELECT ar.id,ar.user_id,ar.course_id,c.title,ar.assessment_type,ar.assessment_scope,ar.score,ar.answers,ar.completed_at,ar.form_hash,ar.instrument_snapshot
 FROM assessment_results ar JOIN courses c ON c.id=ar.course_id JOIN research_participants m ON m.user_id=ar.user_id AND m.study_id=$1
 WHERE ar.course_id=ANY($2::bigint[]) ORDER BY ar.completed_at,ar.id`, id, []int64{study.ControlCourseID, study.AdaptiveCourseID})
	if err != nil {
		return data, err
	}
	defer rows.Close()
	for rows.Next() {
		var r domain.AssessmentResult
		if err := rows.Scan(&r.ID, &r.UserID, &r.CourseID, &r.CourseTitle, &r.Type, &r.Scope, &r.Score, &r.Answers, &r.CompletedAt, &r.FormHash, &r.InstrumentSnapshot); err != nil {
			return data, err
		}
		data.Assessments = append(data.Assessments, r)
	}
	return data, rows.Err()
}

var _ ports.ResearchRepository = (*DB)(nil)

func (db *DB) ResearchAIPageCount(ctx context.Context, courseID int64) (int, error) {
	var count int
	err := db.conn.GetContext(ctx, &count, `SELECT count(*) FROM course_pages p JOIN course_sections s ON s.id=p.section_id WHERE s.course_id=$1 AND p.type='ai_test'`, courseID)
	return count, err
}
