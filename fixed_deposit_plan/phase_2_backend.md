# Phase 2: Models & Repository Layer

This phase covers creating the `FixedDeposit` model and implementing data access methods in the repository layer. It also implements the maturation background job.

## Target Files
1. [fixed_deposit.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/models/fixed_deposit.go) [NEW]
2. [repository.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/repository.go) [MODIFY]
3. [fixed_deposits.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/fixed_deposits.go) [NEW]
4. [process_fixed_deposits.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/funding/funding_jobs/process_fixed_deposits.go) [NEW]
5. [jobs.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/jobs/jobs.go) [MODIFY]

---

## 1. Go Model Definition
Create `server/models/fixed_deposit.go`:

```go
package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type FixedDepositStatus string

const (
	FixedDepositStatusActive    FixedDepositStatus = "active"
	FixedDepositStatusMatured   FixedDepositStatus = "matured"
	FixedDepositStatusWithdrawn FixedDepositStatus = "withdrawn"
)

type FixedDeposit struct {
	tableName string `pg:"fixed_deposits"`

	FixedDepositId               ID[FixedDeposit]     `json:"fixedDepositId" pg:"fixed_deposit_id,notnull,pk"`
	AccountId                    ID[Account]          `json:"-" pg:"account_id,notnull,pk"`
	Account                      *Account             `json:"-" pg:"rel:has-one"`
	SourceBankAccountId          ID[BankAccount]      `json:"sourceBankAccountId" pg:"source_bank_account_id,notnull"`
	SourceBankAccount            *BankAccount         `json:"sourceBankAccount,omitempty" pg:"rel:has-one,fk:source_bank_account_id"`
	FixedBankAccountId           ID[BankAccount]      `json:"fixedBankAccountId" pg:"fixed_bank_account_id,notnull"`
	FixedBankAccount             *BankAccount         `json:"fixedBankAccount,omitempty" pg:"rel:has-one,fk:fixed_bank_account_id"`
	FundingScheduleId            *ID[FundingSchedule] `json:"fundingScheduleId,omitempty" pg:"funding_schedule_id,on_delete:SET NULL"`
	FundingSchedule              *FundingSchedule     `json:"fundingSchedule,omitempty" pg:"rel:has-one"`
	Name                         string               `json:"name" pg:"name,notnull"`
	Amount                       int64                `json:"amount" pg:"amount,notnull,use_zero"`
	InterestRate                 float64              `json:"interestRate" pg:"interest_rate,notnull,use_zero"`
	StartDate                    time.Time            `json:"startDate" pg:"start_date,notnull"`
	EndDate                      time.Time            `json:"endDate" pg:"end_date,notnull"`
	InterestFrequency            string               `json:"interestFrequency" pg:"interest_frequency,notnull"` // "monthly", "quarterly", "end_of_term"
	InterestDestination          string               `json:"interestDestination" pg:"interest_destination,notnull"` // "accumulate", "payout"
	InterestDestinationAccountId *ID[BankAccount]     `json:"interestDestinationAccountId,omitempty" pg:"interest_destination_account_id"`
	InterestDestinationAccount   *BankAccount         `json:"interestDestinationAccount,omitempty" pg:"rel:has-one,fk:interest_destination_account_id"`
	Status                       FixedDepositStatus   `json:"status" pg:"status,notnull"`
	CreatedAt                    time.Time            `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt                    time.Time            `json:"updatedAt" pg:"updated_at,notnull"`
}

func (FixedDeposit) IdentityPrefix() string {
	return "fxdp"
}

var _ pg.BeforeInsertHook = (*FixedDeposit)(nil)
var _ pg.BeforeUpdateHook = (*FixedDeposit)(nil)

func (o *FixedDeposit) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.FixedDepositId.IsZero() {
		o.FixedDepositId = NewID[FixedDeposit]()
	}
	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}
	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = now
	}
	return ctx, nil
}

func (o *FixedDeposit) BeforeUpdate(ctx context.Context) (context.Context, error) {
	o.UpdatedAt = time.Now()
	return ctx, nil
}
```

---

## 2. Repository Layer
### Modify `server/repository/repository.go`
Add the interface definitions under the `BaseRepository` type:
```go
	// Fixed Deposits
	GetFixedDeposits(ctx context.Context, bankAccountId ID[BankAccount]) ([]FixedDeposit, error)
	GetFixedDepositById(ctx context.Context, bankAccountId ID[BankAccount], id ID[FixedDeposit]) (*FixedDeposit, error)
	CreateFixedDeposit(ctx context.Context, fixedDeposit *FixedDeposit) error
	UpdateFixedDeposit(ctx context.Context, fixedDeposit *FixedDeposit) error
	DeleteFixedDeposit(ctx context.Context, bankAccountId ID[BankAccount], id ID[FixedDeposit]) error
```

### Implement Repository `server/repository/fixed_deposits.go`
Create the file and write queries:
```go
package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetFixedDeposits(ctx context.Context, bankAccountId ID[BankAccount]) ([]FixedDeposit, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]FixedDeposit, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("FixedBankAccount").
		Where(`"fixed_deposit"."account_id" = ?`, r.AccountId()).
		Where(`"fixed_deposit"."source_bank_account_id" = ? OR "fixed_deposit"."fixed_bank_account_id" = ?`, bankAccountId, bankAccountId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve fixed deposits")
	}

	span.Status = sentry.SpanStatusOK
	return result, nil
}

func (r *repositoryBase) GetFixedDepositById(ctx context.Context, bankAccountId ID[BankAccount], id ID[FixedDeposit]) (*FixedDeposit, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result FixedDeposit
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("FixedBankAccount").
		Relation("SourceBankAccount").
		Relation("FundingSchedule").
		Where(`"fixed_deposit"."account_id" = ?`, r.AccountId()).
		Where(`"fixed_deposit"."fixed_deposit_id" = ?`, id).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve fixed deposit")
	}

	span.Status = sentry.SpanStatusOK
	return &result, nil
}

func (r *repositoryBase) CreateFixedDeposit(ctx context.Context, fixedDeposit *FixedDeposit) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	fixedDeposit.AccountId = r.AccountId()
	fixedDeposit.CreatedAt = r.clock.Now().UTC()
	fixedDeposit.UpdatedAt = r.clock.Now().UTC()

	if _, err := r.txn.ModelContext(span.Context(), fixedDeposit).Insert(fixedDeposit); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create fixed deposit")
	}

	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) UpdateFixedDeposit(ctx context.Context, fixedDeposit *FixedDeposit) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	fixedDeposit.AccountId = r.AccountId()
	fixedDeposit.UpdatedAt = r.clock.Now().UTC()

	_, err := r.txn.ModelContext(span.Context(), fixedDeposit).
		WherePK().
		Update(fixedDeposit)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update fixed deposit")
	}

	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) DeleteFixedDeposit(ctx context.Context, bankAccountId ID[BankAccount], id ID[FixedDeposit]) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &FixedDeposit{}).
		Where(`"fixed_deposit"."account_id" = ?`, r.AccountId()).
		Where(`"fixed_deposit"."fixed_deposit_id" = ?`, id).
		Delete()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to delete fixed deposit")
	}

	span.Status = sentry.SpanStatusOK
	return nil
}
```

---

## 3. Maturation Background Job
Create the job that checks for active FDs where the `end_date` in UTC has elapsed.

Create `server/funding/funding_jobs/process_fixed_deposits.go`:

```go
package funding_jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

func ProcessFixedDepositMaturitiesCron(ctx queue.Context) error {
	log := ctx.Log()
	log.InfoContext(ctx, "checking for matured fixed deposits")

	// Query active fixed deposits whose end_date has passed
	var maturedDeposits []models.FixedDeposit
	err := ctx.DB().ModelContext(ctx, &maturedDeposits).
		Relation("FixedBankAccount").
		Where(`"fixed_deposit"."status" = ?`, models.FixedDepositStatusActive).
		Where(`"fixed_deposit"."end_date" <= ?`, ctx.Clock().Now().UTC()).
		Select(&maturedDeposits)
	if err != nil {
		return errors.Wrap(err, "failed to fetch matured fixed deposits")
	}

	if len(maturedDeposits) == 0 {
		log.InfoContext(ctx, "no fixed deposits matured at this time")
		return nil
	}

	log.InfoContext(ctx, fmt.Sprintf("processing %d matured fixed deposits", len(maturedDeposits)))

	for _, fd := range maturedDeposits {
		err := ctx.RunInTransaction(ctx, func(txCtx queue.Context) error {
			repo := repository.NewRepositoryFromSession(
				txCtx.Clock(),
				"user_system",
				fd.AccountId,
				txCtx.DB(),
				txCtx.Log(),
			)

			// Load the fixed bank account to get the current balance (accumulated principal + interest)
			fixedAccount, err := repo.GetBankAccount(txCtx, fd.FixedBankAccountId)
			if err != nil {
				return errors.Wrap(err, "failed to load fixed bank account")
			}

			// Balance to transfer back
			balanceToReturn := fixedAccount.CurrentBalance
			if balanceToReturn <= 0 {
				balanceToReturn = fd.Amount // fallback
			}

			// 1. Create debit transaction on Fixed CD account (money leaving)
			txnDebit := models.Transaction{
				BankAccountId: fd.FixedBankAccountId,
				Amount:        balanceToReturn,
				Date:          fd.EndDate,
				Name:          fmt.Sprintf("Maturity Transfer to Source Account"),
				OriginalName:  "Maturity Transfer",
				IsPending:     false,
				Source:        models.TransactionSourceManual,
			}

			// 2. Create credit transaction on Source account (money entering)
			txnCredit := models.Transaction{
				BankAccountId: fd.SourceBankAccountId,
				Amount:        -balanceToReturn,
				Date:          fd.EndDate,
				Name:          fmt.Sprintf("Fixed Deposit Matured: %s", fd.Name),
				OriginalName:  fd.Name,
				IsPending:     false,
				Source:        models.TransactionSourceManual,
			}

			if err := repo.CreateTransaction(txCtx, fd.FixedBankAccountId, &txnDebit); err != nil {
				return errors.Wrap(err, "failed to create debit transaction on CD account")
			}
			if err := repo.CreateTransaction(txCtx, fd.SourceBankAccountId, &txnCredit); err != nil {
				return errors.Wrap(err, "failed to create credit transaction on source account")
			}

			// Update account balances
			sourceAccount, err := repo.GetBankAccount(txCtx, fd.SourceBankAccountId)
			if err == nil {
				sourceAccount.AvailableBalance += balanceToReturn
				sourceAccount.CurrentBalance += balanceToReturn
				_ = repo.UpdateBankAccount(txCtx, sourceAccount)
			}

			fixedAccount.AvailableBalance = 0
			fixedAccount.CurrentBalance = 0
			fixedAccount.Status = models.BankAccountStatusInactive
			fixedAccount.DeletedAt = &fd.EndDate
			if err := repo.UpdateBankAccount(txCtx, fixedAccount); err != nil {
				return errors.Wrap(err, "failed to deactivate CD bank account")
			}

			// Delete recurring funding schedule if interest schedule exists
			if fd.FundingScheduleId != nil {
				_ = repo.DeleteFundingSchedule(txCtx, fd.FixedBankAccountId, *fd.FundingScheduleId)
			}

			// Mark fixed deposit as matured
			fd.Status = models.FixedDepositStatusMatured
			if err := repo.UpdateFixedDeposit(txCtx, &fd); err != nil {
				return errors.Wrap(err, "failed to update fixed deposit status")
			}

			crumbs.Info(txCtx, "processed matured fixed deposit successfully", map[string]any{
				"fixedDepositId": fd.FixedDepositId,
				"amountReturned": balanceToReturn,
			})

			return nil
		})

		if err != nil {
			log.ErrorContext(ctx, "failed to process matured deposit", "fixedDepositId", fd.FixedDepositId, "err", err)
			continue
		}
	}

	return nil
}
```

### Register the Cron Job in `server/jobs/jobs.go`
Modify `RegisterJobs` to register our new background job to run hourly:
```go
		queue.RegisterCron(ctx, processor, funding_jobs.ProcessFixedDepositMaturitiesCron, "0 0 * * * *"),
```
Also register `ProcessFixedDepositMaturities` if you run it asynchronously.

---

## Verification Plan
- Write Go repository unit tests verifying CRUD for `FixedDeposit` objects.
- Mock database values and execute `ProcessFixedDepositMaturitiesCron` verifying proper transfers are created, account deactivated, and funding schedule deleted.
