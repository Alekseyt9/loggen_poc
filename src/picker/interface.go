package picker

// Picker abstracts all choices used during generation (random or deterministic).
type Picker interface {
	// Получение длины по средрему значению
	GeomLen(mean float64) int

	// Сколько раз случится событие, если среднее = lambda
	Poisson(lambda float64) int

	// Глобальное распределение по переходам
	WeightedGlobal(weights []float64) int

	// Локальное распределение по доступным переходам
	WeightedAllowed(fromType int, weights []float64, allow [][]bool) int

	// Событие с вероятностью p
	Bernoulli(p float64) bool

	// Событие 50 на 50
	CoinFlip() bool
}
