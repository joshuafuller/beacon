// Package fuzz provides fuzz testing for DNS-SD service-type validation.
//
// This file hammers responder service-type validation (the serviceTypeRegex in
// responder/service.go, RFC 6763 §7) with a *differential* oracle: an
// independent, regex-free reimplementation of the regex's intended grammar is
// compared against the real validator for millions of inputs. Any divergence
// indicates a validator bug (bad anchoring, newline injection, stray character
// class, etc.) — the failure modes that are easy to miss by eyeballing a regex.
package fuzz

import (
	"strings"
	"testing"

	"github.com/joshuafuller/beacon/responder"
)

// oracleValidServiceType is an independent reference for the grammar the
// responder's serviceTypeRegex (`^_[a-z0-9-]+\._(tcp|udp)\.local$`) is meant to
// accept, implemented WITHOUT a regular expression so it can act as a check on
// the regex itself:
//
//	_<name>._<tcp|udp>.local
//	where <name> is one or more of [a-z0-9-]
//
// Exactly three dot-separated labels, ASCII only, no trailing dot, no newlines.
func oracleValidServiceType(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	name, proto, tld := parts[0], parts[1], parts[2]

	if tld != "local" {
		return false
	}
	if proto != "_tcp" && proto != "_udp" {
		return false
	}
	// name = "_" followed by one-or-more [a-z0-9-]
	if len(name) < 2 || name[0] != '_' {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// FuzzServiceTypeValidation runs the differential oracle against the real
// validator. InstanceName and Port are fixed to valid values so that
// Service.Validate() returns nil iff the SERVICE TYPE is the only thing that
// could be wrong — isolating the regex under test.
//
// Run with: go test -fuzz=FuzzServiceTypeValidation -fuzztime=30s ./tests/fuzz/
func FuzzServiceTypeValidation(f *testing.F) {
	for _, s := range []string{
		"_http._tcp.local",
		"_my-service._tcp.local",
		"_my_service._tcp.local", // embedded underscore — rejected (issue #31)
		"_ipp._udp.local",
		"_http._tcp.local.",      // trailing dot
		"_http._tcp.local\n",     // trailing newline (anchoring probe)
		"_HTTP._tcp.local",       // uppercase
		"_http._sctp.local",      // wrong proto
		"_http._tcp.example.com", // wrong tld
		"_._tcp.local",           // empty name
		"",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, serviceType string) {
		svc := &responder.Service{
			InstanceName: "fuzz",
			ServiceType:  serviceType,
			Port:         8080,
		}
		// Must never panic on any input.
		got := svc.Validate() == nil
		want := oracleValidServiceType(serviceType)
		if got != want {
			t.Fatalf("service-type validation disagrees with oracle for %q: Validate()==nil is %v, oracle is %v", serviceType, got, want)
		}
	})
}
