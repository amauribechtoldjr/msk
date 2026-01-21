build:
	go build -o ./bin/ ./cmd/msk/main.go

install: 
	go install ./cmd/msk/

run:
	go run ./bin/main.exe

config:
	go run ./cmd/msk/main.go c

setp:
	go run ./cmd/msk/main.go add teste123

getp:
	go run ./cmd/msk/main.go p -n "HBO" -g

setp1:
	go run ./cmd/msk/main.go p -n "HBO2" -s "password123"

setp2:
	go run ./cmd/msk/main.go p -n "HBO3" -s "password123"

delp:
	go run ./cmd/msk/main.go p -n "Netflix" -d

list:
	go run ./cmd/msk/main.go p -l
