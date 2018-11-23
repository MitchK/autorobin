package autopilot_test

import (
	"testing"

	"github.com/MitchK/autorobin/lib/autopilot"
	"github.com/MitchK/autorobin/lib/mocks"
	"github.com/MitchK/autorobin/lib/model"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
)

func TestRebalance(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBroker := mocks.NewMockBroker(mockCtrl)

	// fixtures
	a := model.Asset{Symbol: "A"}
	b := model.Asset{Symbol: "B"}
	c := model.Asset{Symbol: "C"}
	d := model.Asset{Symbol: "D"}
	e := model.Asset{Symbol: "E"}
	assets := []model.Asset{a, b, c, d, e}
	availableCash := 130.0
	desiredWeights := model.Weights{
		a: 0.25,
		b: 0.25,
		c: 0.25,
		d: 0.25,
		e: 0.0,
	}
	actualPortfolio := model.Portfolio{
		Weights: model.Weights{
			a: 0.00, // Stock A is missing
			b: 0.25, // already desired
			c: 0.30, // 0.05 more than desired
			d: 0.20, // 0.05 less than desired
			e: 0.25, // undesired stock
		},
		Prices: model.Prices{
			a: 1.0,
			b: 1.0,
			c: 1.0,
			d: 1.0,
			e: 1.0,
		},
		Quantities: model.Quantities{
			a: 0.0,
			b: 25.0,
			c: 30.0,
			d: 20.0,
			e: 25.0,
		},
		TotalValue: 100.0,
	}
	positions := []model.Position{
		model.Position{
			Asset:       a,
			AvgBuyPrice: 0.0,
			Quantity:    actualPortfolio.Quantities[a],
		},
		model.Position{
			Asset:       b,
			AvgBuyPrice: 0.9,
			Quantity:    actualPortfolio.Quantities[b],
		},
		model.Position{
			Asset:       c,
			AvgBuyPrice: 0.9,
			Quantity:    actualPortfolio.Quantities[c],
		},
		model.Position{
			Asset:       d,
			AvgBuyPrice: 0.9,
			Quantity:    actualPortfolio.Quantities[d],
		},
		model.Position{
			Asset:       e,
			AvgBuyPrice: 0.9,
			Quantity:    actualPortfolio.Quantities[e],
		},
	}

	// mock out function calls
	mockBroker.EXPECT().GetAvailableCash().Return(availableCash, nil)
	mockBroker.EXPECT().GetPortfolio(
		gomock.Eq(a),
		gomock.Eq(b),
		gomock.Eq(c),
		gomock.Eq(d),
		gomock.Eq(e)).Return(actualPortfolio, nil)
	mockBroker.EXPECT().GetPositions(
		gomock.Eq(a),
		gomock.Eq(b),
		gomock.Eq(c),
		gomock.Eq(d),
		gomock.Eq(e)).Return(positions, nil)
	autopilot, err := autopilot.NewAutopilot(mockBroker)
	g.Expect(err).To(gomega.BeNil())

	orders, err := autopilot.Rebalance(desiredWeights, true, -1, assets...)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(orders).ToNot(gomega.BeNil())
	g.Expect(orders).ToNot(gomega.BeEmpty())

	orderVolume := 0.0
	for _, order := range orders {
		if order.Type == model.OrderTypeBuy {
			orderVolume += order.Price * order.Quantity
		}
	}
	g.Expect(orderVolume).To(gomega.BeNumerically("~", availableCash))
}
