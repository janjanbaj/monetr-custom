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
		Relation("BankAccount").
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
		Relation("BankAccount").
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
