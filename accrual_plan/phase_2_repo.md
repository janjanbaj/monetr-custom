# Phase 2: Go Backend Models & Repository Layer

This phase defines the Go models for accrual expenses and usage logs, and implements database queries using the repository pattern.

## Context Directories
- [server/models/](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/models)
- [server/repository/](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository)

---

## 1. Model Structures

#### [NEW] [accrual_expense.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/models/accrual_expense.go)
```go
package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type AccrualExpense struct {
	tableName string `pg:"accrual_expenses"`

	AccrualExpenseId ID[AccrualExpense] `json:"accrualExpenseId" pg:"accrual_expense_id,notnull,pk"`
	AccountId        ID[Account]        `json:"-" pg:"account_id,notnull,pk"`
	Account          *Account           `json:"-" pg:"rel:has-one"`
	BankAccountId    ID[BankAccount]    `json:"bankAccountId" pg:"bank_account_id,notnull,pk"`
	BankAccount      *BankAccount       `json:"-" pg:"rel:has-one"`
	TransactionId    *ID[Transaction]   `json:"transactionId,omitempty" pg:"transaction_id,on_delete:SET NULL"`
	Transaction      *Transaction       `json:"transaction,omitempty" pg:"rel:has-one"`
	Name             string             `json:"name" pg:"name,notnull"`
	Description      string             `json:"description,omitempty" pg:"description"`
	Amount           int64              `json:"amount" pg:"amount,notnull,use_zero"`
	StartDate        time.Time          `json:"startDate" pg:"start_date,notnull"`
	EndDate          time.Time          `json:"endDate" pg:"end_date,notnull"`
	CreatedAt        time.Time          `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt        time.Time          `json:"updatedAt" pg:"updated_at,notnull"`

	// Relations (loaded on demand)
	UsageLogs []AccrualUsageLog `json:"usageLogs,omitempty" pg:"rel:has-many,fk:accrual_expense_id"`
}

func (AccrualExpense) IdentityPrefix() string {
	return "acex"
}

var _ pg.BeforeInsertHook = (*AccrualExpense)(nil)
var _ pg.BeforeUpdateHook = (*AccrualExpense)(nil)

func (o *AccrualExpense) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.AccrualExpenseId.IsZero() {
		o.AccrualExpenseId = NewID[AccrualExpense]()
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

func (o *AccrualExpense) BeforeUpdate(ctx context.Context) (context.Context, error) {
	o.UpdatedAt = time.Now()
	return ctx, nil
}
```

#### [NEW] [accrual_usage_log.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/models/accrual_usage_log.go)
```go
package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type AccrualUsageLog struct {
	tableName string `pg:"accrual_usage_logs"`

	AccrualUsageLogId ID[AccrualUsageLog] `json:"accrualUsageLogId" pg:"accrual_usage_log_id,notnull,pk"`
	AccrualExpenseId   ID[AccrualExpense]  `json:"accrualExpenseId" pg:"accrual_expense_id,notnull"`
	AccountId          ID[Account]         `json:"-" pg:"account_id,notnull,pk"`
	BankAccountId      ID[BankAccount]     `json:"bankAccountId" pg:"bank_account_id,notnull,pk"`
	Amount             int64               `json:"amount" pg:"amount,notnull,use_zero"`
	Date               time.Time           `json:"date" pg:"date,notnull"`
	CreatedAt          time.Time           `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt          time.Time           `json:"updatedAt" pg:"updated_at,notnull"`
}

func (AccrualUsageLog) IdentityPrefix() string {
	return "aclg"
}

var _ pg.BeforeInsertHook = (*AccrualUsageLog)(nil)
var _ pg.BeforeUpdateHook = (*AccrualUsageLog)(nil)

func (o *AccrualUsageLog) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.AccrualUsageLogId.IsZero() {
		o.AccrualUsageLogId = NewID[AccrualUsageLog]()
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

func (o *AccrualUsageLog) BeforeUpdate(ctx context.Context) (context.Context, error) {
	o.UpdatedAt = time.Now()
	return ctx, nil
}
```

---

## 2. Repository Interface Updates

#### [MODIFY] [repository.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/repository.go)
Append these methods to `BaseRepository` interface:
```go
	GetAccrualExpenses(ctx context.Context, bankAccountId ID[BankAccount]) ([]AccrualExpense, error)
	GetAccrualExpenseById(ctx context.Context, bankAccountId ID[BankAccount], id ID[AccrualExpense]) (*AccrualExpense, error)
	CreateAccrualExpense(ctx context.Context, expense *AccrualExpense) error
	UpdateAccrualExpense(ctx context.Context, bankAccountId ID[BankAccount], expense *AccrualExpense) error
	DeleteAccrualExpense(ctx context.Context, bankAccountId ID[BankAccount], id ID[AccrualExpense]) error

	GetAccrualUsageLogs(ctx context.Context, bankAccountId ID[BankAccount], expenseId ID[AccrualExpense]) ([]AccrualUsageLog, error)
	CreateAccrualUsageLog(ctx context.Context, log *AccrualUsageLog) error
	DeleteAccrualUsageLog(ctx context.Context, bankAccountId ID[BankAccount], logId ID[AccrualUsageLog]) error
```

---

## 3. Repository Implementation

#### [NEW] [accrual_expense.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/accrual_expense.go)
```go
package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetAccrualExpenses(ctx context.Context, bankAccountId ID[BankAccount]) ([]AccrualExpense, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]AccrualExpense, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("UsageLogs").
		Where(`"accrual_expense"."account_id" = ?`, r.AccountId()).
		Where(`"accrual_expense"."bank_account_id" = ?`, bankAccountId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve accrual expenses")
	}
	span.Status = sentry.SpanStatusOK
	return result, nil
}

func (r *repositoryBase) GetAccrualExpenseById(ctx context.Context, bankAccountId ID[BankAccount], id ID[AccrualExpense]) (*AccrualExpense, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result AccrualExpense
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("UsageLogs").
		Where(`"accrual_expense"."account_id" = ?`, r.AccountId()).
		Where(`"accrual_expense"."bank_account_id" = ?`, bankAccountId).
		Where(`"accrual_expense"."accrual_expense_id" = ?`, id).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve accrual expense by ID")
	}
	span.Status = sentry.SpanStatusOK
	return &result, nil
}

func (r *repositoryBase) CreateAccrualExpense(ctx context.Context, expense *AccrualExpense) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	expense.AccountId = r.AccountId()
	expense.CreatedAt = r.clock.Now().UTC()
	expense.UpdatedAt = expense.CreatedAt

	if _, err := r.txn.ModelContext(span.Context(), expense).Insert(expense); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create accrual expense")
	}
	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) UpdateAccrualExpense(ctx context.Context, bankAccountId ID[BankAccount], expense *AccrualExpense) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	expense.AccountId = r.AccountId()
	expense.BankAccountId = bankAccountId

	_, err := r.txn.ModelContext(span.Context(), expense).
		WherePK().
		Update(expense)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update accrual expense")
	}
	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) DeleteAccrualExpense(ctx context.Context, bankAccountId ID[BankAccount], id ID[AccrualExpense]) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result, err := r.txn.ModelContext(span.Context(), &AccrualExpense{}).
		Where(`"accrual_expense"."account_id" = ?`, r.AccountId()).
		Where(`"accrual_expense"."bank_account_id" = ?`, bankAccountId).
		Where(`"accrual_expense"."accrual_expense_id" = ?`, id).
		Delete()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to delete accrual expense")
	}
	if result.RowsAffected() != 1 {
		span.Status = sentry.SpanStatusDataLoss
		return errors.Errorf("invalid number of accrual expenses deleted: %d", result.RowsAffected())
	}
	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) GetAccrualUsageLogs(ctx context.Context, bankAccountId ID[BankAccount], expenseId ID[AccrualExpense]) ([]AccrualUsageLog, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]AccrualUsageLog, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"accrual_usage_log"."account_id" = ?`, r.AccountId()).
		Where(`"accrual_usage_log"."bank_account_id" = ?`, bankAccountId).
		Where(`"accrual_usage_log"."accrual_expense_id" = ?`, expenseId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve usage logs")
	}
	span.Status = sentry.SpanStatusOK
	return result, nil
}

func (r *repositoryBase) CreateAccrualUsageLog(ctx context.Context, log *AccrualUsageLog) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log.AccountId = r.AccountId()
	log.CreatedAt = r.clock.Now().UTC()
	log.UpdatedAt = log.CreatedAt

	if _, err := r.txn.ModelContext(span.Context(), log).Insert(log); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create usage log")
	}
	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) DeleteAccrualUsageLog(ctx context.Context, bankAccountId ID[BankAccount], logId ID[AccrualUsageLog]) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result, err := r.txn.ModelContext(span.Context(), &AccrualUsageLog{}).
		Where(`"accrual_usage_log"."account_id" = ?`, r.AccountId()).
		Where(`"accrual_usage_log"."bank_account_id" = ?`, bankAccountId).
		Where(`"accrual_usage_log"."accrual_usage_log_id" = ?`, logId).
		Delete()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to delete usage log")
	}
	if result.RowsAffected() != 1 {
		span.Status = sentry.SpanStatusDataLoss
		return errors.Errorf("invalid number of usage logs deleted: %d", result.RowsAffected())
	}
	span.Status = sentry.SpanStatusOK
	return nil
}
```

---

## Verification Plan

### Automated Tests

#### [NEW] [accrual_expense_test.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/accrual_expense_test.go)
Create tests for CRUD operations on accrual expenses and usage logs:
```go
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Add test cases asserting Create, Read, Update, Delete for both tables.
// Ensure relationships cascade correctly (e.g. deleting AccrualExpense drops its UsageLogs).
```

Run test suite:
```bash
go test ./server/repository/... -run TestAccrualExpense -v
```
Ensure all tests pass and models serialize/deserialize correctly.

---

## ✅ Phase 2 Complete — Recap for Next Agent

Phase 2 has been fully implemented. Here is a summary for agents implementing Phase 3+:

### What was created

| File | Purpose |
|------|---------|
| [`server/models/accrual_expense.go`](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/models/accrual_expense.go) | `AccrualExpense` model with `IdentityPrefix() = "acex"`, go-pg `BeforeInsert`/`BeforeUpdate` hooks |
| [`server/models/accrual_usage_log.go`](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/models/accrual_usage_log.go) | `AccrualUsageLog` model with `IdentityPrefix() = "aclg"`, same hooks |
| [`server/repository/repository.go`](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/repository.go) | 8 new methods added to `BaseRepository` interface (5 for expenses, 3 for usage logs) |
| [`server/repository/accrual_expense.go`](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/accrual_expense.go) | Full CRUD implementation on `repositoryBase` using crumbs/sentry spans |
| [`server/repository/accrual_expense_test.go`](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/repository/accrual_expense_test.go) | Tests for all CRUD ops including relation loading |

### Key design decisions
- `AccrualExpense` has a composite PK: `(accrual_expense_id, account_id, bank_account_id)` matching the DB schema from Phase 1
- `AccrualUsageLog` similarly has `(accrual_usage_log_id, account_id, bank_account_id)` as composite PK
- `UsageLogs []AccrualUsageLog` is a has-many relation on `AccrualExpense`, loaded via `.Relation("UsageLogs")` in Get queries
- All repository methods filter by `account_id` (from `r.AccountId()`) and `bank_account_id` for security isolation
- Timestamps are set via `r.clock.Now().UTC()` (deterministic for tests), complemented by go-pg `BeforeInsert` hooks

### Repository interface methods added to `BaseRepository`
```go
GetAccrualExpenses(ctx, bankAccountId) ([]AccrualExpense, error)
GetAccrualExpenseById(ctx, bankAccountId, id) (*AccrualExpense, error)
CreateAccrualExpense(ctx, expense *AccrualExpense) error
UpdateAccrualExpense(ctx, bankAccountId, expense *AccrualExpense) error
DeleteAccrualExpense(ctx, bankAccountId, id) error
GetAccrualUsageLogs(ctx, bankAccountId, expenseId) ([]AccrualUsageLog, error)
CreateAccrualUsageLog(ctx, log *AccrualUsageLog) error
DeleteAccrualUsageLog(ctx, bankAccountId, logId) error
```
