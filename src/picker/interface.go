package picker

// Picker abstracts all choices used during generation (random or deterministic).
type Picker interface {
	GeomLen(mean float64) int
	Poisson(lambda float64) int
	WeightedGlobal(weights []float64) int
	WeightedAllowed(fromType int, weights []float64, allow [][]bool) int
	Bernoulli(p float64) bool
	CoinFlip() bool
}
