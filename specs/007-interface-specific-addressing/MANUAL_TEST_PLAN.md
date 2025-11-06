# Manual Test Plan: Interface-Specific Addressing (RFC 6762 §15)

**Feature**: 007-interface-specific-addressing
**Purpose**: Manual validation on real multi-interface hardware
**Status**: ⏳ **PENDING** (requires physical hardware setup)

---

## Prerequisites

### Hardware Requirements

**Multi-Interface Setup** (Required for T090-T094):
- Laptop with both WiFi and Ethernet interfaces
- Two separate networks (different IP subnets):
  - **Network A** (e.g., WiFi): 10.0.0.0/24
  - **Network B** (e.g., Ethernet): 192.168.1.0/24
- Both interfaces must be active simultaneously
- Networks must be isolated (no routing between them)

**Single-Interface Setup** (Required for T095):
- Machine with only one network interface
- For regression testing

### Platform Requirements (T097-T099)

- **Linux** (Ubuntu/Debian/RHEL) - IP_PKTINFO support
- **macOS** (10.14+) - IP_RECVIF support
- **Windows** (10/11) - Graceful degradation test

### Software Requirements

- Go 1.21+ installed
- Beacon source code (branch: 007-interface-specific-addressing)
- Network packet capture tool:
  - `tcpdump` (Linux/macOS)
  - `Wireshark` (all platforms)
- mDNS query tool:
  - `avahi-browse` (Linux)
  - `dns-sd` (macOS)
  - `avahi-browse` via WSL (Windows)

---

## Test Scenarios

### T090-T094: Multi-Interface Validation

#### T090: Setup Test Environment

**Goal**: Configure laptop with WiFi + Ethernet on different networks

**Steps**:
1. Connect laptop to WiFi network (Network A)
   - Verify IP assignment: `ip addr show` or `ifconfig`
   - Example: WiFi interface `wlan0` gets `10.0.0.50`
2. Connect laptop to Ethernet network (Network B)
   - Verify IP assignment
   - Example: Ethernet interface `eth0` gets `192.168.1.100`
3. Verify both interfaces are UP and have IPs:
   ```bash
   ip addr show wlan0
   ip addr show eth0
   ```
4. Verify networks are isolated (no routing):
   ```bash
   # From Network A, cannot reach Network B gateway
   ping -I wlan0 192.168.1.1  # Should fail or not route
   ```

**Success Criteria**:
- [ ] WiFi interface has IP from Network A subnet
- [ ] Ethernet interface has IP from Network B subnet
- [ ] Both interfaces UP and active
- [ ] Networks isolated (no cross-network routing)

---

#### T091: Register mDNS Service

**Goal**: Start Beacon responder with test service on multi-interface laptop

**Steps**:
1. Build Beacon responder example:
   ```bash
   cd examples/interface-specific
   go build -o interface-test main.go
   ```
2. Run responder:
   ```bash
   sudo ./interface-test
   ```
   (Note: `sudo` required for binding to port 5353)
3. Observe output showing interface detection:
   ```
   === Interface-Specific IP Resolution (RFC 6762 §15) ===
   Available network interfaces:
     [2] wlan0      → [10.0.0.50]
     [3] eth0       → [192.168.1.100]

   ✅ RFC 6762 §15 Compliance: Interface-specific addressing working!
   ```

**Success Criteria**:
- [ ] Responder starts successfully
- [ ] Both WiFi and Ethernet interfaces detected
- [ ] Correct IPs shown for each interface
- [ ] Service registered on both interfaces

---

#### T092: Verify WiFi IP Advertised on WiFi Network

**Goal**: Query from Network A (WiFi) and verify response contains WiFi IP only

**Steps**:
1. **From another device on Network A** (WiFi network):
   ```bash
   # Linux/macOS
   avahi-browse -r _http._tcp --resolve

   # macOS alternative
   dns-sd -B _http._tcp
   ```
2. Capture mDNS traffic on laptop WiFi interface:
   ```bash
   sudo tcpdump -i wlan0 -vv port 5353
   ```
3. Look for mDNS response containing A record
4. Verify A record contains **WiFi IP only** (10.0.0.50)

**Success Criteria**:
- [ ] mDNS response received on WiFi interface
- [ ] A record contains WiFi IP: `10.0.0.50`
- [ ] A record does **NOT** contain Ethernet IP: `192.168.1.100` ✅ **RFC 6762 §15**
- [ ] Response visible in tcpdump output

**Expected tcpdump output** (partial):
```
MultiNIC-Test._http._tcp.local: type A, class IN, addr 10.0.0.50
```

---

#### T093: Verify Ethernet IP Advertised on Ethernet Network

**Goal**: Query from Network B (Ethernet) and verify response contains Ethernet IP only

**Steps**:
1. **From another device on Network B** (Ethernet network):
   ```bash
   avahi-browse -r _http._tcp --resolve
   ```
2. Capture mDNS traffic on laptop Ethernet interface:
   ```bash
   sudo tcpdump -i eth0 -vv port 5353
   ```
3. Look for mDNS response containing A record
4. Verify A record contains **Ethernet IP only** (192.168.1.100)

**Success Criteria**:
- [ ] mDNS response received on Ethernet interface
- [ ] A record contains Ethernet IP: `192.168.1.100`
- [ ] A record does **NOT** contain WiFi IP: `10.0.0.50` ✅ **RFC 6762 §15**
- [ ] Response visible in tcpdump output

**Expected tcpdump output** (partial):
```
MultiNIC-Test._http._tcp.local: type A, class IN, addr 192.168.1.100
```

---

#### T094: Verify Connection Success

**Goal**: Verify clients can connect to service using advertised IPs

**Steps**:
1. **From Network A client** (WiFi):
   - Try connecting to WiFi IP: `curl http://10.0.0.50:8080`
   - Expected: Connection succeeds ✅
   - Try connecting to Ethernet IP: `curl http://192.168.1.100:8080`
   - Expected: Connection fails (network unreachable) ✅
2. **From Network B client** (Ethernet):
   - Try connecting to Ethernet IP: `curl http://192.168.1.100:8080`
   - Expected: Connection succeeds ✅
   - Try connecting to WiFi IP: `curl http://10.0.0.50:8080`
   - Expected: Connection fails (network unreachable) ✅

**Success Criteria**:
- [ ] WiFi clients can connect to WiFi IP
- [ ] WiFi clients **cannot** connect to Ethernet IP (network isolation verified)
- [ ] Ethernet clients can connect to Ethernet IP
- [ ] Ethernet clients **cannot** connect to WiFi IP (network isolation verified)
- [ ] ✅ **This validates the fix prevents Issue #27 connectivity failures**

---

### T095: Single-Interface Regression Test

**Goal**: Verify single-interface machines still work correctly (no regressions)

**Steps**:
1. **On single-interface machine** (or disable all but one interface):
   ```bash
   # Disable WiFi (if testing on laptop)
   sudo ip link set wlan0 down
   ```
2. Run responder:
   ```bash
   cd examples/interface-specific
   sudo ./interface-test
   ```
3. Verify responder detects single interface:
   ```
   Available network interfaces:
     [3] eth0       → [192.168.1.100]
   ```
4. Query from network and verify response contains correct IP
5. Verify connection succeeds

**Success Criteria**:
- [ ] Responder starts successfully with single interface
- [ ] Correct IP detected for single interface
- [ ] mDNS response contains correct IP
- [ ] Clients can connect successfully
- [ ] ✅ **No regression in single-interface behavior**

---

### T096: Document Results

**Goal**: Create manual test report with results

**Steps**:
1. Complete all manual tests (T090-T095)
2. Document results in `MANUAL_TEST_REPORT.md`
3. Include:
   - Test environment details (hardware, networks, IPs)
   - tcpdump output samples
   - Success/failure for each test
   - Screenshots (optional)
   - Any issues encountered

**Success Criteria**:
- [ ] All tests documented with pass/fail status
- [ ] Evidence included (tcpdump output, logs)
- [ ] Any issues or deviations noted

---

## Platform Testing (T097-T099)

### T097: Linux Platform Test

**Goal**: Verify IP_PKTINFO control messages work on Linux

**Platform**: Ubuntu 22.04+ / Debian 11+ / RHEL 8+

**Steps**:
1. Run multi-interface test (T090-T094) on Linux
2. Verify control messages extracted:
   - Add debug logging to `internal/transport/udp.go:214`:
     ```go
     if cm != nil {
         interfaceIndex = cm.IfIndex
         log.Printf("DEBUG: Extracted interfaceIndex=%d via IP_PKTINFO", interfaceIndex)
     }
     ```
3. Verify debug log shows non-zero interface indices
4. Verify A records contain correct per-interface IPs

**Success Criteria**:
- [ ] Control messages successfully extracted via IP_PKTINFO
- [ ] Interface indices non-zero in debug logs
- [ ] Per-interface IP addressing works correctly
- [ ] ✅ **Linux: PASS**

---

### T098: macOS Platform Test

**Goal**: Verify IP_RECVIF control messages work on macOS

**Platform**: macOS 10.14+ (Mojave or later)

**Steps**:
1. Run multi-interface test (T090-T094) on macOS
2. Verify control messages extracted via IP_RECVIF
   - Same debug logging as T097
3. Use `dns-sd` tool for querying:
   ```bash
   dns-sd -B _http._tcp local
   ```
4. Verify A records contain correct per-interface IPs

**Success Criteria**:
- [ ] Control messages successfully extracted via IP_RECVIF
- [ ] Interface indices non-zero in debug logs
- [ ] Per-interface IP addressing works correctly
- [ ] ✅ **macOS: PASS**

---

### T099: Windows Platform Test

**Goal**: Verify graceful degradation on Windows

**Platform**: Windows 10/11

**Steps**:
1. Build Beacon on Windows:
   ```powershell
   go build .\examples\interface-specific\main.go
   ```
2. Run responder (as Administrator):
   ```powershell
   .\main.exe
   ```
3. Verify responder starts successfully
4. Check if control messages available:
   - If interfaceIndex=0 in logs → Graceful degradation working ✅
   - If interfaceIndex>0 → Control messages working ✅
5. Use `avahi-browse` via WSL or Bonjour SDK to query
6. Verify service discoverable and connectable

**Success Criteria**:
- [ ] Responder starts successfully on Windows
- [ ] Graceful degradation works (interfaceIndex=0 fallback)
- [ ] Service discoverable via mDNS
- [ ] Clients can connect to service
- [ ] ✅ **Windows: Graceful degradation verified**

**Note**: Windows control message support varies by version. Graceful degradation (interfaceIndex=0 → getLocalIPv4) ensures functionality even without control messages.

---

### T100: Document Platform Compatibility

**Goal**: Create platform compatibility matrix

**Steps**:
1. Complete platform tests (T097-T099)
2. Create `PLATFORM_COMPATIBILITY.md` with results
3. Include:
   - Platform versions tested
   - Control message support status
   - Test results (pass/fail)
   - Any platform-specific notes

**Template**:

| Platform | Version | Control Messages | Interface Index | RFC 6762 §15 | Status |
|----------|---------|------------------|-----------------|---------------|--------|
| Linux    | Ubuntu 22.04 | IP_PKTINFO ✅ | Extracted ✅ | Compliant ✅ | **PASS** |
| macOS    | 13.0 Ventura | IP_RECVIF ✅ | Extracted ✅ | Compliant ✅ | **PASS** |
| Windows  | 11 22H2 | Graceful degradation ⚠️ | Fallback (0) ⚠️ | Best-effort ⚠️ | **DEGRADED** |

**Success Criteria**:
- [ ] Platform compatibility documented
- [ ] Results for all 3 platforms included
- [ ] Any platform limitations noted

---

## Manual Test Checklist

**Before Starting**:
- [ ] Hardware: Multi-interface laptop available
- [ ] Networks: Two isolated networks configured
- [ ] Software: Beacon built and ready to run
- [ ] Tools: tcpdump/Wireshark, avahi-browse/dns-sd installed

**Multi-Interface Testing**:
- [ ] T090: Test environment setup complete
- [ ] T091: mDNS service registered successfully
- [ ] T092: WiFi IP advertised on WiFi network (RFC 6762 §15 ✅)
- [ ] T093: Ethernet IP advertised on Ethernet network (RFC 6762 §15 ✅)
- [ ] T094: Connections succeed to correct IPs, fail to wrong IPs
- [ ] T095: Single-interface regression test passes

**Platform Testing**:
- [ ] T097: Linux test complete (IP_PKTINFO ✅)
- [ ] T098: macOS test complete (IP_RECVIF ✅)
- [ ] T099: Windows test complete (graceful degradation ⚠️)

**Documentation**:
- [ ] T096: Manual test results documented
- [ ] T100: Platform compatibility matrix complete

---

## Expected Results Summary

**Multi-Interface Behavior** (RFC 6762 §15):
- ✅ Query on WiFi → Response contains WiFi IP only
- ✅ Query on Ethernet → Response contains Ethernet IP only
- ✅ No cross-interface IP leakage
- ✅ Clients can connect to correct IP
- ✅ Clients cannot connect to wrong IP (network isolation)

**Platform Support**:
- ✅ Linux: Full support via IP_PKTINFO
- ✅ macOS: Full support via IP_RECVIF
- ⚠️ Windows: Graceful degradation (interfaceIndex=0 → getLocalIPv4)

**Regression**:
- ✅ Single-interface machines work correctly
- ✅ No existing functionality broken

---

## Troubleshooting

### Issue: Control messages not extracted (interfaceIndex=0)

**Symptoms**: Debug logs show `interfaceIndex=0` on Linux/macOS

**Possible Causes**:
1. Platform doesn't support control messages
2. Socket option not set correctly
3. Kernel version too old

**Solution**:
- Verify kernel version: `uname -r`
- Check socket option set: Look for `SetControlMessage` in logs
- Graceful degradation should still work (falls back to getLocalIPv4)

### Issue: Multiple IPs in response

**Symptoms**: mDNS response contains IPs from multiple interfaces

**Possible Causes**:
1. Implementation not using getIPv4ForInterface
2. interfaceIndex not propagated correctly

**Solution**:
- Check responder logs for interface index
- Verify handleQuery receives correct interfaceIndex
- File bug report with details

### Issue: Cannot bind to port 5353

**Symptoms**: Responder fails to start with "address already in use"

**Possible Causes**:
1. System mDNS daemon already running (Avahi/Bonjour)
2. Another Beacon instance running
3. Insufficient permissions

**Solution**:
- Stop system daemon: `sudo systemctl stop avahi-daemon` (Linux)
- Check for running processes: `sudo lsof -i :5353`
- Run with sudo: `sudo ./interface-test`

---

## Notes

- Manual testing **cannot be automated** - requires physical hardware setup
- Platform testing requires access to Linux, macOS, and Windows machines
- Network isolation is critical for validating RFC 6762 §15 compliance
- Use tcpdump/Wireshark to verify exact mDNS responses
- Document any unexpected behavior for investigation

---

**Status**: Ready for manual testing when hardware available
