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
