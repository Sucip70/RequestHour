package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRepository persists session rows in gm_session.
type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

// InsertSession stores a new session token (column session).
func (r *SessionRepository) InsertSession(ctx context.Context, session string) error {
	_, err := r.pool.Exec(ctx, `insert into gm_session (session) values ($1)`, session)
	return err
}
