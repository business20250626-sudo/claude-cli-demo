// Command slot-web serves the 1x3 slot web UI on a local HTTP port.
package main

import (
	crand "crypto/rand"
	"flag"
	"log"
	mrand "math/rand/v2"
	"net/http"

	"example.com/example-game/internal/httpserver"
	slot "example.com/example-game/slot"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	balance := flag.Int("balance", 100, "starting balance in credits")
	flag.Parse()

	srv := httpserver.New(slot.DefaultConfig(), *balance, productionRNG())

	log.Printf("slot-web listening on http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, srv.Handler()))
}

// productionRNG seeds a ChaCha8 generator from the OS CSPRNG.
func productionRNG() *mrand.Rand {
	var seed [32]byte
	if _, err := crand.Read(seed[:]); err != nil {
		log.Fatalf("seed: %v", err)
	}
	return mrand.New(mrand.NewChaCha8(seed))
}
