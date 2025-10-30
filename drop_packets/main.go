package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go  -target native -cc clang -cflags "-O2 -g -Wall -Werror" bpf ebpf/tcp_drop.c -- -I/usr/include/x86_64-linux-gnu -I/usr/include

func main() {
	var (
		iface = flag.String("interface", "lo", "Network interface to attach to")
		port  = flag.Uint("port", 4040, "TCP port to drop packets on")
	)
	flag.Parse()

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("Failed to remove memlock: %v", err)
	}

	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("Failed to load eBPF objects: %v", err)
	}
	defer objs.Close()

	key := uint32(0)
	value := uint32(*port)
	if err := objs.TargetPort.Put(&key, &value); err != nil {
		log.Fatalf("Failed to update target port: %v", err)
	}

	log.Printf("Configured to drop TCP packets on port: %d\n", port)

	ifaceObj, err := net.InterfaceByName(*iface)
	if err != nil {
		log.Fatalf("Failed to find interface %s: %v", *iface, err)
	}

	xdpLink, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpDropTcpPort,
		Interface: ifaceObj.Index,
		Flags:     link.XDPGenericMode,
	})
	if err != nil {
		log.Fatalf("Failed to attach XDP program: %v", err)
	}
	defer xdpLink.Close()

	log.Printf("Successfully attached XDP program to interface: %s\n", *iface)
	log.Println("Press Ctrl+C to exit...")

	// Wait for a signal to exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("\nDetaching eBPF program and exiting...")
}
