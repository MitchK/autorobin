package tiingo

import "time"

type tiingoError struct {
	Detail string `json:"detail"`
}

type quote struct {
	Date  time.Time `json:"date"`
	Close float64   `json:"close"`
}
