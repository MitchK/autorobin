package broker

import "github.com/MitchK/autorobin/lib/model"

//go:generate mockgen -destination=../mocks/mock_broker.go -package=mocks github.com/MitchK/autorobin/lib/broker Broker

// Broker Broker
type Broker interface {
	Execute(orders ...model.Order) []error
	GetAvailableCash() (float64, error)
	GetPositions(assets ...model.Asset) ([]model.Position, error)
	GetPortfolio(assets ...model.Asset) (model.Portfolio, error)
	GetQuotes(assets ...model.Asset) ([]model.Quote, error)
}
