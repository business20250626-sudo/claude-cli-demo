package slot

import "fmt"

// SpinResult is the flat outcome of a single spin. It is returned by value and
// serializes directly to JSON. Payout and Balance are zero for a pure Engine
// spin; Machine fills them in.
type SpinResult struct {
	Reels      [reelCount]Symbol `json:"reels"`
	IsWin      bool              `json:"isWin"`
	WinSymbol  Symbol            `json:"winSymbol"`
	Multiplier int               `json:"multiplier"`
	Payout     int               `json:"payout"`
	Balance    int               `json:"balance"`
}

// String renders the board and outcome for logs and the CLI.
func (r SpinResult) String() string {
	if r.IsWin {
		return fmt.Sprintf("%v %v %v | win x%d payout=%d balance=%d",
			r.Reels[0], r.Reels[1], r.Reels[2], r.Multiplier, r.Payout, r.Balance)
	}
	return fmt.Sprintf("%v %v %v | no win balance=%d",
		r.Reels[0], r.Reels[1], r.Reels[2], r.Balance)
}
