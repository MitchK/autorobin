package robinhood

import (
	"errors"
	"fmt"
	"math"

	"github.com/MitchK/autorobin/lib/broker"
	"github.com/MitchK/autorobin/lib/model"
	robinhood "github.com/andrewstuart/go-robinhood"
)

// robinhoodBroker robinhoodBroker
type robinhoodBroker struct {
	client *robinhood.Client
}

// NewBroker NewBroker
func NewBroker(username string, password string) (broker.Broker, error) {
	var err error
	broker := &robinhoodBroker{}
	broker.client, err = robinhood.Dial(&robinhood.OAuth{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return broker, nil
}

func (broker *robinhoodBroker) getAccount() (robinhood.Account, error) {
	accounts, err := broker.client.GetAccounts()
	if err != nil {
		return robinhood.Account{}, err
	}
	if len(accounts) == 0 {
		return robinhood.Account{}, errors.New("no account associated with login")
	}
	return accounts[0], nil
}

func (broker *robinhoodBroker) getPortfolio() (robinhood.Portfolio, error) {
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
func (broker *robinhoodBroker) GetAvailableCash() (float64, error) {
	account, err := broker.getAccount()
	if err != nil {
		return 0, err
	}
	return account.MarginBalances.UnallocatedMarginCash, nil
}

func (broker *robinhoodBroker) convertPosition(position robinhood.Position) (model.Position, error) {
	instrument, err := broker.client.GetInstrument(position.Instrument)
	if err != nil {
		return model.Position{}, err
	}

	return model.Position{
		AvgBuyPrice: position.AverageBuyPrice,
		Quantity:    position.Quantity + position.SharesHeldForBuys - position.SharesHeldForSells,
		Asset: model.Asset{
			Symbol: instrument.Symbol,
		},
	}, nil
}

func (broker *robinhoodBroker) convertQuote(quote robinhood.Quote) (model.Quote, error) {
	return model.Quote{
		Asset: model.Asset{
			Symbol: quote.Symbol,
		},
		Price: quote.PreviousClose,
	}, nil
}

// GetQuotes GetQuotes
func (broker *robinhoodBroker) GetQuotes(assets ...model.Asset) ([]model.Quote, error) {
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

// GetPositions GetPositions
func (broker *robinhoodBroker) GetPositions(assets ...model.Asset) ([]model.Position, error) {
	account, err := broker.getAccount()
	if err != nil {
		return nil, err
	}
	positions, err := broker.client.GetPositions(account)
	if err != nil {
		return nil, err
	}
	openPositionsMap := map[model.Asset]model.Position{}
	for _, position := range positions {
		// filter out positions that are closed (i.e. q = 0)
		// as well as positions that were given to us for free (avgBuyPrice = 0)
		if position.Quantity > 0 && position.AverageBuyPrice > 0 {
			convertedPosition, err := broker.convertPosition(position)
			if err != nil {
				return nil, err
			}
			openPositionsMap[convertedPosition.Asset] = convertedPosition
		}
	}
	openPositions := make([]model.Position, len(assets))
	for i, asset := range assets {
		openPositions[i] = openPositionsMap[asset]
	}
	return openPositions, nil
}

// GetPortfolio GetPortfolio
func (broker *robinhoodBroker) GetPortfolio(assets ...model.Asset) (model.Portfolio, error) {
	positions, err := broker.GetPositions(assets...)
	if err != nil {
		return model.Portfolio{}, err
	}

	// Get quantities, include "includeAssets"
	quantities := model.Quantities{}
	for _, position := range positions {
		quantities[position.Asset] = position.Quantity
	}

	// Get quotes for every asset
	quotes, err := broker.GetQuotes(assets...)
	if err != nil {
		return model.Portfolio{}, err
	}
	prices := model.Prices{}
	for _, quote := range quotes {
		prices[quote.Asset] = quote.Price
	}

	// Calculate total
	var total float64
	for _, asset := range assets {
		total += quantities[asset] * prices[asset]
	}

	// Calculate weights
	weights := model.Weights{}
	for _, asset := range assets {
		if total == 0 {
			weights[asset] = 0
		} else {
			quantity := quantities[asset]
			price := prices[asset]
			weights[asset] = (quantity * price / total)
		}
	}

	return model.Portfolio{
		Weights:    weights,
		TotalValue: total,
		Prices:     prices,
		Quantities: quantities,
	}, nil
}

// Execute Execute
func (broker *robinhoodBroker) Execute(orders ...model.Order) []error {
	errors := []error{}
	for _, order := range orders {
		orderStr := fmt.Sprintf("order of %v x %s @ %v (%s)", order.Quantity, order.Asset.Symbol, order.Price, order.Description)
		if order.Quantity < 1 {
			fmt.Println("cannot execute", orderStr, ": robinhood does not support quantities less than 1, skipping order...")
			continue
		}
		fmt.Println("executing", orderStr, "...")
		instr, err := broker.client.GetInstrumentForSymbol(order.Asset.Symbol)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		var orderSide robinhood.OrderSide
		if order.Type == model.OrderTypeBuy {
			orderSide = robinhood.Buy
		} else if order.Type == model.OrderTypeSell {
			orderSide = robinhood.Sell
		} else {
			errors = append(errors, fmt.Errorf("invalid order type: %v", order.Type))
			continue
		}

		orderOpts := robinhood.OrderOpts{
			Side:     orderSide,
			Type:     robinhood.Limit,
			Price:    math.Round(order.Price*100) / 100, // round to nearest
			Quantity: uint64(order.Quantity),
		}
		orderOutput, err := broker.client.Order(instr, orderOpts)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if orderOutput.State != "confirmed" && orderOutput.State != "unconfirmed" {
			errors = append(errors, fmt.Errorf("some issue with order, state: %s, reject reason: %s", orderOutput.State, orderOutput.RejectReason))
			continue
		}
	}
	return nil
}
