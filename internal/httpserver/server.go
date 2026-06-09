// Package httpserver serves a minimal web UI for watching slot spins: an
// embedded HTML page plus JSON /spin and /reset endpoints over one shared,
// mutex-guarded Machine.
package httpserver

import (
	_ "embed"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"sync"

	slot "example.com/example-game/slot"
)

//go:embed index.html
var indexHTML []byte

// Server holds one shared Machine guarded by a mutex (net/http serves
// requests concurrently).
type Server struct {
	mu      sync.Mutex
	cfg     slot.Config
	start   int
	rng     *rand.Rand
	machine *slot.Machine
}

// New returns a Server with a Machine started at startBalance.
func New(cfg slot.Config, startBalance int, rng *rand.Rand) *Server {
	return &Server{
		cfg:     cfg,
		start:   startBalance,
		rng:     rng,
		machine: slot.NewMachine(cfg, startBalance, rng),
	}
}

// Handler returns the HTTP routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/spin", s.handleSpin)
	mux.HandleFunc("/reset", s.handleReset)
	return mux
}

type spinRequest struct {
	Bet int `json:"bet"`
}

type spinResponse struct {
	Reels      []string `json:"reels"`
	IsWin      bool     `json:"isWin"`
	WinSymbol  string   `json:"winSymbol"`
	Multiplier int      `json:"multiplier"`
	Payout     int      `json:"payout"`
	Balance    int      `json:"balance"`
	Error      string   `json:"error,omitempty"`
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func (s *Server) handleSpin(w http.ResponseWriter, r *http.Request) {
	req := spinRequest{Bet: 1}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req) // empty body keeps default bet
	}
	if req.Bet <= 0 {
		req.Bet = 1
	}

	s.mu.Lock()
	res, err := s.machine.Spin(req.Bet)
	balance := s.machine.Balance()
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(spinResponse{Balance: balance, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(toResponse(res))
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.machine = slot.NewMachine(s.cfg, s.start, s.rng)
	balance := s.machine.Balance()
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spinResponse{
		Reels:     []string{"❔", "❔", "❔"},
		WinSymbol: slot.NoSymbol.String(),
		Balance:   balance,
	})
}

// toResponse converts a SpinResult to the wire shape (glyph strings at the
// API boundary, separate from the engine's integer symbols).
func toResponse(res slot.SpinResult) spinResponse {
	reels := make([]string, len(res.Reels))
	for i, sym := range res.Reels {
		reels[i] = sym.String()
	}
	return spinResponse{
		Reels:      reels,
		IsWin:      res.IsWin,
		WinSymbol:  res.WinSymbol.String(),
		Multiplier: res.Multiplier,
		Payout:     res.Payout,
		Balance:    res.Balance,
	}
}
