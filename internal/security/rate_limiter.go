// Package security provides security features including rate limiting
// and source IP validation for mDNS multicast traffic.
package security

import (
	"sync"
	"time"
)

// RateLimitEntry tracks query rate for a single source IP.
// Per F-11 (Security Architecture), this enables per-source-IP rate limiting
// to protect against multicast storms (e.g., Hubitat bug sending 1000+ qps).
type RateLimitEntry struct {
	windowStart    time.Time // Start of current 1-second sliding window
	cooldownExpiry time.Time // When cooldown period ends (zero if not in cooldown)
	lastSeen       time.Time // Last query received (for LRU eviction)
	sourceIP       string    // Source IP address (key in RateLimiter map)
	queryCount     int       // Number of queries in current sliding window
}

// RateLimiter manages per-source-IP rate limiting with bounded map.
// Default configuration: 100 qps threshold, 60s cooldown, 10,000 max entries.
type RateLimiter struct {
	threshold     int                        // Max queries/second per source IP
	cooldown      time.Duration              // Duration to drop packets after threshold exceeded
	maxEntries    int                        // Max number of source IPs tracked
	sources       map[string]*RateLimitEntry // Source IP → RateLimitEntry
	mu            sync.RWMutex               // Protects sources map
	evictionCount uint64                     // Number of LRU evictions (for metrics)

	// now returns the current time. It defaults to time.Now and exists only so
	// that tests can supply a deterministic clock to exercise the time-based
	// boundaries (window expiry, cooldown, LRU ordering, cleanup) exactly.
	now func() time.Time
}

// NewRateLimiter creates a new rate limiter with the given configuration.
// Per FR-026, FR-027, FR-028: Configurable threshold, cooldown, and max entries.
func NewRateLimiter(threshold int, cooldown time.Duration, maxEntries int) *RateLimiter {
	return &RateLimiter{
		threshold:  threshold,
		cooldown:   cooldown,
		maxEntries: maxEntries,
		sources:    make(map[string]*RateLimitEntry),
		now:        time.Now,
	}
}

// Allow checks if a query from the given source IP should be allowed.
// Returns false if the source is in cooldown or exceeds the rate limit threshold.
// Per FR-026, FR-027, FR-028: Implements sliding window rate limiting.
func (rl *RateLimiter) Allow(sourceIP string) bool {
	// Manual unlock required: Must release read lock before acquiring write lock later in function.
	// Lock upgrade pattern - defer would cause deadlock.
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	entry, exists := rl.sources[sourceIP]
	rl.mu.RUnlock()

	if !exists {
		// First query from this source - create entry
		rl.mu.Lock()
		defer rl.mu.Unlock()
		// Check again after acquiring write lock (double-check pattern)
		entry, exists = rl.sources[sourceIP]
		if !exists {
			tNow := rl.now()
			rl.sources[sourceIP] = &RateLimitEntry{
				sourceIP:    sourceIP,
				queryCount:  1,
				windowStart: tNow,
				lastSeen:    tNow,
			}
			// Check if map exceeded maxEntries
			if len(rl.sources) > rl.maxEntries {
				rl.evict()
			}
			return true
		}
		// Entry was created by another goroutine, fall through to check it
	}

	// Update sliding window (needs write lock)
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()

	// Check cooldown (after acquiring lock)
	if !entry.cooldownExpiry.IsZero() && now.Before(entry.cooldownExpiry) {
		return false // In cooldown, drop packet
	}

	// Cooldown has expired or not set, check/reset window
	if !entry.cooldownExpiry.IsZero() && now.After(entry.cooldownExpiry) {
		// Cooldown just expired, reset window
		entry.queryCount = 1
		entry.windowStart = now
		entry.cooldownExpiry = time.Time{} // Clear cooldown
		entry.lastSeen = now
		return true
	}

	// Check if window has expired (>1 second)
	if now.Sub(entry.windowStart) > 1*time.Second {
		// Reset window
		entry.queryCount = 1
		entry.windowStart = now
		entry.cooldownExpiry = time.Time{} // Clear any expired cooldown
	} else {
		// Increment count in current window
		entry.queryCount++
	}

	entry.lastSeen = now

	// Check threshold
	if entry.queryCount > rl.threshold {
		// Exceeded threshold, start cooldown
		entry.cooldownExpiry = now.Add(rl.cooldown)
		return false
	}

	return true
}

// evict performs LRU cleanup when the sources map exceeds maxEntries.
// Removes oldest 10% of entries by lastSeen timestamp.
// MUST be called while holding rl.mu write lock.
func (rl *RateLimiter) evict() {
	// Calculate how many entries to evict (10% of maxEntries)
	evictCount := rl.maxEntries / 10
	if evictCount == 0 {
		evictCount = 1 // Evict at least one entry
	}

	// Collect all entries with their lastSeen timestamp
	type entryWithTime struct {
		ip       string
		lastSeen time.Time
	}

	entries := make([]entryWithTime, 0, len(rl.sources))
	for ip, entry := range rl.sources {
		entries = append(entries, entryWithTime{ip: ip, lastSeen: entry.lastSeen})
	}

	// Sort by lastSeen (oldest first)
	// Using simple bubble-style partial sort for oldest evictCount entries
	for i := 0; i < evictCount && i < len(entries); i++ {
		// Find oldest in remaining entries
		oldestIdx := i
		for j := i + 1; j < len(entries); j++ {
			if entries[j].lastSeen.Before(entries[oldestIdx].lastSeen) {
				oldestIdx = j
			}
		}
		// Swap to position i
		entries[i], entries[oldestIdx] = entries[oldestIdx], entries[i]
	}

	// Evict oldest entries
	evicted := 0
	for i := 0; i < evictCount && i < len(entries); i++ {
		delete(rl.sources, entries[i].ip)
		evicted++
	}

	// evicted is always >= 0 (it only increments from 0), so no guard is needed.
	// G115: conversion is safe — evicted is non-negative and at most maxEntries/10.
	rl.evictionCount += uint64(evicted) //nolint:gosec // G115: bounds checked
}

// Cleanup removes stale entries from the rate limiter map.
// Per FR-031: Called periodically (every 5 minutes) to prevent memory growth.
// Removes entries that haven't been seen in the last minute.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	toDelete := make([]string, 0)

	// Find stale entries (not seen recently)
	for ip, entry := range rl.sources {
		// Remove if not seen in last 1 minute (conservative cleanup)
		// This handles both entries with expired cooldowns and inactive sources
		if now.Sub(entry.lastSeen) > 1*time.Minute {
			toDelete = append(toDelete, ip)
		}
	}

	// Delete stale entries
	for _, ip := range toDelete {
		delete(rl.sources, ip)
	}
}
