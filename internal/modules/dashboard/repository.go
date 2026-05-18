package dashboard

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetStats() (Stats, error) {
	var stats Stats

	query := `
		SELECT
			(SELECT COUNT(*) FROM members) AS total_members,
			(SELECT COUNT(*) FROM members WHERE status = 'active') AS active_members,
			(SELECT COUNT(*) FROM members WHERE status = 'inactive') AS inactive_members,
			(
				SELECT COUNT(*)
				FROM members
				WHERE created_at >= date_trunc('month', NOW())
			) AS new_this_month
	`

	err := r.db.QueryRow(query).Scan(
		&stats.TotalMembers,
		&stats.ActiveMembers,
		&stats.InactiveMembers,
		&stats.NewThisMonth,
	)

	return stats, err
}
