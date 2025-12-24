import { useGame } from './hooks/useGame'
import { Board } from './components/Board'
import { Controls } from './components/Controls'
import { Status } from './components/Status'
import { GameInfo } from './components/GameInfo'
import { Leaderboard } from './components/Leaderboard'

function App() {
  const {
    username,
    setUsername,
    connected,
    status,
    board,
    gameId,
    opponent,
    winner,
    leaderboard,
    connect,
    drop,
    yourTurn
  } = useGame()

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>4 in a Row</h1>
        <p className="subtitle">Real-time Multiplayer Strategy</p>
      </header>

      <main className="game-layout">
        <section className="game-section">
          <Controls 
            username={username}
            setUsername={setUsername}
            onConnect={connect}
            connected={connected}
          />
          <Status message={status} winner={winner} />
          <Board 
            board={board} 
            onDrop={drop} 
            disabled={!connected || !yourTurn || !!winner} 
          />
          <GameInfo gameId={gameId} opponent={opponent} />
        </section>

        <aside className="sidebar">
          <Leaderboard data={leaderboard} />
        </aside>
      </main>
      
      <footer className="app-footer">
        <p>&copy; {new Date().getFullYear()} 4-in-a-row</p>
      </footer>
    </div>
  )
}

export default App
