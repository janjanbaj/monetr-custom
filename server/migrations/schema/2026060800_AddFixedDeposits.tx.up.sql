CREATE TABLE IF NOT EXISTS "fixed_deposits"
(
    "fixed_deposit_id"                VARCHAR(32) NOT NULL,
    "account_id"                      VARCHAR(32) NOT NULL,
    "source_bank_account_id"          VARCHAR(32) NOT NULL,
    "fixed_bank_account_id"           VARCHAR(32) NOT NULL,
    "funding_schedule_id"             VARCHAR(32),
    "name"                            TEXT        NOT NULL,
    "amount"                          BIGINT      NOT NULL DEFAULT 0,
    "interest_rate"                   DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    "start_date"                      TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "end_date"                        TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "interest_frequency"              VARCHAR(32) NOT NULL,
    "interest_destination"            VARCHAR(32) NOT NULL,
    "interest_destination_bank_account_id" VARCHAR(32),
    "status"                          VARCHAR(32) NOT NULL,
    "created_at"                      TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    "updated_at"                      TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT "pk_fixed_deposits" PRIMARY KEY ("fixed_deposit_id", "account_id"),
    CONSTRAINT "fk_fixed_deposits_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    CONSTRAINT "fk_fixed_deposits_source_bank" FOREIGN KEY ("source_bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
    CONSTRAINT "fk_fixed_deposits_fixed_bank" FOREIGN KEY ("fixed_bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
    CONSTRAINT "fk_fixed_deposits_destination_bank" FOREIGN KEY ("interest_destination_bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE SET NULL,
    CONSTRAINT "fk_fixed_deposits_funding_schedule" FOREIGN KEY ("funding_schedule_id", "account_id", "fixed_bank_account_id") REFERENCES "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id") ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS "idx_fixed_deposits_fixed_bank_account" ON "fixed_deposits" ("fixed_bank_account_id", "account_id");
CREATE INDEX IF NOT EXISTS "idx_fixed_deposits_source_bank_account" ON "fixed_deposits" ("source_bank_account_id", "account_id");
