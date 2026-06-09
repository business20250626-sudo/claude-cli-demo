// Command slot-verify prints a theoretical-vs-simulated RTP comparison for the
// default slot configuration.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"text/tabwriter"

	"example.com/example-game/internal/simulator"
	slot "example.com/example-game/slot"
)

func main() {
	spins := flag.Int("spins", 10_000_000, "number of simulated spins")
	seed := flag.Uint64("seed", 1, "RNG seed for the simulation")
	flag.Parse()

	cfg := slot.DefaultConfig()
	rep := simulator.SimulateRTP(cfg, *seed, *spins)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "symbol\tweight\tmult\ttheo.prob\tsim.prob\ttheo.rtp\tsim.rtp\n")
	for _, s := range slot.Symbols() {
		theoProb, _ := cfg.SymbolProbability(s).Float64()
		mult := cfg.Multiplier(s)
		simProb := float64(rep.SymbolHits[s]) / float64(rep.Iterations)
		fmt.Fprintf(tw, "%s\t%d\t%dx\t%.6f\t%.6f\t%.6f\t%.6f\n",
			s, cfg.Weight(s), mult, theoProb, simProb, theoProb*float64(mult), simProb*float64(mult))
	}
	theoHit, _ := cfg.HitFrequency().Float64()
	theoRTP := simulator.AnalyticRTP(cfg)
	fmt.Fprintf(tw, "TOTAL\t\t\thit %.6f\thit %.6f\trtp %.6f\trtp %.6f\n",
		theoHit, rep.MeasuredHit, theoRTP, rep.MeasuredRTP)
	tw.Flush()

	const tol = 0.005
	if math.Abs(rep.MeasuredRTP-theoRTP) <= tol {
		fmt.Printf("\n✓ simulated RTP within %.3f of theoretical (%d spins, seed %d)\n", tol, *spins, *seed)
	} else {
		fmt.Printf("\n✗ simulated RTP off by %.6f (> %.3f)\n", math.Abs(rep.MeasuredRTP-theoRTP), tol)
		os.Exit(1)
	}
}
