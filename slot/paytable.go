package slot

import "math/big"

// evaluate scores a board on the single payline: a 3-of-a-kind pays the
// matched symbol's multiplier; anything else is a loss.
func (c Config) evaluate(board [reelCount]Symbol) (multiplier int, win Symbol) {
	if board[0] == board[1] && board[1] == board[2] {
		s := board[0]
		return c.multipliers[s], s
	}
	return 0, NoSymbol
}

// RTP returns the exact theoretical return-to-player as a reduced rational:
// sum over symbols of multiplier * P(3-of-a-kind).
func (c Config) RTP() *big.Rat {
	sum := new(big.Rat)
	for s := 0; s < NumSymbols; s++ {
		term := new(big.Rat).Mul(
			c.SymbolProbability(Symbol(s)),
			big.NewRat(int64(c.multipliers[s]), 1),
		)
		sum.Add(sum, term)
	}
	return sum
}

// HitFrequency returns the exact probability that a spin wins, as a reduced
// rational: the sum of all 3-of-a-kind probabilities.
func (c Config) HitFrequency() *big.Rat {
	sum := new(big.Rat)
	for s := 0; s < NumSymbols; s++ {
		sum.Add(sum, c.SymbolProbability(Symbol(s)))
	}
	return sum
}
