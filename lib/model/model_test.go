package model_test

import (
	"testing"

	"github.com/MitchK/autorobin/lib/model"
	"github.com/onsi/gomega"
)

func TestDiff(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	portfolio1 := model.Portfolio{
		Weights: model.Weights{
			model.Asset{Symbol: "A"}: 0.25,
			model.Asset{Symbol: "B"}: 0.25,
			model.Asset{Symbol: "C"}: 0.25,
			model.Asset{Symbol: "D"}: 0.25,
		},
	}

	portfolio2 := model.Portfolio{
		Weights: model.Weights{
			model.Asset{Symbol: "B"}: 0.20,
			model.Asset{Symbol: "C"}: 0.20,
			model.Asset{Symbol: "D"}: 0.20,
			model.Asset{Symbol: "E"}: 0.20,
			model.Asset{Symbol: "F"}: 0.20,
		},
	}

	diff := portfolio1.Weights.Diff(portfolio2.Weights)
	g.Expect(len(diff)).To(gomega.Equal(6))
	g.Expect(diff[model.Asset{Symbol: "A"}]).To(gomega.BeNumerically("~", 0.25))
	g.Expect(diff[model.Asset{Symbol: "B"}]).To(gomega.BeNumerically("~", 0.05))
	g.Expect(diff[model.Asset{Symbol: "C"}]).To(gomega.BeNumerically("~", 0.05))
	g.Expect(diff[model.Asset{Symbol: "D"}]).To(gomega.BeNumerically("~", 0.05))
	g.Expect(diff[model.Asset{Symbol: "E"}]).To(gomega.BeNumerically("~", -0.20))
	g.Expect(diff[model.Asset{Symbol: "F"}]).To(gomega.BeNumerically("~", -0.20))

}
