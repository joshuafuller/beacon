.PHONY: run build clean test

# Run the example
run:
	go run main.go

# Build the binary
build:
	go build -o iot-device main.go

# Clean build artifacts
clean:
	rm -f iot-device

# Run tests (if any)
test:
	go test -v ./...
