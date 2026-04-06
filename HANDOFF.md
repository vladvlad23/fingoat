# Session Handoff

## What was completed

### Task 1 — Git repo init + savepoint commit ✅
- Created `.gitignore` covering Go binaries, node_modules, .svelte-kit, build/, .env files, OS/IDE artifacts
- Ran `git init`, staged all relevant source files (excluded node_modules, .claude/)
- Committed at `7dcc3a8` — "Initial commit — FinGoat backend (Go) + frontend (SvelteKit)"
- This is the savepoint. Each subsequent task gets its own commit.

---

## What remains

### Task 2 — Unit + integration tests ⬜
Delegate to `go-backend-engineer` agent. Prompt summary:

**Unit tests (mandatory):**
- `internal/auth/` — JWT generate/validate (expiry, invalid secret, malformed), password hash/check
- `internal/handler/` — each handler via `httptest`, hand-written mocks (no gomock), happy + error paths
- `internal/middleware/` — auth middleware: valid token, missing, expired, invalid
- `internal/model/` — response conversion helpers
- `internal/handler/helpers.go` — parseID, parseNumeric, parseDate edge cases

**Handler mocking strategy:**
- Handlers currently accept `*db.Queries` directly
- Introduce a minimal `Store` interface that `*db.Queries` satisfies
- Only extract methods actually needed for tests — no over-engineering

**Integration tests (if it makes sense):**
- Build tag: `//go:build integration`
- File suffix: `_integration_test.go`
- Use `TEST_DATABASE_URL` env var; skip with `t.Skip` if not set
- Each test cleans up its own data via `t.Cleanup` or rolled-back transactions

**Constraints:**
- No external test libraries unless already in go.mod (check first; testify not currently present)
- Use stdlib: `testing`, `net/http/httptest`, `encoding/json`

**After writing:**
- Run `go test ./...` from `backend/`, fix all failures
- Commit from repo root: `"test: add unit and integration tests for auth, handlers, middleware"`

---

### Task 3 — Summary/stats endpoint ⬜
Delegate to `go-backend-engineer` agent after Task 2 is committed.

New endpoint: `GET /api/summary`
- Protected (Bearer JWT)
- Query params: `from?`, `to?` (date range, default current month)
- Response shape:
  ```json
  {
    "period": { "from": "2026-04-01", "to": "2026-04-30" },
    "totalIncome": "3200.00",
    "totalExpenses": "1450.00",
    "net": "1750.00",
    "byCategory": [
      { "category": "Groceries", "totalExpenses": "320.00", "totalIncome": "0.00" }
    ],
    "goalProgress": [
      { "id": 1, "title": "Emergency Fund", "targetAmount": "5000.00", "currentAmount": "1200.00", "progressPct": 24.0 }
    ]
  }
  ```
- SQL: aggregate transactions for the user in the date range; JOIN with goals for progress
- Add to router in `cmd/server/main.go`
- Write unit tests for the new handler
- Commit: `"feat: add GET /api/summary stats endpoint"`

---

### Task 4 — JWT refresh token ⬜
Delegate to `go-backend-engineer` agent after Task 3 is committed. This touches both backend AND frontend.

**Backend changes:**
- New DB migration: `refresh_tokens` table
  ```sql
  id          BIGSERIAL PRIMARY KEY
  user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
  token_hash  TEXT NOT NULL UNIQUE   -- bcrypt hash of the raw token
  expires_at  TIMESTAMPTZ NOT NULL
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
  ```
- On login/register: generate a random 32-byte refresh token, hash it, store in DB, return raw token to client
- New endpoint: `POST /api/auth/refresh`
  - Accepts refresh token (httpOnly cookie or request body — see frontend note)
  - Validates token hash against DB, checks expiry
  - Issues a new JWT access token (15min expiry — reduce from 24h)
  - Optionally rotates the refresh token (delete old, issue new)
  - Returns new `{ token }` (same shape as login)
- New endpoint: `POST /api/auth/logout`
  - Deletes the refresh token from DB
- Add query methods to `db/` for refresh_tokens (no sqlc binary — write manually like the others)
- Write unit tests

**Frontend changes (`frontend/src/`):**
- `hooks.server.ts`: when a request fails with 401, attempt `POST /api/auth/refresh` using stored refresh token cookie, retry original request if successful, redirect to /login if refresh also fails
- Store refresh token in httpOnly cookie (set by backend response or handled server-side)
- Update `src/lib/api.ts` if needed

**Commit:** `"feat: implement JWT refresh token with rotation"`

---

## Key file locations
- Repo root: `/Users/vladimir.ungureanu/Learning/fingoat-respec/`
- Backend: `backend/`
- Frontend: `frontend/`
- Git: initialized, one commit so far (`7dcc3a8`)

## Agent routing
- Tasks 2, 3 backend parts → `go-backend-engineer`
- Task 4 frontend parts → `sveltekit-frontend-engineer`
- Task 4 coordination → `fullstack-lead` or handle sequentially (backend first, then frontend)
- Always label agent responses with **[agent-name]** prefix before showing to user
