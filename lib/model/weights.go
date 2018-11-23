package model

// Weights Weights
type Weights map[Asset]float64

// Diff Get diff of weights of the intersection. Negative numbers mean that the other portfolio has more, positive means it has less
func (weights Weights) Diff(otherWeights Weights) WeightsDiff {
	diff := WeightsDiff{}
	for asset, weight := range weights {
		diff[asset] = weight
	}
	for asset, weight := range otherWeights {
		if _, ok := diff[asset]; ok {
			diff[asset] -= weight
		} else {
			diff[asset] = -weight
		}
	}
	return diff
}
