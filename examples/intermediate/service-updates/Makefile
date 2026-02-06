.PHONY: run build clean test

# Run the example
run:
	go run main.go

# Build the binary
build:
	go build -o service-updates main.go

# Clean build artifacts
clean:
	rm -f service-updates

# Run tests (if any)
test:
	go test -v ./...
