package portfolioparser

import (
	"io"

	"github.com/MitchK/autorobin/lib/model"
)

// Parser Parser
type Parser interface {
	Parse(reader io.Reader) (*model.Portfolio, []model.Asset, error)
}
