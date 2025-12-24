import React from 'react'

type GameInfoProps = {
  gameId: string
  opponent: string
}

export const GameInfo: React.FC<GameInfoProps> = ({ gameId, opponent }) => {
  if (!gameId) return null
  
  return (
    <div className="game-info">
      <div className="info-item">
        <span className="label">Game ID</span>
        <span className="value">{gameId}</span>
      </div>
      <div className="info-item">
        <span className="label">Opponent</span>
        <span className="value">{opponent || 'Waiting...'}</span>
      </div>
    </div>
  )
}
