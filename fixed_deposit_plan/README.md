# Certificate of Deposit (CD) and Fixed Deposit (FD) Tracking Plan

This directory contains self-contained implementation phases to extend Monetr with support for tracking Fixed Deposits and Certificates of Deposits.

## Architectural Overview
Fixed Deposits are financial assets that grow through locked terms. They are modeled as follows:
1. **The CD Account**: A temporary bank account (`BankAccount` of subtype `cd`) which receives the initial principal from a checking/savings account.
2. **Transfer of Principal**: Two manual transactions (source debit, CD credit) move the money.
3. **Interest Payments**: A `FundingSchedule` with `AutoCreateTransaction` enabled is created to deposit monthly, quarterly, or end-of-term interest back to either the source checking/savings account (payout) or the CD account itself (accumulating).
4. **UTC Calculation**: All date arithmetic (start dates, recurrences, and maturity end dates) are done in UTC to ensure timezones remain consistent if the user travels.
5. **Maturity / Cancel**: An hourly cron job monitors active deposits, transferring funds back and deactivating accounts on maturity. Early withdrawal deactivates the account and cancels the schedules; users manually record penalties.

---

## Plan Status
Refer to [status.md](file:///Users/janeetbajracharya/Desktop/Code/monetr/fixed_deposit_plan/status.md) to check progress and debrief status.

## Phases
1. [Phase 1: DB Schema & Migrations](file:///Users/janeetbajracharya/Desktop/Code/monetr/fixed_deposit_plan/phase_1_db.md)
2. [Phase 2: Models & Repository Layer](file:///Users/janeetbajracharya/Desktop/Code/monetr/fixed_deposit_plan/phase_2_backend.md)
3. [Phase 3: Controllers & REST APIs](file:///Users/janeetbajracharya/Desktop/Code/monetr/fixed_deposit_plan/phase_3_controllers.md)
4. [Phase 4: Frontend API Clients & Hooks](file:///Users/janeetbajracharya/Desktop/Code/monetr/fixed_deposit_plan/phase_4_frontend_api.md)
5. [Phase 5: User Interface and Modals](file:///Users/janeetbajracharya/Desktop/Code/monetr/fixed_deposit_plan/phase_5_ui.md)
