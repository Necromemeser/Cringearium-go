package postgres

import (
	"context"
	"database/sql"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

func (db *DB) FindAll(ctx context.Context) ([]*domain.Course, error) {
	const query = `
		SELECT
			id,
			title,
			theme,
			description,
			price,
			image_id,
			author_id,
			status,
			created_at,
			updated_at
		FROM courses
		WHERE status = 'published'
		ORDER BY id
	`

	rows, err := db.conn.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]*domain.Course, 0)

	for rows.Next() {
		course := new(domain.Course)
		if err := rows.Scan(
			&course.ID,
			&course.Title,
			&course.Theme,
			&course.Description,
			&course.Price,
			&course.ImageID,
			&course.AuthorID,
			&course.Status,
			&course.CreatedAt,
			&course.UpdatedAt,
		); err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

func (db *DB) FindByID(ctx context.Context, id int64) (*domain.CourseDetails, error) {
	course := new(domain.Course)

	const courseQuery = `
		SELECT
			id,
			title,
			theme,
			description,
			price,
			image_id,
			author_id,
			status,
			created_at,
			updated_at
		FROM courses
		WHERE id = $1 AND status = 'published'
	`

	row := db.conn.QueryRowxContext(ctx, courseQuery, id)
	if err := row.Scan(
		&course.ID,
		&course.Title,
		&course.Theme,
		&course.Description,
		&course.Price,
		&course.ImageID,
		&course.AuthorID,
		&course.Status,
		&course.CreatedAt,
		&course.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	sections := make([]domain.SectionDetails, 0)

	const sectionsQuery = `
		SELECT
			id,
			course_id,
			title,
			description,
			position,
			created_at,
			updated_at
		FROM course_sections
		WHERE course_id = $1
		ORDER BY position
	`

	rows, err := db.conn.QueryxContext(ctx, sectionsQuery, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		section := domain.SectionDetails{}
		if err := rows.Scan(
			&section.Section.ID,
			&section.Section.CourseID,
			&section.Section.Title,
			&section.Section.Description,
			&section.Section.Position,
			&section.Section.CreatedAt,
			&section.Section.UpdatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}

		sections = append(sections, section)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	const pagesQuery = `
		SELECT
			id,
			section_id,
			title,
			type,
			content,
			position,
			created_at,
			updated_at
		FROM course_pages
		WHERE section_id = $1
		ORDER BY position
	`

	for i := range sections {
		rows, err := db.conn.QueryxContext(ctx, pagesQuery, sections[i].Section.ID)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			page := domain.Page{}
			if err := rows.Scan(
				&page.ID,
				&page.SectionID,
				&page.Title,
				&page.Type,
				&page.Content,
				&page.Position,
				&page.CreatedAt,
				&page.UpdatedAt,
			); err != nil {
				rows.Close()
				return nil, err
			}

			sections[i].Pages = append(sections[i].Pages, page)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}

	return &domain.CourseDetails{
		Course:   *course,
		Sections: sections,
	}, nil
}

func (db *DB) HasAccess(ctx context.Context, userID, courseID int64) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM course_access
			WHERE user_id = $1 AND course_id = $2
		)
	`

	var exists bool
	if err := db.conn.GetContext(ctx, &exists, query, userID, courseID); err != nil {
		return false, err
	}

	return exists, nil
}

func (db *DB) GrantAccess(ctx context.Context, userID, courseID int64) error {
	const query = `
		INSERT INTO course_access (user_id, course_id)
		VALUES ($1, $2)
	`

	_, err := db.conn.ExecContext(ctx, query, userID, courseID)
	return err
}
