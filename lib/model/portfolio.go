package model

// PortfolioSnapshot PortfolioSnapshot
type PortfolioSnapshot struct {
}

// Portfolio Portfolio
type Portfolio struct {
	Weights    Weights
	Quantities Quantities
	Prices     Prices
	TotalValue float64
	Assets     []Asset
}
