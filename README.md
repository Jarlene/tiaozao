# 跳蚤市场 (Flea Market)

C2C second-hand marketplace - 商品系统 (with reviews/ratings)

## Tech Stack

- **Backend**: Go 1.22+ (net/http + chi router)
- **Database**: PostgreSQL 16 (pgx v5 driver, raw SQL)
- **File Storage**: Local filesystem (MinIO-ready for production)
- **Frontend**: Vanilla HTML/CSS/JS (no build tools needed)
- **Migrations**: Numbered SQL files in `migrations/`
- **Container**: Docker Compose

## Project Structure

```
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── config/          # Configuration (env-driven)
│   ├── middleware/       # Auth stub
│   ├── model/           # Data models
│   ├── handler/         # HTTP handlers
│   ├── repository/      # Data access (SQL)
│   ├── service/         # Business logic + typed errors
│   └── router/          # Route definitions
├── migrations/          # SQL migrations (001_init, 002_comments)
├── web/                 # Frontend (served statically)
├── vendor/              # Vendored Go dependencies
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Quick Start

### Option A: Docker Compose (recommended)

```bash
make docker-up
```

Opens at http://localhost:8080

### Option B: Local development

```bash
# Start PostgreSQL only
docker compose up -d postgres

# Run migrations
psql "$DATABASE_URL" -f migrations/001_init.sql
psql "$DATABASE_URL" -f migrations/002_comments.sql

# Run the application
make dev
```

## Configuration

Copy `.env.example` and adjust values as needed. Notable variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP port |
| `DATABASE_URL` | built from `DB_*` | Postgres DSN |
| `JWT_SECRET` | `tiaozao-dev-secret` | Secret for the auth stub / future real auth |
| `IMAGE_DIR` / `IMAGE_BASE_URL` | `./uploads` | Local image storage |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:8080` | Comma-separated allowed origins (no `*` in production) |

> Note: `.env.example` keys match `internal/config/config.go` exactly.

## API Endpoints

### Health
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |

### Categories
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/categories` | List all categories |
| GET | `/api/categories/:id` | Get category detail |
| POST | `/api/categories` | Create category |
| PUT | `/api/categories/:id` | Update category |
| DELETE | `/api/categories/:id` | Delete category |

### Products
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/products` | List products (`?category_id=&page=&size=&status=&seller_id=`) |
| GET | `/api/products/:id` | Get product detail (includes aggregate `rating`) |
| POST | `/api/products` | Create product |
| PUT | `/api/products/:id` | Update product |
| PUT | `/api/products/:id/delist` | Delist product (status → offline) |
| POST | `/api/products/images` | Upload image (multipart form) |

`status` is an **int enum**: `1` = active (在售), `2` = sold (已售出), `3` = offline (已下架).
Prices are stored as `decimal(10,2)`.

### Comments / Reviews
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/products/:productId/comments` | List reviews + nested replies with aggregate rating |
| POST | `/api/products/:productId/comments` | Create a review (`content`, optional `rating` 1-5) |
| POST | `/api/products/:productId/comments/:commentId/reply` | Reply to a review (1-level nesting) |

Reviews carry an optional `rating` (1-5 stars). The list endpoint returns
a `rating` summary `{average, count}` and product detail exposes the same.

### Auth

All `/api/*` routes use a development auth stub. Set `X-User-Id` header for a specific user ID (defaults to user 1). For token auth, send `Authorization: Bearer dev-token`.

## Frontend Pages

- `/` — Product listing with category filter and pagination
- `/publish.html` — Publish a new product
- `/detail.html?id=N` — Product detail with star rating and reviews

## Architecture Decisions

1. **Auth stub** — JWT/real auth (JAR-7) can replace `internal/middleware/auth.go` without touching handlers
2. **Local file storage** — ImageService abstracts storage; swap to MinIO/S3 by replacing `internal/service/image.go`
3. **Pure SQL** — No ORM; raw SQL for clarity and control, with migrations in `migrations/`
4. **Constructor DI** — every handler/service/repository is wired in `cmd/server/main.go`; no global DB singletons
5. **Typed errors** — services return sentinel errors (`service.ErrNotFound`, `ErrForbidden`, ...) that handlers map to HTTP status codes
6. **Vanilla frontend** — no npm/build step; a Vue/React SPA can be added in `web/` when needed

## Testing

```bash
make test        # go test ./...
```
</think>Let me verify the whole project one more time and check gofmt on all files, then review the git diff:

<｜DSML｜tool_calls>
<｜DSML｜invoke name="exec_command">
<｜DSML｜parameter name="cmd" string="true">cd tiaozao && gofmt -l internal/ cmd/ 2>&1; echo "gofmt clean above (empty=ok)"; GOPROXY=off go build ./... && GOPROXY=off go vet ./... && GOPROXY=off go test ./... 2>&1 | grep -cE "^ok|FAIL"; echo "all go"