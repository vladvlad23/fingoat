# FinGoat — Project Spec

## What This Is

A personal finance REST API + web frontend. Users track financial goals and log transactions against them. The backend manages auth, goals, transactions, and basic balance/progress calculations.

---

## Tech Stack

**Backend:** Go
- Router: `chi` or `echo`
- SQL: `sqlc` (type-safe query generation from plain SQL — no ORM)
- Migrations: `golang-migrate`
- Auth: JWT (HMAC256), BCrypt for password hashing
- Config: environment variables via `.env`

**Frontend:** SvelteKit
- Styling: TailwindCSS
- HTTP client: native `fetch` with a thin wrapper

**Database:** PostgreSQL 16 (Docker for local dev)

**Rationale:** Previous implementation used Kotlin/Ktor with Gradle + KSP + JOOQ, which added significant build-time complexity (codegen requiring a live DB, annotation processing, fat JAR pipeline) for a project of this scale. Go gives comparable type safety at a fraction of the tooling overhead. `sqlc` mirrors the JOOQ philosophy (typed SQL) without the Gradle plugin chain.

---

## Domain Model

### User
```
id            BIGSERIAL PRIMARY KEY
email         TEXT UNIQUE NOT NULL
password_hash TEXT NOT NULL
monthly_income NUMERIC(12, 2) NOT NULL DEFAULT 0.00
created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
```

### Goal
```
id             BIGSERIAL PRIMARY KEY
user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
title          TEXT NOT NULL
target_amount  NUMERIC(12, 2) NOT NULL
current_amount NUMERIC(12, 2) NOT NULL DEFAULT 0.00
deadline       DATE
status         TEXT NOT NULL DEFAULT 'active'  -- active | completed | archived
created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
```
`current_amount` is updated whenever a transaction is linked to this goal.
`status` transitions: active → completed (when current_amount >= target_amount), or manually → archived.

### Transaction
```
id         BIGSERIAL PRIMARY KEY
user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
goal_id    BIGINT REFERENCES goals(id) ON DELETE SET NULL  -- optional link
title      TEXT NOT NULL
amount     NUMERIC(12, 2) NOT NULL  -- always positive
type       TEXT NOT NULL            -- 'income' | 'expense'
category   TEXT                     -- free-text or enum, nullable
date       DATE NOT NULL
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

**Key decisions:**
- `amount` is always positive. Direction is conveyed by `type`.
- `goal_id` is optional — not every transaction is tied to a goal.
- When a transaction with `goal_id` is created, `goals.current_amount` is incremented (or decremented on delete).
- `category` is a nullable free-text field for now. Can be normalized into its own table later if needed.

---

## API Contract

All endpoints under `/api`. Auth endpoints are public; all others require `Authorization: Bearer <token>`.

**Identity is always derived from the JWT — never accepted in the request body.**

### Auth

```
POST /api/auth/register
Body: { email, password }
201: { token }
409: email already exists

POST /api/auth/login
Body: { email, password }
200: { token }
401: invalid credentials
```

### User

```
GET /api/me
200: { id, email, monthlyIncome, createdAt }

PATCH /api/me
Body: { monthlyIncome? }
200: updated user object
```

### Goals

```
GET /api/goals
200: [{ id, title, targetAmount, currentAmount, deadline, status, createdAt }]

POST /api/goals
Body: { title, targetAmount, deadline? }
201: goal object

GET /api/goals/:id
200: goal object
403: not owner
404: not found

PUT /api/goals/:id
Body: { title?, targetAmount?, deadline?, status? }
200: updated goal object
403: not owner
404: not found


DELETE /api/goals/:id
204
403: not owner
404: not found
```

### Transactions

```
GET /api/transactions
Query params: goal_id? (filter by goal), type? (income|expense), from? to? (date range)
200: [{ id, goalId, title, amount, type, category, date, createdAt }]

POST /api/transactions
Body: { title, amount, type, date, goalId?, category? }
201: transaction object
-- if goalId provided, goals.current_amount is updated atomically

GET /api/transactions/:id
200: transaction object
403: not owner
404: not found

PUT /api/transactions/:id
Body: { title?, amount?, type?, date?, goalId?, category? }
200: updated transaction object
403: not owner
404: not found
-- amount/goalId changes must recalculate goal current_amount

DELETE /api/transactions/:id
204
403: not owner
-- if linked to goal, goals.current_amount is decremented
```

---

## Conventions

### Money
- Store as `NUMERIC(12, 2)` in Postgres.
- Serialize as `string` in JSON responses (e.g. `"amount": "1234.56"`) to avoid float precision issues on the client.
- Accept either `string` or `number` in request bodies; parse and validate server-side.

### IDs
- Use `BIGSERIAL` (auto-increment) for all primary keys.
- Expose `id` (integer) in API responses — no separate UUID column. UUIDs added complexity without benefit at this scale.
- Ownership checks: every resource endpoint must verify `resource.user_id == jwt.user_id` before responding.

### Error responses
Consistent shape across all endpoints:
```json
{ "error": "human readable message" }
```

### Dates
- Store timestamps as `TIMESTAMPTZ`.
- Store transaction `date` as `DATE` (no time component — a transaction happened on a day).
- Serialize all timestamps as ISO 8601 strings.

### JWT
- Claim: `user_id` (integer)
- Expiry: 24 hours
- Secret from `JWT_SECRET` env var

---

## What Was Wrong in the Previous Implementation (Don't Repeat)

- `userEmail` was included in request bodies — identity must come from JWT only.
- Ownership checks were applied inconsistently (transactions checked, goals did not).
- `Transaction.uuid` was `String`, `Goal.uuid` was `UUID` — type inconsistency for the same concept.
- `category` existed in `TransactionRequest` but not in the model or schema — silently dropped.
- `amount` on transactions had no direction field — couldn't distinguish income from expense.
- Goals had no `currentAmount` or progress tracking — a goal was just a label with a number.
- DB primary keys had no `GENERATED ALWAYS AS IDENTITY` / `SERIAL` — auto-increment was missing.
- Domain model used nullable `id` (`Long?`) to represent both pre- and post-insert states.
- `GoalRequest` contained a redundant `goalUUID` field on top of the path parameter.
- No `POST /api/auth/register` endpoint — users couldn't be created via the API.

---

## Project Structure (Go Backend)

```
/cmd/server/main.go          -- entry point
/internal/
  auth/                      -- JWT generation, BCrypt
  config/                    -- env config loading
  db/                        -- sqlc generated code
    query/                   -- .sql query files
    migration/               -- .sql migration files
  handler/                   -- HTTP handlers (auth, goals, transactions, user)
  middleware/                 -- JWT auth middleware, logging
  model/                     -- domain types
  service/                   -- business logic (goal progress updates, etc.)
/sqlc.yaml
/docker-compose.yml
/.env.example
```

---

## Local Dev Setup

```bash
docker compose up -d          # start postgres
go run cmd/server/main.go     # start server on :8080
```

Migration and sqlc codegen should be runnable as:
```bash
migrate -path internal/db/migration -database $DATABASE_URL up
sqlc generate
```
