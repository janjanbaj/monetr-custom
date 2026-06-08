package locale

import (
	"sort"
	"strings"

	lclocale "github.com/elliotcourant/go-lclocale"
	"golang.org/x/text/currency"
)

type LConv = lclocale.LConv
type SignPosition = lclocale.SignPosition

func GetCurrencyInternationalFractionalDigits(currencyCode string) (int64, error) {
	if strings.ToUpper(currencyCode) == "NPR" {
		return 2, nil
	}
	digits, err := lclocale.GetCurrencyInternationalFractionalDigits(currencyCode)
	if err == nil {
		return digits, nil
	}

	// Fallback to golang.org/x/text/currency
	iso, parseErr := currency.ParseISO(strings.ToUpper(currencyCode))
	if parseErr != nil {
		return 0, err
	}
	scale, _ := currency.Accounting.Rounding(iso)
	return int64(scale), nil
}

func GetInstalledCurrencies() []string {
	currencies := lclocale.GetInstalledCurrencies()
	hasNPR := false
	for _, c := range currencies {
		if c == "NPR" {
			hasNPR = true
			break
		}
	}
	if !hasNPR {
		copiedCurrencies := make([]string, len(currencies)+1)
		copy(copiedCurrencies, currencies)
		copiedCurrencies[len(currencies)] = "NPR"
		sort.Strings(copiedCurrencies)
		return copiedCurrencies
	}
	return currencies
}

func GetInstalledLocales() []string {
	return lclocale.GetInstalledLocales()
}

func GetLConv(localeString string) (*LConv, error) {
	norm := strings.ReplaceAll(strings.ToLower(localeString), "-", "_")
	if norm == "ne_np" || norm == "en_np" || strings.Contains(norm, "np") {
		return &LConv{
			DecimalPoint:    []byte("."),
			ThousandsSep:    []byte(","),
			IntCurrSymbol:   []byte("NPR "),
			CurrencySymbol:  []byte("Rs"),
			MonDecimalPoint: []byte("."),
			MonThousandsSep: []byte(","),
			IntFracDigits:   2,
			FracDigits:      2,
			PositiveSign:    []byte(""),
			NegativeSign:    []byte("-"),
		}, nil
	}
	return lclocale.GetLConv(localeString)
}
