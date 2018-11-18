package autopilot

import (
	"fmt"

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

// CreateOrders CreateOrders
func (autopilot *Autopilot) CreateOrders(desiredPortfolio model.Portfolio) ([]model.Order, error) {

	// Get available cash
	cash, err := autopilot.broker.GetAvailableCash()
	if err != nil {
		return nil, err
	}
	fmt.Printf("available cash %v\n", cash)

	// Get actual portfolio
	actualPortfolio, err := autopilot.broker.GetPortfolio(desiredPortfolio.Assets...)
	if err != nil {
		return nil, err
	}

	// Create orders from diff
	weightsDiff := desiredPortfolio.Weights.Diff(actualPortfolio.Weights)

	// Create orders from diffs
	orders := []model.Order{}
	buyVolume := 0.0
	for _, asset := range actualPortfolio.Assets {
		volume := actualPortfolio.TotalValue * weightsDiff[asset]
		orderPrice := actualPortfolio.Prices[asset]
		quantity := volume / orderPrice
		var orderType model.OrderType
		var description string
		if quantity >= 0 {
			description = "Purchase of missing stocks"
			orderType = model.OrderTypeBuy
			buyVolume += volume
		} else {
			description = "Sale of excess stocks"
			orderType = model.OrderTypeSell
			quantity *= -1
		}
		orders = append(orders, model.Order{
			Description: description,
			Type:        orderType,
			Quantity:    quantity,
			Price:       orderPrice,
			Asset:       asset,
		})
	}

	// we cannot further use cash that we need for buying
	cash -= buyVolume

	// Do we still have cash left to allocate?
	if cash > 0 {
		fmt.Printf("allocating remaining cash %v\n", cash)
		for _, asset := range desiredPortfolio.Assets {
			volume := cash * desiredPortfolio.Weights[asset]
			orderPrice := actualPortfolio.Prices[asset]
			quantity := volume / orderPrice
			orderType := model.OrderTypeBuy
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
