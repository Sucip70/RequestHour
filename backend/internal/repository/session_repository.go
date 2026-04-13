package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
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
	_, err := r.pool.Exec(ctx, `insert into tr_session (session) values ($1)`, session)
	return err
}

// GetSessionGames loads the integer array games for a session token.
// If no row exists, found is false and games is an empty slice.
func (r *SessionRepository) GetSessionGames(ctx context.Context, session string) (found bool, games []int, err error) {
	var raw []int32
	err = r.pool.QueryRow(ctx,
		`select coalesce(games, '{}') from tr_session where session = $1`,
		session,
	).Scan(&raw)
	if err == pgx.ErrNoRows {
		return false, []int{}, nil
	}
	if err != nil {
		return false, nil, err
	}
	games = make([]int, len(raw))
	for i, v := range raw {
		games[i] = int(v)
	}
	return true, games, nil
}
