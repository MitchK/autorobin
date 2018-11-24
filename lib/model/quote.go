package model

import "time"

// Quote Quote
type Quote struct {
	Asset Asset
	Price float64
	Date  time.Time
}
