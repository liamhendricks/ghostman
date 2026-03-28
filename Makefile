.PHONY: build test test-race lint-boundary clean

build:
	go build -o ghostman ./cmd/ghostman

test:
	go test ./...

test-race:
	go test -race ./...

lint-boundary:
	@grep -r "spf13/cobra\|spf13/viper" pkg/ && (echo "VIOLATION: cobra/viper import found in pkg/" && exit 1) || echo "boundary: ok"

clean:
	go clean ./...
