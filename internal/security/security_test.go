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
	// Create RateLimiter with maxEntries=100 (small for testing).
	rl := NewRateLimiter(100, 60*time.Second, 100)

	// Use a deterministic clock that advances per query so each source has a
	// strictly increasing lastSeen. Without this the test is flaky on platforms
	// with coarse timer resolution (e.g. Windows ~15ms): all entries get the
	// same time.Now() value, making LRU eviction order arbitrary so the "newest"
	// entry can be evicted.
	clk := newManualClock()
	rl.now = clk.Now

	// Send queries from 150 unique source IPs.
	for i := 0; i < 150; i++ {
		sourceIP := fmt.Sprintf("192.168.1.%d", i)
		rl.Allow(sourceIP)
		clk.advance(time.Millisecond)
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

	// Test LRU behavior: Add a new source (newest lastSeen), verify it survives
	// eviction — the eviction it triggers must remove the OLDEST entries, not it.
	clk.advance(time.Millisecond)
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

		// 172.16.0.0/12 range boundaries (pins ip4[1] >= 16 && ip4[1] <= 31).
		// Without these, mutation testing showed the `<= 31` / `>= 16` edges survive.
		{"172.15 just below range", "172.15.255.255", false}, // 15 < 16 → not private
		{"172.16 lower edge", "172.16.0.0", true},            // exactly 16 → private
		{"172.31 upper edge", "172.31.255.255", true},        // exactly 31 → private
		{"172.32 just above range", "172.32.0.0", false},     // 32 > 31 → not private
		// 192.168.0.0/16 neighbours (pins ip4[1] == 168).
		{"192.167 not private", "192.167.1.1", false},
		{"192.169 not private", "192.169.1.1", false},
		// 10/8 boundary neighbours (pins ip4[0] == 10).
		{"9.x not private", "9.255.255.255", false},
		{"11.x not private", "11.0.0.0", false},
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

// TestNewSourceFilter_CachesInterfaceAddrs exercises the real NewSourceFilter
// constructor (the other SourceFilter tests build the struct literally and so
// never run it). On an interface whose Addrs() succeeds, the constructor MUST
// cache the interface's *net.IPNet addresses — not take the empty error-path
// fallback. Mutation testing flagged the `if err != nil` branch as uncovered;
// this asserts the success path populates ifaceAddrs.
func TestNewSourceFilter_CachesInterfaceAddrs(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Skipf("cannot list interfaces in this environment: %v", err)
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		wantIPNets := 0
		for _, a := range addrs {
			if _, ok := a.(*net.IPNet); ok {
				wantIPNets++
			}
		}
		if wantIPNets == 0 {
			continue // need an interface with at least one IPNet to be decisive
		}

		sf, err := NewSourceFilter(iface)
		if err != nil {
			t.Fatalf("NewSourceFilter(%s) returned error: %v", iface.Name, err)
		}
		if len(sf.ifaceAddrs) != wantIPNets {
			t.Fatalf("NewSourceFilter(%s) cached %d addrs, want %d (success path must cache, not take the empty fallback)",
				iface.Name, len(sf.ifaceAddrs), wantIPNets)
		}
		return // one decisive interface is enough
	}
	t.Skip("no interface with an IPNet address available to exercise the constructor")
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

// manualClock is a deterministic clock for exercising the rate limiter's
// time-based boundaries (window expiry, LRU ordering, cleanup) exactly — wall
// clock cannot hit an edge like "elapsed == 1s" reliably.
type manualClock struct{ t time.Time }

func (c *manualClock) Now() time.Time          { return c.t }
func (c *manualClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newManualClock() *manualClock {
	return &manualClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
}

// queryCountOf returns the current sliding-window count for a source (test-only
// white-box inspection).
func queryCountOf(rl *RateLimiter, ip string) (int, bool) {
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	defer rl.mu.RUnlock()
	e, ok := rl.sources[ip]
	if !ok {
		return 0, false
	}
	return e.queryCount, true
}

func hasSource(rl *RateLimiter, ip string) bool {
	rl.mu.RLock() // nosemgrep: beacon-mutex-defer-unlock
	defer rl.mu.RUnlock()
	_, ok := rl.sources[ip]
	return ok
}

// TestRateLimiter_WindowExpiry_Boundary pins the sliding-window reset edge
// (rate_limiter.go: `now.Sub(entry.windowStart) > 1*time.Second`). At exactly
// 1 second the window must NOT reset; strictly past 1 second it must. A fake
// clock lets us land on the edge precisely, killing the boundary mutant that
// wall-clock tests cannot.
func TestRateLimiter_WindowExpiry_Boundary(t *testing.T) {
	clk := newManualClock()
	rl := NewRateLimiter(100, time.Hour, 10000) // high threshold so we never cool down
	rl.now = clk.Now
	const src = "203.0.113.7"

	rl.Allow(src)            // create: count=1, windowStart=T0
	clk.advance(time.Second) // now == windowStart + exactly 1s
	rl.Allow(src)            // 1s is NOT > 1s → must increment, not reset
	if c, _ := queryCountOf(rl, src); c != 2 {
		t.Fatalf("queryCount = %d at exactly windowStart+1s; window must NOT reset (expected 2)", c)
	}

	clk.advance(time.Nanosecond) // now strictly past 1s relative to windowStart
	rl.Allow(src)                // now > 1s → must reset window to count=1
	if c, _ := queryCountOf(rl, src); c != 1 {
		t.Fatalf("queryCount = %d strictly past windowStart+1s; window must reset (expected 1)", c)
	}
}

// TestRateLimiter_Cleanup_Boundary pins the staleness edge in Cleanup()
// (`now.Sub(entry.lastSeen) > 1*time.Minute`). At exactly one minute the entry
// must be kept; strictly past one minute it must be removed.
func TestRateLimiter_Cleanup_Boundary(t *testing.T) {
	clk := newManualClock()
	rl := NewRateLimiter(100, time.Hour, 10000)
	rl.now = clk.Now
	const src = "203.0.113.8"

	rl.Allow(src) // lastSeen = T0
	clk.advance(time.Minute)
	rl.Cleanup()
	if !hasSource(rl, src) {
		t.Fatalf("entry removed at exactly lastSeen+1m; it must be kept (boundary is strictly > 1m)")
	}

	clk.advance(time.Nanosecond)
	rl.Cleanup()
	if hasSource(rl, src) {
		t.Fatalf("entry kept strictly past lastSeen+1m; it must be removed")
	}
}

// TestRateLimiter_Evict_RemovesOldest pins the LRU selection in evict(): it must
// remove the entries with the OLDEST lastSeen, not the newest. A distinct,
// monotonically advancing timestamp per source makes the ordering unambiguous,
// killing the `lastSeen.Before(...)` comparison mutant and the partial-sort /
// delete-loop boundary mutants.
func TestRateLimiter_Evict_RemovesOldest(t *testing.T) {
	clk := newManualClock()
	const maxEntries = 20 // evictCount = maxEntries/10 = 2
	rl := NewRateLimiter(1000, time.Hour, maxEntries)
	rl.now = clk.Now

	// Fill exactly maxEntries sources, each 1s newer than the last.
	for i := 0; i < maxEntries; i++ {
		rl.Allow(fmt.Sprintf("198.51.100.%d", i))
		clk.advance(time.Second)
	}
	// One more source trips eviction of the 2 oldest (ip 0 and ip 1).
	rl.Allow("198.51.100.200")

	if rl.evictionCount != 2 {
		t.Fatalf("evictionCount = %d; expected 2 (maxEntries/10)", rl.evictionCount)
	}
	if hasSource(rl, "198.51.100.0") {
		t.Errorf("oldest source .0 survived eviction; evict() must remove oldest-by-lastSeen")
	}
	if hasSource(rl, "198.51.100.1") {
		t.Errorf("second-oldest source .1 survived eviction; evict() must remove oldest-by-lastSeen")
	}
	if !hasSource(rl, "198.51.100.2") {
		t.Errorf("source .2 was evicted; only the 2 OLDEST should be removed")
	}
	if !hasSource(rl, "198.51.100.200") {
		t.Errorf("newest source was evicted; evict() must remove oldest, not newest")
	}
}

// TestRateLimiter_Allow_ThresholdBoundary pins the EXACT threshold edge in
// Allow() (rate_limiter.go:110, `entry.queryCount > rl.threshold`).
//
// Mutation testing showed the existing exceeds-threshold test asserts only
// "allowedCount <= 100", which lets a `>`→`>=` (or `>`→`>=` boundary) mutant
// survive. Here threshold=5 and the queries are issued in a tight loop (well
// within the 1s window, so the window never resets): exactly the first
// `threshold` queries must be allowed and the very next one denied.
func TestRateLimiter_Allow_ThresholdBoundary(t *testing.T) {
	const threshold = 5
	rl := NewRateLimiter(threshold, 60*time.Second, 10000)
	const src = "192.168.1.200"

	for i := 1; i <= threshold; i++ {
		if !rl.Allow(src) {
			t.Fatalf("query %d/%d denied; the first %d queries must be allowed (queryCount %d is not > threshold %d)", i, threshold, threshold, i, threshold)
		}
	}
	// The (threshold+1)th query is the first to exceed the limit.
	if rl.Allow(src) {
		t.Fatalf("query %d was allowed; it exceeds threshold %d (queryCount %d > %d) and must be denied", threshold+1, threshold, threshold+1, threshold)
	}
}

// TestRateLimiter_Eviction_Boundary pins the maxEntries eviction edge
// (rate_limiter.go:67, `len(rl.sources) > rl.maxEntries`) and the eviction
// arithmetic (`evictCount := rl.maxEntries / 10`, rate_limiter.go:124).
//
// Filling exactly maxEntries distinct sources must NOT evict; one more must
// evict exactly maxEntries/10 entries. The exact post-eviction size kills both
// the boundary mutant (>= would evict early) and the arithmetic mutants
// (*,+,- all yield a different evict count and thus a different final size).
func TestRateLimiter_Eviction_Boundary(t *testing.T) {
	const maxEntries = 100
	rl := NewRateLimiter(1000, 60*time.Second, maxEntries) // high threshold: never rate-limit

	for i := 0; i < maxEntries; i++ {
		rl.Allow(fmt.Sprintf("10.0.%d.%d", i/256, i%256))
	}
	if got := rl.evictionCount; got != 0 {
		t.Fatalf("evictionCount = %d after exactly maxEntries=%d sources; expected 0 (len == maxEntries is not > maxEntries)", got, maxEntries)
	}
	if got := len(rl.sources); got != maxEntries {
		t.Fatalf("sources size = %d after filling to maxEntries; expected %d", got, maxEntries)
	}

	// One more distinct source pushes len to maxEntries+1 (> maxEntries) -> evict 10%.
	rl.Allow("10.99.99.99")
	const wantEvicted = maxEntries / 10 // 10
	if got := rl.evictionCount; got != wantEvicted {
		t.Fatalf("evictionCount = %d after exceeding maxEntries; expected exactly maxEntries/10 = %d", got, wantEvicted)
	}
	if got, want := len(rl.sources), maxEntries+1-wantEvicted; got != want {
		t.Fatalf("sources size = %d after eviction; expected %d (%d added - %d evicted)", got, want, maxEntries+1, wantEvicted)
	}
}

// TestRateLimiter_Eviction_MinimumOne pins the "evict at least one" guard
// (rate_limiter.go:125, `if evictCount == 0 { evictCount = 1 }`) for small
// maxEntries where maxEntries/10 == 0.
func TestRateLimiter_Eviction_MinimumOne(t *testing.T) {
	const maxEntries = 5 // maxEntries/10 == 0 -> must clamp to 1
	rl := NewRateLimiter(1000, 60*time.Second, maxEntries)

	for i := 0; i <= maxEntries; i++ { // maxEntries+1 distinct sources
		rl.Allow(fmt.Sprintf("172.16.0.%d", i))
	}
	if got := rl.evictionCount; got != 1 {
		t.Fatalf("evictionCount = %d for maxEntries=%d; expected exactly 1 (maxEntries/10 rounds to 0, clamped to 1)", got, maxEntries)
	}
	if got := len(rl.sources); got != maxEntries {
		t.Fatalf("sources size = %d after eviction; expected %d", got, maxEntries)
	}
}
