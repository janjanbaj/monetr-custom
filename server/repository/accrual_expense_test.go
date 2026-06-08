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

func TestRepositoryBase_AccrualExpense_CreateAndRead(t *testing.T) {
	t.Run("create and get by id", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		expense := &models.AccrualExpense{
			BankAccountId: bankAccount.BankAccountId,
			Name:          "Bulk Rice Purchase",
			Description:   "50 lb bag of jasmine rice",
			Amount:        5000, // $50.00 in cents
			StartDate:     now,
			EndDate:       now.Add(90 * 24 * time.Hour),
		}

		err := repo.CreateAccrualExpense(context.Background(), expense)
		require.NoError(t, err, "should create accrual expense without error")
		require.NotEmpty(t, expense.AccrualExpenseId, "ID should be populated after insert")

		fetched, err := repo.GetAccrualExpenseById(context.Background(), bankAccount.BankAccountId, expense.AccrualExpenseId)
		require.NoError(t, err, "should retrieve accrual expense by ID")
		assert.Equal(t, expense.AccrualExpenseId, fetched.AccrualExpenseId)
		assert.Equal(t, "Bulk Rice Purchase", fetched.Name)
		assert.Equal(t, int64(5000), fetched.Amount)
	})

	t.Run("list expenses for bank account", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		for _, name := range []string{"Rice", "Toiletries", "Annual Subscription"} {
			err := repo.CreateAccrualExpense(context.Background(), &models.AccrualExpense{
				BankAccountId: bankAccount.BankAccountId,
				Name:          name,
				Amount:        1000,
				StartDate:     now,
				EndDate:       now.Add(30 * 24 * time.Hour),
			})
			require.NoError(t, err, "should create expense %s", name)
		}

		expenses, err := repo.GetAccrualExpenses(context.Background(), bankAccount.BankAccountId)
		require.NoError(t, err, "should list accrual expenses")
		assert.Len(t, expenses, 3, "should have 3 expenses")
	})

	t.Run("does not return expenses for a different bank account", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount1 := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		bankAccount2 := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		require.NoError(t, repo.CreateAccrualExpense(context.Background(), &models.AccrualExpense{
			BankAccountId: bankAccount1.BankAccountId,
			Name:          "Expense for Account 1",
			Amount:        1000,
			StartDate:     now,
			EndDate:       now.Add(30 * 24 * time.Hour),
		}))

		expenses, err := repo.GetAccrualExpenses(context.Background(), bankAccount2.BankAccountId)
		require.NoError(t, err)
		assert.Empty(t, expenses, "should not return expenses from a different bank account")
	})
}

func TestRepositoryBase_AccrualExpense_Update(t *testing.T) {
	t.Run("update name and amount", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		expense := &models.AccrualExpense{
			BankAccountId: bankAccount.BankAccountId,
			Name:          "Original Name",
			Amount:        1000,
			StartDate:     now,
			EndDate:       now.Add(30 * 24 * time.Hour),
		}
		require.NoError(t, repo.CreateAccrualExpense(context.Background(), expense))

		expense.Name = "Updated Name"
		expense.Amount = 2500
		err := repo.UpdateAccrualExpense(context.Background(), bankAccount.BankAccountId, expense)
		require.NoError(t, err, "should update accrual expense")

		fetched, err := repo.GetAccrualExpenseById(context.Background(), bankAccount.BankAccountId, expense.AccrualExpenseId)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", fetched.Name)
		assert.Equal(t, int64(2500), fetched.Amount)
	})
}

func TestRepositoryBase_AccrualExpense_Delete(t *testing.T) {
	t.Run("delete expense", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		expense := &models.AccrualExpense{
			BankAccountId: bankAccount.BankAccountId,
			Name:          "To Delete",
			Amount:        500,
			StartDate:     now,
			EndDate:       now.Add(30 * 24 * time.Hour),
		}
		require.NoError(t, repo.CreateAccrualExpense(context.Background(), expense))

		err := repo.DeleteAccrualExpense(context.Background(), bankAccount.BankAccountId, expense.AccrualExpenseId)
		require.NoError(t, err, "should delete accrual expense")

		expenses, err := repo.GetAccrualExpenses(context.Background(), bankAccount.BankAccountId)
		require.NoError(t, err)
		assert.Empty(t, expenses, "no expenses should remain after deletion")
	})

	t.Run("delete nonexistent expense returns error", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		fakeId := models.NewID[models.AccrualExpense]()
		err := repo.DeleteAccrualExpense(context.Background(), bankAccount.BankAccountId, fakeId)
		assert.Error(t, err, "deleting a nonexistent expense should return an error")
	})
}

func TestRepositoryBase_AccrualUsageLog_CRUD(t *testing.T) {
	t.Run("create and list usage logs", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		expense := &models.AccrualExpense{
			BankAccountId: bankAccount.BankAccountId,
			Name:          "Toiletries",
			Amount:        3000,
			StartDate:     now,
			EndDate:       now.Add(60 * 24 * time.Hour),
		}
		require.NoError(t, repo.CreateAccrualExpense(context.Background(), expense))

		usageLog1 := &models.AccrualUsageLog{
			AccrualExpenseId: expense.AccrualExpenseId,
			BankAccountId:    bankAccount.BankAccountId,
			Amount:           500,
			Date:             now.Add(24 * time.Hour),
		}
		usageLog2 := &models.AccrualUsageLog{
			AccrualExpenseId: expense.AccrualExpenseId,
			BankAccountId:    bankAccount.BankAccountId,
			Amount:           750,
			Date:             now.Add(48 * time.Hour),
		}

		require.NoError(t, repo.CreateAccrualUsageLog(context.Background(), usageLog1))
		require.NoError(t, repo.CreateAccrualUsageLog(context.Background(), usageLog2))
		require.NotEmpty(t, usageLog1.AccrualUsageLogId)
		require.NotEmpty(t, usageLog2.AccrualUsageLogId)

		logs, err := repo.GetAccrualUsageLogs(context.Background(), bankAccount.BankAccountId, expense.AccrualExpenseId)
		require.NoError(t, err, "should retrieve usage logs")
		assert.Len(t, logs, 2, "should have 2 usage logs")
	})

	t.Run("usage logs loaded with expense relation", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		expense := &models.AccrualExpense{
			BankAccountId: bankAccount.BankAccountId,
			Name:          "Soap",
			Amount:        1500,
			StartDate:     now,
			EndDate:       now.Add(30 * 24 * time.Hour),
		}
		require.NoError(t, repo.CreateAccrualExpense(context.Background(), expense))

		usageLog := &models.AccrualUsageLog{
			AccrualExpenseId: expense.AccrualExpenseId,
			BankAccountId:    bankAccount.BankAccountId,
			Amount:           300,
			Date:             now.Add(5 * 24 * time.Hour),
		}
		require.NoError(t, repo.CreateAccrualUsageLog(context.Background(), usageLog))

		// Fetch expense and confirm UsageLogs relation is populated
		fetched, err := repo.GetAccrualExpenseById(context.Background(), bankAccount.BankAccountId, expense.AccrualExpenseId)
		require.NoError(t, err)
		require.Len(t, fetched.UsageLogs, 1, "expense should carry its usage logs via relation")
		assert.Equal(t, usageLog.AccrualUsageLogId, fetched.UsageLogs[0].AccrualUsageLogId)
	})

	t.Run("delete usage log", func(t *testing.T) {
		clk := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clk)
		link := fixtures.GivenIHaveAManualLink(t, clk, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clk, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clk, user.UserId, user.AccountId, db, log)

		now := clk.Now()
		expense := &models.AccrualExpense{
			BankAccountId: bankAccount.BankAccountId,
			Name:          "Paper Towels",
			Amount:        2000,
			StartDate:     now,
			EndDate:       now.Add(45 * 24 * time.Hour),
		}
		require.NoError(t, repo.CreateAccrualExpense(context.Background(), expense))

		usageLog := &models.AccrualUsageLog{
			AccrualExpenseId: expense.AccrualExpenseId,
			BankAccountId:    bankAccount.BankAccountId,
			Amount:           200,
			Date:             now.Add(3 * 24 * time.Hour),
		}
		require.NoError(t, repo.CreateAccrualUsageLog(context.Background(), usageLog))

		err := repo.DeleteAccrualUsageLog(context.Background(), bankAccount.BankAccountId, usageLog.AccrualUsageLogId)
		require.NoError(t, err, "should delete usage log")

		logs, err := repo.GetAccrualUsageLogs(context.Background(), bankAccount.BankAccountId, expense.AccrualExpenseId)
		require.NoError(t, err)
		assert.Empty(t, logs, "no usage logs should remain after deletion")
	})
}
