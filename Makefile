build:
	go build -o ./bin/ ./cmd/msk/main.go

install: 
	go install ./cmd/msk

run:
	go run ./cmd/msk/main.go

config:
	go run ./cmd/msk/main.go config -m "master-key-example"
