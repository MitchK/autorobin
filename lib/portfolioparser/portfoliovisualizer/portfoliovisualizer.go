package portfoliovisualizer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/MitchK/autorobin/lib/model"
	"github.com/MitchK/autorobin/lib/portfolioparser"
)

const (
	skipLines = 3
)

// Parser Parser
type parser struct{}

// NewParser NewParser
func NewParser() portfolioparser.Parser {
	return &parser{}
}

// Parse Parse
func (parser *parser) Parse(reader io.Reader, portfolio *model.Portfolio) error {
	r := csv.NewReader(reader)
	r.FieldsPerRecord = -1 // allow uneven fields

	lines, err := r.ReadAll()
	if err != nil {
		return err
	}
	portfolio.Weights = model.Weights{}

	start := false
	for i, line := range lines {
		if line[0] == "Ticker" {
			start = true
			continue
		}
		if !start {
			continue
		}
		if line[0] == "Portfolio Performance" {
			break
		}
		percentage, err := strconv.ParseFloat(
			strings.Replace(
				strings.Trim(
					line[2], " ",
				), "%", "", -1,
			),
			64,
		)
		if err != nil {
			return fmt.Errorf("csv parser: error at line %v: %s", i+1, err)
		}
		asset := model.Asset{
			Symbol: strings.Trim(line[0], " "),
		}
		portfolio.Assets = append(portfolio.Assets, asset)
		portfolio.Weights[asset] = percentage / 100.0
	}

	return nil
}
