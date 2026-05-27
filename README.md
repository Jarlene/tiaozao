# 跳蚤市场 (Flea Market)

C2C second-hand marketplace - 商品系统

## Tech Stack

- **Backend**: Go 1.22+ (net/http + chi router)
- **Database**: PostgreSQL 16 (pgx v5 driver)
- **File Storage**: Local filesystem (MinIO-ready for production)
- **Frontend**: Vanilla HTML/CSS/JS (no build tools needed)
- **Container**: Docker Compose

## Project Structure

```
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── config/          # Configuration
│   ├── middleware/       # Auth stub
│   ├── model/           # Data models
│   ├── handler/         # HTTP handlers
│   ├── repository/      # Data access (SQL)
│   ├── service/         # Business logic
│   └── router/          # Route definitions
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

# Run the application
make dev
```

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
| GET | `/api/products` | List products (supports `?category_id=&page=&size=&seller_id=`) |
| GET | `/api/products/:id` | Get product detail |
| POST | `/api/products` | Create product |
| PUT | `/api/products/:id` | Update product |
| PUT | `/api/products/:id/delist` | Delist product |
| POST | `/api/products/images` | Upload image (multipart form) |

### Orders
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/orders` | Create order |
| GET | `/api/orders` | List orders (`?role=buyer\|seller&status=&page=&size=`) |
| GET | `/api/orders/:id` | Get order detail |
| PUT | `/api/orders/:id/status` | Transition order status (`{"action":"pay\|ship\|receive\|cancel\|refund"}`) |

### Auth

All `/api/*` routes use a development auth stub. Set `X-User-Id` header for a specific user ID, or it defaults to user 1. For token auth, send `Authorization: Bearer dev-token`.

## Frontend Pages

- `/` — Product listing with category filter and pagination
- `/publish.html` — Publish a new product
- `/detail.html?id=N` — Product detail page
- `/orders.html` — Order list (buyer/seller toggle + status filter)
- `/order-detail.html?id=N` — Order detail page with timeline and actions

## Database

```bash
# Run migrations
make migrate
```

## Architecture Decisions

1. **Auth stub** — JWT/real auth (JAR-7) can replace `internal/middleware/auth.go` without touching handlers
2. **Local file storage** — ImageService abstracts storage; swap to MinIO/S3 by replacing `internal/service/image.go`
3. **Pure SQL** — No ORM; raw SQL for clarity and control
4. **Vanilla frontend** — No npm/build step; a Vue/React SPA can be added in `web/` when needed
