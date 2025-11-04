//go:build unix

package transport

import (
	"runtime"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

// TestSetSocketOptions_Linux verifies SO_REUSEADDR and SO_REUSEPORT are set on Linux.
// Per F-9 REQ-F9-2: Linux kernel 3.9+ requires both options for Avahi coexistence.
func TestSetSocketOptions_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-specific test")
	}

	// Create a UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		t.Fatalf("Failed to create socket: %v", err)
	}
	defer func() { _ = syscall.Close(fd) }()

	// Call setSocketOptions
	if err := setSocketOptions(uintptr(fd)); err != nil {
		t.Fatalf("setSocketOptions() failed: %v", err)
	}

	// Verify SO_REUSEADDR is set
	reuseAddr, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR)
	if err != nil {
		t.Fatalf("Failed to get SO_REUSEADDR: %v", err)
	}
	if reuseAddr != 1 {
		t.Errorf("SO_REUSEADDR = %d, want 1", reuseAddr)
	}

	// Verify SO_REUSEPORT is set (or gracefully unavailable on old kernels)
	reusePort, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT)
	if err != nil && err != unix.ENOPROTOOPT {
		t.Fatalf("Failed to get SO_REUSEPORT: %v", err)
	}
	if err == nil && reusePort != 1 {
		t.Errorf("SO_REUSEPORT = %d, want 1", reusePort)
	}

	// Verify kernel version detection works
	version := getKernelVersion()
	if version == "" || version == "unknown" {
		t.Errorf("getKernelVersion() returned %q, expected valid version string", version)
	}
	t.Logf("Linux kernel version: %s", version)
}

// TestSetSocketOptions_macOS verifies SO_REUSEADDR and SO_REUSEPORT are set on macOS.
// Per F-9 REQ-F9-2: macOS requires both options for Bonjour coexistence.
func TestSetSocketOptions_macOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-specific test")
	}

	// Create a UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		t.Fatalf("Failed to create socket: %v", err)
	}
	defer func() { _ = syscall.Close(fd) }()

	// Call setSocketOptions
	if err := setSocketOptions(uintptr(fd)); err != nil {
		t.Fatalf("setSocketOptions() failed: %v", err)
	}

	// Verify SO_REUSEADDR is set
	reuseAddr, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR)
	if err != nil {
		t.Fatalf("Failed to get SO_REUSEADDR: %v", err)
	}
	if reuseAddr != 1 {
		t.Errorf("SO_REUSEADDR = %d, want 1", reuseAddr)
	}

	// Verify SO_REUSEPORT is set (macOS always supports it)
	reusePort, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT)
	if err != nil {
		t.Fatalf("Failed to get SO_REUSEPORT: %v", err)
	}
	if reusePort != 1 {
		t.Errorf("SO_REUSEPORT = %d, want 1", reusePort)
	}

	// Kernel version not applicable on macOS
	version := getKernelVersion()
	if version != "" {
		t.Logf("macOS kernel version: %s (informational only)", version)
	}
}

// NOTE: PlatformControl functionality tested via:
// 1. Platform-specific tests above (TestSetSocketOptions_*)
// 2. Integration test: tests/integration/avahi_coexistence_test.go
// 3. Windows test: socket_windows_test.go
// Additional unit test would be redundant.
