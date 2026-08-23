package member

import (
	"database/sql"
	"errors"
	"hash/fnv"
	"strconv"

	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

type TransactionRepository interface {
	CandidateFinder
	AcquireDuplicateLock(normalizedName string) error
	Create(Member) (Member, error)
	GetByID(id int64) (Member, error)
	Update(Member) (Member, error)
	CreateDuplicateAudits([]DuplicateAudit) error
}

type transactionRepository struct {
	tx *sql.Tx
}

type dbExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) RunInTransaction(fn func(TransactionRepository) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	repo := &transactionRepository{tx: tx}
	if err := fn(repo); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *Repository) Create(m Member) (Member, error) {
	return createMember(r.db, m)
}

func (r *transactionRepository) Create(m Member) (Member, error) {
	return createMember(r.tx, m)
}

func createMember(db dbExecutor, m Member) (Member, error) {
	normalized := NormalizeMember(m)
	query := `
		INSERT INTO members (
			name, email, phone, status,
			member_since, baptized, baptism_date,
			church_role, marital_status, origin_denomination, congregation,
			membership_course_completed, membership_course_completed_at,
			contacted, contacted_at,
			address, address_number, address_complement, neighborhood, city, state,
			created_at, updated_at, created_by,
			normalized_name, normalized_phone, normalized_address,
			normalized_address_number, normalized_neighborhood, normalized_city,
			normalized_congregation
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,
			$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31
		)
		RETURNING id
	`

	err := db.QueryRow(
		query,
		m.Name, m.Email, m.Phone, m.Status,
		m.MemberSince, m.Baptized, m.BaptismDate,
		m.ChurchRole, m.MaritalStatus, m.OriginDenomination, m.Congregation,
		m.MembershipCourseCompleted, m.MembershipCourseCompletedAt,
		m.Contacted, m.ContactedAt,
		m.Address, m.AddressNumber, m.AddressComplement, m.Neighborhood, m.City, m.State,
		m.CreatedAt, m.UpdatedAt, m.CreatedBy,
		normalized.Name, normalized.Phone, normalized.Address,
		normalized.AddressNumber, normalized.Neighborhood, normalized.City,
		normalized.Congregation,
	).Scan(&m.ID)

	return m, err
}

func (r *Repository) Find(filters map[string]string, limit, offset int) ([]Member, error) {
	query := memberSelect + ` WHERE 1=1`
	args := []any{}
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

	query += " ORDER BY id DESC LIMIT $" + strconv.Itoa(argID)
	args = append(args, limit)
	argID++
	query += " OFFSET $" + strconv.Itoa(argID)
	args = append(args, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Member, 0)
	for rows.Next() {
		member, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, member)
	}
	return result, rows.Err()
}

func (r *Repository) Count(filters map[string]string) (int, error) {
	query := `SELECT COUNT(*) FROM members WHERE 1=1`
	args := []any{}
	argID := 1

	if name, ok := filters["name"]; ok && name != "" {
		query += " AND LOWER(name) LIKE '%' || LOWER($" + strconv.Itoa(argID) + ") || '%'"
		args = append(args, name)
		argID++
	}
	if status, ok := filters["status"]; ok && status != "" {
		query += " AND status = $" + strconv.Itoa(argID)
		args = append(args, status)
	}

	var total int
	err := r.db.QueryRow(query, args...).Scan(&total)
	return total, err
}

func (r *Repository) GetByID(id int64) (Member, error) {
	return getMemberByID(r.db, id)
}

func (r *transactionRepository) GetByID(id int64) (Member, error) {
	return scanMember(r.tx.QueryRow(memberSelect+` WHERE id = $1 FOR UPDATE`, id))
}

func getMemberByID(db dbExecutor, id int64) (Member, error) {
	return scanMember(db.QueryRow(memberSelect+` WHERE id = $1`, id))
}

func (r *Repository) Update(m Member) (Member, error) {
	return updateMember(r.db, m)
}

func (r *transactionRepository) Update(m Member) (Member, error) {
	return updateMember(r.tx, m)
}

func updateMember(db dbExecutor, m Member) (Member, error) {
	normalized := NormalizeMember(m)
	query := `
		UPDATE members SET
			name=$1, email=$2, phone=$3, status=$4,
			member_since=$5, baptized=$6, baptism_date=$7,
			church_role=$8, marital_status=$9, origin_denomination=$10,
			congregation=$11, membership_course_completed=$12,
			membership_course_completed_at=$13, contacted=$14, contacted_at=$15,
			address=$16, address_number=$17, address_complement=$18,
			neighborhood=$19, city=$20, state=$21, updated_at=$22,
			normalized_name=$23, normalized_phone=$24, normalized_address=$25,
			normalized_address_number=$26, normalized_neighborhood=$27,
			normalized_city=$28, normalized_congregation=$29
		WHERE id=$30
		RETURNING id
	`

	err := db.QueryRow(
		query,
		m.Name, m.Email, m.Phone, m.Status,
		m.MemberSince, m.Baptized, m.BaptismDate,
		m.ChurchRole, m.MaritalStatus, m.OriginDenomination,
		m.Congregation, m.MembershipCourseCompleted,
		m.MembershipCourseCompletedAt, m.Contacted, m.ContactedAt,
		m.Address, m.AddressNumber, m.AddressComplement,
		m.Neighborhood, m.City, m.State, m.UpdatedAt,
		normalized.Name, normalized.Phone, normalized.Address,
		normalized.AddressNumber, normalized.Neighborhood,
		normalized.City, normalized.Congregation, m.ID,
	).Scan(&m.ID)

	return m, err
}

func (r *Repository) SoftDelete(id int64) error {
	result, err := r.db.Exec(`UPDATE members SET status=$1, updated_at=NOW() WHERE id=$2`, Inactive, id)
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

func (r *Repository) FindDuplicateCandidates(normalized NormalizedMember, excludeID int64) ([]Member, error) {
	return findDuplicateCandidates(r.db, normalized, excludeID)
}

func (r *transactionRepository) FindDuplicateCandidates(normalized NormalizedMember, excludeID int64) ([]Member, error) {
	return findDuplicateCandidates(r.tx, normalized, excludeID)
}

func findDuplicateCandidates(db dbExecutor, normalized NormalizedMember, excludeID int64) ([]Member, error) {
	query := `
		SELECT id, name, phone, congregation, address, address_number, neighborhood, city
		FROM members
		WHERE ($3 = 0 OR id <> $3)
		  AND (
			($1 <> '' AND normalized_name = $1)
			OR ($2 <> '' AND normalized_phone = $2)
		  )
		ORDER BY id DESC
	`
	rows, err := db.Query(query, normalized.Name, normalized.Phone, excludeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Member, 0)
	for rows.Next() {
		var member Member
		if err := rows.Scan(
			&member.ID, &member.Name, &member.Phone, &member.Congregation,
			&member.Address, &member.AddressNumber, &member.Neighborhood, &member.City,
		); err != nil {
			return nil, err
		}
		result = append(result, member)
	}
	return result, rows.Err()
}

func (r *transactionRepository) AcquireDuplicateLock(normalizedName string) error {
	if normalizedName == "" {
		return nil
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(normalizedName))
	_, err := r.tx.Exec(`SELECT pg_advisory_xact_lock($1)`, int64(hasher.Sum64()))
	return err
}

func (r *transactionRepository) CreateDuplicateAudits(audits []DuplicateAudit) error {
	for _, audit := range audits {
		_, err := r.tx.Exec(`
			INSERT INTO member_duplicate_audits (
				member_id, candidate_member_id, confirmed_by, score,
				matched_fields, operation, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,NOW())
		`,
			audit.MemberID, audit.CandidateMemberID, audit.ConfirmedBy,
			audit.Score, pq.Array(audit.MatchedFields), audit.Operation,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

const memberSelect = `
	SELECT id, name, email, phone, status,
		member_since, baptized, baptism_date,
		church_role, marital_status, origin_denomination, congregation,
		membership_course_completed, membership_course_completed_at,
		contacted, contacted_at,
		address, address_number, address_complement, neighborhood, city, state,
		created_at, updated_at, created_by
	FROM members
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMember(row rowScanner) (Member, error) {
	var member Member
	err := row.Scan(
		&member.ID, &member.Name, &member.Email, &member.Phone, &member.Status,
		&member.MemberSince, &member.Baptized, &member.BaptismDate,
		&member.ChurchRole, &member.MaritalStatus, &member.OriginDenomination,
		&member.Congregation, &member.MembershipCourseCompleted,
		&member.MembershipCourseCompletedAt, &member.Contacted, &member.ContactedAt,
		&member.Address, &member.AddressNumber, &member.AddressComplement,
		&member.Neighborhood, &member.City, &member.State,
		&member.CreatedAt, &member.UpdatedAt, &member.CreatedBy,
	)
	return member, err
}
