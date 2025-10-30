package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target native -cc clang -cflags "-O2 -g -Wall -Werror" -type process_config bpf ebpf/process_filter.bpf.c -- -I/usr/include/x86_64-linux-gnu

type ProcessConfig struct {
	ProcessName [16]byte
	AllowedPort uint16
}

func main() {
	var (
		port        = flag.Int("port", 4040, "TCP port to allow (all others blocked for tracked process)")
		cgroupPath  = flag.String("cgroup", "/sys/fs/cgroup", "Path to cgroup to attach to")
		processName = flag.String("process", "", "Process name to track")
	)
	flag.Parse()

	if *processName == "" {
		log.Fatal("Error: -process flag is required")
	}

	if len(*processName) > 15 {
		log.Fatal("Error: process name must be 15 characters or less")
	}

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal("Failed to remove memlock limit:", err)
	}

	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("Failed to load eBPF objects: %v", err)
	}
	defer objs.Close()

	var config ProcessConfig
	copy(config.ProcessName[:], *processName)
	config.AllowedPort = uint16(*port)

	key := uint32(0)
	if err := objs.ProcessFilterConfig.Put(&key, &config); err != nil {
		log.Fatalf("Failed to configure process filter: %v", err)
	}

	log.Printf("Filtering process '%s' on port %d", *processName, *port)

	// Attach all eBPF programs to cgroup
	programs := []struct {
		name   string
		attach ebpf.AttachType
		prog   *ebpf.Program
	}{
		{"connect4", ebpf.AttachCGroupInet4Connect, objs.FilterConnect4},
		{"connect6", ebpf.AttachCGroupInet6Connect, objs.FilterConnect6},
		{"bind4", ebpf.AttachCGroupInet4Bind, objs.FilterBind4},
		{"bind6", ebpf.AttachCGroupInet6Bind, objs.FilterBind6},
	}

	for _, p := range programs {
		lnk, err := link.AttachCgroup(link.CgroupOptions{
			Path:    *cgroupPath,
			Attach:  p.attach,
			Program: p.prog,
		})
		if err != nil {
			log.Fatalf("Failed to attach %s: %v", p.name, err)
		}
		defer lnk.Close()
	}
	log.Printf("Attached all filters to cgroup %s", *cgroupPath)

	log.Println("\nFilter active! Press Ctrl+C to stop")
	log.Println("Monitor: sudo cat /sys/kernel/debug/tracing/trace_pipe | grep BLOCKING")

	// Wait for signal to exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("\nReceived signal, exiting...")
}
