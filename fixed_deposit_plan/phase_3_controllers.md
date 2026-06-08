# Phase 3: Controllers & REST APIs

This phase covers registering routes and implementing REST controllers to handle fixed deposit operations (creation, retrieval, cancellation/early withdrawal, deletion).

## Target Files
1. [routes.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/controller/routes.go) [MODIFY]
2. [fixed_deposits.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/controller/fixed_deposits.go) [NEW]

---

## 1. Register Routes in `server/controller/routes.go`
Add endpoints to the `billed` routing group:
```go
	// Fixed Deposits
	billed.GET("/bank_accounts/:bankAccountId/fixed_deposits", c.getFixedDeposits)
	billed.GET("/bank_accounts/:bankAccountId/fixed_deposits/:fixedDepositId", c.getFixedDepositById)
	billed.POST("/bank_accounts/:bankAccountId/fixed_deposits", c.postFixedDeposit)
	billed.POST("/bank_accounts/:bankAccountId/fixed_deposits/:fixedDepositId/withdraw", c.postFixedDepositWithdraw)
	billed.DELETE("/bank_accounts/:bankAccountId/fixed_deposits/:fixedDepositId", c.deleteFixedDeposit)
```

---

## 2. Implement Controller `server/controller/fixed_deposits.go`
Create the REST controller:

```go
package controller

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type NewFixedDepositRequest struct {
	Name                         string     `json:"name"`
	Amount                       int64      `json:"amount"`
	InterestRate                 float64    `json:"interestRate"`
	StartDate                    time.Time  `json:"startDate"`
	TermMonths                   int        `json:"termMonths"`
	InterestFrequency            string     `json:"interestFrequency"`            // "monthly", "quarterly", "end_of_term"
	InterestDestination          string     `json:"interestDestination"`          // "accumulate", "payout"
	InterestDestinationAccountId *ID[BankAccount] `json:"interestDestinationAccountId"` // required if payout
}

func (c *Controller) getFixedDeposits(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	deposits, err := repo.GetFixedDeposits(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to get fixed deposits")
	}

	return ctx.JSON(http.StatusOK, deposits)
}

func (c *Controller) getFixedDepositById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fixedDepositId, err := ParseID[FixedDeposit](ctx.Param("fixedDepositId"))
	if err != nil || fixedDepositId.IsZero() {
		return c.badRequest(ctx, "must specify a valid fixed deposit Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	deposit, err := repo.GetFixedDepositById(c.getContext(ctx), bankAccountId, fixedDepositId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to get fixed deposit details")
	}

	return ctx.JSON(http.StatusOK, deposit)
}

func (c *Controller) postFixedDeposit(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	var req NewFixedDepositRequest
	if err := ctx.Bind(&req); err != nil {
		return c.invalidJson(ctx)
	}

	if req.Name == "" {
		return c.badRequest(ctx, "Fixed deposit must have a name")
	}
	if req.Amount <= 0 {
		return c.badRequest(ctx, "Fixed deposit amount must be greater than 0")
	}
	if req.TermMonths <= 0 {
		return c.badRequest(ctx, "Term months must be greater than 0")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Fetch source bank account
	sourceAccount, err := repo.GetBankAccount(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to find source bank account")
	}

	if sourceAccount.AvailableBalance < req.Amount {
		return c.badRequest(ctx, "Insufficient available balance in source account to fund the fixed deposit")
	}

	// 1. Calculate dates in UTC to keep consistent with relocation
	startUTC := req.StartDate.UTC()
	endUTC := startUTC.AddDate(0, req.TermMonths, 0)

	// 2. Create the temporary CD bank account
	cdAccount := BankAccount{
		LinkId:         sourceAccount.LinkId,
		Currency:       sourceAccount.Currency,
		Name:           fmt.Sprintf("Fixed Deposit: %s", req.Name),
		OriginalName:   fmt.Sprintf("Fixed Deposit: %s", req.Name),
		AccountType:    DepositoryBankAccountType,
		AccountSubType: CDBankAccountSubType,
		Status:         BankAccountStatusActive,
		LastUpdated:    c.Clock.Now().UTC(),
	}

	if err := repo.CreateBankAccounts(c.getContext(ctx), &cdAccount); err != nil {
		return c.wrapPgError(ctx, err, "failed to create CD bank account")
	}

	// 3. Perform the principal transfer (Transactions)
	txnDebit := Transaction{
		BankAccountId: sourceAccount.BankAccountId,
		Amount:        req.Amount, // positive means money leaving
		Date:          startUTC,
		Name:          fmt.Sprintf("Transfer to FD: %s", req.Name),
		OriginalName:  "Fixed Deposit Transfer",
		IsPending:     false,
		Source:        TransactionSourceManual,
	}

	txnCredit := Transaction{
		BankAccountId: cdAccount.BankAccountId,
		Amount:        -req.Amount, // negative means money entering
		Date:          startUTC,
		Name:          "Initial Funding",
		OriginalName:  "Fixed Deposit Funding",
		IsPending:     false,
		Source:        TransactionSourceManual,
	}

	if err := repo.CreateTransaction(c.getContext(ctx), sourceAccount.BankAccountId, &txnDebit); err != nil {
		return c.wrapPgError(ctx, err, "failed to create debit transaction on source account")
	}
	if err := repo.CreateTransaction(c.getContext(ctx), cdAccount.BankAccountId, &txnCredit); err != nil {
		return c.wrapPgError(ctx, err, "failed to create credit transaction on CD account")
	}

	// Update account balances
	sourceAccount.AvailableBalance -= req.Amount
	sourceAccount.CurrentBalance -= req.Amount
	_ = repo.UpdateBankAccount(c.getContext(ctx), sourceAccount)

	cdAccount.AvailableBalance = req.Amount
	cdAccount.CurrentBalance = req.Amount
	_ = repo.UpdateBankAccount(c.getContext(ctx), &cdAccount)

	// 4. Create interest payout FundingSchedule if rate > 0
	var scheduleId *ID[FundingSchedule]
	if req.InterestRate > 0 {
		var interestDestId ID[BankAccount]
		if req.InterestDestination == "payout" && req.InterestDestinationAccountId != nil {
			interestDestId = *req.InterestDestinationAccountId
		} else {
			interestDestId = cdAccount.BankAccountId // default is accumulate
		}

		// Calculate interest payment amount based on frequency
		var interestAmount int64
		var count int
		var interval int
		var ruleStartDate time.Time

		switch req.InterestFrequency {
		case "monthly":
			interestAmount = int64(math.Round(float64(req.Amount) * (req.InterestRate / 100.0) / 12.0))
			count = req.TermMonths
			interval = 1
			ruleStartDate = startUTC.AddDate(0, 1, 0)
		case "quarterly":
			interestAmount = int64(math.Round(float64(req.Amount) * (req.InterestRate / 100.0) / 4.0))
			count = req.TermMonths / 3
			if count < 1 {
				count = 1
			}
			interval = 3
			ruleStartDate = startUTC.AddDate(0, 3, 0)
		case "end_of_term":
			interestAmount = int64(math.Round(float64(req.Amount) * (req.InterestRate / 100.0) * (float64(req.TermMonths) / 12.0)))
			count = 1
			interval = 1
			ruleStartDate = endUTC
		}

		if interestAmount > 0 {
			// Generate RRule string
			rruleStr := fmt.Sprintf("DTSTART:%s\nRRULE:FREQ=MONTHLY;INTERVAL=%d;COUNT=%d",
				ruleStartDate.Format("20060102T150405Z"), interval, count)

			ruleset, err := NewRuleSet(rruleStr)
			if err == nil {
				interestSchedule := FundingSchedule{
					BankAccountId:         interestDestId,
					Name:                  fmt.Sprintf("Interest Payment - %s", req.Name),
					Description:           fmt.Sprintf("Interest on %s Fixed Deposit", req.Name),
					RuleSet:               ruleset,
					AutoCreateTransaction: true,
					EstimatedDeposit:      &interestAmount,
					NextRecurrence:        ruleStartDate,
					NextRecurrenceOriginal: ruleStartDate,
				}

				if err := repo.CreateFundingSchedule(c.getContext(ctx), &interestSchedule); err == nil {
					scheduleId = &interestSchedule.FundingScheduleId
				}
			}
		}
	}

	// 5. Store Fixed Deposit details
	fd := FixedDeposit{
		SourceBankAccountId:          sourceAccount.BankAccountId,
		FixedBankAccountId:           cdAccount.BankAccountId,
		FundingScheduleId:            scheduleId,
		Name:                         req.Name,
		Amount:                       req.Amount,
		InterestRate:                 req.InterestRate,
		StartDate:                    startUTC,
		EndDate:                      endUTC,
		InterestFrequency:            req.InterestFrequency,
		InterestDestination:          req.InterestDestination,
		InterestDestinationAccountId: req.InterestDestinationAccountId,
		Status:                       FixedDepositStatusActive,
	}

	if err := repo.CreateFixedDeposit(c.getContext(ctx), &fd); err != nil {
		return c.wrapPgError(ctx, err, "failed to store fixed deposit record")
	}

	return ctx.JSON(http.StatusOK, fd)
}

func (c *Controller) postFixedDepositWithdraw(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fixedDepositId, err := ParseID[FixedDeposit](ctx.Param("fixedDepositId"))
	if err != nil || fixedDepositId.IsZero() {
		return c.badRequest(ctx, "must specify a valid fixed deposit Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Fetch deposit details
	fd, err := repo.GetFixedDepositById(c.getContext(ctx), bankAccountId, fixedDepositId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to locate fixed deposit")
	}

	if fd.Status != FixedDepositStatusActive {
		return c.badRequest(ctx, "Only active fixed deposits can be cancelled/withdrawn early")
	}

	// Load CD account to retrieve its balance (principal + any accumulated interest)
	fixedAccount, err := repo.GetBankAccount(c.getContext(ctx), fd.FixedBankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to load CD bank account details")
	}

	balanceToReturn := fixedAccount.CurrentBalance
	if balanceToReturn <= 0 {
		balanceToReturn = fd.Amount
	}

	// 1. Create debit transaction on Fixed CD account (money leaving)
	txnDebit := Transaction{
		BankAccountId: fd.FixedBankAccountId,
		Amount:        balanceToReturn,
		Date:          c.Clock.Now().UTC(),
		Name:          "Early Withdrawal Return",
		OriginalName:  "Early Withdrawal",
		IsPending:     false,
		Source:        TransactionSourceManual,
	}

	// 2. Create credit transaction on Source checking/savings (money entering)
	txnCredit := Transaction{
		BankAccountId: fd.SourceBankAccountId,
		Amount:        -balanceToReturn,
		Date:          c.Clock.Now().UTC(),
		Name:          fmt.Sprintf("Fixed Deposit Cancelled: %s", fd.Name),
		OriginalName:  fd.Name,
		IsPending:     false,
		Source:        TransactionSourceManual,
	}

	if err := repo.CreateTransaction(c.getContext(ctx), fd.FixedBankAccountId, &txnDebit); err != nil {
		return c.wrapPgError(ctx, err, "failed to create debit transaction on CD account")
	}
	if err := repo.CreateTransaction(c.getContext(ctx), fd.SourceBankAccountId, &txnCredit); err != nil {
		return c.wrapPgError(ctx, err, "failed to create credit transaction on source account")
	}

	// Update balances
	sourceAccount, err := repo.GetBankAccount(c.getContext(ctx), fd.SourceBankAccountId)
	if err == nil {
		sourceAccount.AvailableBalance += balanceToReturn
		sourceAccount.CurrentBalance += balanceToReturn
		_ = repo.UpdateBankAccount(c.getContext(ctx), sourceAccount)
	}

	fixedAccount.AvailableBalance = 0
	fixedAccount.CurrentBalance = 0
	fixedAccount.Status = BankAccountStatusInactive
	now := c.Clock.Now().UTC()
	fixedAccount.DeletedAt = &now
	_ = repo.UpdateBankAccount(c.getContext(ctx), fixedAccount)

	// Cancel/delete the interest schedule
	if fd.FundingScheduleId != nil {
		_ = repo.DeleteFundingSchedule(c.getContext(ctx), fd.FixedBankAccountId, *fd.FundingScheduleId)
	}

	// Set status to withdrawn
	fd.Status = FixedDepositStatusWithdrawn
	if err := repo.UpdateFixedDeposit(c.getContext(ctx), fd); err != nil {
		return c.wrapPgError(ctx, err, "failed to update fixed deposit status")
	}

	return ctx.JSON(http.StatusOK, fd)
}

func (c *Controller) deleteFixedDeposit(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fixedDepositId, err := ParseID[FixedDeposit](ctx.Param("fixedDepositId"))
	if err != nil || fixedDepositId.IsZero() {
		return c.badRequest(ctx, "must specify a valid fixed deposit Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Simply deletes the FD history record. It does NOT rollback transactions.
	if err := repo.DeleteFixedDeposit(c.getContext(ctx), bankAccountId, fixedDepositId); err != nil {
		return c.wrapPgError(ctx, err, "failed to delete fixed deposit")
	}

	return ctx.NoContent(http.StatusOK)
}
```

---

## Verification Plan
- Create `server/controller/fixed_deposits_test.go` and mock Echo context calls.
- Submit `POST` calls with valid fields, asserting 200 OK and validating that:
  - CD bank account is created.
  - Initial debit/credit transactions are correctly formed.
  - Funding schedules are successfully registered in UTC timezone.
