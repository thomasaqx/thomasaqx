-- 001_initial.sql
-- Initial schema for the finance application

CREATE TABLE IF NOT EXISTS accounts (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT        NOT NULL,
    name        VARCHAR(100)  NOT NULL,
    type        VARCHAR(20)   NOT NULL CHECK (type IN ('checking', 'savings', 'wallet', 'invest')),
    balance     NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency    CHAR(3)       NOT NULL DEFAULT 'BRL',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);

CREATE TABLE IF NOT EXISTS categories (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT        NOT NULL,
    name        VARCHAR(100)  NOT NULL,
    description TEXT          NOT NULL DEFAULT '',
    color       VARCHAR(7)    NOT NULL DEFAULT '#000000',
    icon        VARCHAR(50)   NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);

CREATE TABLE IF NOT EXISTS transactions (
    id               BIGSERIAL PRIMARY KEY,
    account_id       BIGINT        NOT NULL REFERENCES accounts(id),
    to_account_id    BIGINT        REFERENCES accounts(id),
    category_id      BIGINT        REFERENCES categories(id),
    type             VARCHAR(20)   NOT NULL CHECK (type IN ('income', 'expense', 'transfer')),
    status           VARCHAR(20)   NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    amount           NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    description      TEXT          NOT NULL DEFAULT '',
    transaction_date TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id    ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type          ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_date          ON transactions(transaction_date);

CREATE TABLE IF NOT EXISTS budgets (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT        NOT NULL,
    category_id  BIGINT        NOT NULL REFERENCES categories(id),
    amount       NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    spent        NUMERIC(15,2) NOT NULL DEFAULT 0,
    period_year  INT           NOT NULL,
    period_month INT           NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, category_id, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_budgets_user_period ON budgets(user_id, period_year, period_month);
