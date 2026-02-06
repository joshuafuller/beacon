.PHONY: run build clean test

# Run the example
run:
	go run *.go

# Build the binary
build:
	go build -o multi-interface-bridge *.go

# Clean build artifacts
clean:
	rm -f multi-interface-bridge

# Run tests (if any)
test:
	go test -v ./...
