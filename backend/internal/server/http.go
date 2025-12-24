package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/rishirajmaheshwari/4-in-a-row/internal/analytics"
	"github.com/rishirajmaheshwari/4-in-a-row/internal/config"
	"github.com/rishirajmaheshwari/4-in-a-row/internal/game"
	"github.com/rishirajmaheshwari/4-in-a-row/internal/storage"
)

type Server struct {
	cfg       config.Config
	manager   *game.Manager
	repo      *storage.Repository
	producer  *analytics.Producer
	upgrader  websocket.Upgrader
	clients   map[string]map[string]*wsClient // gameID -> username -> client
	mu        sync.Mutex
}

type wsClient struct {
	username string
	conn     *websocket.Conn
	game     *game.Game
	player   int
}

func New(cfg config.Config, manager *game.Manager, repo *storage.Repository, producer *analytics.Producer) *Server {
	return &Server{
		cfg:      cfg,
		manager:  manager,
		repo:     repo,
		producer: producer,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, allowed := range cfg.AllowedOrigins {
					if origin == allowed {
						return true
					}
				}
				return origin == ""
			},
		},
		clients: make(map[string]map[string]*wsClient),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/leaderboard", s.handleLeaderboard)
	mux.HandleFunc("/ws", s.handleWS)
	return mux
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.repo.Leaderboard(ctx, 20)
	if err != nil {
		log.Printf("leaderboard error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}

	botInfo := game.PlayerInfo{Username: "bot", IsBot: true}
	g, playerIdx, existing := s.manager.WaitForMatch(username, time.Duration(s.cfg.BotWaitSeconds)*time.Second, botInfo)
	client := &wsClient{username: username, conn: conn, game: g, player: playerIdx}
	s.registerClient(client)

	// Send initial state to the joining client (with reconnect flag), then broadcast to all with correct turn flags
	_ = conn.WriteJSON(game.ServerMessage{Type: "state", GameID: g.ID, State: g, YourTurn: g.Turn == playerIdx, Opponent: opponentName(g, username), Reconnect: existing})
	s.broadcastState(g, "")

	s.produceEvent(context.Background(), "joined", g.ID, map[string]string{"player": username})

	go s.readLoop(client)
}

func (s *Server) readLoop(c *wsClient) {
	defer s.unregisterClient(c)

	reconnectDeadline := time.After(time.Duration(s.cfg.ReconnectSeconds) * time.Second)
	for {
		var msg game.ClientMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			log.Printf("read error: %v", err)
			return
		}

		switch msg.Type {
		case "move":
			board, winner, err := c.game.ApplyMove(c.username, msg.Column)
			if err != nil {
				_ = c.conn.WriteJSON(game.ServerMessage{Type: "error", Error: err.Error()})
				continue
			}

			state := c.game.Snapshot()
			s.broadcastState(&state, "")
			if winner != 0 || board.IsFull() {
				winnerName := ""
				if winner != 0 {
					winnerName = c.game.Players[winner-1].Username
				}
				s.persistFinish(state, winnerName)
				s.produceEvent(context.Background(), "finished", state.ID, map[string]interface{}{"winner": winnerName, "moves": state.Moves})
			}

			// Bot move when needed
			opp := opponentName(c.game, c.username)
			if opp == "bot" && !c.game.Done && c.game.CurrentPlayer().IsBot {
				s.doBotMove(c.game)
			}

		case "ping":
			_ = c.conn.WriteJSON(game.ServerMessage{Type: "pong"})
		case "reconnect":
			reconnectDeadline = time.After(time.Duration(s.cfg.ReconnectSeconds) * time.Second)
		default:
			_ = c.conn.WriteJSON(game.ServerMessage{Type: "error", Error: "unknown message"})
		}
	}

	<-reconnectDeadline
}

func (s *Server) doBotMove(g *game.Game) {
	bot := game.NewBot(g.PlayerIndex("bot"), g.PlayerIndex(opponentName(g, "bot")), nil)
	col := bot.ChooseMove(g.Board)
	g.ApplyMove("bot", col)
	state := g.Snapshot()
	s.produceEvent(context.Background(), "bot_move", g.ID, map[string]interface{}{"column": col})
	s.broadcastState(&state, "")
}

func (s *Server) persistFinish(state game.Game, winner string) {
	ctx := context.Background()
	movesBytes, _ := json.Marshal(state.Moves)
	rec := storage.FinishedGame{
		ID:         state.ID,
		Player1:    state.Players[0].Username,
		Player2:    state.Players[1].Username,
		Winner:     winner,
		Moves:      movesBytes,
		CreatedAt:  state.CreatedAt,
		FinishedAt: state.UpdatedAt,
	}
	if err := s.repo.SaveFinishedGame(ctx, rec); err != nil {
		log.Printf("persist finish: %v", err)
	}
	s.manager.Finish(state.ID)
}

func (s *Server) produceEvent(ctx context.Context, eventType, gameID string, payload interface{}) {
	if s.producer == nil {
		return
	}
	err := s.producer.Emit(ctx, analytics.Event{Type: eventType, GameID: gameID, Payload: payload, OccurredAt: time.Now()})
	if err != nil {
		log.Printf("analytics emit: %v", err)
	}
}

func opponentName(g *game.Game, username string) string {
	if g.Players[0].Username == username {
		return g.Players[1].Username
	}
	return g.Players[0].Username
}

func (s *Server) registerClient(c *wsClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.clients[c.game.ID]; !ok {
		s.clients[c.game.ID] = make(map[string]*wsClient)
	}
	// Replace previous connection for this username if present
	if existing, ok := s.clients[c.game.ID][c.username]; ok {
		existing.conn.Close()
	}
	s.clients[c.game.ID][c.username] = c
}

func (s *Server) unregisterClient(c *wsClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if gameClients, ok := s.clients[c.game.ID]; ok {
		delete(gameClients, c.username)
		if len(gameClients) == 0 {
			delete(s.clients, c.game.ID)
		}
	}
	c.conn.Close()

	// If opponent remains and player does not reconnect within window, forfeit.
	go s.maybeForfeit(c.game, c.username)
}

func (s *Server) maybeForfeit(g *game.Game, username string) {
	time.Sleep(time.Duration(s.cfg.ReconnectSeconds) * time.Second)
	if s.isConnected(g.ID, username) {
		return
	}
	if g.Done {
		return
	}
	state, opponent := g.Forfeit(username)
	s.persistFinish(state, opponent)
	s.broadcastState(&state, "forfeit")
}

func (s *Server) isConnected(gameID, username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if gameClients, ok := s.clients[gameID]; ok {
		_, present := gameClients[username]
		return present
	}
	return false
}

func (s *Server) broadcastState(state *game.Game, message string) {
	s.mu.Lock()
	clients := s.clients[state.ID]
	s.mu.Unlock()
	current := state.CurrentPlayer().Username
	for _, cl := range clients {
		yourTurn := current == cl.username
		msg := game.ServerMessage{Type: "state", GameID: state.ID, State: state, YourTurn: yourTurn, Opponent: opponentName(state, cl.username), Message: message}
		_ = cl.conn.WriteJSON(msg)
	}
}
