import { useEffect, useRef, useState, useCallback } from 'react'

export type Player = {
  username: string
}

export type BoardState = {
  cells: number[][]
}

export type GameState = {
  id: string
  players: Player[]
  board: BoardState
  turn: number
  winner: number
  done: boolean
}

export type LeaderboardEntry = {
  username: string
  wins: number
}

export const useGame = () => {
  const [username, setUsername] = useState('')
  const [connected, setConnected] = useState(false)
  const [status, setStatus] = useState('Enter a username to start')
  const [board, setBoard] = useState<number[][]>(Array.from({ length: 6 }, () => Array(7).fill(0)))
  const [gameId, setGameId] = useState('')
  const [yourTurn, setYourTurn] = useState(false)
  const [opponent, setOpponent] = useState('')
  const [winner, setWinner] = useState('')
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])

  const socketRef = useRef<WebSocket | null>(null)

  const connect = useCallback(() => {
    if (!username) return

    // Close existing connection if any
    if (socketRef.current) {
      socketRef.current.close()
    }

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${location.host}/ws?username=${encodeURIComponent(username)}`
    const ws = new WebSocket(wsUrl)
    socketRef.current = ws

    ws.onopen = () => {
      setConnected(true)
      setStatus('Waiting for match...')
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'state' && msg.state) {
          const state: GameState = msg.state
          setGameId(msg.gameId || state.id)
          setBoard(state.board.cells)
          setYourTurn(msg.yourTurn)
          setOpponent(msg.opponent || '')
          
          if (state.winner) {
             setWinner(state.players[state.winner - 1].username)
          } else {
             setWinner('')
          }

          if (state.done && !state.winner) {
            setStatus('Draw')
          } else if (state.done && state.winner) {
            const winnerName = state.players[state.winner - 1].username
            setStatus(winnerName === username ? 'You won!' : `${winnerName} won`)
            // Refresh leaderboard on game end
            fetchLeaderboard()
          } else {
            setStatus(msg.yourTurn ? 'Your turn' : `Waiting for ${msg.opponent || 'opponent'}`)
          }
        } else if (msg.type === 'error') {
          setStatus(msg.error)
        }
      } catch (e) {
        console.error("Failed to parse message", e)
      }
    }

    ws.onclose = () => {
      setConnected(false)
      setStatus('Disconnected')
      setBoard(Array.from({ length: 6 }, () => Array(7).fill(0)))
      setGameId('')
      setOpponent('')
      setWinner('')
    }
  }, [username])

  const sendMove = useCallback((col: number) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) return
    socketRef.current.send(JSON.stringify({ type: 'move', column: col }))
  }, [])

  const drop = useCallback((col: number) => {
    if (!yourTurn || winner) return
    sendMove(col)
  }, [yourTurn, winner, sendMove])

  const fetchLeaderboard = useCallback(async () => {
    try {
      const res = await fetch('/leaderboard')
      if (res.ok) {
        const data = await res.json()
        setLeaderboard(data)
      }
    } catch (err) {
      console.error(err)
    }
  }, [])

  useEffect(() => {
    fetchLeaderboard()
    return () => {
      if (socketRef.current) {
        socketRef.current.close()
      }
    }
  }, [fetchLeaderboard])

  return {
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
    yourTurn // Exported for UI indicators
  }
}
