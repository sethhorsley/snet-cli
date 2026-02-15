.PHONY: build build-prod install install-prod clean test release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_PKG := github.com/seth4242/snet/internal/buildinfo

# Development build (default) - uses localhost:3001
LDFLAGS_DEV := -ldflags "-X $(BUILD_PKG).Version=$(VERSION) -X $(BUILD_PKG).Mode=development"

# Production build - uses seth4242.net
LDFLAGS_PROD := -ldflags "-X $(BUILD_PKG).Version=$(VERSION) -X $(BUILD_PKG).Mode=production"

# Default build is development mode
build:
	@echo "Building snet (development mode)..."
	go build $(LDFLAGS_DEV) -o bin/snet .

# Production build
build-prod:
	@echo "Building snet (production mode)..."
	go build $(LDFLAGS_PROD) -o bin/snet .

# Install locally (development mode)
install:
	@echo "Installing snet (development mode)..."
	go install $(LDFLAGS_DEV) .

# Install for production use
install-prod:
	@echo "Installing snet (production mode)..."
	go install $(LDFLAGS_PROD) .

clean:
	rm -rf bin/

test:
	go test -v ./...

# Build for all platforms (production mode)
release:
	@echo "Building release binaries (production mode)..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS_PROD) -o bin/snet-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS_PROD) -o bin/snet-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS_PROD) -o bin/snet-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS_PROD) -o bin/snet-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS_PROD) -o bin/snet-windows-amd64.exe .
