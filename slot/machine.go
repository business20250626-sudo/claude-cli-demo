package slot

import (
	"fmt"
	"math/rand/v2"
)

// Machine wraps an Engine with credit accounting. It is not safe for
// concurrent use; guard it with a mutex if shared across goroutines.
type Machine struct {
	engine  *Engine
	balance int
}

// NewMachine returns a Machine with the given starting balance (in credits)
// and injected RNG.
func NewMachine(cfg Config, balance int, rng *rand.Rand) *Machine {
	return &Machine{engine: NewEngine(cfg, rng), balance: balance}
}

// Balance returns the current credit balance.
func (m *Machine) Balance() int { return m.balance }

// Spin validates the bet, runs one engine spin, checks result invariants, and
// settles the balance (balance += payout - bet). On any error the balance is
// left unchanged.
func (m *Machine) Spin(bet int) (SpinResult, error) {
	switch {
	case bet <= 0:
		return SpinResult{}, fmt.Errorf("%w: %d", ErrInvalidBet, bet)
	case bet > m.balance:
		return SpinResult{}, ErrInsufficientBalance
	}
	res := m.engine.Spin()
	res.Payout = bet * res.Multiplier
	if err := validate(res); err != nil {
		return SpinResult{}, err
	}
	m.balance += res.Payout - bet
	res.Balance = m.balance
	return res, nil
}

// validate guards against engine/paytable scoring bugs before crediting.
func validate(res SpinResult) error {
	switch {
	case res.Multiplier < 0:
		return fmt.Errorf("slot: negative multiplier %d", res.Multiplier)
	case res.IsWin != (res.Multiplier > 0):
		return fmt.Errorf("slot: win flag %v inconsistent with multiplier %d", res.IsWin, res.Multiplier)
	case res.IsWin && res.WinSymbol == NoSymbol:
		return fmt.Errorf("slot: win recorded with no winning symbol")
	}
	return nil
}
