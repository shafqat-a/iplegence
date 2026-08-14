.PHONY: test build lookup validate serve docker docker-run

test:
	go test ./...

build:
	go run ./cmd/build

lookup:
	go run ./cmd/lookup $(IP)

validate:
	go run ./cmd/validate dist/Superior-IP.mmdb

serve:
	go run ./cmd/serve

docker:
	docker build -t iplegence:latest .

docker-run:
	docker run --rm -p 8080:8080 iplegence:latest
