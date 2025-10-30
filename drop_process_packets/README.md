# Problem 2
```
Write an eBPF code to allow traffic only at a specific TCP port (default 4040) for a given process name (for e.g, "myprocess"). All the traffic to all other ports for only that process should be dropped.
```

## eBPF Process-Based Port Filter
This eBPF program allows traffic only at a specific TCP port (default 4040) for a given process name (e.g., "python3", "myprocess"). All traffic to/from all other ports for that process is blocked in **both directions** (inbound and outbound).

## Features
-  Filter by process name
-  Allow traffic only on specified port (default: 4040)
-  Block inbound traffic (bind operations) on other ports
-  Block outbound traffic (connect operations) on other ports
-  All other processes are unaffected


## Building
```bash
make clean
make build
```

## Usage
```bash
# Filter process "python3" to only use port 4040
sudo ./cgroup-tcp-drop -process python3 -port 4040
```


### Manual Testing
```bash
# Terminal 1: Start the filter
sudo ./cgroup-tcp-drop -process python3 -port 4040

# Terminal 2: Try to start server on blocked port (will fail)
python3 -m http.server 8080
# PermissionError: [Errno 1] Operation not permitted

# Terminal 2: Start server on allowed port (will succeed)
python3 -m http.server 4040
# Serving HTTP on :: port 4040 ...
```

#### Test Outbound Blocking
```bash
# Terminal 1: Start the filer
sudo ./cgroup-tcp-drop -process python3 -port 9000

# Terminal 2: Start server on allowed port
python3 -m http.server 9000

# Terminal 2: Run test script
python3 test.py
```

### Monitor Blocks in Real-Time
```bash
sudo cat /sys/kernel/debug/tracing/trace_pipe | grep BLOCKING
```
