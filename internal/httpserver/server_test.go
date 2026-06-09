package httpserver_test

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"example.com/example-game/internal/httpserver"
	slot "example.com/example-game/slot"
)

// spinResp mirrors the server's JSON wire shape for decoding in tests.
type spinResp struct {
	Reels      []string `json:"reels"`
	IsWin      bool     `json:"isWin"`
	WinSymbol  string   `json:"winSymbol"`
	Multiplier int      `json:"multiplier"`
	Payout     int      `json:"payout"`
	Balance    int      `json:"balance"`
	Error      string   `json:"error"`
}

func newTestServer(t *testing.T, balance int) *httptest.Server {
	t.Helper()
	srv := httpserver.New(slot.DefaultConfig(), balance, rand.New(rand.NewPCG(1, 2)))
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func postSpin(t *testing.T, ts *httptest.Server, body string) spinResp {
	t.Helper()
	resp, err := http.Post(ts.URL+"/spin", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /spin: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var out spinResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out
}

func TestSpinEndpoint(t *testing.T) {
	ts := newTestServer(t, 100)
	out := postSpin(t, ts, `{"bet":1}`)
	if len(out.Reels) != 3 {
		t.Errorf("len(reels) = %d, want 3", len(out.Reels))
	}
	// Consistency at the wire boundary (bet == 1).
	if out.IsWin != (out.Multiplier > 0) {
		t.Errorf("isWin=%v inconsistent with multiplier=%d", out.IsWin, out.Multiplier)
	}
	if out.Payout != out.Multiplier {
		t.Errorf("payout=%d != multiplier=%d for bet 1", out.Payout, out.Multiplier)
	}
	if want := 100 - 1 + out.Payout; out.Balance != want {
		t.Errorf("balance=%d, want %d", out.Balance, want)
	}
}

func TestSpinDefaultsBetWhenMissingOrNonPositive(t *testing.T) {
	for _, body := range []string{``, `{}`, `{"bet":0}`, `{"bet":-3}`} {
		ts := newTestServer(t, 100)
		out := postSpin(t, ts, body)
		// bet defaults to 1, so balance moves by at most the bet of 1 (minus 1, plus payout).
		if out.Balance != 100-1+out.Payout {
			t.Errorf("body %q: balance=%d, want %d (bet should default to 1)", body, out.Balance, 100-1+out.Payout)
		}
	}
}

func TestSpinInsufficientBalanceReturnsError(t *testing.T) {
	ts := newTestServer(t, 1)
	// First spin (bet 1) is allowed; drain isn't guaranteed, so over-bet directly.
	resp, err := http.Post(ts.URL+"/spin", "application/json", strings.NewReader(`{"bet":1000000}`))
	if err != nil {
		t.Fatalf("POST /spin: %v", err)
	}
	defer resp.Body.Close()
	var out spinResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Error == "" {
		t.Errorf("expected non-empty error for over-bet, got none")
	}
	if out.Balance != 1 {
		t.Errorf("balance changed on rejected bet: %d, want 1", out.Balance)
	}
}

func TestReset(t *testing.T) {
	ts := newTestServer(t, 100)
	postSpin(t, ts, `{"bet":1}`) // move balance away from 100

	resp, err := http.Post(ts.URL+"/reset", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /reset: %v", err)
	}
	defer resp.Body.Close()
	var r spinResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Balance != 100 {
		t.Errorf("reset balance = %d, want 100", r.Balance)
	}
	if len(r.Reels) != 3 || r.Reels[0] != "❔" {
		t.Errorf("reset reels = %v, want three placeholders", r.Reels)
	}
}

func TestIndexServed(t *testing.T) {
	ts := newTestServer(t, 100)
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestUnknownPathNotFound(t *testing.T) {
	ts := newTestServer(t, 100)
	resp, err := http.Get(ts.URL + "/nope")
	if err != nil {
		t.Fatalf("GET /nope: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// TestConcurrentSpinReset verifies the mutex contract the server documents.
// Run with `go test -race ./internal/httpserver/` to exercise the detector.
func TestConcurrentSpinReset(t *testing.T) {
	ts := newTestServer(t, 1_000_000)
	const workers, iters = 32, 50
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				var path, body string
				if (w+i)%5 == 0 {
					path, body = "/reset", `{}`
				} else {
					path, body = "/spin", `{"bet":1}`
				}
				resp, err := http.Post(ts.URL+path, "application/json", strings.NewReader(body))
				if err != nil {
					t.Errorf("POST %s: %v", path, err)
					return
				}
				resp.Body.Close()
			}
		}(w)
	}
	wg.Wait()
}
