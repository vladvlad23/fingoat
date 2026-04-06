CREATE TABLE users (
  id             BIGSERIAL PRIMARY KEY,
  email          TEXT UNIQUE NOT NULL,
  password_hash  TEXT NOT NULL,
  monthly_income NUMERIC(12,2) NOT NULL DEFAULT 0.00,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE goals (
  id             BIGSERIAL PRIMARY KEY,
  user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title          TEXT NOT NULL,
  target_amount  NUMERIC(12,2) NOT NULL,
  current_amount NUMERIC(12,2) NOT NULL DEFAULT 0.00,
  deadline       DATE,
  status         TEXT NOT NULL DEFAULT 'active',
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  goal_id    BIGINT REFERENCES goals(id) ON DELETE SET NULL,
  title      TEXT NOT NULL,
  amount     NUMERIC(12,2) NOT NULL,
  type       TEXT NOT NULL,
  category   TEXT,
  date       DATE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
