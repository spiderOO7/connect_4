package game

import "errors"

var (
	ErrNotYourTurn = errors.New("not your turn")
	ErrNotYourGame = errors.New("not part of this game")
)
