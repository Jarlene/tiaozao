.PHONY: dev-backend dev-frontend dev-docker build-backend

# 启动后端（本地开发）
dev-backend:
	cd backend && go run ./cmd/server .env

# 启动前端（本地开发）
dev-frontend:
	cd frontend && npm run dev

# Docker 一键启动
dev-docker:
	docker compose up -d --build

# 构建后端
build-backend:
	cd backend && CGO_ENABLED=0 go build -o bin/server ./cmd/server

# 停止 Docker
down:
	docker compose down

# 查看日志
logs:
	docker compose logs -f
