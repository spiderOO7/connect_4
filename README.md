# 4 in a Row/Connet 4

Real-time Connect Four with WebSockets matchmaking, bot fallback, Postgres persistence, Kafka analytics, and a React (Vite) UI.

## Features
- Real-time play over WebSockets with reconnect support and graceful forfeit after a timeout
- Automatic bot opponent after a configurable wait when no human is found
- Leaderboard persisted in Postgres
- Analytics events published to Kafka topic `game-analytics` (sample consumer included)
- React/Vite frontend that connects to the backend WebSocket and leaderboard API

## Prerequisites
- Go 1.22+
- Node 18+
- Docker (for Postgres + Kafka via docker-compose)

## Run locally

```bash
# 1) start infra (Postgres + Kafka)
docker-compose up -d

# 2) backend
cd backend
POSTGRES_URL="postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable" \
KAFKA_BROKERS="localhost:9092" \
go run ./cmd/server

# 3) frontend
cd ../frontend
npm install
VITE_BACKEND_ORIGIN=http://localhost:8080 npm run dev
```

Backend listens on `:8080`. Vite dev server listens on `:5173` and talks directly to the backend origin you configure.

## Environment variables

Backend
- `PORT` (default `8080`)
- `POSTGRES_URL` (default `postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable`)
- `KAFKA_BROKERS` (comma-separated; empty disables analytics; default `localhost:9092`)
- `ALLOWED_ORIGINS` (comma-separated CORS allowlist; default includes local Vite and the hosted demo)
- `BOT_WAIT_SECONDS` (seconds to wait before assigning a bot; default `10`)
- `RECONNECT_SECONDS` (grace period before a disconnected player forfeits; default `30`)

Frontend
- `VITE_BACKEND_ORIGIN` (backend base URL; defaults to the hosted demo URL; set to `http://localhost:8080` for local dev)

## API

### WebSocket: `/ws?username=<name>`
- Connects a player; if the same username reconnects, the server restores the game until `RECONNECT_SECONDS` expires.
- Client messages: `{ "type": "move", "column": 3 }`, `{ "type": "ping" }`, `{ "type": "reconnect" }`.
- Server messages: `{ "type": "state", gameId, state, yourTurn, opponent, reconnect, message }` or `{ "type": "error", error }`.
- State payload includes board cells, players, whose turn, winner, and move history.

### HTTP
- `GET /leaderboard` → `[{ "username": "alice", "wins": 5 }, ...]` (top 20 by wins)
- `GET /healthz` → `ok`

## Game flow
1) Connect over WebSocket with a username.
2) If another player is waiting, you are matched; otherwise after `BOT_WAIT_SECONDS` a bot joins.
3) Moves are column numbers 0-6; server broadcasts full state after each move.
4) Win detection handles horizontal/vertical/diagonal streaks of four; draw when the board is full.
5) Disconnects: if a player does not reconnect within `RECONNECT_SECONDS`, the opponent wins by forfeit.

## Persistence
- Postgres table `games` stores finished games with players, winner, moves (JSON), created/finished timestamps.
- Leaderboard aggregates wins from this table.

## Analytics
- When `KAFKA_BROKERS` is set, events are emitted to topic `game-analytics` (producer in `internal/analytics`).
- Events include types like `joined`, `bot_move`, and `finished` with payloads containing game id, winner, and moves.

### Sample consumer
Run the bundled consumer to print analytics events:
```bash
cd backend
KAFKA_BROKERS="localhost:9092" go run ./cmd/consumer
```

## Deployment
- Backend: build a Go binary or container; expose port `8080` (or `PORT`).
- Frontend: `npm run build` then serve the static `dist/` and set `VITE_BACKEND_ORIGIN` to point at the backend.
- Configure CORS via `ALLOWED_ORIGINS`; provision Postgres and Kafka if you want persistence/analytics.
