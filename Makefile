.PHONY: build test vet fmt verify web clean

build:                    ## compile everything
	go build ./...

test:                     ## run all tests
	go test ./...

vet:                      ## static analysis
	go vet ./...

fmt:                      ## report files needing gofmt (empty == clean)
	gofmt -l .

verify:                   ## probability/RTP comparison report
	go run ./cmd/slot-verify -spins=10000000 -seed=1

web:                      ## start the web UI on http://localhost:8080
	go run ./cmd/slot-web

clean:
	go clean
	rm -rf dist
