# Fixed Deposit Plan Status

Use this file to track implementation progress and store succinct debrief details from each completed phase. This enables subsequent agents to resume work without parsing all phase-specific files.

## Progress Checklist

- [x] Phase 1: DB Schema & Migrations
- [x] Phase 2: Go Models & Repository Layer
- [x] Phase 3: Controllers & REST APIs
- [x] Phase 4: Frontend API Clients & Hooks
- [x] Phase 5: User Interface and Modals

---

## Phase Debriefs & Context

This section contains critical context accumulated as phases are completed.

### Phase 1: DB Schema & Migrations
*Status: Completed*
*Key Details for subsequent phases:*
- Added `fixed_deposits` table with columns: `fixed_deposit_id`, `account_id`, `source_bank_account_id`, `fixed_bank_account_id`, `funding_schedule_id` (nullable), `name`, `amount`, `interest_rate`, `start_date`, `end_date`, `interest_frequency`, `interest_destination`, `interest_destination_account_id` (nullable), `status`.
- Primary Key: `(fixed_deposit_id, account_id)`.
- Foreign Keys added for `accounts`, `bank_accounts` (source, fixed, and destination), and `funding_schedules`.
- Created indexes on `fixed_bank_account_id` and `source_bank_account_id`.
- Migrations tested successfully and `schema_test.go` updated to expect version `2026060800`.

### Phase 2: Go Models & Repository Layer
*Status: Completed*
*Key Details for subsequent phases:*
- Created Go Model `models.FixedDeposit` mapping to `fixed_deposits` table:
  - Fields: `FixedDepositId ID[FixedDeposit]`, `AccountId ID[Account]`, `SourceBankAccountId ID[BankAccount]`, `BankAccountId ID[BankAccount]` (mapped to pg column `fixed_bank_account_id`), `FundingScheduleId *ID[FundingSchedule]`, `Name string`, `Amount int64`, `InterestRate float64`, `StartDate time.Time`, `EndDate time.Time`, `InterestFrequency string`, `InterestDestination string`, `InterestDestinationBankAccountId *ID[BankAccount]`, `Status FixedDepositStatus`, `CreatedAt`, `UpdatedAt`.
- Added repository methods to `BaseRepository` interface and `repositoryBase`:
  - `GetFixedDeposits(ctx context.Context, bankAccountId ID[BankAccount]) ([]FixedDeposit, error)`
  - `GetFixedDepositById(ctx context.Context, bankAccountId ID[BankAccount], id ID[FixedDeposit]) (*FixedDeposit, error)`
  - `CreateFixedDeposit(ctx context.Context, fixedDeposit *FixedDeposit) error`
  - `UpdateFixedDeposit(ctx context.Context, fixedDeposit *FixedDeposit) error`
  - `DeleteFixedDeposit(ctx context.Context, bankAccountId ID[BankAccount], id ID[FixedDeposit]) error`
- Created maturation background job: `ProcessFixedDepositMaturitiesCron` registered in `server/jobs/jobs.go` to run hourly. Handles active deposits where `end_date` has passed, transfers balances back to `source_bank_account_id`, deactivates the CD bank account, clears and deletes the recurring interest funding schedule, and marks status as `matured`.

### Phase 3: Controllers & REST APIs
*Status: Completed*
*Key Details for subsequent phases:*
- Registered API routes under the authenticated/billed routing group (`server/controller/routes.go`):
  - `GET /bank_accounts/:bankAccountId/fixed_deposits`: returns all fixed deposits for the bank account
  - `GET /bank_accounts/:bankAccountId/fixed_deposits/:fixedDepositId`: returns details for a specific fixed deposit
  - `POST /bank_accounts/:bankAccountId/fixed_deposits`: creates a fixed deposit (validates name, amount > 0, termMonths > 0, frequency, destination; creates CD bank account; initiates manual debit/credit transfer transactions; updates balances; schedules interest payout `FundingSchedule` under the CD bank account if interestRate > 0)
  - `POST /bank_accounts/:bankAccountId/fixed_deposits/:fixedDepositId/withdraw`: early withdrawal (returns balance/principal to source account, sets CD bank account status to inactive/deleted, updates fixed deposit status to `withdrawn`, deletes the interest schedule)
  - `DELETE /bank_accounts/:bankAccountId/fixed_deposits/:fixedDepositId`: deletes the historical fixed deposit record from the database (does not roll back transactions)
- Fixed an unrelated routing regression in `routes.go` by restoring the missing `DELETE /api/bank_accounts/:bankAccountId/spending/:spendingId` route so `TestDeleteSpending` tests pass cleanly.

### Phase 4: Frontend API Clients & Hooks
*Status: Completed*
*Key Details for subsequent phases:*
- Created `FixedDeposit` domain model in `interface/src/models/FixedDeposit.ts` representing the Fixed Deposit structure and relations (including `fixedBankAccount`).
- Implemented hooks in `interface/src/hooks/useFixedDeposits.ts`:
  - `useFixedDeposits(bankAccountId)`: Query hook fetching all fixed deposits for a bank account.
  - `useCreateFixedDeposit(bankAccountId)`: Mutation hook for opening a new fixed deposit. Invalidates fixed deposits and bank accounts query data.
  - `useWithdrawFixedDeposit(bankAccountId)`: Mutation hook for early withdrawal of a fixed deposit. Invalidates fixed deposits and bank accounts query data.
  - `useDeleteFixedDeposit(bankAccountId)`: Mutation hook for deleting historical fixed deposit records.

### Phase 5: User Interface and Modals
*Status: Completed*
*Key Details for successive phases:*
- Created a robust Fixed Deposits & CDs management dashboard at `/bank/:bankAccountId/fixed_deposits`.
- Integrated a real-time account Analytics Overview section showing Free Balance, Active Fixed Deposits total, Combined Net Worth, and Projected Worth including expected interest payouts at maturity.
- Designed `NewFixedDepositModal.tsx` allowing users to configure term length, rates, start date, payout frequency, and payout destination account.
- Provided elegant visual cues including progress bars for active deposit maturity tracker and clear state listings.
