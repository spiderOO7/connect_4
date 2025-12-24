package game

import "math/rand"

type Bot struct {
	Mark       int
	Opponent   int
	RandomSeed *rand.Rand
}

func NewBot(mark int, opponent int, rng *rand.Rand) *Bot {
	return &Bot{Mark: mark, Opponent: opponent, RandomSeed: rng}
}

// ChooseMove picks a column using simple strategy: win if possible, block opponent, prefer center, otherwise first available.
func (b *Bot) ChooseMove(board Board) int {
	// 1) Can we win now?
	if col, ok := b.findWinningMove(board, b.Mark); ok {
		return col
	}
	// 2) Can we block opponent immediate win?
	if col, ok := b.findWinningMove(board, b.Opponent); ok {
		return col
	}
	// 3) Prefer center column if available
	center := Columns / 2
	if b.canPlay(board, center) {
		return center
	}
	// 4) Prefer columns towards center
	order := []int{3, 2, 4, 1, 5, 0, 6}
	for _, col := range order {
		if b.canPlay(board, col) {
			return col
		}
	}
	// 5) Fallback any available
	for c := 0; c < Columns; c++ {
		if b.canPlay(board, c) {
			return c
		}
	}
	return -1
}

func (b *Bot) canPlay(board Board, col int) bool {
	if col < 0 || col >= Columns {
		return false
	}
	return board.Cells[0][col] == 0
}

func (b *Bot) findWinningMove(board Board, mark int) (int, bool) {
	for c := 0; c < Columns; c++ {
		if !b.canPlay(board, c) {
			continue
		}
		temp := board
		_, _ = temp.Drop(c, mark)
		if temp.Winner() == mark {
			return c, true
		}
	}
	return 0, false
}
