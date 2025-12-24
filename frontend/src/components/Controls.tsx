import React from 'react'

type ControlsProps = {
  username: string
  setUsername: (name: string) => void
  onConnect: () => void
  connected: boolean
}

export const Controls: React.FC<ControlsProps> = ({ username, setUsername, onConnect, connected }) => {
  return (
    <div className="controls">
      <div className="input-group">
        <input
          type="text"
          placeholder="Enter Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          disabled={connected}
          maxLength={15}
        />
        <button 
          onClick={onConnect} 
          disabled={connected || !username.trim()}
          className={connected ? 'connected' : ''}
        >
          {connected ? 'Connected' : 'Play Now'}
        </button>
      </div>
    </div>
  )
}
