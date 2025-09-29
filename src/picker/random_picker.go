package picker

import (
	"math"
	"math/rand"
)

// RandomPicker uses math/rand to implement Picker.
type RandomPicker struct{ rnd *rand.Rand }

func NewRandomPicker(seed int64) *RandomPicker {
	return &RandomPicker{rnd: rand.New(rand.NewSource(seed))}
}

func (p *RandomPicker) GeomLen(mean float64) int {
	if mean < 1 {
		mean = 1
	}

	pv := 1.0 / mean
	k := 1
	for p.rnd.Float64() > pv {
		k++
	}

	return k
}

func (p *RandomPicker) Poisson(lambda float64) int {
	if lambda <= 0 {
		return 0
	}

	L := math.Exp(-lambda)
	k := 0
	acc := 1.0

	for {
		k++
		acc *= p.rnd.Float64()
		if acc <= L {
			break
		}
	}

	return k - 1
}

func (p *RandomPicker) WeightedGlobal(w []float64) int {
	sum := 0.0
	for _, x := range w {
		sum += x
	}

	if sum <= 0 {
		return 0
	}

	u := p.rnd.Float64() * sum
	acc := 0.0

	for i, x := range w {
		acc += x
		if u <= acc {
			return i
		}
	}

	return len(w) - 1
}

func (p *RandomPicker) WeightedAllowed(from int, weights []float64, allow [][]bool) (int, bool) {
    row := allow[from]
    sum := 0.0
    idx := make([]int, 0, len(row))
    wts := make([]float64, 0, len(row))

	for t, ok := range row {
		if ok {
			w := weights[t]
			if w > 0 {
				sum += w
				idx = append(idx, t)
				wts = append(wts, w)
			}
		}
	}

    if len(idx) == 0 {
        return -1, false
    }

	u := p.rnd.Float64() * sum
	acc := 0.0

	for i, id := range idx {
		acc += wts[i]
		if u <= acc {
            return id, true
        }
    }

    return idx[len(idx)-1], true
}

func (p *RandomPicker) Bernoulli(prob float64) bool { return prob > 0 && p.rnd.Float64() < prob }

func (p *RandomPicker) CoinFlip() bool { return p.rnd.Float64() < 0.5 }
