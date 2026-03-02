# rh-diag

`rh-diag` is a robust system diagnostic CLI tool purposely built for **Red Hat Enterprise Linux 9** and **Rocky Linux 9**. It acts as a fast, user-space triage utility to quickly identify common misconfigurations or capacity bottlenecks before they cause outages.

It is designed to be lightweight, natively analyzing kernel states and running system tools securely without requiring full `root` privileges where possible.

## Features

- **DNS Diagnostics (`rh-diag dns`)**: Actively parses `/etc/resolv.conf` and executes synthetic UDP resolution queries directly against configured nameservers to detect split-brain DNS failures or dropped traffic on port 53.
- **Networking & Routing (`rh-diag net`)**: Enumerates active IPv4/IPv6 interfaces, organically discovers the active default gateway from the OS routing table, and verifies L2 switching and outbound NAT via ICMP pings.
- **Firewalld State (`rh-diag firewall`)**: Safely parses active zones and list allowed ports and services exposed by the `firewalld` daemon, failing gracefully if the daemon is inactive.
- **System Resources (`rh-diag sysres`)**: Checks memory, CPU load averages, and local disk usage directly from the `/proc` filesystem and native Golang statfs syscalls—blazing fast and strictly avoiding resource-heavy external commands like `top` or `df`.

## Installation

Ensure you have Go 1.21+ installed.

```bash
git clone https://github.com/darrylonvinyl/rh-diag.git
cd rh-diag
go build -o rh-diag main.go
sudo mv rh-diag /usr/local/bin/
```

## Usage

Run the main command to view available modules:

```bash
rh-diag --help
```

### Examples

**Check System Resource Capacity:**
```bash
$ rh-diag sysres
Running System Resource diagnostics...

--- Memory & Swap ---
Memory Total      : 14.46 GB
Memory Available  : 2.77 GB
Swap Enabled      : No

--- CPU Load Average ---
1-Minute Load     : 2.78
5-Minute Load     : 1.75
15-Minute Load    : 1.56

--- Persistent Local Disks ---
[/] Usage: 45.2% (150.21 GB / 332.00 GB)

Summary: Core system capacity (Memory/Disks) appears within normal operating thresholds.
```

**Check Network and Gateway Status:**
```bash
$ rh-diag net
Running Networking & Routing diagnostics...

--- Active Interfaces ---
[lo] Flags: up|loopback|running
  -> IP: 127.0.0.1
  -> IP: ::1
[eth0] Flags: up|broadcast|multicast|running
  -> IP: 172.17.0.2

--- Routing & Connectivity ---
Default Gateway : 172.17.0.1
Gateway Ping    : [PASS] (Reachable)
Internet Ping   : [PASS] (Reachable against 8.8.8.8)

Summary: Hardware and basic system routing look HEALTHY.
```

### Developing & Testing

A `.devcontainer` is provided for VS Code users, but if you just have Docker installed locally, you can use the standard Make targets to handle dependency injection and isolated testing without cluttering your host machine.

**Run the unified test suite natively in Linux:**
```bash
make test
```

**Compile the binary for Linux locally (without needing Go installed):**
```bash
make build
```
