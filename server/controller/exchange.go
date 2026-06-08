package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/currency"
)

type ExchangeRateResponse struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
}

// getExchangeRate processes GET requests to fetch currency exchange rates
// between the 'from' and 'to' parameters.
func (c *Controller) getExchangeRate(ctx echo.Context) error {
	from := strings.ToUpper(strings.TrimSpace(ctx.QueryParam("from")))
	to := strings.ToUpper(strings.TrimSpace(ctx.QueryParam("to")))

	if from == "" || to == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Both 'from' and 'to' currency parameters are required")
	}

	rate, err := currency.GetExchangeRate(from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, ExchangeRateResponse{
		From: from,
		To:   to,
		Rate: rate,
	})
}
