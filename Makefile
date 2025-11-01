run:
	go run ./cmd/api

test:
	go test ./... -v

tidy:
	go mod tidy
