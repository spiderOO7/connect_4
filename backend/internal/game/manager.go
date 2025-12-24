package game

import (
	"sync"
	"time"
)

type matchResult struct {
	game      *Game
	playerIdx int
}

type waitEntry struct {
	username string
	ch       chan matchResult
}

type Manager struct {
	mu        sync.Mutex
	waiting   *waitEntry
	active    map[string]*Game      // gameID -> game
	userGames map[string]string     // username -> gameID
}

func NewManager() *Manager {
	return &Manager{
		active:    make(map[string]*Game),
		userGames: make(map[string]string),
	}
}

// WaitForMatch blocks until a match is ready or timeout triggers a bot game.
// Returns game, playerIdx (1 or 2), and a boolean indicating if the game already existed.
func (m *Manager) WaitForMatch(username string, timeout time.Duration, botInfo PlayerInfo) (*Game, int, bool) {
	// Rejoin existing game if present
	if g, idx, ok := m.findExisting(username); g != nil {
		return g, idx, ok
	}

	m.mu.Lock()
	if m.waiting == nil {
		ch := make(chan matchResult, 1)
		m.waiting = &waitEntry{username: username, ch: ch}
		m.mu.Unlock()

		select {
		case res := <-ch:
			return res.game, res.playerIdx, false
		case <-time.After(timeout):
			g := NewGame(PlayerInfo{Username: username}, botInfo)
			m.registerGame(g)
			return g, playerOne, false
		}
	}

	// If someone is already waiting, match immediately.
	waiting := m.waiting
	if waiting.username == username {
		m.mu.Unlock()
		g, idx, ok := m.findExisting(username)
		return g, idx, ok
	}
	m.waiting = nil
	p1 := PlayerInfo{Username: waiting.username}
	p2 := PlayerInfo{Username: username}
	g := NewGame(p1, p2)
	m.registerGame(g)
	m.mu.Unlock()

	waiting.ch <- matchResult{game: g, playerIdx: playerOne}
	return g, playerTwo, false
}

func (m *Manager) findExisting(username string) (*Game, int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if gameID, ok := m.userGames[username]; ok {
		if g, exists := m.active[gameID]; exists {
			return g, g.PlayerIndex(username), true
		}
	}
	return nil, 0, false
}

func (m *Manager) registerGame(g *Game) {
	m.active[g.ID] = g
	for _, p := range g.Players {
		if p.Username != "" {
			m.userGames[p.Username] = g.ID
		}
	}
}

// Finish cleans up active maps once persisted.
func (m *Manager) Finish(gameID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.active[gameID]
	if !ok {
		return
	}
	delete(m.active, gameID)
	for _, p := range g.Players {
		if p.Username != "" {
			delete(m.userGames, p.Username)
		}
	}
}

// ActiveGame returns a game by ID if present.
func (m *Manager) ActiveGame(gameID string) *Game {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active[gameID]
}
