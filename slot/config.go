package slot

import "math/big"

// reelCount is the number of reels on the single payline (1x3).
const reelCount = 3

// Config is the static description of a slot machine: its reels and the
// 3-of-a-kind multiplier for each symbol. Build one with NewConfig or
// DefaultConfig; it carries no runtime state and is safe to copy.
type Config struct {
	reels       [reelCount]reel
	multipliers [NumSymbols]int
}

// NewConfig builds a Config whose three reels are identical, from a
// symbol->weight map and a symbol->multiplier map.
func NewConfig(weights, multipliers map[Symbol]int) Config {
	reel := newReel(weights)
	var c Config
	for i := range c.reels {
		c.reels[i] = reel
	}
	for s, m := range multipliers {
		c.multipliers[s] = m
	}
	return c
}

// DefaultConfig returns the reference 5-symbol configuration whose exact RTP
// is 949/1000 (94.9%) and hit frequency is 191/2000 (9.55%).
func DefaultConfig() Config {
	return NewConfig(
		map[Symbol]int{Cherry: 8, Lemon: 6, Bell: 3, Star: 2, Seven: 1},
		map[Symbol]int{Cherry: 6, Lemon: 12, Bell: 24, Star: 80, Seven: 640},
	)
}

// Weight returns the weight of a symbol on the first reel (reels are identical
// in the default and NewConfig configurations).
func (c Config) Weight(s Symbol) int { return c.reels[0].weights[s] }

// Multiplier returns the 3-of-a-kind payout multiplier for a symbol.
func (c Config) Multiplier(s Symbol) int { return c.multipliers[s] }

// SymbolProbability returns the exact probability of a 3-of-a-kind of the
// given symbol across all reels.
func (c Config) SymbolProbability(s Symbol) *big.Rat {
	p := big.NewRat(1, 1)
	for _, r := range c.reels {
		p.Mul(p, big.NewRat(int64(r.weights[s]), int64(r.total)))
	}
	return p
}
