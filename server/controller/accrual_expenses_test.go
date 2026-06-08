package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/fixtures"
	. "github.com/monetr/monetr/server/models"
)

func TestAccrualExpense(t *testing.T) {
	t.Run("GET accrual list is authenticated", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		// Call without cookie
		response := e.GET("/api/bank_accounts/{bankAccountId}/accrual").
			WithPath("bankAccountId", bank.BankAccountId).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("post validation failures", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		now := time.Now()

		// 1. Empty name
		response := e.POST("/api/bank_accounts/{bankAccountId}/accrual").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":      "",
				"amount":    1000,
				"startDate": now.Format(time.RFC3339),
				"endDate":   now.Add(24 * time.Hour).Format(time.RFC3339),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("accrual expense must have a name")

		// 2. Negative/zero amount
		response = e.POST("/api/bank_accounts/{bankAccountId}/accrual").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":      "Toiletries",
				"amount":    0,
				"startDate": now.Format(time.RFC3339),
				"endDate":   now.Add(24 * time.Hour).Format(time.RFC3339),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("accrual expense amount must be greater than 0")

		// 3. Start date after end date
		response = e.POST("/api/bank_accounts/{bankAccountId}/accrual").
			WithPath("bankAccountId", bank.BankAccountId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":      "Toiletries",
				"amount":    1000,
				"startDate": now.Add(24 * time.Hour).Format(time.RFC3339),
				"endDate":   now.Format(time.RFC3339),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("start date cannot be after end date")
	})

	t.Run("CRUD happy path", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		startDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

		var expenseId string

		{ // Create
			response := e.POST("/api/bank_accounts/{bankAccountId}/accrual").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":        "Toiletries Bulk",
					"description": "Costco toiletries run",
					"amount":      15000,
					"startDate":   startDate.Format(time.RFC3339),
					"endDate":     endDate.Format(time.RFC3339),
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.accrualExpenseId").String().NotEmpty()
			response.JSON().Path("$.bankAccountId").String().IsEqual(bank.BankAccountId.String())
			response.JSON().Path("$.name").String().IsEqual("Toiletries Bulk")
			response.JSON().Path("$.description").String().IsEqual("Costco toiletries run")
			response.JSON().Path("$.amount").Number().IsEqual(15000)
			response.JSON().Path("$.startDate").String().NotEmpty()
			response.JSON().Path("$.endDate").String().NotEmpty()

			expenseId = response.JSON().Path("$.accrualExpenseId").String().Raw()
		}

		{ // List
			response := e.GET("/api/bank_accounts/{bankAccountId}/accrual").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(1)
			response.JSON().Path("$[0].accrualExpenseId").String().IsEqual(expenseId)
			response.JSON().Path("$[0].name").String().IsEqual("Toiletries Bulk")
		}

		{ // Get by ID
			response := e.GET("/api/bank_accounts/{bankAccountId}/accrual/{accrualExpenseId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("accrualExpenseId", expenseId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.accrualExpenseId").String().IsEqual(expenseId)
			response.JSON().Path("$.name").String().IsEqual("Toiletries Bulk")
		}

		{ // Update
			response := e.PUT("/api/bank_accounts/{bankAccountId}/accrual/{accrualExpenseId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("accrualExpenseId", expenseId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":        "Toiletries Bulk Updated",
					"description": "Costco toiletries run updated",
					"amount":      20000,
					"startDate":   startDate.Add(24 * time.Hour).Format(time.RFC3339),
					"endDate":     endDate.Add(24 * time.Hour).Format(time.RFC3339),
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.accrualExpenseId").String().IsEqual(expenseId)
			response.JSON().Path("$.name").String().IsEqual("Toiletries Bulk Updated")
			response.JSON().Path("$.description").String().IsEqual("Costco toiletries run updated")
			response.JSON().Path("$.amount").Number().IsEqual(20000)
		}

		{ // Update validation checks (StartDate after EndDate)
			response := e.PUT("/api/bank_accounts/{bankAccountId}/accrual/{accrualExpenseId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("accrualExpenseId", expenseId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"startDate": endDate.Add(48 * time.Hour).Format(time.RFC3339),
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("start date cannot be after end date")
		}

		{ // Create usage log (exempt)
			response := e.POST("/api/bank_accounts/{bankAccountId}/accrual/{accrualExpenseId}/usage_logs").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("accrualExpenseId", expenseId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"amount": 0,
					"date":   startDate.Add(48 * time.Hour).Format(time.RFC3339),
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.accrualUsageLogId").String().NotEmpty()
			response.JSON().Path("$.amount").Number().IsEqual(0)
			response.JSON().Path("$.accrualExpenseId").String().IsEqual(expenseId)

			logId := response.JSON().Path("$.accrualUsageLogId").String().Raw()

			// Get expense and verify usage log relation is returned
			respGet := e.GET("/api/bank_accounts/{bankAccountId}/accrual/{accrualExpenseId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("accrualExpenseId", expenseId).
				WithCookie(TestCookieName, token).
				Expect()

			respGet.Status(http.StatusOK)
			respGet.JSON().Path("$.usageLogs").Array().Length().IsEqual(1)
			respGet.JSON().Path("$.usageLogs[0].accrualUsageLogId").String().IsEqual(logId)

			// Delete usage log
			respDel := e.DELETE("/api/bank_accounts/{bankAccountId}/usage_logs/{accrualUsageLogId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("accrualUsageLogId", logId).
				WithCookie(TestCookieName, token).
				Expect()

			respDel.Status(http.StatusOK)
			respDel.Body().IsEmpty()
		}

		{ // Delete
			response := e.DELETE("/api/bank_accounts/{bankAccountId}/accrual/{accrualExpenseId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("accrualExpenseId", expenseId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		{ // List after deletion
			response := e.GET("/api/bank_accounts/{bankAccountId}/accrual").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}
	})
}
