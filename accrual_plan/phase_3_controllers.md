# Phase 3: Go Backend REST Controllers & Routes

This phase defines the REST API endpoints inside Echo controller handlers, validates payloads, and wires them into standard routes.

## Context Directories
- [server/controller/](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/controller)

---

## 1. Controller Endpoint Implementation

#### [NEW] [accrual_expenses.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/controller/accrual_expenses.go)
```go
package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) getAccrualExpenses(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account ID")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	expenses, err := repo.GetAccrualExpenses(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve accrual expenses")
	}

	return ctx.JSON(http.StatusOK, expenses)
}

func (c *Controller) getAccrualExpenseById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account ID")
	}

	expenseId, err := ParseID[AccrualExpense](ctx.Param("accrualExpenseId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid accrual expense ID")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	expense, err := repo.GetAccrualExpenseById(c.getContext(ctx), bankAccountId, expenseId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve accrual expense")
	}

	return ctx.JSON(http.StatusOK, expense)
}

func (c *Controller) postAccrualExpense(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account ID")
	}

	var payload struct {
		TransactionId *ID[Transaction] `json:"transactionId"`
		Name          string           `json:"name"`
		Description   string           `json:"description"`
		Amount        int64            `json:"amount"`
		StartDate     time.Time        `json:"startDate"`
		EndDate       time.Time        `json:"endDate"`
	}

	if err := ctx.Bind(&payload); err != nil {
		return c.invalidJson(ctx)
	}

	if payload.Name == "" {
		return c.badRequest(ctx, "accrual expense must have a name")
	}
	if payload.Amount <= 0 {
		return c.badRequest(ctx, "accrual expense amount must be greater than 0")
	}
	if payload.StartDate.After(payload.EndDate) {
		return c.badRequest(ctx, "start date cannot be after end date")
	}

	expense := &AccrualExpense{
		BankAccountId: bankAccountId,
		TransactionId: payload.TransactionId,
		Name:          payload.Name,
		Description:   payload.Description,
		Amount:        payload.Amount,
		StartDate:     payload.StartDate,
		EndDate:       payload.EndDate,
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.CreateAccrualExpense(c.getContext(ctx), expense); err != nil {
		return c.wrapPgError(ctx, err, "failed to create accrual expense")
	}

	return ctx.JSON(http.StatusOK, expense)
}

func (c *Controller) putAccrualExpense(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account ID")
	}

	expenseId, err := ParseID[AccrualExpense](ctx.Param("accrualExpenseId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid accrual expense ID")
	}

	var payload struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Amount      int64     `json:"amount"`
		StartDate   time.Time `json:"startDate"`
		EndDate     time.Time `json:"endDate"`
	}

	if err := ctx.Bind(&payload); err != nil {
		return c.invalidJson(ctx)
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	expense, err := repo.GetAccrualExpenseById(c.getContext(ctx), bankAccountId, expenseId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not find existing accrual expense")
	}

	if payload.Name != "" {
		expense.Name = payload.Name
	}
	expense.Description = payload.Description
	if payload.Amount > 0 {
		expense.Amount = payload.Amount
	}
	if !payload.StartDate.IsZero() {
		expense.StartDate = payload.StartDate
	}
	if !payload.EndDate.IsZero() {
		expense.EndDate = payload.EndDate
	}

	if expense.StartDate.After(expense.EndDate) {
		return c.badRequest(ctx, "start date cannot be after end date")
	}

	if err := repo.UpdateAccrualExpense(c.getContext(ctx), bankAccountId, expense); err != nil {
		return c.wrapPgError(ctx, err, "failed to update accrual expense")
	}

	return ctx.JSON(http.StatusOK, expense)
}

func (c *Controller) deleteAccrualExpense(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account ID")
	}

	expenseId, err := ParseID[AccrualExpense](ctx.Param("accrualExpenseId"))
	if err != nil {
		return c.badRequest(ctx, "must specify a valid accrual expense ID")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.DeleteAccrualExpense(c.getContext(ctx), bankAccountId, expenseId); err != nil {
		return c.wrapPgError(ctx, err, "failed to delete accrual expense")
	}

	return ctx.NoContent(http.StatusOK)
}
```

---

## 2. Routes Configuration

#### [MODIFY] [routes.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/controller/routes.go)
Inside `RegisterRoutes` under `billed := authed.Group("", c.requireActiveSubscriptionMiddleware)` (around line 337), add routes for the accrual resource:
```go
	// Accrual Expenses
	billed.GET("/bank_accounts/:bankAccountId/accrual", c.getAccrualExpenses)
	billed.GET("/bank_accounts/:bankAccountId/accrual/:accrualExpenseId", c.getAccrualExpenseById)
	billed.POST("/bank_accounts/:bankAccountId/accrual", c.postAccrualExpense)
	billed.PUT("/bank_accounts/:bankAccountId/accrual/:accrualExpenseId", c.putAccrualExpense)
	billed.DELETE("/bank_accounts/:bankAccountId/accrual/:accrualExpenseId", c.deleteAccrualExpense)
```

---

## Verification Plan

### Automated Tests

#### [NEW] [accrual_expenses_test.go](file:///Users/janeetbajracharya/Desktop/Code/monetr/server/controller/accrual_expenses_test.go)
Create HTTP integration tests ensuring:
1. `GET /bank_accounts/:bankAccountId/accrual` is authenticated.
2. Payload validation fires bad request for empty names, negative amounts, or mismatched date bounds.
3. CRUD returns expected JSON outputs matching `AccrualExpense` model serialization rules.

Run test suite:
```bash
go test ./server/controller/... -run TestAccrualExpense -v
```
Ensure all API handlers execute successfully.
