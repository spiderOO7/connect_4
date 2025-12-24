package game

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	playerOne = 1
	playerTwo = 2
)

// PlayerInfo holds lightweight player metadata used across the session.
type PlayerInfo struct {
	Username string `json:"username"`
	IsBot    bool   `json:"isBot"`
}

// Move represents a player move request or broadcast payload.
type Move struct {
	Column int    `json:"column"`
	By     string `json:"by"`
}

// Game models the in-memory game state.
type Game struct {
	ID        string     `json:"id"`
	Board     Board      `json:"board"`
	Players   [2]PlayerInfo `json:"players"`
	Turn      int        `json:"turn"` // 1 or 2
	Winner    int        `json:"winner"` // 0 none
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Done      bool       `json:"done"`
	Moves     []Move     `json:"moves"`
	mu        sync.RWMutex
}

func NewGame(p1 PlayerInfo, p2 PlayerInfo) *Game {
	return &Game{
		ID:        uuid.NewString(),
		Board:     NewBoard(),
		Players:   [2]PlayerInfo{p1, p2},
		Turn:      playerOne,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Moves:     make([]Move, 0, Rows*Columns),
	}
}

// PlayerIndex maps a username to player slot 1 or 2.
func (g *Game) PlayerIndex(username string) int {
	if g.Players[0].Username == username {
		return playerOne
	}
	if g.Players[1].Username == username {
		return playerTwo
	}
	return 0
}

func (g *Game) CurrentPlayer() PlayerInfo {
	return g.Players[g.Turn-1]
}

func (g *Game) ApplyMove(username string, col int) (Board, int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Done {
		return g.Board, g.Winner, nil
	}
	idx := g.PlayerIndex(username)
	if idx == 0 {
		return g.Board, g.Winner, ErrNotYourGame
	}
	if idx != g.Turn {
		return g.Board, g.Winner, ErrNotYourTurn
	}
	if _, err := g.Board.Drop(col, idx); err != nil {
		return g.Board, g.Winner, err
	}
	g.Moves = append(g.Moves, Move{Column: col, By: username})
	winner := g.Board.Winner()
	if winner != 0 {
		g.Winner = winner
		g.Done = true
	} else if g.Board.IsFull() {
		g.Done = true
	}
	if !g.Done {
		if g.Turn == playerOne {
			g.Turn = playerTwo
		} else {
			g.Turn = playerOne
		}
	}
	g.UpdatedAt = time.Now()
	return g.Board, g.Winner, nil
}

// Snapshot returns a copy usable for transport without mutex locking leaks.
func (g *Game) Snapshot() Game {
	g.mu.RLock()
	defer g.mu.RUnlock()
	copyGame := *g
	return copyGame
}

// Forfeit marks the game as finished due to a disconnect timeout.
func (g *Game) Forfeit(loser string) (Game, string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Done {
		return *g, ""
	}
	opponent := opponentOf(g, loser)
	g.Winner = g.PlayerIndex(opponent)
	g.Done = true
	g.UpdatedAt = time.Now()
	return *g, opponent
}

func opponentOf(g *Game, username string) string {
	if g.Players[0].Username == username {
		return g.Players[1].Username
	}
	return g.Players[0].Username
}
