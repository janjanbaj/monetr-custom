package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type AccrualUsageLog struct {
	tableName string `pg:"accrual_usage_logs"`

	AccrualUsageLogId ID[AccrualUsageLog] `json:"accrualUsageLogId" pg:"accrual_usage_log_id,notnull,pk"`
	AccrualExpenseId  ID[AccrualExpense]  `json:"accrualExpenseId" pg:"accrual_expense_id,notnull"`
	AccountId         ID[Account]         `json:"-" pg:"account_id,notnull,pk"`
	BankAccountId     ID[BankAccount]     `json:"bankAccountId" pg:"bank_account_id,notnull,pk"`
	Amount            int64               `json:"amount" pg:"amount,notnull,use_zero"`
	Date              time.Time           `json:"date" pg:"date,notnull"`
	CreatedAt         time.Time           `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt         time.Time           `json:"updatedAt" pg:"updated_at,notnull"`
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
