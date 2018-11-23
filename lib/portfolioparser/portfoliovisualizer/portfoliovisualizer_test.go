package portfoliovisualizer_test

import (
	"strings"
	"testing"

	"github.com/MitchK/autorobin/lib/model"
	"github.com/MitchK/autorobin/lib/model/fixtures"
	"github.com/MitchK/autorobin/lib/portfolioparser/portfoliovisualizer"
	"github.com/onsi/gomega"
)

func TestParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	weights, assets, err := portfoliovisualizer.NewParser().Parse(strings.NewReader(fixtures.ExamplePVCSVStr(t)))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(weights)).To(gomega.Equal(4))
	g.Expect(len(assets)).To(gomega.Equal(4))

	g.Expect(weights[model.Asset{Symbol: "TRQ"}]).To(gomega.BeNumerically("~", 0.3890))
	g.Expect(weights[model.Asset{Symbol: "LRAD"}]).To(gomega.BeNumerically("~", 0.3603))
	g.Expect(weights[model.Asset{Symbol: "ISIG"}]).To(gomega.BeNumerically("~", 0.1626))
	g.Expect(weights[model.Asset{Symbol: "CELH"}]).To(gomega.BeNumerically("~", 0.0880))

}
