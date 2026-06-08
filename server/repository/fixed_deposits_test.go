package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryBase_FixedDeposits(t *testing.T) {
	clock := clock.NewMock()
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

	user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
	link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
	sourceAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
	fixedAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

	repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db, log)

	t.Run("Create and Get By ID", func(t *testing.T) {
		fd := models.FixedDeposit{
			SourceBankAccountId: sourceAccount.BankAccountId,
			BankAccountId:       fixedAccount.BankAccountId,
			Name:                "FD 1",
			Amount:              500000,
			InterestRate:        0.05,
			StartDate:           clock.Now().UTC(),
			EndDate:             clock.Now().Add(365 * 24 * time.Hour).UTC(),
			InterestFrequency:   "monthly",
			InterestDestination: "payout",
			Status:              models.FixedDepositStatusActive,
		}

		err := repo.CreateFixedDeposit(context.Background(), &fd)
		require.NoError(t, err)
		assert.NotZero(t, fd.FixedDepositId)

		fetched, err := repo.GetFixedDepositById(context.Background(), sourceAccount.BankAccountId, fd.FixedDepositId)
		require.NoError(t, err)
		assert.Equal(t, fd.Name, fetched.Name)
		assert.Equal(t, fd.Amount, fetched.Amount)
		assert.Equal(t, fd.Status, fetched.Status)
		assert.Equal(t, sourceAccount.BankAccountId, fetched.SourceBankAccountId)
		assert.Equal(t, fixedAccount.BankAccountId, fetched.BankAccountId)
	})

	t.Run("Get All", func(t *testing.T) {
		fds, err := repo.GetFixedDeposits(context.Background(), sourceAccount.BankAccountId)
		require.NoError(t, err)
		assert.Len(t, fds, 1)
	})

	t.Run("Update", func(t *testing.T) {
		fds, err := repo.GetFixedDeposits(context.Background(), sourceAccount.BankAccountId)
		require.NoError(t, err)
		require.Len(t, fds, 1)

		fd := fds[0]
		fd.Name = "Updated Name"
		fd.Status = models.FixedDepositStatusMatured

		err = repo.UpdateFixedDeposit(context.Background(), &fd)
		require.NoError(t, err)

		fetched, err := repo.GetFixedDepositById(context.Background(), sourceAccount.BankAccountId, fd.FixedDepositId)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", fetched.Name)
		assert.Equal(t, models.FixedDepositStatusMatured, fetched.Status)
	})

	t.Run("Delete", func(t *testing.T) {
		fds, err := repo.GetFixedDeposits(context.Background(), sourceAccount.BankAccountId)
		require.NoError(t, err)
		require.Len(t, fds, 1)

		err = repo.DeleteFixedDeposit(context.Background(), sourceAccount.BankAccountId, fds[0].FixedDepositId)
		require.NoError(t, err)

		fdsAfter, err := repo.GetFixedDeposits(context.Background(), sourceAccount.BankAccountId)
		require.NoError(t, err)
		assert.Len(t, fdsAfter, 0)
	})
}
