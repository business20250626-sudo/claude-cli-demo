package slot

import "math/rand/v2"

// Engine is the pure spin core: a Config plus an injected RNG. It holds no
// money state, so it is the unit the simulator and RTP math exercise.
type Engine struct {
	cfg Config
	rng *rand.Rand
}

// NewEngine returns an Engine bound to cfg and the injected RNG. Use
// rand.NewPCG for reproducible sequences (tests, simulation) and
// rand.NewChaCha8 with a crypto/rand seed in production.
func NewEngine(cfg Config, rng *rand.Rand) *Engine {
	return &Engine{cfg: cfg, rng: rng}
}

// Config returns the engine's configuration.
func (e *Engine) Config() Config { return e.cfg }

// Spin draws one symbol per reel and evaluates the payline. Payout and Balance
// are left zero; Machine sets them.
func (e *Engine) Spin() SpinResult {
	var board [reelCount]Symbol
	for i := range board {
		board[i] = e.cfg.reels[i].spin(e.rng)
	}
	mult, win := e.cfg.evaluate(board)
	return SpinResult{
		Reels:      board,
		IsWin:      mult > 0,
		WinSymbol:  win,
		Multiplier: mult,
	}
}
