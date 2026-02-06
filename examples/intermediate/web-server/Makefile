.PHONY: run build clean test

# Run the example
run:
	go run main.go

# Build the binary
build:
	go build -o web-server main.go

# Clean build artifacts
clean:
	rm -f web-server

# Run tests (if any)
test:
	go test -v ./...
