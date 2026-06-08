package currency_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/monetr/monetr/server/currency"
	"github.com/stretchr/testify/assert"
)

func TestGetExchangeRate_ShortCircuit(t *testing.T) {
	rate, err := currency.GetExchangeRate("USD", "USD")
	assert.NoError(t, err)
	assert.Equal(t, 1.0, rate)

	rate, err = currency.GetExchangeRate("NPR", "NPR")
	assert.NoError(t, err)
	assert.Equal(t, 1.0, rate)
}

func TestGetExchangeRate_Validation(t *testing.T) {
	_, err := currency.GetExchangeRate("", "USD")
	assert.Error(t, err)

	_, err = currency.GetExchangeRate("USD", "")
	assert.Error(t, err)
}

func TestGetExchangeRate_Regex(t *testing.T) {
	// Sample JS payload snippet from Google Finance
	samplePayload := `window.IJ_values = [ ... ]; AF_initDataCallback({key: 'ds:3', data:[[["/g/11bvvzbj2p",null,"NPR / USD",3,null,[0.006539959,-4.4541E-5,-0.006764522,4,6,2],null,0.0065845]]]});`

	rateRegex := regexp.MustCompile(`"NPR\s*/\s*USD"\s*,\s*\d+\s*,\s*null\s*,\s*\[\s*([0-9.]+e?-?[0-9]*)`)
	matches := rateRegex.FindStringSubmatch(samplePayload)
	
	assert.Len(t, matches, 2)
	assert.Equal(t, "0.006539959", matches[1])
}

func TestGetExchangeRate_FromFile(t *testing.T) {
	// Check if our locally saved test_gf.html is present
	wd, err := os.Getwd()
	if err != nil {
		t.Skip("Could not determine working directory")
	}

	// Resolve the path to test_gf.html in the project root
	rootPath := filepath.Join(wd, "..", "..")
	testHtmlPath := filepath.Join(rootPath, "test_gf.html")

	if _, err := os.Stat(testHtmlPath); os.IsNotExist(err) {
		t.Skip("test_gf.html not found, skipping parsing verification test")
	}

	bodyBytes, err := os.ReadFile(testHtmlPath)
	assert.NoError(t, err)

	rateRegex := regexp.MustCompile(`"NPR\s*/\s*USD"\s*,\s*\d+\s*,\s*null\s*,\s*\[\s*([0-9.]+e?-?[0-9]*)`)
	matches := rateRegex.FindSubmatch(bodyBytes)
	
	assert.True(t, len(matches) >= 2, "Should find matching pattern in test_gf.html")
	assert.Equal(t, "0.006539959", string(matches[1]))
}
