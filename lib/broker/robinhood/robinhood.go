package robinhood

import (
	"errors"
	"fmt"
	"math"

	"github.com/MitchK/autorobin/lib/broker"
	"github.com/MitchK/autorobin/lib/model"
	robinhood "github.com/andrewstuart/go-robinhood"
)

// Broker Broker
type Broker struct {
	client *robinhood.Client
}

// NewBroker NewBroker
func NewBroker(username string, password string) (broker.Broker, error) {
	var err error
	broker := &Broker{}
	broker.client, err = robinhood.Dial(&robinhood.OAuth{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return broker, nil
}

func (broker *Broker) getAccount() (robinhood.Account, error) {
	accounts, err := broker.client.GetAccounts()
	if err != nil {
		return robinhood.Account{}, err
	}
	if len(accounts) == 0 {
		return robinhood.Account{}, errors.New("no account associated with login")
	}
	return accounts[0], nil
}

func (broker *Broker) getPortfolio() (robinhood.Portfolio, error) {
	portfolios, err := broker.client.GetPortfolios()
	if err != nil {
		return robinhood.Portfolio{}, err
	}
	if len(portfolios) == 0 {
		return robinhood.Portfolio{}, errors.New("no account associated with login")
	}
	return portfolios[0], nil
}

// GetAvailableCash GetAvailableCash
func (broker *Broker) GetAvailableCash() (float64, error) {
	account, err := broker.getAccount()
	if err != nil {
		return 0, err
	}
	return account.MarginBalances.UnallocatedMarginCash, nil
}

func (broker *Broker) convertPosition(position robinhood.Position) (model.Position, error) {
	instrument, err := broker.client.GetInstrument(position.Instrument)
	if err != nil {
		return model.Position{}, err
	}
	return model.Position{
		AvgBuyPrice: position.AverageBuyPrice,
		Quantity:    position.Quantity,
		Asset: model.Asset{
			Symbol: instrument.Symbol,
		},
	}, nil
}

func (broker *Broker) convertQuote(quote robinhood.Quote) (model.Quote, error) {
	return model.Quote{
		Asset: model.Asset{
			Symbol: quote.Symbol,
		},
		LastTradePrice: quote.LastTradePrice,
	}, nil
}

// GetQuotes GetQuotes
func (broker *Broker) GetQuotes(assets ...model.Asset) ([]model.Quote, error) {
	symbols := []string{}
	for _, asset := range assets {
		symbols = append(symbols, asset.Symbol)
	}
	quotes, err := broker.client.GetQuote(symbols...)
	if err != nil {
		return nil, err
	}
	retQuotes := []model.Quote{}
	for _, quote := range quotes {
		convertedQuote, err := broker.convertQuote(quote)
		if err != nil {
			return nil, err
		}
		retQuotes = append(retQuotes, convertedQuote)
	}
	return retQuotes, nil
}

// GetOpenPositions GetOpenPositions
func (broker *Broker) GetOpenPositions() ([]model.Position, error) {
	account, err := broker.getAccount()
	if err != nil {
		return nil, err
	}
	positions, err := broker.client.GetPositions(account)
	if err != nil {
		return nil, err
	}
	openPositions := []model.Position{}
	for _, position := range positions {
		// filter out positions that are closed (i.e. q = 0)
		// as well as positions that were given to us for free (avgBuyPrice = 0)
		if position.Quantity > 0 && position.AverageBuyPrice > 0 {
			convertedPosition, err := broker.convertPosition(position)
			if err != nil {
				return nil, err
			}
			openPositions = append(openPositions, convertedPosition)
		}
	}
	return openPositions, nil
}

// GetPortfolio GetPortfolio
func (broker *Broker) GetPortfolio(includeAssets ...model.Asset) (model.Portfolio, error) {
	openPositions, err := broker.GetOpenPositions()
	if err != nil {
		return model.Portfolio{}, err
	}

	// Get quantities, include "includeAssets"
	quantities := model.Quantities{}
	assets := []model.Asset{}
	for _, openPosition := range openPositions {
		quantities[openPosition.Asset] = openPosition.Quantity
		assets = append(assets, openPosition.Asset)
	}
	for _, asset := range includeAssets { // merge with includeAssets
		if _, ok := quantities[asset]; !ok {
			quantities[asset] = 0.0
			assets = append(assets, asset)
		}
	}

	// Get quotes for every asset
	quotes, err := broker.GetQuotes(assets...)
	if err != nil {
		return model.Portfolio{}, err
	}
	prices := model.Prices{}
	for _, quote := range quotes {
		prices[quote.Asset] = quote.LastTradePrice
	}

	// Calculate total
	var total float64
	for _, asset := range assets {
		total += quantities[asset] * prices[asset]
	}

	// Calculate weights
	weights := model.Weights{}
	for _, asset := range assets {
		weights[asset] = (quantities[asset] * prices[asset] / total)
	}

	return model.Portfolio{
		Weights:    weights,
		TotalValue: total,
		Prices:     prices,
		Quantities: quantities,
		Assets:     assets,
	}, nil
}

// Execute Execute
func (broker *Broker) Execute(orders ...model.Order) error {

	for _, order := range orders {
		orderStr := fmt.Sprintf("order of %v x %s @ %v (%s)", order.Quantity, order.Asset.Symbol, order.Price, order.Description)
		if order.Quantity < 1 {
			fmt.Println("cannot execute", orderStr, ": robinhood does not support quantities less than 1, skipping order...")
			continue
		}
		fmt.Println("executing", orderStr, "...")
		instr, err := broker.client.GetInstrumentForSymbol(order.Asset.Symbol)
		if err != nil {
			return err
		}

		var orderSide robinhood.OrderSide
		if order.Type == model.OrderTypeBuy {
			orderSide = robinhood.Buy
		} else if order.Type == model.OrderTypeSell {
			orderSide = robinhood.Sell
		} else {
			return fmt.Errorf("invalid order type: %v", order.Type)
		}

		orderSide = robinhood.Sell

		orderOutput, err := broker.client.Order(instr, robinhood.OrderOpts{
			Side:     orderSide,
			Type:     robinhood.Limit,
			Price:    math.Round(order.Price*100) / 100, // round to nearest
			Quantity: uint64(order.Quantity),
		})
		if err != nil {
			return err
		}
		if orderOutput.State != "confirmed" && orderOutput.State != "unconfirmed" {
			return fmt.Errorf("some issue with order, state: %s, reject reason: %s", orderOutput.State, orderOutput.RejectReason)
		}
	}
	return nil
}
