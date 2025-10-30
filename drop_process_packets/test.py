#!/usr/bin/env python3
"""
Automated test script for eBPF process filtering.
Continuously attempts outbound connections on different ports.
"""
import socket
import time

ports = [4040, 8080, 9000, 3000, 5000]

while True:
    for port in ports:
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(0.5)
            sock.connect(('127.0.0.1', port))
            print(f"{port}: Connected")
            sock.close()
        except socket.timeout:
            print(f"{port}: Timeout")
        except ConnectionRefusedError:
            print(f"{port}: Refused")
        except PermissionError:
            print(f"{port}: BLOCKED")
        except Exception as e:
            print(f"{port}: {e}")
    print()
    time.sleep(2)
