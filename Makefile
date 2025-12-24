up:
	docker-compose up -d

down:
	docker-compose down

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm install && npm run dev

fmt:
	cd backend && go fmt ./...
	test -d frontend && cd frontend && npm run lint || true
