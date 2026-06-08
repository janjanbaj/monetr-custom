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

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

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

	span.Data = map[string]any{
		"accountId":        r.AccountId(),
		"bankAccountId":    bankAccountId,
		"accrualExpenseId": id,
	}

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
	span.Data = map[string]any{
		"accountId":        r.AccountId(),
		"bankAccountId":    expense.BankAccountId,
		"accrualExpenseId": expense.AccrualExpenseId,
	}
	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) UpdateAccrualExpense(ctx context.Context, bankAccountId ID[BankAccount], expense *AccrualExpense) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	expense.AccountId = r.AccountId()
	expense.BankAccountId = bankAccountId
	expense.UpdatedAt = r.clock.Now().UTC()

	span.Data = map[string]any{
		"accountId":        r.AccountId(),
		"bankAccountId":    bankAccountId,
		"accrualExpenseId": expense.AccrualExpenseId,
	}

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

	span.Data = map[string]any{
		"accountId":        r.AccountId(),
		"bankAccountId":    bankAccountId,
		"accrualExpenseId": id,
	}

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

	span.Data = map[string]any{
		"accountId":        r.AccountId(),
		"bankAccountId":    bankAccountId,
		"accrualExpenseId": expenseId,
	}

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
	span.Data = map[string]any{
		"accountId":         r.AccountId(),
		"bankAccountId":     log.BankAccountId,
		"accrualExpenseId":  log.AccrualExpenseId,
		"accrualUsageLogId": log.AccrualUsageLogId,
	}
	span.Status = sentry.SpanStatusOK
	return nil
}

func (r *repositoryBase) DeleteAccrualUsageLog(ctx context.Context, bankAccountId ID[BankAccount], logId ID[AccrualUsageLog]) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":         r.AccountId(),
		"bankAccountId":     bankAccountId,
		"accrualUsageLogId": logId,
	}

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
