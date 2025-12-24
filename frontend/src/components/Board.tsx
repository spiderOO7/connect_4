import React from 'react'

type BoardProps = {
  board: number[][]
  onDrop: (col: number) => void
  disabled: boolean
}

export const Board: React.FC<BoardProps> = ({ board, onDrop, disabled }) => {
  return (
    <div className="board-container">
      <div className="board">
        {board.map((row, r) => (
          <div key={r} className="board-row">
            {row.map((cell, c) => (
              <div 
                key={`${r}-${c}`} 
                className={`cell ${disabled ? 'disabled' : ''}`} 
                onClick={() => !disabled && onDrop(c)}
              >
                <div className={`disc ${cell === 1 ? 'p1' : cell === 2 ? 'p2' : ''}`} />
              </div>
            ))}
          </div>
        ))}
      </div>
      {/* Visual column indicators could go here */}
    </div>
  )
}
