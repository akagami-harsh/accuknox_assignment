//go:build ignore

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define IPPROTO_TCP 6

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u32);
} target_port SEC(".maps");

#define TOTAL_SIZE (sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct tcphdr))

SEC("xdp")
int xdp_drop_tcp_port(struct xdp_md *ctx) {
    void *data_start = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    
    struct iphdr *ip = data_start + sizeof(struct ethhdr);
    struct tcphdr *tcp = data_start + sizeof(struct ethhdr) + sizeof(struct iphdr);

    if (data_start + TOTAL_SIZE > data_end) {
        return XDP_PASS;
    }

    if (ip->protocol != IPPROTO_TCP) {
        return XDP_PASS;
    }

    __u32 key = 0;
    __u32 *port = bpf_map_lookup_elem(&target_port, &key);
    
    if (port == NULL) {
        return XDP_PASS;
    }
    
    __u16 drop_port = (__u16)*port;

    if (tcp->source == bpf_htons(drop_port) || tcp->dest == bpf_htons(drop_port)) {
        bpf_printk("XDP: Dropping TCP packet on port %d\n", drop_port);
        return XDP_DROP;
    }
    
    return XDP_PASS;
}

char LICENSE[] SEC("license") = "GPL";
