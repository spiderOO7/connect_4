package game

// Inbound messages from clients.
type ClientMessage struct {
	Type   string `json:"type"`
	Column int    `json:"column,omitempty"`
}

// Outbound events to clients.
type ServerMessage struct {
	Type      string      `json:"type"`
	GameID    string      `json:"gameId,omitempty"`
	State     *Game       `json:"state,omitempty"`
	Error     string      `json:"error,omitempty"`
	YourTurn  bool        `json:"yourTurn,omitempty"`
	Opponent  string      `json:"opponent,omitempty"`
	Reconnect bool        `json:"reconnect,omitempty"`
	Message   string      `json:"message,omitempty"`
}
