# AGENTS.md — ShikaGari Developer Instructions

## Quick Start

```bash
# Install deps
go mod download

# Configure env (required: DB_PASSWORD, JWT_SECRET)
cp .env.example .env
# edit .env

# Run dev server (port 8080)
go run cmd/api/main.go

# Or with live reload (requires Air)
go install github.com/air-verse/air@latest
air

# Health check
curl http://localhost:8080/health
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run server (`go run cmd/api/main.go`) |
| `make dev` | Run with live reload (`air`) |
| `make build` | Compile to `./bin/shikagari` |
| `make test` | Run all tests (`go test ./... -v -count=1`) |
| `make test-coverage` | Tests + HTML coverage report |
| `make lint` | Run `golangci-lint run ./...` |
| `make tidy` | `go mod tidy && go mod verify` |
| `make migrate` | Apply raw SQL migrations (requires `DATABASE_URL`) |
| `make clean` | Remove `bin/`, `tmp/`, `uploads/` |

## Required Environment Variables

| Variable | Required | Default | Notes |
|----------|----------|---------|-------|
| `DB_PASSWORD` | **Yes** | — | PostgreSQL password |
| `JWT_SECRET` | **Yes** | — | Min 32 chars, HS256 signing key |
| `APP_ENV` | No | `development` | `development` or `production` |
| `APP_PORT` | No | `8080` | HTTP port |
| `DB_HOST` | No | `localhost` | |
| `DB_PORT` | No | `5432` | |
| `DB_USER` | No | `shikagari_user` | |
| `DB_NAME` | No | `shikagari_db` | |
| `DB_SSLMODE` | No | `disable` | Use `require` in production |
| `JWT_EXPIRY_HOURS` | No | `72` | Token TTL |
| `UPLOAD_DIR` | No | `./uploads` | Local file storage |
| `MAX_FILE_SIZE_MB` | No | `5` | Upload limit |
| `CORS_ALLOWED_ORIGINS` | No | `http://localhost:3000` | Comma-separated list |

> `.env` loads only when `APP_ENV != "production"` (see `config/config.go:60`)

## Architecture Overview

```
cmd/api/main.go          # Entrypoint → config → DB → router → server
internal/router/router.go   # DI wiring + ALL route registration (single source of truth)
internal/handler/        # Gin controllers (HTTP layer)
internal/service/        # Business logic
internal/repository/     # Data access (interfaces + postgres impl)
internal/middleware/     # Auth, CORS, logger, RBAC
pkg/                     # Shared utilities (db, jwt, hash, upload, validator, response)
migrations/              # Raw SQL (run manually via `make migrate` or psql)
```

- **Entry point**: `cmd/api/main.go`
- **All routes defined in**: `internal/router/router.go` — search here for endpoint paths
- **DI container**: `wireDependencies()` in `router.go` constructs repos → services → handlers
- **Auth middleware**: `Authenticate` (strict) vs `OptionalAuthenticate` (soft) in `internal/middleware/auth_middleware.go`
- **RBAC helpers**: `RequireSeller()`, `RequireAdmin()`, `RequireRole(...)` in `internal/middleware/role_middleware.go`

## Testing

- Single test file: `go test -v ./internal/handler/... -run TestAuthHandler_ListSecurityEvents`
- All tests: `make test` or `go test ./... -v -count=1`
- Coverage: `make test-coverage` → opens `coverage.html`
- Tests use Gin test mode (`gin.SetMode(gin.TestMode)`) and `httptest`
- Mock services implemented inline per test file (see `auth_handler_test.go`)
- No test DB — unit tests mock repository interfaces

## Database & Migrations

- **ORM**: GORM with `pgcrypto` for UUIDs
- **Timezone**: `Africa/Nairobi` (hardcoded in DSN, `config/config.go:41`)
- **Migrations**: Raw SQL files in `migrations/` — apply manually:
  ```bash
  export DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable"
  make migrate
  ```
- **AutoMigrate**: Runs on every start in development (`pkg/database/database.go`)
- **Soft deletes**: `users` and `listings` use `gorm.DeletedAt` for referential integrity

## Key Conventions

- **Repository pattern**: Interfaces in `internal/repository/interfaces/`, implementations in `internal/repository/postgres/`
- **Not found = `nil, nil`**: Repo methods return `(nil, nil)` when record missing — callers check for `nil`
- **Response envelope**: All HTTP responses use `pkg/response` wrapper (`Success`, `Data`, `Message`, `Error`)
- **JWT claims**: `user_id`, `email`, `role`, `is_verified`, `session_id`, `exp`
- **Roles**: `buyer`, `seller` (private/individual), `dealer` (business), `admin` (admin assigned manually in DB)
- **Registration**: `POST /api/v1/auth/register` accepts `role: buyer | seller | dealer` (defaults `buyer`); frontend shows Buy / Sell cards, Sell expands to Dealer / Private-seller sub-cards
- **Seller flow**: Register with role → Create matching profile (`/dealers/profile` needs `dealer`; `/sellers/profile` allows `buyer` as upgrade path + `seller`) → Admin approves → Can create listings (`POST /listings` allows `seller` + `dealer` + `admin`)
- **Prices**: Kenyan Shillings (KES), stored as integers
- **Locations**: Profile locations validated as Kenyan counties (`kenyacounty` rule, list in `internal/domain/location.go`); listing search still uses the legacy city list
- **File uploads**: Local `./uploads/` by default; swap `UploadService.UploadImage` for S3

## Common Tasks

### Add a new endpoint
1. Define DTO in `internal/dto/`
2. Add repo interface + impl in `internal/repository/`
3. Add service in `internal/service/`
4. Add handler in `internal/handler/`
5. Wire in `router.go:wireDependencies()` and register route in a `register*Routes` function

### Run linter
```bash
golangci-lint run ./...
# Or: make lint
```

### Build for production
```bash
make build
# Binary at ./bin/shikagari
# Run with: APP_ENV=production ./bin/shikagari
```

## Gotchas

- `.gitignore` only ignores `frontend/` and `.vscode/` — `bin/`, `tmp/`, `uploads/`, `.env` are **not** ignored
- No CI/CD workflows (`.github/workflows/` empty)
- `golangci-lint` must be installed separately
- `make migrate` requires `DATABASE_URL` env var set
- Admin role cannot be self-assigned — must be set directly in DB
- Uploads served statically at `/uploads/*` via Gin (`router.go:35`) — replace with CDN in prod