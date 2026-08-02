package member

import (
	"database/sql"
	"errors"
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
		church_role, marital_status, origin_denomination, congregation,
		membership_course_completed, membership_course_completed_at,
		contacted, contacted_at,
		address, address_number, address_complement, neighborhood, city, state,
		created_at, updated_at, created_by
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
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
		m.Congregation,
		m.MembershipCourseCompleted,
		m.MembershipCourseCompletedAt,
		m.Contacted,
		m.ContactedAt,
		m.Address,
		m.AddressNumber,
		m.AddressComplement,
		m.Neighborhood,
		m.City,
		m.State,
		m.CreatedAt,
		m.UpdatedAt,
		m.CreatedBy,
	).Scan(&m.ID)

	return m, err
}

func (r *Repository) Find(filters map[string]string, limit, offset int) ([]Member, error) {
	query := `
	SELECT id, name, email, phone, status,
		member_since, baptized, baptism_date,
		church_role, marital_status, origin_denomination, congregation,
		membership_course_completed, membership_course_completed_at,
		contacted, contacted_at,
		address, address_number, address_complement, neighborhood, city, state,
		created_at, updated_at, created_by
	FROM members
	WHERE 1=1
	`

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

	query += " ORDER BY id DESC"

	query += " LIMIT $" + strconv.Itoa(argID)
	args = append(args, limit)
	argID++

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
			&m.MemberSince,
			&m.Baptized,
			&m.BaptismDate,
			&m.ChurchRole,
			&m.MaritalStatus,
			&m.OriginDenomination,
			&m.Congregation,
			&m.MembershipCourseCompleted,
			&m.MembershipCourseCompletedAt,
			&m.Contacted,
			&m.ContactedAt,
			&m.Address,
			&m.AddressNumber,
			&m.AddressComplement,
			&m.Neighborhood,
			&m.City,
			&m.State,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

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

func (r *Repository) GetByID(id int64) (Member, error) {
	var member Member

	query := `
		SELECT id, name, email, phone, status,
			member_since, baptized, baptism_date,
			church_role, marital_status, origin_denomination, congregation,
			membership_course_completed, membership_course_completed_at,
			contacted, contacted_at,
			address, address_number, address_complement, neighborhood, city, state,
			created_at, updated_at, created_by
		FROM members
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&member.ID,
		&member.Name,
		&member.Email,
		&member.Phone,
		&member.Status,
		&member.MemberSince,
		&member.Baptized,
		&member.BaptismDate,
		&member.ChurchRole,
		&member.MaritalStatus,
		&member.OriginDenomination,
		&member.Congregation,
		&member.MembershipCourseCompleted,
		&member.MembershipCourseCompletedAt,
		&member.Contacted,
		&member.ContactedAt,
		&member.Address,
		&member.AddressNumber,
		&member.AddressComplement,
		&member.Neighborhood,
		&member.City,
		&member.State,
		&member.CreatedAt,
		&member.UpdatedAt,
		&member.CreatedBy,
	)

	return member, err
}

func (r *Repository) Update(m Member) (Member, error) {
	query := `
		UPDATE members
		SET
			name = $1,
			email = $2,
			phone = $3,
			status = $4,
			church_role = $5,
			marital_status = $6,
			origin_denomination = $7,
			congregation = $8,
			address = $9,
			address_number = $10,
			address_complement = $11,
			neighborhood = $12,
			city = $13,
			state = $14,
			updated_at = $15
		WHERE id = $16
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		m.Name,
		m.Email,
		m.Phone,
		m.Status,
		m.ChurchRole,
		m.MaritalStatus,
		m.OriginDenomination,
		m.Congregation,
		m.Address,
		m.AddressNumber,
		m.AddressComplement,
		m.Neighborhood,
		m.City,
		m.State,
		m.UpdatedAt,
		m.ID,
	).Scan(&m.ID)

	return m, err
}

func (r *Repository) SoftDelete(id int64) error {
	query := `
		UPDATE members
		SET
			status = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(query, Inactive, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("member not found")
	}

	return nil
}

func (r *Repository) CreateMany(members []Member) error {
	query := `
		INSERT INTO members (
		name, email, phone, status,
		member_since, baptized, baptism_date,
		church_role, marital_status, origin_denomination, congregation,
		membership_course_completed, membership_course_completed_at,
		contacted, contacted_at,
		address, address_number, address_complement, neighborhood, city, state,
		created_at, updated_at, created_by
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
	RETURNING id
	`

	for _, m := range members {
		_, err := r.db.Exec(
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
			m.Congregation,
			m.MembershipCourseCompleted,
			m.MembershipCourseCompletedAt,
			m.Contacted,
			m.ContactedAt,
			m.Address,
			m.AddressNumber,
			m.AddressComplement,
			m.Neighborhood,
			m.City,
			m.State,
			m.CreatedAt,
			m.UpdatedAt,
			m.CreatedBy,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
