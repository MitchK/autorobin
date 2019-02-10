package backtest

import (
	"fmt"
	"path"
	"time"

	"github.com/MitchK/autorobin/lib/autopilot"
	"github.com/MitchK/autorobin/lib/broker/fake"
	"github.com/MitchK/autorobin/lib/data/tiingo"
	"github.com/MitchK/autorobin/lib/model"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// Run Backtest rebalancing strategy
func Run(desiredWeights model.Weights, assets []model.Asset, tiingoToken string, output string) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	// get quotes
	fmt.Println("Fetching quote data from last year from Tiingo...")
	adapter := tiingo.NewAdapter(tiingoToken)
	now := time.Now()
	data, err := adapter.GetDailyAsc(now.AddDate(-1, 0, 0), now, assets...)
	if err != nil {
		return err
	}

	chartData := []interface{}{}

	fmt.Println("Simulating...")

	// Simulating HOLD strategy...
	portfolioQuotes, err := simulate(desiredWeights, data, assets, false)
	chartData = append(chartData, "HOLD Portfolio + cash")
	chartData = append(chartData, toXYs(portfolioQuotes))

	// Simulating REBALANCE strategy
	portfolioQuotes, err = simulate(desiredWeights, data, assets, true)
	chartData = append(chartData, "REBALANCE Portfolio + cash")
	chartData = append(chartData, toXYs(portfolioQuotes))

	p.Title.Text = "Backtest"
	p.X.Label.Text = "Period"
	p.Y.Label.Text = "Value"
	err = plotutil.AddLinePoints(p, chartData...)
	if err != nil {
		return err
	}

	// Save the plot to a PNG file.
	fullPath := path.Join(output, "points.png")
	fmt.Printf("Saving outcome as %v...", fullPath)
	if err := p.Save(40*vg.Centimeter, 20*vg.Centimeter, fullPath); err != nil {
		return err
	}
	return nil
}

// randomPoints returns some random x, y points.
func toXYs(quotes []model.Quote) plotter.XYs {
	pts := make(plotter.XYs, len(quotes))
	var value float64
	for i := range quotes {
		if i == 0 {
			value = 1.0
		} else {
			curr := quotes[i].Price
			prev := quotes[i-1].Price
			value *= (1.0 + (curr-prev)/prev)
		}
		pts[i].X = float64(i)
		pts[i].Y = value
	}
	return pts
}

func simulate(desiredWeights model.Weights, data [][]model.Quote, assets []model.Asset, rebalance bool) ([]model.Quote, error) {
	numAssets := len(assets)
	periods := len(data[0])

	// Transpose data
	dataT := make([][]model.Quote, periods)
	for period := 0; period < periods; period++ {
		dataT[period] = make([]model.Quote, numAssets)
		for i := range assets {
			tmp := data[i]
			quote := tmp[period]
			dataT[period][i] = quote
		}
	}

	// Run back testing
	broker := fake.NewBroker(10000.0)
	pilot, err := autopilot.NewAutopilot(broker)
	if err != nil {
		return nil, err
	}
	portfolioQuotes := make([]model.Quote, periods)
	for period := 0; period < periods; period++ {
		broker.SetQuotes(dataT[period]...)
		cash, err := broker.GetAvailableCash()
		if err != nil {
			return nil, err
		}
		currentPortfolio, err := broker.GetPortfolio(assets...)
		if err != nil {
			return nil, err
		}

		if period == 0 || rebalance {
			orders, err := pilot.Rebalance(desiredWeights, false, assets...)
			if err != nil {
				return nil, err
			}
			if len(orders) > 0 {
				errs := broker.Execute(orders...)
				if len(errs) > 0 {
					for _, err := range errs {
						return nil, err
					}
				}
				cash, err = broker.GetAvailableCash()
				if err != nil {
					return nil, err
				}
				currentPortfolio, err = broker.GetPortfolio(assets...)
				if err != nil {
					return nil, err
				}
			}
		}
		portfolioQuotes[period] = model.Quote{Price: currentPortfolio.TotalValue + cash}
	}
	return portfolioQuotes, nil
}
