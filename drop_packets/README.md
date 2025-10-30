# TCP Packet Dropper solution for eBPF Assignment

## Problem Statement
```
Write an eBPF code to drop the TCP packets on a port (def: 4040). Additionally, if you can make the port number configurable from the userspace, that will be a big plus.
```

An eBPF program that drops TCP packets on a configurable port using XDP hooks.

## Features

- Drop TCP packets on a specific port (default: 4040)
- It uses XDP(Xpress Data Path) to filter out traffic
- Port is configurable from userspace via command-line flag
- Drops packets on both source and destination ports

## Usage

Run with default settings (port 4040, loopback interface):
```bash
sudo ./tcp-drop
```

Drop packets on a custom port:
```bash
sudo ./tcp-drop -port 8080
```

Specify network interface:
```bash
sudo ./tcp-drop -interface lo -port 4040
```

## Testing

### Terminal 1: Run the eBPF program
```bash
sudo ./tcp-drop -port 4040
```

### Terminal 2: Start a server on the blocked port
```bash
nc -l 4040
```

### Terminal 3: Try to connect (this should fail/timeout)
```bash
nc localhost 4040
# or
curl http://localhost:4040
```

### Monitor dropped packets
```bash
sudo cat /sys/kernel/debug/tracing/trace_pipe | grep "Dropping TCP"
```
