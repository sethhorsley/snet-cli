.PHONY: build install clean test

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/seth4242/snet/cmd.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/snet .

install:
	go install $(LDFLAGS) .

clean:
	rm -rf bin/

test:
	go test -v ./...

# Build for all platforms
release:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/snet-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/snet-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/snet-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/snet-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/snet-windows-amd64.exe .
