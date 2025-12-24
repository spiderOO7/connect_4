import React from 'react'
import { LeaderboardEntry } from '../hooks/useGame'

type LeaderboardProps = {
  data: LeaderboardEntry[]
}

export const Leaderboard: React.FC<LeaderboardProps> = ({ data }) => {
  return (
    <div className="leaderboard-card">
      <h3>🏆 Leaderboard</h3>
      <div className="leaderboard-list">
        {data.length === 0 ? (
          <div className="empty-state">No records yet</div>
        ) : (
          data.map((entry, index) => (
            <div key={entry.username} className="leaderboard-item">
              <span className="rank">#{index + 1}</span>
              <span className="name">{entry.username}</span>
              <span className="score">{entry.wins} wins</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
