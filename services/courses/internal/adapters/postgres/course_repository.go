package postgres

import (
	"context"
	"database/sql"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func (db *DB) FindAll(ctx context.Context) ([]*domain.Course, error) {
	const query = `
		SELECT id, title, COALESCE(theme, ''), COALESCE(description, ''), price,
			image_id, author_id, status, created_at, updated_at
		FROM courses
		WHERE status = 'published'
		ORDER BY id
	`

	rows, err := db.conn.QueryxContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	courses := make([]*domain.Course, 0)
	for rows.Next() {
		course := new(domain.Course)
		if err := rows.Scan(&course.ID, &course.Title, &course.Theme, &course.Description, &course.Price,
			&course.ImageID, &course.AuthorID, &course.Status, &course.CreatedAt, &course.UpdatedAt); err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return courses, nil
}

func (db *DB) FindByID(ctx context.Context, id int64) (*domain.CourseDetails, error) {
	course := new(domain.Course)
	const courseQuery = `
		SELECT id, title, COALESCE(theme, ''), COALESCE(description, ''), price,
			image_id, author_id, status, created_at, updated_at
		FROM courses WHERE id = $1 AND status = 'published'
	`
	if err := db.conn.QueryRowxContext(ctx, courseQuery, id).Scan(
		&course.ID, &course.Title, &course.Theme, &course.Description, &course.Price,
		&course.ImageID, &course.AuthorID, &course.Status, &course.CreatedAt, &course.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}

	sections := make([]domain.SectionDetails, 0)
	const sectionsQuery = `
		SELECT id, course_id, title, COALESCE(description, ''), position, created_at, updated_at
		FROM course_sections WHERE course_id = $1 ORDER BY position
	`
	rows, err := db.conn.QueryxContext(ctx, sectionsQuery, id)
	if err != nil { return nil, err }
	for rows.Next() {
		section := domain.SectionDetails{}
		if err := rows.Scan(&section.Section.ID, &section.Section.CourseID, &section.Section.Title,
			&section.Section.Description, &section.Section.Position, &section.Section.CreatedAt, &section.Section.UpdatedAt); err != nil {
			rows.Close(); return nil, err
		}
		sections = append(sections, section)
	}
	if err := rows.Err(); err != nil { rows.Close(); return nil, err }
	rows.Close()

	const pagesQuery = `
		SELECT id, section_id, title, type, COALESCE(content, ''), position, created_at, updated_at
		FROM course_pages WHERE section_id = $1 ORDER BY position
	`
	for i := range sections {
		rows, err := db.conn.QueryxContext(ctx, pagesQuery, sections[i].Section.ID)
		if err != nil { return nil, err }
		for rows.Next() {
			page := domain.Page{}
			if err := rows.Scan(&page.ID, &page.SectionID, &page.Title, &page.Type, &page.Content,
				&page.Position, &page.CreatedAt, &page.UpdatedAt); err != nil {
				rows.Close(); return nil, err
			}
			sections[i].Pages = append(sections[i].Pages, page)
		}
		if err := rows.Err(); err != nil { rows.Close(); return nil, err }
		rows.Close()
	}

	return &domain.CourseDetails{Course: *course, Sections: sections}, nil
}

func (db *DB) FindEnrolled(ctx context.Context, userID int64) ([]*domain.Course, error) {
	const query = `
		SELECT c.id, c.title, COALESCE(c.theme, ''), COALESCE(c.description, ''), c.price,
			c.image_id, c.author_id, c.status, c.created_at, c.updated_at
		FROM courses c
		JOIN course_access ca ON ca.course_id = c.id
		WHERE ca.user_id = $1 AND c.status = 'published'
		ORDER BY ca.granted_at DESC
	`
	rows, err := db.conn.QueryxContext(ctx, query, userID)
	if err != nil { return nil, err }
	defer rows.Close()

	courses := make([]*domain.Course, 0)
	for rows.Next() {
		course := new(domain.Course)
		if err := rows.Scan(&course.ID, &course.Title, &course.Theme, &course.Description, &course.Price,
			&course.ImageID, &course.AuthorID, &course.Status, &course.CreatedAt, &course.UpdatedAt); err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return courses, nil
}

func (db *DB) HasAccess(ctx context.Context, userID, courseID int64) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM course_access WHERE user_id = $1 AND course_id = $2)`
	var exists bool
	if err := db.conn.GetContext(ctx, &exists, query, userID, courseID); err != nil { return false, err }
	return exists, nil
}

func (db *DB) GrantAccess(ctx context.Context, userID, courseID int64) error {
	const query = `INSERT INTO course_access (user_id, course_id) VALUES ($1, $2) ON CONFLICT (user_id, course_id) DO NOTHING`
	_, err := db.conn.ExecContext(ctx, query, userID, courseID)
	return err
}

func (db *DB) GetCompletedPages(ctx context.Context, userID, courseID int64) ([]int64, error) {
	const query = `
		SELECT pp.page_id
		FROM page_progress pp
		JOIN course_pages p ON p.id = pp.page_id
		JOIN course_sections s ON s.id = p.section_id
		WHERE pp.user_id = $1 AND s.course_id = $2
		ORDER BY p.position
	`
	var pageIDs []int64
	if err := db.conn.SelectContext(ctx, &pageIDs, query, userID, courseID); err != nil { return nil, err }
	return pageIDs, nil
}

func (db *DB) CompletePage(ctx context.Context, userID, pageID int64) error {
	const query = `INSERT INTO page_progress (user_id, page_id) VALUES ($1, $2) ON CONFLICT (user_id, page_id) DO NOTHING`
	_, err := db.conn.ExecContext(ctx, query, userID, pageID)
	return err
}
