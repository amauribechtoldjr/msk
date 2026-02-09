build:
	go build -o ./bin/ ./cmd/msk/main.go

run-build:
	go run ./bin/main.exe

run:
	go run ./cmd/msk/main.go

install: 
	go install ./cmd/msk/

test:
	go test ./...

test-v:
	go test ./... -v
