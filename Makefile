run:
	go run ./cmd/api .

# need gcc for this to work for sql
build:
	go build ./cmd/api
	./cmd/api