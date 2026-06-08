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
