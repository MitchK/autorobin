package autopilot

import (
	"math"

	"github.com/MitchK/autorobin/lib/broker"
	"github.com/MitchK/autorobin/lib/model"
)

// Autopilot Autopilot
type Autopilot struct {
	broker broker.Broker
}

// NewAutopilot NewAutopilot
func NewAutopilot(broker broker.Broker) (*Autopilot, error) {
	return &Autopilot{
		broker: broker,
	}, nil
}

// GetBroker GetBroker
func (autopilot *Autopilot) GetBroker() broker.Broker {
	return autopilot.broker
}

// Rebalance Rebalance
func (autopilot *Autopilot) Rebalance(desiredPortfolio model.Portfolio, partials bool, minReturn float64, assets ...model.Asset) ([]model.Order, error) {

	// Get available cash
	availableCash, err := autopilot.broker.GetAvailableCash()
	if err != nil {
		return nil, err
	}

	// Get actual portfolio
	actualPortfolio, err := autopilot.broker.GetPortfolio(assets...)
	if err != nil {
		return nil, err
	}

	// Get open positions
	positions, err := autopilot.broker.GetPositions(assets...)
	if err != nil {
		return nil, err
	}

	// Create orders from diff
	weightsDiff := desiredPortfolio.Weights.Diff(actualPortfolio.Weights)
	// fmt.Printf("desired portfolio weights: %+v\n", desiredPortfolio.Weights)
	// fmt.Printf("diff: %+v\n", weightsDiff)

	// Create orders from diffs
	orders := []model.Order{}
	for i, asset := range assets {
		var description string
		var orderType model.OrderType

		// Get current market price
		orderPrice := actualPortfolio.Prices[asset]
		position := positions[i]

		// determine how much we can use to buy new stocks
		volume := actualPortfolio.TotalValue * weightsDiff[asset]
		if volume >= availableCash {
			volume = availableCash
		}
		if volume >= 0 {
			description = "Purchase of missing stocks"
			orderType = model.OrderTypeBuy
			availableCash -= volume
		} else {
			description = "Sale of excess stocks"
			orderType = model.OrderTypeSell
			volume *= -1

			if minReturn > 0 {
				ret := (orderPrice - position.AvgBuyPrice) / position.AvgBuyPrice
				if ret < minReturn {
					continue
				}
			}
		}

		maxQuantity := volume / orderPrice // max possible quantity we can buy
		if !partials {
			maxQuantity = math.Floor(maxQuantity)
			if maxQuantity < 1 {
				continue
			}
		}
		orders = append(orders, model.Order{
			Description: description,
			Type:        orderType,
			Quantity:    maxQuantity,
			Price:       orderPrice,
			Asset:       asset,
		})
	}

	// Do we still have cash left to allocate?
	if availableCash > 0 {
		for _, asset := range assets {
			volume := availableCash * desiredPortfolio.Weights[asset]
			orderPrice := actualPortfolio.Prices[asset]
			quantity := volume / orderPrice
			orderType := model.OrderTypeBuy
			if !partials {
				quantity = math.Floor(quantity)
				if quantity < 1 {
					continue
				}
			}
			orders = append(orders, model.Order{
				Description: "Allocation of unbound cash",
				Type:        orderType,
				Quantity:    quantity,
				Price:       orderPrice,
				Asset:       asset,
			})
		}
	}

	return orders, nil
}
