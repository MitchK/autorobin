package model

const (
	// OrderTypeBuy OrderTypeBuy
	OrderTypeBuy OrderType = iota + 1

	// OrderTypeSell OrderTypeSell
	OrderTypeSell
)

// OrderType OrderType
type OrderType int

// Order Order
type Order struct {
	Description string
	Type        OrderType
	Asset       Asset
	Price       float64
	Quantity    float64
}
