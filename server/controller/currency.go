package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	locale "github.com/monetr/monetr/server/locale"
)

func (c *Controller) listCurrencies(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, locale.GetInstalledCurrencies())
}
