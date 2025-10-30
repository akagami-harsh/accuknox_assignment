.PHONY: generate build clean

generate:
	@echo "Generating eBPF bytecode..."
	go generate

build: generate
	@echo "Building cgroup-tcp-drop..."
	go build -o cgroup-tcp-drop

clean:
	@echo "Cleaning up..."
	rm -f cgroup-tcp-drop bpfnew_*.go bpfnew_*.o
