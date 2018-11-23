package fake_test

import (
	"testing"

	"github.com/MitchK/autorobin/lib/broker"
	"github.com/MitchK/autorobin/lib/broker/fake"
	"github.com/MitchK/autorobin/lib/broker/fake/fixtures"
	"github.com/MitchK/autorobin/lib/model"
	"github.com/onsi/gomega"
)

var (
	initialCash = 10000.0
	aapl        = model.Asset{
		Symbol: "AAPL",
	}
	sap = model.Asset{
		Symbol: "SAP",
	}
	orcl = model.Asset{
		Symbol: "ORCL",
	}
	googl = model.Asset{
		Symbol: "GOOGL",
	}
	assets = []model.Asset{
		aapl, sap, orcl, googl,
	}
)

func newBroker(t *testing.T) broker.Broker {
	broker := fake.NewBroker(initialCash)
	quotes := make([]model.Quote, len(assets))
	for i, asset := range assets {
		quotes[i] = fixtures.GetQuotes(t, asset.Symbol)[0]
	}
	broker.SetQuotes(quotes...)
	return broker
}

func TestAvailableCash(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	broker := newBroker(t)

	availableCash, err := broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(availableCash).To(gomega.Equal(initialCash))
}

func TestGetQuotes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	broker := newBroker(t)

	quotes, err := broker.GetQuotes(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(quotes)).To(gomega.Equal(4))

	for i, asset := range assets {
		g.Expect(quotes[i].Asset).To(gomega.Equal(asset))
		g.Expect(quotes[i].Price).To(gomega.BeNumerically(">", 0))
	}
}

func TestBuySell(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	broker := newBroker(t)

	// Buy
	errs := broker.Execute(model.Order{
		Asset:    googl,
		Quantity: 1.0,
		Price:    100.0,
		Type:     model.OrderTypeBuy,
	})
	g.Expect(len(errs)).To(gomega.Equal(0))
	availableCash, err := broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(availableCash).To(gomega.BeNumerically("~", initialCash-100))
	positions, err := broker.GetPositions(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(positions)).To(gomega.Equal(4))
	for _, position := range positions {
		g.Expect(position.Asset).ToNot(gomega.Equal(model.Asset{}))
		if position.Asset == googl {
			g.Expect(position.AvgBuyPrice).To(gomega.Equal(100.0))
			g.Expect(position.Quantity).To(gomega.Equal(1.0))
		} else {
			g.Expect(position.AvgBuyPrice).To(gomega.Equal(0.0))
			g.Expect(position.Quantity).To(gomega.Equal(0.0))
		}
	}

	// Sell
	errs = broker.Execute(model.Order{
		Asset:    googl,
		Quantity: 1,
		Price:    100,
		Type:     model.OrderTypeSell,
	})
	g.Expect(len(errs)).To(gomega.Equal(0))
	availableCash, err = broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(availableCash).To(gomega.BeNumerically("~", initialCash))
	positions, err = broker.GetPositions(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(positions)).To(gomega.Equal(4))
	for _, position := range positions {
		g.Expect(position.AvgBuyPrice).To(gomega.Equal(0.0))
		g.Expect(position.Quantity).To(gomega.Equal(0.0))
		g.Expect(position.Asset).ToNot(gomega.Equal(model.Asset{}))
	}
}

func TestBuyMoreSellSome(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	broker := newBroker(t)

	// Buy
	errs := broker.Execute(model.Order{
		Asset:    googl,
		Quantity: 1,
		Price:    50,
		Type:     model.OrderTypeBuy,
	})
	g.Expect(len(errs)).To(gomega.Equal(0))
	availableCash, err := broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(availableCash).To(gomega.BeNumerically("~", initialCash-50))
	positions, err := broker.GetPositions(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(positions)).To(gomega.Equal(4))

	// Buy more
	errs = broker.Execute(model.Order{
		Asset:    googl,
		Quantity: 1,
		Price:    50,
		Type:     model.OrderTypeBuy,
	})
	g.Expect(len(errs)).To(gomega.Equal(0))
	availableCash, err = broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(availableCash).To(gomega.BeNumerically("~", initialCash-100))
	positions, err = broker.GetPositions(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(positions)).To(gomega.Equal(4))

	// Sell some
	errs = broker.Execute(model.Order{
		Asset:    googl,
		Quantity: 1,
		Price:    20,
		Type:     model.OrderTypeSell,
	})
	g.Expect(len(errs)).To(gomega.Equal(0))
	availableCash, err = broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(availableCash).To(gomega.BeNumerically("~", initialCash-80))
	positions, err = broker.GetPositions(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(positions)).To(gomega.Equal(4))

}

func TestGetPortfolio(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	broker := newBroker(t)

	quotes, err := broker.GetQuotes(googl, sap)
	g.Expect(err).To(gomega.BeNil())

	quantity := 1.0
	portfolioValue := 0.0

	// Buy googl
	errs := broker.Execute(model.Order{
		Asset:    googl,
		Quantity: quantity,
		Price:    quotes[0].Price,
		Type:     model.OrderTypeBuy,
	})
	g.Expect(len(errs)).To(gomega.Equal(0))
	availableCash, err := broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	portfolioValue += quotes[0].Price * quantity
	g.Expect(availableCash).To(gomega.BeNumerically("~", initialCash-portfolioValue))
	positions, err := broker.GetPositions(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(positions)).To(gomega.Equal(4))

	// Buy sap
	errs = broker.Execute(model.Order{
		Asset:    sap,
		Quantity: quantity,
		Price:    quotes[1].Price,
		Type:     model.OrderTypeBuy,
	})
	g.Expect(len(errs)).To(gomega.Equal(0))
	availableCash, err = broker.GetAvailableCash()
	g.Expect(err).To(gomega.BeNil())
	portfolioValue += quotes[1].Price * quantity
	g.Expect(availableCash).To(gomega.BeNumerically("~", initialCash-portfolioValue))
	positions, err = broker.GetPositions(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(positions)).To(gomega.Equal(4))

	// Get portfolio
	portfolio, err := broker.GetPortfolio(assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(portfolio.Weights[sap]).To(gomega.BeNumerically(">", 0))
	g.Expect(portfolio.Weights[googl]).To(gomega.BeNumerically(">", 0))
	g.Expect(portfolio.Weights[googl] + portfolio.Weights[sap]).To(gomega.BeNumerically("~", 1))
	g.Expect(portfolio.TotalValue).To(gomega.BeNumerically("~", portfolioValue))
	g.Expect(portfolio.Prices[googl]).To(gomega.BeNumerically("~", quotes[0].Price))
	g.Expect(portfolio.Prices[sap]).To(gomega.BeNumerically("~", quotes[1].Price))
	g.Expect(portfolio.Quantities[googl]).To(gomega.BeNumerically("~", 1))
	g.Expect(portfolio.Quantities[sap]).To(gomega.BeNumerically("~", 1))
}
