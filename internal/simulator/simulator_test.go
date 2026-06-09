package simulator_test

import (
	"math"
	"testing"

	"example.com/example-game/internal/simulator"
	slot "example.com/example-game/slot"
)

func TestSimulateDeterministic(t *testing.T) {
	a := simulator.SimulateRTP(slot.DefaultConfig(), 7, 50_000)
	b := simulator.SimulateRTP(slot.DefaultConfig(), 7, 50_000)
	if a != b {
		t.Errorf("same-seed reports differ:\n a=%+v\n b=%+v", a, b)
	}
}

func TestSimulateRTPMatchesAnalytic(t *testing.T) {
	cfg := slot.DefaultConfig()
	const iters = 2_000_000
	rep := simulator.SimulateRTP(cfg, 1, iters)

	if rep.TotalBet != iters {
		t.Errorf("TotalBet = %d, want %d", rep.TotalBet, iters)
	}

	// Self-consistency: per-symbol hits must sum to total hits.
	sum := 0
	for _, h := range rep.SymbolHits {
		sum += h
	}
	if sum != rep.Hits {
		t.Errorf("sum(SymbolHits) = %d, want Hits = %d", sum, rep.Hits)
	}

	// Hit frequency is the stable estimator (low variance), so pin it tightly.
	analyticHit, _ := cfg.HitFrequency().Float64() // 0.0955
	const hitTol = 0.002
	if math.Abs(rep.MeasuredHit-analyticHit) > hitTol {
		t.Errorf("MeasuredHit=%.6f analyticHit=%.6f, diff > %v", rep.MeasuredHit, analyticHit, hitTol)
	}

	// RTP variance is dominated by the rare 640x jackpot, so use a looser,
	// variance-justified tolerance (SE ~= 0.005 at 2M spins; this is ~3 SE).
	analyticRTP := simulator.AnalyticRTP(cfg) // 0.949
	const rtpTol = 0.015
	if math.Abs(rep.MeasuredRTP-analyticRTP) > rtpTol {
		t.Errorf("MeasuredRTP=%.6f analyticRTP=%.6f, diff > %v", rep.MeasuredRTP, analyticRTP, rtpTol)
	}
}
