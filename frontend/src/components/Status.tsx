import React from 'react'

type StatusProps = {
  message: string
  winner: string
}

export const Status: React.FC<StatusProps> = ({ message, winner }) => {
  return (
    <div className={`status-bar ${winner ? 'has-winner' : ''}`}>
      <div className="status-message">{message}</div>
    </div>
  )
}
