package game

import "errors"

const (
	Rows    = 6
	Columns = 7
)

var (
	errColumnFull = errors.New("column is full")
	errBadColumn  = errors.New("invalid column")
)

type Board struct {
	Cells [Rows][Columns]int `json:"cells"`
}

func NewBoard() Board {
	return Board{}
}

// Drop places a disc for the player (1 or 2) in the given column. Returns the row index used.
func (b *Board) Drop(col int, player int) (int, error) {
	if col < 0 || col >= Columns {
		return -1, errBadColumn
	}
	for row := Rows - 1; row >= 0; row-- {
		if b.Cells[row][col] == 0 {
			b.Cells[row][col] = player
			return row, nil
		}
	}
	return -1, errColumnFull
}

func (b *Board) IsFull() bool {
	for c := 0; c < Columns; c++ {
		if b.Cells[0][c] == 0 {
			return false
		}
	}
	return true
}

// Winner returns the winning player number (1 or 2), or 0 if no winner.
func (b *Board) Winner() int {
	directions := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	for r := 0; r < Rows; r++ {
		for c := 0; c < Columns; c++ {
			player := b.Cells[r][c]
			if player == 0 {
				continue
			}
			for _, d := range directions {
				if b.streak(r, c, d[0], d[1], player) {
					return player
				}
			}
		}
	}
	return 0
}

func (b *Board) streak(r, c, dr, dc, player int) bool {
	for i := 1; i < 4; i++ {
		rr := r + dr*i
		cc := c + dc*i
		if rr < 0 || rr >= Rows || cc < 0 || cc >= Columns {
			return false
		}
		if b.Cells[rr][cc] != player {
			return false
		}
	}
	return true
}
