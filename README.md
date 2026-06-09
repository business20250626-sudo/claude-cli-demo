# example-game — 1×3 Slot

A small, stdlib-only Go 1×3 slot machine: a reusable library, an RTP
verification tool, and a web viewer.

## Use as a library

```go
import (
    "math/rand/v2"
    slot "example.com/example-game/slot"
)

m := slot.NewMachine(slot.DefaultConfig(), 100, rand.New(rand.NewPCG(1, 2)))
res, err := m.Spin(1) // err is slot.ErrInvalidBet / slot.ErrInsufficientBalance
_ = res               // res.Reels, res.IsWin, res.Multiplier, res.Payout, res.Balance
```

The default configuration has an exact RTP of 94.9% (`slot.DefaultConfig().RTP()` → `949/1000`)
and a hit frequency of 9.55% (`191/2000`).

## Commands

| Command | What |
|---------|------|
| `make web` / `go run ./cmd/slot-web` | Web UI at http://localhost:8080 |
| `make verify` / `go run ./cmd/slot-verify` | Theoretical-vs-simulated RTP table |

## Develop

| Command | What |
|---------|------|
| `make build` | `go build ./...` |
| `make test`  | `go test ./...` |
| `make vet`   | `go vet ./...` |
| `make fmt`   | `gofmt -l .` (empty output = formatted) |

Zero external dependencies; requires Go 1.25+.
