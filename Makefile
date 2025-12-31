build:
	go build -o ./bin/ ./cmd/msk/main.go

install: 
	go install ./cmd/msk

run:
	go run ./cmd/msk/main.go

config:
	go run ./cmd/msk/main.go c

set-p:
	go run ./cmd/msk/main.go p -n "Netflix" -s "password123"

del-p:
	go run ./cmd/msk/main.go p -n "Netflix" -d