package currency

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)



type cacheEntry struct {
	rate      float64
	timestamp time.Time
}

var (
	cache   = make(map[string]cacheEntry)
	cacheMu sync.RWMutex
)

// GetExchangeRate returns the exchange rate between the `from` currency and the `to` currency
// using Google Finance data. Results are cached in-memory for 30 minutes to respect limits.
func GetExchangeRate(from, to string) (float64, error) {
	if from == "" || to == "" {
		return 0, fmt.Errorf("currency codes cannot be empty")
	}

	if from == to {
		return 1.0, nil
	}

	cacheKey := fmt.Sprintf("%s-%s", from, to)

	cacheMu.RLock()
	entry, found := cache[cacheKey]
	cacheMu.RUnlock()

	if found && time.Since(entry.timestamp) < 30*time.Minute {
		return entry.rate, nil
	}

	rate, err := fetchExchangeRate(from, to)
	if err != nil {
		return 0, err
	}

	cacheMu.Lock()
	cache[cacheKey] = cacheEntry{
		rate:      rate,
		timestamp: time.Now(),
	}
	cacheMu.Unlock()

	return rate, nil
}

func fetchExchangeRate(from, to string) (float64, error) {
	url := fmt.Sprintf("https://www.google.com/finance/quote/%s-%s", from, to)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create http request: %w", err)
	}

	// Set a standard browser User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request to Google Finance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("google finance responded with status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Matches specifically: "FROM / TO", <digits>, null, [<rate_value>, ...
	pattern := fmt.Sprintf(`"%s\s*/\s*%s"\s*,\s*\d+\s*,\s*null\s*,\s*\[\s*([0-9.]+e?-?[0-9]*)`, from, to)
	rateRegex, err := regexp.Compile(pattern)
	if err != nil {
		return 0, fmt.Errorf("failed to compile exchange rate regex: %w", err)
	}

	matches := rateRegex.FindSubmatch(bodyBytes)
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not find exchange rate data on Google Finance page for %s-%s", from, to)
	}

	rateStr := string(matches[1])
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse exchange rate '%s' as float: %w", rateStr, err)
	}

	return rate, nil
}
