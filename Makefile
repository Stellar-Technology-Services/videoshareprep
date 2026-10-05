BINARY  := vidprep
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo dev)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build run install uninstall test lint fmt tidy clean release snapshot deps

build:
	go build $(LDFLAGS) -o $(BINARY) .

run:
	go run . $(ARGS)

install: build
	mv $(BINARY) /usr/local/bin/$(BINARY)

uninstall:
	rm -f /usr/local/bin/$(BINARY)

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
	rm -rf dist/

deps:
	@echo "==> Go dependencies"
	go mod download
	@echo ""
	@echo "==> Transcription backend"
	@if command -v mlx_whisper >/dev/null 2>&1; then \
		echo "  mlx_whisper: $(shell mlx_whisper --version 2>/dev/null || echo installed)"; \
	else \
		echo "  mlx_whisper: not found — run: pip install mlx-whisper"; \
	fi
	@if command -v whisper >/dev/null 2>&1; then \
		echo "  whisper:     $(shell whisper --version 2>/dev/null || echo installed)"; \
	else \
		echo "  whisper:     not found — run: brew install openai-whisper"; \
	fi

snapshot:
	goreleaser build --snapshot --clean

release:
	goreleaser release --clean
