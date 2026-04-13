package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SongRepository reads gm_songs.
type SongRepository struct {
	pool *pgxpool.Pool
}

func NewSongRepository(pool *pgxpool.Pool) *SongRepository {
	return &SongRepository{pool: pool}
}

// RandomSongIDsExcluding returns up to limit distinct song ids not present in exclude.
func (r *SongRepository) RandomSongIDsExcluding(ctx context.Context, exclude []int, limit int) ([]int, error) {
	if limit < 1 {
		return nil, fmt.Errorf("limit must be positive")
	}
	rows, err := r.pool.Query(ctx,
		`select id from gm_songs where not (id = any($1::int[])) order by random() limit $2`,
		exclude,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, int(id))
	}
	return ids, rows.Err()
}

// GetSongTitleLink returns title and link for a song id.
func (r *SongRepository) GetSongTitleLink(ctx context.Context, id int) (title, link string, err error) {
	err = r.pool.QueryRow(ctx,
		`select title, link from gm_songs where id = $1`,
		id,
	).Scan(&title, &link)
	if err == pgx.ErrNoRows {
		return "", "", fmt.Errorf("get song %d: %w", id, pgx.ErrNoRows)
	}
	return title, link, err
}
