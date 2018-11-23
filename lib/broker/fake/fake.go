package fake

import (
	"errors"
	"fmt"

	"github.com/MitchK/autorobin/lib/model"
)

// Fake Fake
type Fake struct {
	cash   float64
	quotes map[model.Asset]model.Quote

	positions map[model.Asset]*model.Position
}

// NewBroker NewBroker
func NewBroker(cash float64) *Fake {
	return &Fake{
		cash:      cash,
		positions: map[model.Asset]*model.Position{},
	}
}

// SetQuotes SetQuotes
func (fake *Fake) SetQuotes(quotes ...model.Quote) {
	fake.quotes = map[model.Asset]model.Quote{}
	for _, quote := range quotes {
		fake.quotes[quote.Asset] = quote
	}
}

// Execute Execute
func (fake *Fake) Execute(orders ...model.Order) []error {

	errs := []error{}

	for _, order := range orders {
		if order.Type == model.OrderTypeBuy {
			fmt.Printf("BUY %v x %s @ %v (%s)\n", order.Quantity, order.Asset.Symbol, order.Price, order.Description)
		} else if order.Type == model.OrderTypeSell {
			fmt.Printf("SELL %v x %s @ %v (%s)\n", order.Quantity, order.Asset.Symbol, order.Price, order.Description)
		}

		if order.Asset == (model.Asset{}) {
			errs = append(errs, errors.New("cannot execute order: no asset set"))
			continue
		}
		asset := order.Asset
		if order.Price <= 0 {
			errs = append(errs, fmt.Errorf("cannot execute order of %s: invalid price: %v", asset.Symbol, order.Price))
			continue
		}
		if order.Quantity < 1 {
			errs = append(errs, fmt.Errorf("cannot execute order of %s: quantity less than 1", asset.Symbol))
			continue
		}
		if order.Type == 0 {
			errs = append(errs, fmt.Errorf("cannot execute order of %s: order type not set", asset.Symbol))
			continue
		}
		position, exists := fake.positions[asset]
		if !exists {
			if order.Type == model.OrderTypeSell {
				errs = append(errs, fmt.Errorf("cannot execute sell order of %s: no open positions", asset.Symbol))
				continue
			}
			fake.positions[asset] = &model.Position{
				Asset: asset,
			}
			position = fake.positions[asset]
		}
		if order.Type == model.OrderTypeBuy {
			total := order.Quantity * order.Price
			if total > fake.cash {
				errs = append(errs, fmt.Errorf("cannot execute buy order of %s: not enough cash", asset.Symbol))
				continue
			}
			position.AvgBuyPrice = (position.AvgBuyPrice*position.Quantity + order.Price*order.Quantity) / (position.Quantity + order.Quantity)
			position.Quantity += order.Quantity
			fake.cash -= order.Quantity * order.Price
		} else if order.Type == model.OrderTypeSell {
			if order.Quantity > position.Quantity {
				errs = append(errs, fmt.Errorf("cannot execute sell order of %s: not enough positions to sell", asset.Symbol))
				continue
			}
			position.AvgBuyPrice = (position.AvgBuyPrice*position.Quantity - order.Price*order.Quantity) / (position.Quantity - order.Quantity)
			position.Quantity -= order.Quantity
			fake.cash += order.Quantity * order.Price
			if position.Quantity == 0 {
				delete(fake.positions, asset)
			}
		} else {
			errs = append(errs, fmt.Errorf("invalid order type: %v", order.Type))
			continue
		}
	}
	return errs
}

// GetAvailableCash GetAvailableCash
func (fake *Fake) GetAvailableCash() (float64, error) {
	return fake.cash, nil
}

// GetPositions GetPositions
func (fake *Fake) GetPositions(assets ...model.Asset) ([]model.Position, error) {
	positions := make([]model.Position, len(assets))
	for i, asset := range assets {
		position, exists := fake.positions[asset]
		if !exists || position == nil {
			position = &model.Position{
				Asset: asset,
			}
		}
		positions[i] = *position
	}
	return positions, nil
}

// GetPortfolio GetPortfolio
func (fake *Fake) GetPortfolio(assets ...model.Asset) (model.Portfolio, error) {
	positions, err := fake.GetPositions(assets...)
	if err != nil {
		return model.Portfolio{}, err
	}

	// Get quantities, include
	quantities := model.Quantities{}
	for _, position := range positions {
		quantities[position.Asset] = position.Quantity
	}

	// Get quotes for every asset
	quotes, err := fake.GetQuotes(assets...)
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
			weight := (quantities[asset] * prices[asset] / total)
			weights[asset] = weight
		}
	}

	return model.Portfolio{
		Weights:    weights,
		TotalValue: total,
		Prices:     prices,
		Quantities: quantities,
	}, nil
}

// GetQuotes GetQuotes
func (fake *Fake) GetQuotes(assets ...model.Asset) ([]model.Quote, error) {
	quotes := make([]model.Quote, len(assets))
	for i, asset := range assets {
		quote, exists := fake.quotes[asset]
		if !exists {
			return nil, fmt.Errorf("could not get quote, asset %s does not exist", asset)
		}
		quotes[i] = quote
	}
	return quotes, nil
}
