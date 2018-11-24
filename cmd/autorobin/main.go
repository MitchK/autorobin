package main

import (
	"bufio"
	"fmt"
	"time"

	"github.com/MitchK/autorobin/lib/autopilot"
	robinhoodBroker "github.com/MitchK/autorobin/lib/broker/robinhood"
	"github.com/MitchK/autorobin/lib/data/tiingo"
	"github.com/MitchK/autorobin/lib/portfolioparser/portfoliovisualizer"

	"github.com/MitchK/autorobin/lib/broker/fake"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"os"

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

func main() {
	csvFile, err := os.Open(os.Getenv("PV_CSV_FILE"))
	if err != nil {
		panic(err)
	}
	parser := portfoliovisualizer.NewParser()
	desiredWeights, assets, err := parser.Parse(bufio.NewReader(csvFile)) // TODO, connect with PV profile
	if err != nil {
		panic(err)
	}

	backtest(desiredWeights, assets, true)
}

// randomPoints returns some random x, y points.
func toXYs(quotes []float64) plotter.XYs {
	pts := make(plotter.XYs, len(quotes))
	var value float64
	for i := range quotes {
		if i == 0 {
			value = 1.0
		} else {
			curr := quotes[i]
			prev := quotes[i-1]
			value *= (1.0 + (curr-prev)/prev)
		}
		pts[i].X = float64(i)
		pts[i].Y = value
	}
	return pts
}

// func getQuotes(filePath string, symbol string) ([]model.Quote, error) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r := csv.NewReader(file)
// 	records, err := r.ReadAll()
// 	if err != nil {
// 		return nil, err
// 	}
// 	quotes := []model.Quote{}
// 	for i := len(records) - 1; i >= 0; i-- {
// 		if i == 0 {
// 			continue // skip header
// 		}
// 		record := records[i]
// 		price, err := strconv.ParseFloat(record[close], 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		quotes = append(quotes, model.Quote{
// 			Asset: model.Asset{
// 				Symbol: symbol,
// 			},
// 			Price: price,
// 		})
// 	}
// 	return quotes, nil
// }

func run(desiredWeights model.Weights, assets []model.Asset) {
	// Connect to Robinhood
	broker, err := robinhoodBroker.NewBroker(os.Getenv("ROBINHOOD_USERNAME"), os.Getenv("ROBINHOOD_PASSWORD"))
	if err != nil {
		panic(err)
	}
	fmt.Println("connected to robinhood")

	// Create autopilot
	autopilot, err := autopilot.NewAutopilot(broker)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	orders, err := autopilot.Rebalance(desiredWeights, false, -1, assets...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	errs := broker.Execute(orders...)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}

func backtest(desiredWeights model.Weights, assets []model.Asset, rebalance bool) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// get quotes
	adapter := tiingo.NewAdapter(os.Getenv("TIINGO_TOKEN"))
	now := time.Now()
	data, err := adapter.GetDailyAsc(now.AddDate(-1, 0, 0), now, assets...)
	if err != nil {
		panic(err)
	}

	chartData := []interface{}{}
	portfolioQuotes, err := simulate(desiredWeights, data, assets, false)
	chartData = append(chartData, "HOLD Portfolio + cash")
	chartData = append(chartData, toXYs(portfolioQuotes))

	portfolioQuotes, err = simulate(desiredWeights, data, assets, true)
	chartData = append(chartData, "REBALANCE Portfolio + cash")
	chartData = append(chartData, toXYs(portfolioQuotes))

	p.Title.Text = "Backtest"
	p.X.Label.Text = "Period"
	p.Y.Label.Text = "Value"
	err = plotutil.AddLinePoints(p, chartData...)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(40*vg.Centimeter, 20*vg.Centimeter, "points.png"); err != nil {
		panic(err)
	}
}

func simulate(desiredWeights model.Weights, data [][]model.Quote, assets []model.Asset, rebalance bool) ([]float64, error) {
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
	broker := fake.NewBroker(100000.0)
	pilot, err := autopilot.NewAutopilot(broker)
	if err != nil {
		return nil, err
	}
	portfolioQuotes := make([]float64, periods)
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
			orders, err := pilot.Rebalance(desiredWeights, false, 0.06, assets...)
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
		portfolioQuotes[period] = currentPortfolio.TotalValue + cash
	}
	return portfolioQuotes, nil
}
