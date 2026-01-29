build:
	go build -o ./bin/ ./cmd/msk/main.go

install: 
	go install ./cmd/msk/

run:
	go run ./bin/main.exe

test:
	go test ./...
