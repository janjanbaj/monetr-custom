# Phase 1: DB Schema & Migrations

This phase implements the database migrations to create the tables required for accrual expenses and their corresponding manual usage logs.

## Context Directory
- [server/migrations/schema/](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/migrations/schema)

---

## Proposed Database Schema

We need two tables:
1. `accrual_expenses`: Stores the metadata, total value, and date range for the bulk item or subscription.
2. `accrual_usage_logs`: Stores manual usage entries that record specific consumption amounts on specific days.

### Up Migration File

#### [NEW] [2026060700_AddAccrualExpenses.tx.up.sql](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/migrations/schema/2026060700_AddAccrualExpenses.tx.up.sql)
```sql
CREATE TABLE IF NOT EXISTS "accrual_expenses"
(
    "accrual_expense_id" VARCHAR(32) NOT NULL,
    "account_id"         VARCHAR(32) NOT NULL,
    "bank_account_id"    VARCHAR(32) NOT NULL,
    "transaction_id"     VARCHAR(32),
    "name"               TEXT        NOT NULL,
    "description"        TEXT,
    "amount"             BIGINT      NOT NULL DEFAULT 0,
    "start_date"         TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "end_date"           TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "created_at"         TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "updated_at"         TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT "pk_accrual_expenses" PRIMARY KEY ("accrual_expense_id", "account_id", "bank_account_id"),
    CONSTRAINT "fk_accrual_expenses_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    CONSTRAINT "fk_accrual_expenses_bank_accounts" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
    CONSTRAINT "fk_accrual_expenses_transactions" FOREIGN KEY ("transaction_id", "account_id", "bank_account_id") REFERENCES "transactions" ("transaction_id", "account_id", "bank_account_id") ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS "idx_accrual_expenses_bank_account" ON "accrual_expenses" ("bank_account_id", "account_id");

CREATE TABLE IF NOT EXISTS "accrual_usage_logs"
(
    "accrual_usage_log_id" VARCHAR(32) NOT NULL,
    "accrual_expense_id"   VARCHAR(32) NOT NULL,
    "account_id"           VARCHAR(32) NOT NULL,
    "bank_account_id"      VARCHAR(32) NOT NULL,
    "amount"               BIGINT      NOT NULL DEFAULT 0,
    "date"                 TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "created_at"           TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "updated_at"           TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT "pk_accrual_usage_logs" PRIMARY KEY ("accrual_usage_log_id", "account_id", "bank_account_id"),
    CONSTRAINT "fk_accrual_usage_logs_accrual_expenses" FOREIGN KEY ("accrual_expense_id", "account_id", "bank_account_id") REFERENCES "accrual_expenses" ("accrual_expense_id", "account_id", "bank_account_id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_accrual_usage_logs_expense" ON "accrual_usage_logs" ("accrual_expense_id", "account_id", "bank_account_id");
```

### Down Migration File

#### [NEW] [2026060700_AddAccrualExpenses.tx.down.sql](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/migrations/schema/2026060700_AddAccrualExpenses.tx.down.sql)
```sql
DROP TABLE IF EXISTS "accrual_usage_logs";
DROP TABLE IF EXISTS "accrual_expenses";
```

---

## Verification Plan

1. **Local Migration Run**: Start the backend which automatically runs embedded migrations:
   ```bash
   make develop
   ```
2. Check the logs for `database upgraded` message showing version `2026060700` was successfully applied.
3. Query the Postgres database (e.g. via psql or PgAdmin) to verify the schema tables `accrual_expenses` and `accrual_usage_logs` exist and have correct foreign key constraints.
