package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

type FinishedGame struct {
	ID         string          `json:"id"`
	Player1    string          `json:"player1"`
	Player2    string          `json:"player2"`
	Winner     string          `json:"winner"`
	Moves      json.RawMessage `json:"moves"`
	CreatedAt  time.Time       `json:"createdAt"`
	FinishedAt time.Time       `json:"finishedAt"`
}

type LeaderboardRow struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
}

func NewRepository(ctx context.Context, url string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	repo := &Repository{pool: pool}
	if err := repo.init(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return repo, nil
}

func (r *Repository) init(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS games (
	id TEXT PRIMARY KEY,
	player1 TEXT,
	player2 TEXT,
	winner TEXT,
	moves JSONB,
	created_at TIMESTAMPTZ,
	finished_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_games_winner ON games(winner);
`)
	return err
}

func (r *Repository) SaveFinishedGame(ctx context.Context, g FinishedGame) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO games (id, player1, player2, winner, moves, created_at, finished_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (id) DO NOTHING;
`, g.ID, g.Player1, g.Player2, g.Winner, g.Moves, g.CreatedAt, g.FinishedAt)
	return err
}

func (r *Repository) Leaderboard(ctx context.Context, limit int) ([]LeaderboardRow, error) {
	rows, err := r.pool.Query(ctx, `
SELECT winner, COUNT(*) AS wins
FROM games
WHERE winner <> '' AND winner IS NOT NULL
GROUP BY winner
ORDER BY wins DESC
LIMIT $1;
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []LeaderboardRow
	for rows.Next() {
		var row LeaderboardRow
		if err := rows.Scan(&row.Username, &row.Wins); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *Repository) Close() {
	r.pool.Close()
}
