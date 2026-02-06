module github.com/joshuafuller/beacon/examples/docker-deployment

go 1.21

require github.com/joshuafuller/beacon v0.0.0

// Use local Beacon code instead of remote
replace github.com/joshuafuller/beacon => ../../..
