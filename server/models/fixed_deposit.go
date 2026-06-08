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

	FixedDepositId                 ID[FixedDeposit]     `json:"fixedDepositId" pg:"fixed_deposit_id,notnull,pk"`
	AccountId                      ID[Account]          `json:"-" pg:"account_id,notnull,pk"`
	Account                        *Account             `json:"-" pg:"rel:has-one"`
	SourceBankAccountId            ID[BankAccount]      `json:"sourceBankAccountId" pg:"source_bank_account_id,notnull"`
	SourceBankAccount              *BankAccount         `json:"sourceBankAccount,omitempty" pg:"rel:has-one,fk:source_"`
	BankAccountId                  ID[BankAccount]      `json:"fixedBankAccountId" pg:"fixed_bank_account_id,notnull"`
	BankAccount                    *BankAccount         `json:"fixedBankAccount,omitempty" pg:"rel:has-one,fk:fixed_"`
	FundingScheduleId              *ID[FundingSchedule] `json:"fundingScheduleId,omitempty" pg:"funding_schedule_id,on_delete:SET NULL"`
	FundingSchedule                *FundingSchedule     `json:"fundingSchedule,omitempty" pg:"rel:has-one,fk:fixed_"`
	Name                           string               `json:"name" pg:"name,notnull"`
	Amount                         int64                `json:"amount" pg:"amount,notnull,use_zero"`
	InterestRate                   float64              `json:"interestRate" pg:"interest_rate,notnull,use_zero"`
	StartDate                      time.Time            `json:"startDate" pg:"start_date,notnull"`
	EndDate                        time.Time            `json:"endDate" pg:"end_date,notnull"`
	InterestFrequency              string               `json:"interestFrequency" pg:"interest_frequency,notnull"` // "monthly", "quarterly", "end_of_term"
	InterestDestination            string               `json:"interestDestination" pg:"interest_destination,notnull"` // "accumulate", "payout"
	InterestDestinationBankAccountId *ID[BankAccount]     `json:"interestDestinationAccountId,omitempty" pg:"interest_destination_bank_account_id"`
	InterestDestinationAccount     *BankAccount         `json:"interestDestinationAccount,omitempty" pg:"rel:has-one,fk:interest_destination_"`
	Status                         FixedDepositStatus   `json:"status" pg:"status,notnull"`
	CreatedAt                      time.Time            `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt                      time.Time            `json:"updatedAt" pg:"updated_at,notnull"`
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
