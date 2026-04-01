.PHONY: build test test-short lint integration clean vet

# Unset GOROOT so the go binary uses its own stdlib (handles GVM conflicts)
unexport GOROOT

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
