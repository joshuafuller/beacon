module github.com/joshuafuller/beacon/examples/basic/multi-service

go 1.24.0

toolchain go1.24.13

require github.com/joshuafuller/beacon v0.0.0

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)

replace github.com/joshuafuller/beacon => ../../..
