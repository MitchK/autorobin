package fixtures

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/MitchK/autorobin/lib/model"
)

const (
	date = iota
	close
	volume
	open
	high
	low
)

// GetQuotes GetQuotes
func GetQuotes(t *testing.T, symbol string) []model.Quote {
	path := filepath.Join("fixtures", symbol+".csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	quotes := []model.Quote{}
	for i, record := range records {
		if i == 0 {
			continue // skip header
		}
		price, err := strconv.ParseFloat(record[close], 64)
		if err != nil {
			t.Fatal(err)
		}
		quotes = append(quotes, model.Quote{
			Asset: model.Asset{
				Symbol: symbol,
			},
			Price: price,
		})
	}
	return quotes
}
