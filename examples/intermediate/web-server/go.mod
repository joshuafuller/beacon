module github.com/joshuafuller/beacon/examples/intermediate/web-server

go 1.21

require github.com/joshuafuller/beacon v0.0.0

require (
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

// Use local Beacon code instead of remote
replace github.com/joshuafuller/beacon => ../../..
