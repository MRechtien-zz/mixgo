BINARY := bin/mixgo
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build build-pure build-portmidi test lint fmt tidy run clean

all: build

# Default build includes the PortMidi input adapter.
build: build-portmidi

# Pure Go build: GPIO input only, no PortMidi/cgo required.
build-pure:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/mixgo

# Full build including the PortMidi input adapter.
build-portmidi:
	go build -tags portmidi -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/mixgo

test:
	go test -race ./...

lint:
	@files="$$(gofmt -l cmd internal)"; if [ -n "$$files" ]; then echo "gofmt needed:"; echo "$$files"; exit 1; fi
	go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then staticcheck ./...; else echo "staticcheck not installed - skipping"; fi

fmt:
	gofmt -w cmd internal

tidy:
	go mod tidy

run:
	go run ./cmd/mixgo --config config/config-cq-gpio.yml

clean:
	rm -rf bin
