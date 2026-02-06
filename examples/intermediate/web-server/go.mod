module github.com/joshuafuller/beacon/examples/intermediate/web-server

go 1.24.0

toolchain go1.24.13

require github.com/joshuafuller/beacon v0.0.0

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)

// Use local Beacon code instead of remote
replace github.com/joshuafuller/beacon => ../../..
