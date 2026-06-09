package slot

import "math/rand/v2"

// reel is a single reel described by a per-symbol integer weight. The symbol
// frequency is the probability model: P(symbol) = weight / total.
type reel struct {
	weights [NumSymbols]int
	total   int
}

// newReel builds a reel from a symbol->weight map. It panics if the total
// weight is non-positive, which is a construction-time programmer error.
func newReel(weights map[Symbol]int) reel {
	var r reel
	for s, w := range weights {
		r.weights[s] = w
		r.total += w
	}
	if r.total <= 0 {
		panic("slot: reel has non-positive total weight")
	}
	return r
}

// symbolForRoll maps a roll in [0,total) to a symbol via cumulative weights.
// Every symbol with positive weight is reachable; there is no modulo bias.
func (r reel) symbolForRoll(x int) Symbol {
	for s := 0; s < NumSymbols; s++ {
		if x < r.weights[s] {
			return Symbol(s)
		}
		x -= r.weights[s]
	}
	panic("slot: roll out of range for reel total")
}

// spin draws one symbol using the injected RNG.
func (r reel) spin(rng *rand.Rand) Symbol {
	return r.symbolForRoll(rng.IntN(r.total))
}
