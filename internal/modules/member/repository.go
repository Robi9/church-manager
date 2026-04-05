package member

import (
	"database/sql"
	"log"
	"strconv"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(m Member) (Member, error) {
	query := `
	INSERT INTO members (
		name, email, phone, status,
		member_since, baptized, baptism_date,
		church_role, marital_status, origin_denomination,
		membership_course_completed, membership_course_completed_at,
		contacted, contacted_at,
		created_at, updated_at, created_by
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	RETURNING id
	`

	err := r.db.QueryRow(
		query,
		m.Name,
		m.Email,
		m.Phone,
		m.Status,
		m.MemberSince,
		m.Baptized,
		m.BaptismDate,
		m.ChurchRole,
		m.MaritalStatus,
		m.OriginDenomination,
		m.MembershipCourseCompleted,
		m.MembershipCourseCompletedAt,
		m.Contacted,
		m.ContactedAt,
		m.CreatedAt,
		m.UpdatedAt,
		m.CreatedBy,
	).Scan(&m.ID)

	return m, err
}

func (r *Repository) Find(filters map[string]string, limit, offset int) ([]Member, error) {
	query := `
	SELECT id, name, email, phone, status, created_at, updated_at
	FROM members
	WHERE 1=1
	`

	args := []interface{}{}
	argID := 1

	// filtro por nome
	if name, ok := filters["name"]; ok && name != "" {
		query += " AND LOWER(name) LIKE '%' || LOWER($" + strconv.Itoa(argID) + ") || '%'"
		args = append(args, name)
		argID++
	}

	// filtro por status
	if status, ok := filters["status"]; ok && status != "" {
		query += " AND status = $" + strconv.Itoa(argID)
		args = append(args, status)
		argID++
	}

	query += " ORDER BY id DESC"

	// limit
	query += " LIMIT $" + strconv.Itoa(argID)
	args = append(args, limit)
	argID++

	// offset
	query += " OFFSET $" + strconv.Itoa(argID)
	args = append(args, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Member

	for rows.Next() {
		var m Member

		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Email,
			&m.Phone,
			&m.Status,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		log.Println("member found:", m.ID)

		result = append(result, m)
	}

	return result, nil
}

func (r *Repository) Count(filters map[string]string) (int, error) {
	query := `SELECT COUNT(*) FROM members WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if name, ok := filters["name"]; ok && name != "" {
		query += " AND LOWER(name) LIKE '%' || LOWER($" + strconv.Itoa(argID) + ") || '%'"
		args = append(args, name)
		argID++
	}

	if status, ok := filters["status"]; ok && status != "" {
		query += " AND status = $" + strconv.Itoa(argID)
		args = append(args, status)
		argID++
	}

	var total int
	err := r.db.QueryRow(query, args...).Scan(&total)
	return total, err
}
