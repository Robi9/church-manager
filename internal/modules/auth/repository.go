package auth

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(email string) (User, error) {
	var user User

	err := r.db.QueryRow(
		`SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1 AND disabled_at IS NULL`,
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

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (r *Repository) CreateSession(userID int64, digest string, expires time.Time) error {
	result, err := r.db.Exec(`INSERT INTO sessions (token_hash, user_id, expires_at)
 SELECT $1, id, $3 FROM users WHERE id = $2 AND disabled_at IS NULL`, digest, userID, expires)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	_, err = r.db.Exec(`DELETE FROM sessions WHERE expires_at <= NOW()`)
	return err
}

func (r *Repository) ValidateSession(ctx context.Context, digest string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT s.user_id FROM sessions s JOIN users u ON u.id = s.user_id
 WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.disabled_at IS NULL AND s.created_at >= u.updated_at`, digest).Scan(&id)
	return id, err
}

func (r *Repository) RevokeSession(ctx context.Context, digest string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, digest)
	return err
}

func (r *Repository) RevokeAllSessions(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}
