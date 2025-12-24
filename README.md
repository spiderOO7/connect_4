# 4 in a Row

Real-time Connect Four built with Go backend (WebSockets, bot fallback, Postgres persistence, Kafka analytics) and React (Vite) frontend.

## Prerequisites
- Go 1.22+
- Node 18+
- Docker (for Postgres + Kafka via docker-compose)

## Run locally

```bash
# start infra (Postgres + Kafka)
docker-compose up -d

# backend
cd backend
go run ./cmd/server

# frontend
cd frontend
npm install
npm run dev
```

Backend listens on `:8080`, frontend on `:5173` with proxy to backend.

## Environment variables
- `PORT` (default `8080`)
- `POSTGRES_URL` (default `postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable`)
- `KAFKA_BROKERS` (default `localhost:9092`)
- `ALLOWED_ORIGINS` (default `http://localhost:5173,http://localhost:3000`)
- `BOT_WAIT_SECONDS` (default `10`)
- `RECONNECT_SECONDS` (default `30`)

## Game flow
- WebSocket `GET /ws?username=alice` matches you with another player; after 10s if none, a bot joins.
- Move via `{"type":"move","column":3}` messages. State snapshots broadcast after each move.
- Finished games stored in Postgres; leaderboard available at `GET /leaderboard`.
- Analytics events published to Kafka topic `game-analytics`.

## Kafka consumer (bonus)
A simple consumer service can be added to read `game-analytics` and log/aggregate metrics (not yet implemented here).

## Deployment
Containerize backend and frontend separately or serve frontend statically from any host; point it to backend WebSocket/HTTP origin and set CORS via `ALLOWED_ORIGINS`.
