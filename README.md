# FinGoat

Personal finance tracker — set goals, log transactions, and watch your progress.

**Stack:** Go backend · SvelteKit frontend · PostgreSQL 16

---

## Prerequisites

- [Go](https://golang.org/dl/) 1.25+
- [Node.js](https://nodejs.org/) 20+ with npm
- [Docker](https://www.docker.com/) (for PostgreSQL)
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) CLI

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## Quick start

### 1. Start the database

```bash
cd backend
docker compose up -d
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env — set JWT_SECRET to something secure
```

Default `.env.example` values work out of the box with the Docker Compose database:

```
DATABASE_URL=postgres://fingoat:fingoat@localhost:5432/fingoat?sslmode=disable
JWT_SECRET=your-secret-here
PORT=8080
```

### 3. Run migrations

```bash
migrate -path internal/db/migration -database "$DATABASE_URL" up
```

### 4. Start the backend

```bash
go run ./cmd/server/
# Listening on :8080
```

### 5. Start the frontend

```bash
cd ../frontend
npm install
npm run dev
# Listening on http://localhost:5173
```

---

## Project layout

```
fingoat/
├── backend/
│   ├── cmd/server/main.go          # entry point
│   ├── internal/
│   │   ├── auth/                   # JWT generation, bcrypt
│   │   ├── config/                 # env config loading
│   │   ├── db/
│   │   │   ├── migration/          # SQL migration files (golang-migrate)
│   │   │   └── query/              # .sql query files (sqlc source)
│   │   ├── handler/                # HTTP handlers
│   │   ├── middleware/             # JWT auth, logging
│   │   ├── model/                  # domain types
│   │   ├── service/                # business logic
│   │   └── stores/                 # store interfaces
│   ├── docker-compose.yml
│   ├── sqlc.yaml
│   └── .env.example
└── frontend/
    └── src/                        # SvelteKit app
```

---

## API overview

All endpoints are under `/api`. Auth endpoints are public; everything else requires `Authorization: Bearer <token>`.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Create account → `{ token }` |
| POST | `/api/auth/login` | Sign in → `{ token }` |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Invalidate refresh token |
| GET | `/api/me` | Current user profile |
| PATCH | `/api/me` | Update monthly income |
| GET | `/api/goals` | List goals |
| POST | `/api/goals` | Create goal |
| GET/PUT/DELETE | `/api/goals/:id` | Read / update / delete goal |
| GET | `/api/transactions` | List transactions (filterable) |
| POST | `/api/transactions` | Create transaction |
| GET/PUT/DELETE | `/api/transactions/:id` | Read / update / delete transaction |
| GET | `/api/summary` | Aggregated stats for a date range |

See [`backend/SPEC.md`](backend/SPEC.md) for the full API contract, domain model, and conventions.

---

## Running tests

```bash
cd backend
go test ./...
```

Integration tests require a live database and are skipped automatically if `TEST_DATABASE_URL` is not set:

```bash
TEST_DATABASE_URL="postgres://fingoat:fingoat@localhost:5432/fingoat?sslmode=disable" \
  go test -tags integration ./...
```

---

## Regenerating SQL code (sqlc)

If you modify query files under `internal/db/query/`, regenerate the Go code:

```bash
cd backend
sqlc generate
```

Install sqlc: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
