package main_test

import (
	"testing"

	"github.com/MitchK/autorobin/lib/autopilot"
	"github.com/MitchK/autorobin/lib/broker/fake"
	"github.com/MitchK/autorobin/lib/model"

	"github.com/onsi/gomega"
)

func TestRebalance(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Test data
	cash := 100.0
	a := model.Asset{
		Symbol: "A",
	}
	b := model.Asset{
		Symbol: "B",
	}
	c := model.Asset{
		Symbol: "C",
	}
	assets := []model.Asset{a, b, c}
	desiredWeights := model.Weights{
		a: .5,
		b: .25,
		c: .25,
	}

	broker := fake.NewBroker(cash)

	// Set quotes
	quotes := make([]model.Quote, len(assets))
	for i, asset := range assets {
		quotes[i] = model.Quote{
			Asset: asset,
			Price: 1.0,
		}
	}
	broker.SetQuotes(quotes...)

	// Create auto pilot
	autopilot, err := autopilot.NewAutopilot(broker)
	g.Expect(err).To(gomega.BeNil())

	// Rebalance
	partials := true
	orders, err := autopilot.Rebalance(desiredWeights, partials, assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(orders).ToNot(gomega.BeEmpty())
	g.Expect(len(orders)).To(gomega.Equal(3))
	for _, order := range orders {
		if order.Asset == a {
			g.Expect(order.Description).To(gomega.Equal("Allocation of unbound cash"))
			g.Expect(order.Price).To(gomega.BeNumerically("~", 1))
			g.Expect(order.Quantity).To(gomega.BeNumerically("~", 50))
			g.Expect(order.Type).To(gomega.Equal(model.OrderTypeBuy))
		} else if order.Asset == b {
			g.Expect(order.Description).To(gomega.Equal("Allocation of unbound cash"))
			g.Expect(order.Price).To(gomega.BeNumerically("~", 1))
			g.Expect(order.Quantity).To(gomega.BeNumerically("~", 25))
			g.Expect(order.Type).To(gomega.Equal(model.OrderTypeBuy))
		} else if order.Asset == c {
			g.Expect(order.Description).To(gomega.Equal("Allocation of unbound cash"))
			g.Expect(order.Price).To(gomega.BeNumerically("~", 1))
			g.Expect(order.Quantity).To(gomega.BeNumerically("~", 25))
			g.Expect(order.Type).To(gomega.Equal(model.OrderTypeBuy))
		} else {
			t.Errorf("unexpected asset %s", order.Asset)
		}
	}
	// errs := broker.Execute(orders...)
	// g.Expect(len(errs)).To(gomega.BeZero())
	// cash, err = broker.GetAvailableCash()
	// g.Expect(err).To(gomega.BeNil())
	// g.Expect(cash).To(gomega.BeNumerically("~", 0))

	// // // Rebalance again, there should be no orders
	// // orders, err = autopilot.Rebalance(desiredWeights, partials, minReturn, assets...)
	// // g.Expect(err).To(gomega.BeNil())
	// // g.Expect(orders).To(gomega.BeEmpty())

	// // Simulate a price change in stock market
	// quotes = []model.Quote{
	// 	model.Quote{
	// 		Asset: a,
	// 		Price: 2,
	// 	},
	// 	model.Quote{
	// 		Asset: b,
	// 		Price: 1,
	// 	},
	// 	model.Quote{
	// 		Asset: c,
	// 		Price: 1,
	// 	},
	// }
	// broker.SetQuotes(quotes...)

	// // Rebalance and execute orders
	// orders, err = autopilot.Rebalance(desiredWeights, partials, minReturn, assets...)
	// g.Expect(err).To(gomega.BeNil())
	// errs = broker.Execute(orders...)
	// for _, err := range errs {
	// 	fmt.Println(err)
	// }
	// cash, err = broker.GetAvailableCash()
	// g.Expect(err).To(gomega.BeNil())
	// g.Expect(cash).To(gomega.BeNumerically("~", 25))
	// // p, err := broker.GetPortfolio(assets...)
	// // g.Expect(err).To(gomega.BeNil())
	// // g.Expect(len(errs)).To(gomega.BeZero())
	// orders, err = autopilot.Rebalance(desiredWeights, partials, minReturn, assets...)
	// g.Expect(err).To(gomega.BeNil())
	// g.Expect(orders).ToNot(gomega.BeEmpty())
	// errs = broker.Execute(orders...)
	// for _, err := range errs {
	// 	fmt.Println(err)
	// }
	// // p, err = broker.GetPortfolio(assets...)
	// // g.Expect(err).To(gomega.BeNil())
}
