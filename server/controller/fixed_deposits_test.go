package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/migrations"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixedDepositController(t *testing.T) {
	// Ensure migrations are run on the test database
	db := testutils.GetPgDatabase(t)
	log := testutils.GetLog(t)
	migrations.RunMigrations(t.Context(), log, db)

	t.Run("GET list and GET by ID are authenticated", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		// List without token
		e.GET("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			Expect().
			Status(http.StatusUnauthorized)

		// Detail without token
		e.GET("/api/bank_accounts/{bankAccountId}/fixed_deposits/fxdp_0123456789abcdef01234567").
			WithPath("bankAccountId", bank.BankAccountId).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("validation failures", func(t *testing.T) {
		app, e := NewTestApplication(t)
		db := testutils.GetPgDatabase(t)
		log := testutils.GetLog(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		// Set initial balance
		repo := repository.NewRepositoryFromSession(app.Clock, user.UserId, user.AccountId, db, log)
		bank.AvailableBalance = 10000
		bank.CurrentBalance = 10000
		require.NoError(t, repo.UpdateBankAccount(t.Context(), &bank))

		// 1. Empty name
		e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                "",
				"amount":              5000,
				"interestRate":        5.5,
				"termMonths":          12,
				"interestFrequency":   "monthly",
				"interestDestination": "accumulate",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Path("$.error").String().IsEqual("Fixed deposit must have a name")

		// 2. Non-positive amount
		e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                "CD 1",
				"amount":              0,
				"interestRate":        5.5,
				"termMonths":          12,
				"interestFrequency":   "monthly",
				"interestDestination": "accumulate",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Path("$.error").String().IsEqual("Fixed deposit amount must be greater than 0")

		// 3. Non-positive term
		e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                "CD 1",
				"amount":              5000,
				"interestRate":        5.5,
				"termMonths":          0,
				"interestFrequency":   "monthly",
				"interestDestination": "accumulate",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Path("$.error").String().IsEqual("Term months must be greater than 0")

		// 4. Invalid interest frequency
		e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                "CD 1",
				"amount":              5000,
				"interestRate":        5.5,
				"termMonths":          12,
				"interestFrequency":   "yearly",
				"interestDestination": "accumulate",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Path("$.error").String().IsEqual("interest frequency must be monthly, quarterly, or end_of_term")

		// 5. Invalid interest destination
		e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                "CD 1",
				"amount":              5000,
				"interestRate":        5.5,
				"termMonths":          12,
				"interestFrequency":   "monthly",
				"interestDestination": "somewhere_else",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Path("$.error").String().IsEqual("interest destination must be accumulate or payout")

		// 6. Payout destination missing destination ID
		e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                "CD 1",
				"amount":              5000,
				"interestRate":        5.5,
				"termMonths":          12,
				"interestFrequency":   "monthly",
				"interestDestination": "payout",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Path("$.error").String().IsEqual("must specify a valid interest destination account ID for payout")

		// 7. Insufficient balance
		e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                "CD 1",
				"amount":              20000,
				"interestRate":        5.5,
				"termMonths":          12,
				"interestFrequency":   "monthly",
				"interestDestination": "accumulate",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Path("$.error").String().IsEqual("Insufficient available balance in source account to fund the fixed deposit")
	})

	t.Run("Create, List, Details, Withdraw, Delete - Accumulate Path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		db := testutils.GetPgDatabase(t)
		log := testutils.GetLog(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		// Set initial balance
		repo := repository.NewRepositoryFromSession(app.Clock, user.UserId, user.AccountId, db, log)
		bank.AvailableBalance = 10000
		bank.CurrentBalance = 10000
		require.NoError(t, repo.UpdateBankAccount(t.Context(), &bank))

		startDate := app.Clock.Now().UTC()

		var fixedDepositId string
		var fixedBankAccountId string

		{ // 1. Create Fixed Deposit
			response := e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":                "My Accumulate CD",
					"amount":              5000,
					"interestRate":        4.5,
					"startDate":           startDate.Format(time.RFC3339),
					"termMonths":          12,
					"interestFrequency":   "monthly",
					"interestDestination": "accumulate",
				}).
				Expect().
				Status(http.StatusOK)

			response.JSON().Path("$.fixedDepositId").String().NotEmpty()
			response.JSON().Path("$.sourceBankAccountId").String().IsEqual(bank.BankAccountId.String())
			response.JSON().Path("$.fixedBankAccountId").String().NotEmpty()
			response.JSON().Path("$.fundingScheduleId").String().NotEmpty()
			response.JSON().Path("$.name").String().IsEqual("My Accumulate CD")
			response.JSON().Path("$.amount").Number().IsEqual(5000)
			response.JSON().Path("$.interestRate").Number().IsEqual(4.5)
			response.JSON().Path("$.interestFrequency").String().IsEqual("monthly")
			response.JSON().Path("$.interestDestination").String().IsEqual("accumulate")
			response.JSON().Path("$.status").String().IsEqual("active")

			fixedDepositId = response.JSON().Path("$.fixedDepositId").String().Raw()
			fixedBankAccountId = response.JSON().Path("$.fixedBankAccountId").String().Raw()

			// Verify source account balance is deducted
			src, err := repo.GetBankAccount(t.Context(), bank.BankAccountId)
			require.NoError(t, err)
			assert.Equal(t, int64(5000), src.AvailableBalance)
			assert.Equal(t, int64(5000), src.CurrentBalance)

			// Verify CD account created with correct type/balance
			cdId, err := ParseID[BankAccount](fixedBankAccountId)
			require.NoError(t, err)
			cd, err := repo.GetBankAccount(t.Context(), cdId)
			require.NoError(t, err)
			assert.Equal(t, "Fixed Deposit: My Accumulate CD", cd.Name)
			assert.Equal(t, CDBankAccountSubType, cd.AccountSubType)
			assert.Equal(t, DepositoryBankAccountType, cd.AccountType)
			assert.Equal(t, int64(5000), cd.AvailableBalance)
			assert.Equal(t, int64(5000), cd.CurrentBalance)

			// Verify transactions exist
			// Debit transaction on source bank account
			var sourceTxs []Transaction
			err = db.ModelContext(t.Context(), &sourceTxs).
				Where("bank_account_id = ?", bank.BankAccountId).
				Select(&sourceTxs)
			require.NoError(t, err)
			require.Len(t, sourceTxs, 1)
			assert.Equal(t, int64(5000), sourceTxs[0].Amount)
			assert.Equal(t, "Transfer to FD: My Accumulate CD", sourceTxs[0].Name)

			// Credit transaction on CD account
			var cdTxs []Transaction
			err = db.ModelContext(t.Context(), &cdTxs).
				Where("bank_account_id = ?", cd.BankAccountId).
				Select(&cdTxs)
			require.NoError(t, err)
			require.Len(t, cdTxs, 1)
			assert.Equal(t, int64(-5000), cdTxs[0].Amount)
			assert.Equal(t, "Initial Funding", cdTxs[0].Name)
		}

		{ // 2. GET list
			response := e.GET("/api/bank_accounts/{bankAccountId}/fixed_deposits").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect().
				Status(http.StatusOK)

			response.JSON().Array().Length().IsEqual(1)
			response.JSON().Path("$[0].fixedDepositId").String().IsEqual(fixedDepositId)
		}

		{ // 3. GET details
			response := e.GET("/api/bank_accounts/{bankAccountId}/fixed_deposits/{fixedDepositId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("fixedDepositId", fixedDepositId).
				WithCookie(TestCookieName, token).
				Expect().
				Status(http.StatusOK)

			response.JSON().Path("$.fixedDepositId").String().IsEqual(fixedDepositId)
			response.JSON().Path("$.name").String().IsEqual("My Accumulate CD")
		}

		{ // 4. Withdraw early
			response := e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits/{fixedDepositId}/withdraw").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("fixedDepositId", fixedDepositId).
				WithCookie(TestCookieName, token).
				Expect().
				Status(http.StatusOK)

			response.JSON().Path("$.fixedDepositId").String().IsEqual(fixedDepositId)
			response.JSON().Path("$.status").String().IsEqual("withdrawn")

			// Source account balance should be restored to 10000
			src, err := repo.GetBankAccount(t.Context(), bank.BankAccountId)
			require.NoError(t, err)
			assert.Equal(t, int64(10000), src.AvailableBalance)
			assert.Equal(t, int64(10000), src.CurrentBalance)

			// CD account should be inactive, 0 balance, and deleted_at populated
			cdId, _ := ParseID[BankAccount](fixedBankAccountId)
			cd, err := repo.GetBankAccount(t.Context(), cdId)
			require.NoError(t, err)
			assert.Equal(t, int64(0), cd.AvailableBalance)
			assert.Equal(t, int64(0), cd.CurrentBalance)
			assert.Equal(t, BankAccountStatusInactive, cd.Status)
			assert.NotNil(t, cd.DeletedAt)

			// Transaction created to return money:
			// Debit on CD account (+5000), credit on source account (-5000)
			var sourceTxs []Transaction
			err = db.ModelContext(t.Context(), &sourceTxs).
				Where("bank_account_id = ?", bank.BankAccountId).
				Order("created_at ASC").
				Select(&sourceTxs)
			require.NoError(t, err)
			require.Len(t, sourceTxs, 2)
			assert.Equal(t, int64(-5000), sourceTxs[1].Amount)
			assert.Equal(t, "Fixed Deposit Cancelled: My Accumulate CD", sourceTxs[1].Name)
		}

		{ // 5. Delete historical record
			e.DELETE("/api/bank_accounts/{bankAccountId}/fixed_deposits/{fixedDepositId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("fixedDepositId", fixedDepositId).
				WithCookie(TestCookieName, token).
				Expect().
				Status(http.StatusOK)

			// Verify GET detail fails or record is not in DB anymore
			fdId, _ := ParseID[FixedDeposit](fixedDepositId)
			_, err := repo.GetFixedDepositById(t.Context(), bank.BankAccountId, fdId)
			assert.Error(t, err)
		}
	})

	t.Run("Create with Payout Destination", func(t *testing.T) {
		app, e := NewTestApplication(t)
		db := testutils.GetPgDatabase(t)
		log := testutils.GetLog(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		payoutDestBank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, SavingsBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		// Set initial balance
		repo := repository.NewRepositoryFromSession(app.Clock, user.UserId, user.AccountId, db, log)
		bank.AvailableBalance = 10000
		bank.CurrentBalance = 10000
		require.NoError(t, repo.UpdateBankAccount(t.Context(), &bank))

		startDate := app.Clock.Now().UTC()

		response := e.POST("/api/bank_accounts/{bankAccountId}/fixed_deposits").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":                         "My Payout CD",
				"amount":                       8000,
				"interestRate":                 5.0,
				"startDate":                    startDate.Format(time.RFC3339),
				"termMonths":                   6,
				"interestFrequency":            "quarterly",
				"interestDestination":          "payout",
				"interestDestinationAccountId": payoutDestBank.BankAccountId,
			}).
			Expect().
			Status(http.StatusOK)

		response.JSON().Path("$.interestDestinationAccountId").String().IsEqual(payoutDestBank.BankAccountId.String())
		response.JSON().Path("$.interestDestination").String().IsEqual("payout")
		schedIdStr := response.JSON().Path("$.fundingScheduleId").String().Raw()

		// Verify schedule is created under CD bank account ID
		schedId, err := ParseID[FundingSchedule](schedIdStr)
		require.NoError(t, err)

		fixedBankAccountIdStr := response.JSON().Path("$.fixedBankAccountId").String().Raw()
		fixedBankAccountId, err := ParseID[BankAccount](fixedBankAccountIdStr)
		require.NoError(t, err)

		sched, err := repo.GetFundingSchedule(t.Context(), fixedBankAccountId, schedId)
		require.NoError(t, err)
		assert.Equal(t, fixedBankAccountId, sched.BankAccountId)
		assert.Equal(t, "Interest Payment - My Payout CD", sched.Name)
		assert.Equal(t, int64(100), *sched.EstimatedDeposit) // 8000 * 0.05 / 4 = 100
	})
}
