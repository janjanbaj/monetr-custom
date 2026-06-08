package funding_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/funding/funding_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProcessFixedDepositMaturitiesCron(t *testing.T) {
	t.Run("no fixed deposits to process", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		mockCtx := mockgen.NewMockContext(ctrl)
		mockCtx.EXPECT().Clock().Return(clock).AnyTimes()
		mockCtx.EXPECT().DB().Return(db).AnyTimes()
		mockCtx.EXPECT().Log().Return(log).AnyTimes()

		err := funding_jobs.ProcessFixedDepositMaturitiesCron(
			mockqueue.NewMockContext(mockCtx),
		)
		assert.NoError(t, err)
	})

	t.Run("matured fixed deposits are processed", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		sourceAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fixedAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		// Set initial balances
		sourceAccount.AvailableBalance = 1000
		sourceAccount.CurrentBalance = 1000
		fixedAccount.AvailableBalance = 5000
		fixedAccount.CurrentBalance = 5000

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db, log)
		require.NoError(t, repo.UpdateBankAccount(t.Context(), &sourceAccount))
		require.NoError(t, repo.UpdateBankAccount(t.Context(), &fixedAccount))

		// Create funding schedule for interest
		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		fundingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", clock.Now())
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              user.AccountId,
			BankAccountId:          fixedAccount.BankAccountId,
			Name:                   "FD Interest",
			Description:            "FD Interest",
			RuleSet:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		// Create Fixed Deposit
		startDate := clock.Now().UTC()
		endDate := clock.Now().Add(30 * 24 * time.Hour).UTC()
		fd := testutils.MustInsert(t, models.FixedDeposit{
			AccountId:           user.AccountId,
			SourceBankAccountId: sourceAccount.BankAccountId,
			BankAccountId:       fixedAccount.BankAccountId,
			FundingScheduleId:   &fundingSchedule.FundingScheduleId,
			Name:                "FD Test",
			Amount:              5000,
			InterestRate:        0.05,
			StartDate:           startDate,
			EndDate:             endDate,
			InterestFrequency:   "monthly",
			InterestDestination: "payout",
			Status:              models.FixedDepositStatusActive,
		})

		// Move time forward past end date
		clock.Set(endDate.Add(1 * time.Hour))

		mockCtx := mockgen.NewMockContext(ctrl)
		mockCtx.EXPECT().Clock().Return(clock).AnyTimes()
		mockCtx.EXPECT().DB().Return(db).AnyTimes()
		mockCtx.EXPECT().Log().Return(log).AnyTimes()
		mockCtx.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		err := funding_jobs.ProcessFixedDepositMaturitiesCron(
			mockqueue.NewMockContext(mockCtx),
		)
		require.NoError(t, err)

		// Verify fixed deposit is marked matured
		updatedFd, err := repo.GetFixedDepositById(t.Context(), sourceAccount.BankAccountId, fd.FixedDepositId)
		require.NoError(t, err)
		assert.Equal(t, models.FixedDepositStatusMatured, updatedFd.Status)

		// Verify source bank account balance updated (+5000)
		updatedSource, err := repo.GetBankAccount(t.Context(), sourceAccount.BankAccountId)
		require.NoError(t, err)
		assert.Equal(t, int64(6000), updatedSource.CurrentBalance)
		assert.Equal(t, int64(6000), updatedSource.AvailableBalance)

		// Verify fixed bank account balance is 0, inactive, deleted_at is set
		updatedFixed, err := repo.GetBankAccount(t.Context(), fixedAccount.BankAccountId)
		require.NoError(t, err)
		assert.Equal(t, int64(0), updatedFixed.CurrentBalance)
		assert.Equal(t, int64(0), updatedFixed.AvailableBalance)
		assert.Equal(t, models.BankAccountStatusInactive, updatedFixed.Status)
		assert.NotNil(t, updatedFixed.DeletedAt)
		assert.True(t, updatedFixed.DeletedAt.Equal(endDate))

		// Verify funding schedule deleted
		_, err = repo.GetFundingSchedule(t.Context(), fixedAccount.BankAccountId, fundingSchedule.FundingScheduleId)
		assert.Error(t, err)

		// Verify transactions created:
		// Debit on fixedAccount
		// Credit on sourceAccount
		var txs []models.Transaction
		err = db.ModelContext(t.Context(), &txs).
			Where("bank_account_id = ?", fixedAccount.BankAccountId).
			Select(&txs)
		require.NoError(t, err)
		require.Len(t, txs, 1)
		assert.Equal(t, int64(5000), txs[0].Amount)
		assert.Equal(t, "Maturity Transfer to Source Account", txs[0].Name)

		err = db.ModelContext(t.Context(), &txs).
			Where("bank_account_id = ?", sourceAccount.BankAccountId).
			Select(&txs)
		require.NoError(t, err)
		require.Len(t, txs, 1)
		assert.Equal(t, int64(-5000), txs[0].Amount)
		assert.Equal(t, "Fixed Deposit Matured: FD Test", txs[0].Name)
	})
}
