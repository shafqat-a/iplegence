.PHONY: test build lookup validate

test:
	go test ./...

build:
	go run ./cmd/build

lookup:
	go run ./cmd/lookup $(IP)

validate:
	go run ./cmd/validate dist/Superior-IP.mmdb
