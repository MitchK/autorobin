package autopilot

import (
	"fmt"
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
func (autopilot *Autopilot) Rebalance(desiredWeights model.Weights, partials bool, assets ...model.Asset) ([]model.Order, error) {

	// Get available cash
	availableCash, err := autopilot.broker.GetAvailableCash()
	if err != nil {
		return nil, err
	}
	fmt.Println("Unallocated cash:", availableCash)

	// Get actual portfolio
	actualPortfolio, err := autopilot.broker.GetPortfolio(assets...)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Current portfolio value for assets %v: %v\n", assets, actualPortfolio.TotalValue)

	// Create orders from diff
	weightsDiff := desiredWeights.Diff(actualPortfolio.Weights)

	// Create orders from diffs
	orders := []model.Order{}
	for _, asset := range assets {
		var description string
		var orderType model.OrderType

		// Get current market price
		orderPrice := actualPortfolio.Prices[asset]

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
		}

		maxQuantity := volume / orderPrice // max possible quantity we can buy
		if maxQuantity <= 0 {
			continue
		}
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
			volume := availableCash * desiredWeights[asset]
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
