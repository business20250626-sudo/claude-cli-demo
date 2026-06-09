package slot

// Symbol is a reel symbol identified by a small integer.
type Symbol int

// Reel symbols, in canonical order.
const (
	Cherry Symbol = iota
	Lemon
	Bell
	Star
	Seven
)

// NumSymbols is the number of distinct reel symbols.
const NumSymbols = 5

// NoSymbol is the sentinel returned when there is no winning symbol.
const NoSymbol Symbol = -1

// Symbols returns all reel symbols in canonical order.
func Symbols() []Symbol {
	return []Symbol{Cherry, Lemon, Bell, Star, Seven}
}

// String returns the display glyph for the symbol.
func (s Symbol) String() string {
	switch s {
	case Cherry:
		return "🍒"
	case Lemon:
		return "🍋"
	case Bell:
		return "🔔"
	case Star:
		return "⭐"
	case Seven:
		return "7️⃣"
	default:
		return "-"
	}
}
