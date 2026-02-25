VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/amauribechtoldjr/msk/internal/build.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o ./bin/ ./cmd/msk/main.go

run-build:
	go run ./bin/main.exe

run:
	go run ./cmd/msk/main.go

install:
	go install $(LDFLAGS) ./cmd/msk/

test:
	go test ./...

test-v:
	go test ./... -v

clean-vault:
	rm -Rf ./vault/
