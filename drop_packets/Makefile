.PHONY: build generate

generate:
	@echo "Generating eBPF bytecode..."
	go generate

build: generate
	@echo "Building tcp-drop..."
	go build -o tcp-drop

clean:
	@echo "Cleaning up..."
	rm -f tcp-drop bpf_bpfel*.go bpf_*.o
