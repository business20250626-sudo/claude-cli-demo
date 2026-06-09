package slot

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand/v2"
	"testing"
)

// --- symbol ---

func TestSymbolString(t *testing.T) {
	tests := []struct {
		name string
		in   Symbol
		want string
	}{
		{"cherry", Cherry, "🍒"},
		{"lemon", Lemon, "🍋"},
		{"bell", Bell, "🔔"},
		{"star", Star, "⭐"},
		{"seven", Seven, "7️⃣"},
		{"none", NoSymbol, "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("Symbol(%d).String() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSymbolsLenMatchesNumSymbols(t *testing.T) {
	if got := len(Symbols()); got != NumSymbols {
		t.Errorf("len(Symbols()) = %d, want NumSymbols = %d", got, NumSymbols)
	}
}

// --- reels ---

func TestNewReelTotal(t *testing.T) {
	r := newReel(map[Symbol]int{Cherry: 8, Lemon: 6, Bell: 3, Star: 2, Seven: 1})
	if r.total != 20 {
		t.Errorf("total = %d, want 20", r.total)
	}
}

// symbolForRoll must map every roll in [0,total) to a symbol, with each
// symbol covering exactly its weight, in canonical order.
func TestSymbolForRollDistribution(t *testing.T) {
	r := newReel(map[Symbol]int{Cherry: 8, Lemon: 6, Bell: 3, Star: 2, Seven: 1})
	counts := map[Symbol]int{}
	for x := 0; x < r.total; x++ {
		counts[r.symbolForRoll(x)]++
	}
	want := map[Symbol]int{Cherry: 8, Lemon: 6, Bell: 3, Star: 2, Seven: 1}
	for s, w := range want {
		if counts[s] != w {
			t.Errorf("symbol %v covered %d rolls, want %d", s, counts[s], w)
		}
	}
	// Boundaries.
	if got := r.symbolForRoll(0); got != Cherry {
		t.Errorf("roll 0 = %v, want Cherry", got)
	}
	if got := r.symbolForRoll(r.total - 1); got != Seven {
		t.Errorf("roll %d = %v, want Seven", r.total-1, got)
	}
}

func TestReelSpinDeterministic(t *testing.T) {
	r := newReel(map[Symbol]int{Cherry: 8, Lemon: 6, Bell: 3, Star: 2, Seven: 1})
	a := rand.New(rand.NewPCG(1, 2))
	b := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < 100; i++ {
		if r.spin(a) != r.spin(b) {
			t.Fatalf("same-seed reels diverged at i=%d", i)
		}
	}
}

// --- config ---

func TestDefaultConfigAccessors(t *testing.T) {
	c := DefaultConfig()
	wantW := map[Symbol]int{Cherry: 8, Lemon: 6, Bell: 3, Star: 2, Seven: 1}
	wantM := map[Symbol]int{Cherry: 6, Lemon: 12, Bell: 24, Star: 80, Seven: 640}
	for s, w := range wantW {
		if got := c.Weight(s); got != w {
			t.Errorf("Weight(%v) = %d, want %d", s, got, w)
		}
	}
	for s, m := range wantM {
		if got := c.Multiplier(s); got != m {
			t.Errorf("Multiplier(%v) = %d, want %d", s, got, m)
		}
	}
}

func TestNewConfigIdenticalReels(t *testing.T) {
	c := NewConfig(
		map[Symbol]int{Cherry: 1, Seven: 1},
		map[Symbol]int{Cherry: 2, Seven: 50},
	)
	for i := 0; i < reelCount; i++ {
		if c.reels[i].total != 2 {
			t.Errorf("reel %d total = %d, want 2", i, c.reels[i].total)
		}
	}
	if c.Multiplier(Seven) != 50 {
		t.Errorf("Multiplier(Seven) = %d, want 50", c.Multiplier(Seven))
	}
}

// --- paytable / analytic RTP ---

func TestEvaluate(t *testing.T) {
	c := DefaultConfig()
	tests := []struct {
		name     string
		board    [reelCount]Symbol
		wantMult int
		wantWin  Symbol
	}{
		{"three sevens", [reelCount]Symbol{Seven, Seven, Seven}, 640, Seven},
		{"three cherries", [reelCount]Symbol{Cherry, Cherry, Cherry}, 6, Cherry},
		{"three bells", [reelCount]Symbol{Bell, Bell, Bell}, 24, Bell},
		{"two of a kind", [reelCount]Symbol{Cherry, Cherry, Lemon}, 0, NoSymbol},
		{"all different", [reelCount]Symbol{Cherry, Lemon, Bell}, 0, NoSymbol},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMult, gotWin := c.evaluate(tt.board)
			if gotMult != tt.wantMult || gotWin != tt.wantWin {
				t.Errorf("evaluate(%v) = (%d,%v), want (%d,%v)",
					tt.board, gotMult, gotWin, tt.wantMult, tt.wantWin)
			}
		})
	}
}

func TestAnalyticRTPExact(t *testing.T) {
	c := DefaultConfig()
	if c.RTP().Cmp(big.NewRat(949, 1000)) != 0 {
		t.Errorf("RTP() = %s, want 949/1000", c.RTP().RatString())
	}
	if c.HitFrequency().Cmp(big.NewRat(191, 2000)) != 0 {
		t.Errorf("HitFrequency() = %s, want 191/2000", c.HitFrequency().RatString())
	}
}

// --- engine ---

func TestEngineSpinDeterministic(t *testing.T) {
	cfg := DefaultConfig()
	a := NewEngine(cfg, rand.New(rand.NewPCG(42, 7)))
	b := NewEngine(cfg, rand.New(rand.NewPCG(42, 7)))
	for i := 0; i < 200; i++ {
		if a.Spin() != b.Spin() {
			t.Fatalf("same-seed engines diverged at i=%d", i)
		}
	}
}

func TestEngineSpinValidBoardAndWinFlag(t *testing.T) {
	eng := NewEngine(DefaultConfig(), rand.New(rand.NewPCG(1, 2)))
	for i := 0; i < 1000; i++ {
		res := eng.Spin()
		for _, s := range res.Reels {
			if s < Cherry || s > Seven {
				t.Fatalf("invalid symbol %v on board %v", s, res.Reels)
			}
		}
		if res.IsWin != (res.Multiplier > 0) {
			t.Fatalf("IsWin=%v inconsistent with Multiplier=%d", res.IsWin, res.Multiplier)
		}
		if res.IsWin && res.WinSymbol == NoSymbol {
			t.Fatalf("win with NoSymbol on board %v", res.Reels)
		}
	}
}

func TestEngineForcedWin(t *testing.T) {
	// A single-symbol config makes every spin a guaranteed 3-of-a-kind.
	cfg := NewConfig(map[Symbol]int{Seven: 1}, map[Symbol]int{Seven: 10})
	eng := NewEngine(cfg, rand.New(rand.NewPCG(1, 2)))
	res := eng.Spin()
	if res.Reels != [reelCount]Symbol{Seven, Seven, Seven} {
		t.Errorf("Reels = %v, want three Seven", res.Reels)
	}
	if !res.IsWin || res.Multiplier != 10 || res.WinSymbol != Seven {
		t.Errorf("got win=%v mult=%d sym=%v, want true/10/Seven",
			res.IsWin, res.Multiplier, res.WinSymbol)
	}
}

// --- machine ---

func TestMachineBetValidation(t *testing.T) {
	m := NewMachine(DefaultConfig(), 10, rand.New(rand.NewPCG(1, 2)))
	if _, err := m.Spin(0); !errors.Is(err, ErrInvalidBet) {
		t.Errorf("Spin(0) err = %v, want ErrInvalidBet", err)
	}
	if _, err := m.Spin(-5); !errors.Is(err, ErrInvalidBet) {
		t.Errorf("Spin(-5) err = %v, want ErrInvalidBet", err)
	}
	if _, err := m.Spin(11); !errors.Is(err, ErrInsufficientBalance) {
		t.Errorf("Spin(11) err = %v, want ErrInsufficientBalance", err)
	}
	if m.Balance() != 10 {
		t.Errorf("balance changed on rejected bets: %d, want 10", m.Balance())
	}
}

func TestMachineBalanceArithmeticInvariant(t *testing.T) {
	m := NewMachine(DefaultConfig(), 1000, rand.New(rand.NewPCG(99, 1)))
	for i := 0; i < 500 && m.Balance() > 0; i++ {
		prev := m.Balance()
		res, err := m.Spin(1)
		if err != nil {
			t.Fatalf("spin %d: %v", i, err)
		}
		if res.Payout != res.Multiplier { // bet == 1
			t.Fatalf("payout %d != multiplier %d at i=%d", res.Payout, res.Multiplier, i)
		}
		if want := prev - 1 + res.Payout; m.Balance() != want {
			t.Fatalf("balance = %d, want %d at i=%d", m.Balance(), want, i)
		}
		if res.Balance != m.Balance() {
			t.Fatalf("res.Balance %d != m.Balance() %d", res.Balance, m.Balance())
		}
		// Loss path pinned independently of res.Payout: a non-winning spin
		// must pay nothing and cost exactly the bet.
		if !res.IsWin {
			if res.Payout != 0 || res.Multiplier != 0 {
				t.Fatalf("loss with payout=%d mult=%d at i=%d, want 0/0", res.Payout, res.Multiplier, i)
			}
			if m.Balance() != prev-1 {
				t.Fatalf("loss balance = %d, want %d at i=%d", m.Balance(), prev-1, i)
			}
		}
	}
}

func TestMachineForcedWinBalance(t *testing.T) {
	cfg := NewConfig(map[Symbol]int{Cherry: 1}, map[Symbol]int{Cherry: 5})
	m := NewMachine(cfg, 100, rand.New(rand.NewPCG(1, 2)))
	res, err := m.Spin(2)
	if err != nil {
		t.Fatalf("Spin: %v", err)
	}
	if res.Payout != 10 || m.Balance() != 108 { // 100 - 2 + 10
		t.Errorf("payout=%d balance=%d, want 10 and 108", res.Payout, m.Balance())
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		res     SpinResult
		wantErr bool
	}{
		{"valid win", SpinResult{IsWin: true, Multiplier: 5, WinSymbol: Cherry}, false},
		{"valid loss", SpinResult{IsWin: false, Multiplier: 0, WinSymbol: NoSymbol}, false},
		{"negative multiplier", SpinResult{Multiplier: -1}, true},
		{"win flag without multiplier", SpinResult{IsWin: true, Multiplier: 0}, true},
		{"multiplier without win flag", SpinResult{IsWin: false, Multiplier: 5}, true},
		{"win without winning symbol", SpinResult{IsWin: true, Multiplier: 5, WinSymbol: NoSymbol}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.res)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%+v) error = %v, wantErr %v", tt.res, err, tt.wantErr)
			}
		})
	}
}

// --- runnable example + benchmark ---

func ExampleMachine_Spin() {
	rng := rand.New(rand.NewPCG(1, 2))
	// A single-symbol config makes the output deterministic for the doc.
	cfg := NewConfig(
		map[Symbol]int{Seven: 1},
		map[Symbol]int{Seven: 10},
	)
	m := NewMachine(cfg, 100, rng)
	res, _ := m.Spin(5)
	fmt.Printf("reels=%v win=%v payout=%d balance=%d\n",
		res.Reels, res.IsWin, res.Payout, res.Balance)
	// Output: reels=[7️⃣ 7️⃣ 7️⃣] win=true payout=50 balance=145
}

func BenchmarkSpin(b *testing.B) {
	eng := NewEngine(DefaultConfig(), rand.New(rand.NewPCG(1, 2)))
	b.ReportAllocs()
	for b.Loop() {
		_ = eng.Spin()
	}
}
