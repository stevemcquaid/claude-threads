.PHONY: build test lint integration clean vet

build:
	go build -o bin/claude-threads ./cmd/claude-threads

test:
	go test ./... -v

test-short:
	go test ./... -short -v

lint:
	golangci-lint run ./...

integration:
	go test ./tests/integration/... -v -timeout 10m

clean:
	rm -rf bin/

vet:
	go vet ./...
