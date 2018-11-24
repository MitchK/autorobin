package data

import (
	"time"

	"github.com/MitchK/autorobin/lib/model"
)

// Adapter Adapter
type Adapter interface {
	GetDailyDesc(from, to time.Time, assets ...model.Asset) ([]model.Quote, error)
}
