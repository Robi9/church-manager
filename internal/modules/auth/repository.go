package auth

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(email string) (User, error) {
	var user User

	err := r.db.QueryRow(
		`SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func (r *Repository) Create(user User) (User, error) {
	err := r.db.QueryRow(
		`INSERT INTO users (email, password, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		user.Email,
		user.Password,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return user, err
}

func (r *Repository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		email,
	).Scan(&exists)

	return exists, err
}
