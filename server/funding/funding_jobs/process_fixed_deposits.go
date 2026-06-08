package funding_jobs

import (
	"fmt"

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
		Relation("BankAccount").
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
			fixedAccount, err := repo.GetBankAccount(txCtx, fd.BankAccountId)
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
				BankAccountId: fd.BankAccountId,
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

			if err := repo.CreateTransaction(txCtx, fd.BankAccountId, &txnDebit); err != nil {
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
				scheduleId := *fd.FundingScheduleId
				fd.FundingScheduleId = nil
				if err := repo.UpdateFixedDeposit(txCtx, &fd); err != nil {
					return errors.Wrap(err, "failed to clear funding schedule id from fixed deposit")
				}
				log.InfoContext(txCtx, "attempting to delete funding schedule", "scheduleId", scheduleId, "bankAccountId", fd.BankAccountId)
				if err := repo.DeleteFundingSchedule(txCtx, fd.BankAccountId, scheduleId); err != nil {
					log.ErrorContext(txCtx, "failed to delete funding schedule", "err", err)
					return errors.Wrap(err, "failed to delete funding schedule")
				}
			} else {
				log.InfoContext(txCtx, "funding schedule id is nil")
			}

			// Mark fixed deposit as matured
			fd.Status = models.FixedDepositStatusMatured
			if err := repo.UpdateFixedDeposit(txCtx, &fd); err != nil {
				return errors.Wrap(err, "failed to update fixed deposit status")
			}

			crumbs.Debug(txCtx, "processed matured fixed deposit successfully", map[string]any{
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
