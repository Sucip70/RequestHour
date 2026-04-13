package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRepository persists session rows in tr_session.
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

// GetSessionState loads games and current question id for a session.
// If no row exists, found is false.
func (r *SessionRepository) GetSessionState(ctx context.Context, session string) (found bool, games []int, current *int, err error) {
	var raw []int32
	var cur sql.NullInt32
	err = r.pool.QueryRow(ctx,
		`select coalesce(games, '{}'), current from tr_session where session = $1`,
		session,
	).Scan(&raw, &cur)
	if err == pgx.ErrNoRows {
		return false, []int{}, nil, nil
	}
	if err != nil {
		return false, nil, nil, err
	}
	games = make([]int, len(raw))
	for i, v := range raw {
		games[i] = int(v)
	}
	if cur.Valid {
		v := int(cur.Int32)
		current = &v
	}
	return true, games, current, nil
}

// SetCurrent sets tr_session.current to the active quiz song id.
func (r *SessionRepository) SetCurrent(ctx context.Context, session string, songID int) error {
	tag, err := r.pool.Exec(ctx,
		`update tr_session set current = $2 where session = $1`,
		session, songID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// AppendGameClearCurrent appends songID to games and clears current.
func (r *SessionRepository) AppendGameClearCurrent(ctx context.Context, session string, songID int) error {
	tag, err := r.pool.Exec(ctx,
		`update tr_session set games = array_append(coalesce(games, '{}'), $2::int), current = null where session = $1`,
		session, songID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ResetGames clears games and current (wrong answer).
func (r *SessionRepository) ResetGames(ctx context.Context, session string) error {
	tag, err := r.pool.Exec(ctx,
		`update tr_session set games = '{}', current = null where session = $1`,
		session,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
