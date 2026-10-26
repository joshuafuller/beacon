# Bonjour Conformance Test (BCT) harness

`examples/bct` is the device-under-test program for Apple's Bonjour
Conformance Test. BCT runs on a Mac. It tests a **separate** device on the same
network: it probes, queries, injects name conflicts and checks the replies
against RFC 6762 / RFC 6763. Products that do not use Apple's Bonjour code must
pass it to use the Bonjour name.

## What you need

- A Mac with BCT installed. Download it from
  <https://developer.apple.com/bonjour/> (Apple developer sign-in). Version
  1.5.x is current.
- A second device on the **same layer-2 network** as the Mac (same switch or
  Wi-Fi, no router in between). Use a Linux VM with a bridged NIC, or a
  Raspberry Pi.
  - WSL2 with default NAT networking does not work: LAN multicast does not
    reach it.
  - Turn off Avahi or any other mDNS responder on the device, or BCT tests the
    wrong stack.
  - On Raspberry Pi Wi-Fi, turn off power saving (`iw wlan0 set power_save off`).
- Root (or `CAP_NET_RAW`) on the device is not needed for Beacon, but BCT's
  link-local address tests change the device's IPv4 configuration.

## Run

On the device:

```bash
go build -o bct ./examples/bct
./bct -host beacon-bct -name "Beacon BCT" -type _http._tcp.local -port 8080
```

Flags: `-host` (host label, `.local` is added), `-name`, `-type`, `-port`,
`-ipv6` (also answer on `ff02::fb`).

On the Mac (command used by brutella/dnssd with BCT 1.5.0; run
`./BonjourConformanceTest -h` to check the flags of your version):

```bash
sudo ./BonjourConformanceTest -S -M h -DD -E <router-ip>
```

A full run takes several hours. Results are written to `ConformanceTestResults`.
The harness log prints every probe, announcement and forced rename, so you can
match it against the results file.

## Smoke test without a Mac

From another host on the same LAN (or the same host), a plain DNS client
exercises the legacy unicast path (RFC 6762 §6.7):

```bash
dig -p 5353 @<device-ip> _http._tcp.local PTR
dig -p 5353 @<device-ip> 'Beacon\032BCT._http._tcp.local' SRV
dig -p 5353 @<device-ip> beacon-bct.local A
```

Each answer must have TTL 10 and the `aa` flag.

## Known gaps BCT is likely to report

- The host name (`beacon-bct.local`) is not probed; only service names are
  (RFC 6762 §8.1).
- Multicast responses over 9000 bytes set the TC bit (RFC 6762 §18.5).
- Service-type enumeration does not answer legacy unicast queries.
