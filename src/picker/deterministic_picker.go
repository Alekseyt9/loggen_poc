package picker

// DeterministicPicker implements Picker with smooth weighted round-robin and rounding.
type DeterministicPicker struct {
	numTypes     int
	defGlobal    []float64
	defAllowed   [][]float64
	accBernoulli float64
	coin         bool
}

func NewDeterministicPicker(numTypes int) *DeterministicPicker {
	return &DeterministicPicker{numTypes: numTypes}
}

func (d *DeterministicPicker) GeomLen(mean float64) int {
	if mean < 1 {
		return 1
	}

	k := int(mean + 0.5)
	if k < 1 {
		k = 1
	}

	return k
}

func (d *DeterministicPicker) Poisson(lambda float64) int {
	if lambda <= 0 {
		return 0
	}

	return int(lambda + 0.5)
}

func (d *DeterministicPicker) stepDeficits(def []float64, weights []float64, allowMask []bool) int {
	sum := 0.0
	for i, w := range weights {
		if allowMask == nil || allowMask[i] {
			sum += w
		}
	}

	if sum <= 0 {
		for i := range weights {
			if allowMask == nil || allowMask[i] {
				return i
			}
		}
		return 0
	}

	for i, w := range weights {
		if allowMask == nil || allowMask[i] {
			def[i] += w / sum
		}
	}

	best, bestVal := -1, -1.0
	for i := range weights {
		if (allowMask == nil || allowMask[i]) && def[i] > bestVal {
			bestVal, best = def[i], i
		}
	}

	if best >= 0 {
		def[best] -= 1.0
		return best
	}

	return 0
}

func (d *DeterministicPicker) ensureAllowed(from int) {
	if d.defAllowed == nil || len(d.defAllowed) != d.numTypes {
		d.defAllowed = make([][]float64, d.numTypes)
	}

	if d.defAllowed[from] == nil || len(d.defAllowed[from]) != d.numTypes {
		d.defAllowed[from] = make([]float64, d.numTypes)
	}
}

func (d *DeterministicPicker) WeightedGlobal(weights []float64) int {
	if d.defGlobal == nil || len(d.defGlobal) != len(weights) {
		d.defGlobal = make([]float64, len(weights))
	}

	return d.stepDeficits(d.defGlobal, weights, nil)
}

func (d *DeterministicPicker) WeightedAllowed(from int, weights []float64, allow [][]bool) int {
	d.ensureAllowed(from)

	return d.stepDeficits(d.defAllowed[from], weights, allow[from])
}

func (d *DeterministicPicker) Bernoulli(p float64) bool {
	if p <= 0 {
		return false
	}

	if p >= 1 {
		return true
	}

	d.accBernoulli += p

	if d.accBernoulli >= 1 {
		d.accBernoulli -= 1
		return true
	}

	return false
}

func (d *DeterministicPicker) CoinFlip() bool { d.coin = !d.coin; return d.coin }
