# Bonjour Conformance Test (BCT) harness

`examples/bct` is the device-under-test program for Apple's Bonjour
Conformance Test. BCT runs on a Mac. It tests a **separate** device on the same
network: it probes, queries, injects name conflicts and checks the replies
against RFC 3927, RFC 6762 and RFC 6763. Products that do not use Apple's
Bonjour code must pass it to use the Bonjour name.

Facts below are from the BCT 1.5.4 `README.txt` and
`Conformance Test Outline.txt` in the DMG.

## What you need

- A Mac (OS X 10.8 or later) with BCT. Download it from
  <https://developer.apple.com/bonjour/> (Apple developer sign-in). BCT is an
  x86_64 macOS program: on Apple Silicon it needs Rosetta 2. It does not run on
  Linux or WSL; it uses macOS BPF capture and Apple private frameworks.
- The Mac on **wired Ethernet**. Apple: "No wireless connection is allowed on
  the test machine side."
- The device on the same hub/switch or a Wi-Fi access point in bridge mode (no
  NAT or routing between them). The test network should be isolated from the
  Internet with no other devices on it.
  - A Linux VM with a bridged NIC or a Raspberry Pi works. WSL2 with default
    NAT networking does not: LAN multicast does not reach it.
  - Stop Avahi or any other mDNS responder on the device, or BCT tests the
    wrong stack.
  - On Raspberry Pi Wi-Fi, turn off power saving (`iw wlan0 set power_save off`).
- Phase I (link-local addressing, RFC 3927) tests the device's **operating
  system**, not Beacon. With no DHCP server, the device must self-assign a
  169.254.x.x address (e.g. NetworkManager `ipv4.link-local enabled`, or
  `avahi-autoipd`). BCT only accepts mDNS packets from the address the device
  ends Phase I with.

## Run

On the device (no root needed for Beacon itself):

```bash
go build -o bct ./examples/bct
./bct -host beacon-bct -name "Beacon BCT" -type _http._tcp.local -port 8080
```

Flags: `-host` (host label, `.local` is added), `-name`, `-type`, `-port`,
`-ipv6` (also answer on `ff02::fb`).

On the Mac, as root (`./BonjourConformanceTest -h` lists every option):

```bash
# Full run, wired interface en0, results to a file:
sudo ./BonjourConformanceTest -I en0 -F ~/bct-results.txt
# mDNS phase only, hot-plug subtest disabled (no cable pulls):
sudo ./BonjourConformanceTest -I en0 -M h -F ~/bct-results.txt
```

Useful options from BCT 1.5.4: `-L` / `-M` / `-N` select the link-local,
mDNS and interoperability phases, and letters after them **disable** subtests
(`-M h` = no hot-plugging). `-S` stops at the first failure. `-D` / `-DD`
enable debug logging. `-Aip` / `-Amac` ignore traffic from the access point.
`-E "<name>"` ignores a name for the whole test. BCT also writes `BCT.pcap`.

The chattiness subtest (III.2) alone can take up to 24 hours; a full run
takes several hours at least. The harness log prints probes, announcements
and a rename forced during registration.

## Smoke test without a Mac

From another host on the same LAN, a plain DNS client exercises the legacy
unicast path (RFC 6762 §6.7):

```bash
dig -p 5353 @<device-ip> _http._tcp.local PTR
dig -p 5353 @<device-ip> 'Beacon\032BCT._http._tcp.local' SRV
dig -p 5353 @<device-ip> beacon-bct.local A
dig -p 5353 @<device-ip> _services._dns-sd._udp.local PTR
```

Each answer must have TTL 10, the `aa` flag and no cache-flush bit.

## Known gaps BCT is likely to report

Tracked for the next release (see the PR/issue that adds BCT readiness):

- The host name is not probed, and probes carry no Authority section, so BCT
  may not recognise them as probes (II.1, RFC 6762 §8.1).
- Conflicts are not detected during probing on a real network, nor after
  announcement (II.2, II.6, RFC 6762 §8.1, §9).
- Registration gives up after 10 renames; BCT denies more than 15 (II.3).
- Shared (PTR) answers are not delayed 20-120 ms and back-to-back queries are
  not aggregated (II.8, II.15); TC queries are not held 400-500 ms (II.10).
- Type ANY (255) queries are not answered (II.7, II.8).
- mDNS packets use the OS default IP TTL, not 255 (III.3, warning).
- No hot-plug handling (II.17): run with `-M h`.
- Manual name change (II.16) needs a rename mechanism the harness does not
  have yet.
