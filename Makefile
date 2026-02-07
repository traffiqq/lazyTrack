.PHONY: build test lint clean

build:
	go build -o lazytrack ./cmd/lazytrack

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -f lazytrack
