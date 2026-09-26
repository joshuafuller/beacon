module github.com/joshuafuller/beacon/examples/advanced/iot-device

go 1.25.0

require github.com/joshuafuller/beacon v0.0.0

require (
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
)

// Use local Beacon code instead of remote
replace github.com/joshuafuller/beacon => ../../..
