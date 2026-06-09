package slot

import "errors"

// Exported sentinel errors for the betting API.
var (
	// ErrInsufficientBalance is returned when the bet exceeds the balance.
	ErrInsufficientBalance = errors.New("slot: insufficient balance")
	// ErrInvalidBet is returned when the bet is not positive.
	ErrInvalidBet = errors.New("slot: invalid bet")
)
