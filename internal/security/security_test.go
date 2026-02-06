package security

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// NOTE: This test file intentionally uses mu.RLock() without defer for
// testing internal state inspection. All locks are immediately followed
// by unlock on the next line. This is safe for tests.
// nosemgrep: beacon-mutex-defer-unlock

// TestRateLimiter_Allow_NormalLoad verifies rate limiter allows traffic under threshold.
// Per F-11 REQ-F11-2: Default 100 qps threshold should allow legitimate high-volume traffic.
func TestRateLimiter_Allow_NormalLoad(t *testing.T) {
	// Create RateLimiter with threshold=100
	rl := NewRateLimiter(100, 60*time.Second, 10000)

	sourceIP := "192.168.1.50"

	// Send 50 queries from same source IP (well under 100 qps threshold)
	for i := 0; i < 50; i++ {
		allowed := rl.Allow(sourceIP)
		if !allowed {
			t.Errorf("Query %d was blocked but should be allowed (under 100 qps threshold)", i+1)
		}
	}

	// Verify no cooldown triggered (entry should exist but no cooldown)
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	entry, exists := rl.sources[sourceIP]
	rl.mu.RUnlock()

	if !exists {
		t.Fatal("Expected entry to exist for source IP")
	}

	if !entry.cooldownExpiry.IsZero() {
		t.Errorf("Expected no cooldown, but cooldownExpiry is set to %v", entry.cooldownExpiry)
	}

	if entry.queryCount > 100 {
		t.Errorf("Expected queryCount <= 100, got %d", entry.queryCount)
	}
}

// TestRateLimiter_Allow_ExceedsThreshold verifies rate limiter blocks flooding sources.
// Per F-11 REQ-F11-2: >100 qps triggers cooldown.
func TestRateLimiter_Allow_ExceedsThreshold(t *testing.T) {
	// Create RateLimiter with threshold=100, cooldown=60s
	rl := NewRateLimiter(100, 60*time.Second, 10000)

	sourceIP := "192.168.1.100"

	allowedCount := 0
	blockedCount := 0

	// Send 150 queries from same source IP within 1 second (exceeds 100 qps threshold)
	for i := 0; i < 150; i++ {
		allowed := rl.Allow(sourceIP)
		if allowed {
			allowedCount++
		} else {
			blockedCount++
		}
	}

	// Verify first ~100 allowed, remaining blocked
	if allowedCount > 100 {
		t.Errorf("Expected at most 100 queries allowed, got %d", allowedCount)
	}

	if blockedCount == 0 {
		t.Error("Expected some queries to be blocked, but all were allowed")
	}

	// Verify cooldown triggered
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	entry, exists := rl.sources[sourceIP]
	rl.mu.RUnlock()

	if !exists {
		t.Fatal("Expected entry to exist for source IP")
	}

	if entry.cooldownExpiry.IsZero() {
		t.Error("Expected cooldown to be triggered, but cooldownExpiry is zero")
	}

	if entry.cooldownExpiry.Before(time.Now()) {
		t.Error("Expected cooldown to be in the future")
	}
}

// TestRateLimiter_Cooldown verifies cooldown period drops packets.
// Per F-11 REQ-F11-3: 60s default cooldown.
func TestRateLimiter_Cooldown(t *testing.T) {
	// Create RateLimiter with threshold=10, cooldown=500ms (short for testing)
	rl := NewRateLimiter(10, 500*time.Millisecond, 10000)

	sourceIP := "192.168.1.150"

	// Trigger cooldown by exceeding threshold
	for i := 0; i < 20; i++ {
		rl.Allow(sourceIP)
	}

	// Verify all queries blocked during cooldown
	for i := 0; i < 5; i++ {
		allowed := rl.Allow(sourceIP)
		if allowed {
			t.Errorf("Query %d was allowed but should be blocked during cooldown", i+1)
		}
	}

	// Wait for cooldown to expire (500ms + 100ms buffer)
	time.Sleep(600 * time.Millisecond)

	// After cooldown expires, verify queries allowed again
	allowed := rl.Allow(sourceIP)
	if !allowed {
		t.Error("Query was blocked after cooldown expired, but should be allowed")
	}

	// Verify cooldown was cleared
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	entry, exists := rl.sources[sourceIP]
	rl.mu.RUnlock()

	if !exists {
		t.Fatal("Expected entry to exist for source IP")
	}

	// After cooldown expires and new query arrives, cooldownExpiry should either be zero
	// or in the past (expired)
	if !entry.cooldownExpiry.IsZero() && entry.cooldownExpiry.After(time.Now()) {
		t.Errorf("Expected cooldown to be expired, but cooldownExpiry is %v", entry.cooldownExpiry)
	}
}

// TestRateLimiter_BoundedMap verifies LRU eviction at 10,000 entries.
// Per F-11 REQ-F11-4: Prevent memory exhaustion.
func TestRateLimiter_BoundedMap(t *testing.T) {
	// Create RateLimiter with maxEntries=100 (small for testing)
	rl := NewRateLimiter(100, 60*time.Second, 100)

	// Send queries from 150 unique source IPs
	for i := 0; i < 150; i++ {
		sourceIP := fmt.Sprintf("192.168.1.%d", i)
		rl.Allow(sourceIP)
	}

	// Verify map size never exceeds 100
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	mapSize := len(rl.sources)
	evictionCount := rl.evictionCount
	rl.mu.RUnlock()

	if mapSize > 100 {
		t.Errorf("Expected map size <= 100, got %d", mapSize)
	}

	// Verify eviction occurred (we added 150 sources but max is 100)
	if evictionCount == 0 {
		t.Error("Expected evictionCount > 0 after exceeding maxEntries, but got 0")
	}

	// Test LRU behavior: Add a new source, verify it's in the map
	newestIP := "10.0.0.1"
	rl.Allow(newestIP)

	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	_, exists := rl.sources[newestIP]
	rl.mu.RUnlock()

	if !exists {
		t.Error("Expected newest entry to exist after eviction")
	}
}

// TestRateLimiter_Cleanup verifies periodic cleanup removes stale entries.
// Per F-11 REQ-F11-5: Cleanup every 5 minutes.
func TestRateLimiter_Cleanup(t *testing.T) {
	// Create RateLimiter
	rl := NewRateLimiter(100, 60*time.Second, 10000)

	staleIP1 := "192.168.1.1"
	staleIP2 := "192.168.1.2"
	activeIP := "192.168.1.3"

	// Add stale entries (simulate old traffic)
	rl.Allow(staleIP1)
	rl.Allow(staleIP2)

	// Manually age these entries by updating their lastSeen to >1 minute ago
	rl.mu.Lock() // nosemgrep: beacon-mutex-defer-unlock
	if entry, exists := rl.sources[staleIP1]; exists {
		entry.lastSeen = time.Now().Add(-2 * time.Minute)
	}
	if entry, exists := rl.sources[staleIP2]; exists {
		entry.lastSeen = time.Now().Add(-2 * time.Minute)
	}
	rl.mu.Unlock()

	// Add active IP (recent traffic)
	rl.Allow(activeIP)

	// Get initial map size
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	initialSize := len(rl.sources)
	rl.mu.RUnlock()

	if initialSize != 3 {
		t.Fatalf("Expected 3 entries before cleanup, got %d", initialSize)
	}

	// Trigger cleanup
	rl.Cleanup()

	// After cleanup, verify stale entries removed
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	afterSize := len(rl.sources)
	_, staleExists1 := rl.sources[staleIP1]
	_, staleExists2 := rl.sources[staleIP2]
	_, activeExists := rl.sources[activeIP]
	rl.mu.RUnlock()

	// Stale entries should be removed
	if staleExists1 {
		t.Error("Expected stale entry 1 to be removed, but it still exists")
	}
	if staleExists2 {
		t.Error("Expected stale entry 2 to be removed, but it still exists")
	}

	// Active entry should be retained (seen recently)
	if !activeExists {
		t.Error("Expected active entry to be retained, but it was removed")
	}

	// Map size should decrease after cleanup (from 3 to 1)
	if afterSize != 1 {
		t.Errorf("Expected map size=1 after cleanup, got %d", afterSize)
	}
}

// NOTE: Original test skeletons (T067-T070) removed.
// Actual implementations use _Agent4 suffix (see below).

// TestIsPrivate verifies private IP range detection.
// Helper function used by SourceFilter.IsValid().
func TestIsPrivate(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"10.x private", "10.0.0.1", true},
		{"172.16-31 private", "172.16.0.1", true},
		{"192.168 private", "192.168.1.1", true},
		{"Public IP", "8.8.8.8", false},
		{"Link-local", "169.254.1.1", false}, // Link-local is NOT private range
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			got := IsPrivate(ip)
			if got != tt.want {
				t.Errorf("IsPrivate(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// ===== USER STORY 4: SOURCE FILTER TESTS (T067-T070) =====
// These tests are part of Agent 4's implementation of link-local source filtering

// TestSourceFilter_IsValid_LinkLocal_Agent4 verifies link-local IPs are accepted.
// Per RFC 6762 §2: mDNS is link-local scope (169.254.0.0/16).
// Task T067
func TestSourceFilter_IsValid_LinkLocal_Agent4(t *testing.T) {
	// Create a mock interface
	iface := net.Interface{
		Index: 1,
		Name:  "eth0",
		Flags: net.FlagUp | net.FlagMulticast,
	}

	// Create source filter
	sf, err := NewSourceFilter(iface)
	if err != nil {
		t.Fatalf("NewSourceFilter() failed: %v", err)
	}

	// Test various link-local IPs (169.254.0.0/16)
	linkLocalIPs := []string{
		"169.254.1.1",
		"169.254.255.254",
		"169.254.0.1",
		"169.254.123.45",
	}

	for _, ipStr := range linkLocalIPs {
		t.Run(ipStr, func(t *testing.T) {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", ipStr)
			}

			if !sf.IsValid(ip) {
				t.Errorf("IsValid(%s) = false, want true (link-local IP should be accepted per RFC 6762 §2)", ipStr)
			}
		})
	}
}

// TestSourceFilter_IsValid_SameSubnet_Agent4 verifies same-subnet IPs are accepted.
// Per F-11 REQ-F11-1: Accept packets from same subnet as interface.
// Task T068
func TestSourceFilter_IsValid_SameSubnet_Agent4(t *testing.T) {
	// Create interface
	iface := net.Interface{
		Index: 1,
		Name:  "eth0",
		Flags: net.FlagUp | net.FlagMulticast,
	}

	// Manually create SourceFilter with known subnet (192.168.1.0/24)
	_, ipnet, err := net.ParseCIDR("192.168.1.100/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	sf := &SourceFilter{
		iface:      iface,
		ifaceAddrs: []net.IPNet{*ipnet},
	}

	// Test IPs in same subnet (should be accepted)
	sameSubnetIPs := []string{
		"192.168.1.1",
		"192.168.1.50",
		"192.168.1.100",
		"192.168.1.254",
	}

	for _, ipStr := range sameSubnetIPs {
		t.Run("same_"+ipStr, func(t *testing.T) {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", ipStr)
			}

			if !sf.IsValid(ip) {
				t.Errorf("IsValid(%s) = false, want true (IP is in same subnet 192.168.1.0/24)", ipStr)
			}
		})
	}

	// Test IPs in different subnet (should be rejected)
	differentSubnetIPs := []string{
		"192.168.2.50",
		"10.0.1.1",
	}

	for _, ipStr := range differentSubnetIPs {
		t.Run("diff_"+ipStr, func(t *testing.T) {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", ipStr)
			}

			if sf.IsValid(ip) {
				t.Errorf("IsValid(%s) = true, want false (IP is NOT in same subnet)", ipStr)
			}
		})
	}
}

// TestSourceFilter_IsValid_RejectsRoutedIP_Agent4 verifies non-link-local IPs are rejected.
// Per F-11 REQ-F11-1: Reject packets from routed IPs (e.g., 8.8.8.8).
// Task T069
func TestSourceFilter_IsValid_RejectsRoutedIP_Agent4(t *testing.T) {
	iface := net.Interface{
		Index: 1,
		Name:  "eth0",
		Flags: net.FlagUp | net.FlagMulticast,
	}

	_, ipnet, err := net.ParseCIDR("192.168.1.100/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	sf := &SourceFilter{
		iface:      iface,
		ifaceAddrs: []net.IPNet{*ipnet},
	}

	// Test routed/public IPs that are NOT link-local and NOT same subnet
	routedIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
	}

	for _, ipStr := range routedIPs {
		t.Run(ipStr, func(t *testing.T) {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", ipStr)
			}

			if sf.IsValid(ip) {
				t.Errorf("IsValid(%s) = true, want false (routed IP should be rejected)", ipStr)
			}
		})
	}
}

// TestSourceFilter_IsValid_RejectsDifferentSubnet_Agent4 verifies different-subnet IPs are rejected.
// Per F-11 REQ-F11-1: mDNS scope is link-local, not inter-subnet.
// Task T070
func TestSourceFilter_IsValid_RejectsDifferentSubnet_Agent4(t *testing.T) {
	iface := net.Interface{
		Index: 1,
		Name:  "eth0",
		Flags: net.FlagUp | net.FlagMulticast,
	}

	_, ipnet, err := net.ParseCIDR("10.0.1.100/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	sf := &SourceFilter{
		iface:      iface,
		ifaceAddrs: []net.IPNet{*ipnet},
	}

	// Test private IPs in different subnets
	differentSubnetIPs := []string{
		"10.0.2.50",
		"10.1.1.1",
		"192.168.1.1",
	}

	for _, ipStr := range differentSubnetIPs {
		t.Run(ipStr, func(t *testing.T) {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", ipStr)
			}

			if sf.IsValid(ip) {
				t.Errorf("IsValid(%s) = true, want false (IP is in different subnet than 10.0.1.0/24)", ipStr)
			}
		})
	}

	// Verify IPs in the SAME subnet are still accepted
	sameSubnetIP := "10.0.1.50"
	ip := net.ParseIP(sameSubnetIP)
	if !sf.IsValid(ip) {
		t.Errorf("IsValid(%s) = false, want true (IP is in same subnet 10.0.1.0/24)", sameSubnetIP)
	}
}
