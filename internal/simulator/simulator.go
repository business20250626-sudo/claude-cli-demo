// Package simulator cross-checks the analytic RTP of a slot.Config with a
// fixed-seed Monte-Carlo run. The analytic value is authoritative; the
// simulation is a deterministic, non-flaky sanity check.
package simulator

import (
	"math/rand/v2"

	slot "example.com/example-game/slot"
)

// Report holds the aggregate results of a Monte-Carlo run (bet = 1 per spin).
type Report struct {
	Iterations  int
	TotalBet    int
	TotalWin    int
	Hits        int
	SymbolHits  [slot.NumSymbols]int
	MeasuredRTP float64
	MeasuredHit float64
}

// AnalyticRTP returns the exact theoretical RTP of cfg as a float64.
func AnalyticRTP(cfg slot.Config) float64 {
	f, _ := cfg.RTP().Float64()
	return f
}

// SimulateRTP runs `iterations` seeded spins (bet = 1) on a fresh Engine and
// returns the aggregate Report. The seed makes the run fully reproducible.
func SimulateRTP(cfg slot.Config, seed uint64, iterations int) Report {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	eng := slot.NewEngine(cfg, rng)
	rep := Report{Iterations: iterations}
	for i := 0; i < iterations; i++ {
		res := eng.Spin()
		rep.TotalBet++
		if res.IsWin {
			rep.Hits++
			rep.TotalWin += res.Multiplier // bet == 1
			rep.SymbolHits[res.WinSymbol]++
		}
	}
	if rep.TotalBet > 0 {
		rep.MeasuredRTP = float64(rep.TotalWin) / float64(rep.TotalBet)
		rep.MeasuredHit = float64(rep.Hits) / float64(rep.TotalBet)
	}
	return rep
}
