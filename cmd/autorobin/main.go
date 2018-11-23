package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/MitchK/autorobin/lib/autopilot"
	"github.com/MitchK/autorobin/lib/portfolioparser/portfoliovisualizer"

	"github.com/MitchK/autorobin/lib/broker/fake"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"

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

func getQuotes(filePath string, symbol string) ([]model.Quote, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	quotes := []model.Quote{}
	for i := len(records) - 1; i >= 0; i-- {
		if i == 0 {
			continue // skip header
		}
		record := records[i]
		price, err := strconv.ParseFloat(record[close], 64)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, model.Quote{
			Asset: model.Asset{
				Symbol: symbol,
			},
			Price: price,
		})
	}
	return quotes, nil
}

func main() {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// add asset charts
	csvfiles := strings.Split(os.Getenv("CSV_FILES"), ",")
	chartData := []interface{}{}
	numAssets := len(csvfiles)
	periods := 0
	assets := make([]model.Asset, len(csvfiles))
	data := make(map[model.Asset][]model.Quote, len(csvfiles))
	for i, filePath := range csvfiles {
		_, fileName := filepath.Split(filePath)
		symbol := strings.Split(fileName, ".")[0]
		quotes, err := getQuotes(filePath, symbol)
		if err != nil {
			panic(err)
		}
		if periods == 0 || len(quotes) < periods {
			periods = len(quotes)
		}
		assets[i] = model.Asset{Symbol: symbol}
		data[assets[i]] = quotes
		chartData = append(chartData, symbol)
		chartData = append(chartData, toXYs(quotes))
	}

	// Transpose data
	dataT := make([][]model.Quote, periods)
	for period := 0; period < periods; period++ {
		dataT[period] = make([]model.Quote, numAssets)
		for i, asset := range assets {
			quote := data[asset][period]
			dataT[period][i] = quote
		}
	}

	// Read desired portfolio
	// desiredPortfolio := model.Portfolio{
	// 	Weights: model.Weights{
	// 		model.Asset
	// 	},
	// }
	csvFile, err := os.Open(os.Getenv("PV_CSV_FILE"))
	if err != nil {
		panic(err)
	}
	parser := portfoliovisualizer.NewParser()
	desiredPortfolio, assets, err := parser.Parse(bufio.NewReader(csvFile)) // TODO, connect with PV profile
	if err != nil {
		panic(err)
	}

	// Run back testing
	broker := fake.NewBroker(100000.0)
	pilot, err := autopilot.NewAutopilot(broker)
	if err != nil {
		panic(err)
	}

	portfolioQuotes := make([]model.Quote, periods)
	for period := 0; period < periods; period++ {
		fmt.Printf("----------------- Period %v-----------------\n", period)
		broker.SetQuotes(dataT[period]...)
		fmt.Println(broker.GetQuotes(assets...))
		cash, err := broker.GetAvailableCash()
		if err != nil {
			panic(err)
		}
		currentPortfolio, err := broker.GetPortfolio(assets...)
		if err != nil {
			panic(err)
		}
		fmt.Println("before: total value is", currentPortfolio.TotalValue+cash, "(assets:", currentPortfolio.TotalValue, ", cash:", cash, ")")
		orders, err := pilot.Rebalance(*desiredPortfolio, false, -1, assets...)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(orders) == 0 {
			fmt.Printf("%v: no orders created\n", period)
			continue
		}
		errs := broker.Execute(orders...)
		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Printf("%v: %s\n", period, err)
				panic(errs)
			}
		}
		cash, err = broker.GetAvailableCash()
		if err != nil {
			panic(err)
		}
		currentPortfolio, err = broker.GetPortfolio(assets...)
		if err != nil {
			panic(err)
		}
		portfolioQuotes[period] = model.Quote{
			Price: currentPortfolio.TotalValue + cash,
		}
		fmt.Println("after: total value is", currentPortfolio.TotalValue+cash, "(assets:", currentPortfolio.TotalValue, ", cash:", cash, ")")
		fmt.Println("-----------------------------------------")
	}

	chartData = append(chartData, "Portfolio + cash")
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

	// // Connect to Robinhood
	// broker, err := broker_robinhood.NewBroker(os.Getenv("ROBINHOOD_USERNAME"), os.Getenv("ROBINHOOD_PASSWORD"))
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("connected to robinhood")

	// // Create autopilot
	// autopilot, err := autopilot.NewAutopilot(broker)
	// if err != nil {
	// 	panic(err)
	// }

	// orders, err := autopilot.CreateOrders(desiredPortfolio)
}
