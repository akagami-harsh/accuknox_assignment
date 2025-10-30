//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define TASK_COMM_LEN 16
#define MAX_PROCESS_NAME_LEN 16


struct process_config {
    char process_name[MAX_PROCESS_NAME_LEN];
    __u16 allowed_port;
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct process_config);
} process_filter_config SEC(".maps");

// Helper to compare strings
static __always_inline int str_equal(const char *s1, const char *s2, int len) {
    for (int i = 0; i < len; i++) {
        if (s1[i] != s2[i]) {
            return 0;
        }
        if (s1[i] == '\0') {
            return 1;
        }
    }
    return 1;
}

// Filtering logic for both IPv4 and IPv6
static __always_inline int filter_connect(struct bpf_sock_addr *ctx)
{
    __u32 config_key = 0;
    struct process_config *config = bpf_map_lookup_elem(&process_filter_config, &config_key);
    if (!config || config->process_name[0] == '\0') {
        return 1;
    }

    char comm[TASK_COMM_LEN];
    bpf_get_current_comm(&comm, sizeof(comm));

    if (!str_equal(comm, config->process_name, MAX_PROCESS_NAME_LEN)) {
        return 1;
    }

    __u16 dest_port = bpf_ntohs(ctx->user_port);

    if (dest_port == config->allowed_port) {
        return 1;  // Allow
    }

    bpf_printk("BLOCKING connect to port %d\n", dest_port);
    return 0;
}

// Filter for outbound IPv4 connections
SEC("cgroup/connect4")
int filter_connect4(struct bpf_sock_addr *ctx)
{
    return filter_connect(ctx);
}

// Filter for outbound IPv6 connections
SEC("cgroup/connect6")
int filter_connect6(struct bpf_sock_addr *ctx)
{
    return filter_connect(ctx);
}

// filtering logic for bind operations
static __always_inline int filter_bind(struct bpf_sock_addr *ctx)
{
    __u32 config_key = 0;
    struct process_config *config = bpf_map_lookup_elem(&process_filter_config, &config_key);
    if (!config || config->process_name[0] == '\0') {
        return 1;
    }

    char comm[TASK_COMM_LEN];
    bpf_get_current_comm(&comm, sizeof(comm));

    if (!str_equal(comm, config->process_name, MAX_PROCESS_NAME_LEN)) {
        return 1;
    }

    __u16 bind_port = bpf_ntohs(ctx->user_port);

    if (bind_port == config->allowed_port || bind_port == 0) {
        return 1;  // Allow
    }

    bpf_printk("BLOCKING bind to port %d\n", bind_port);
    return 0;
}

// Filter for IPv4 bind operations
SEC("cgroup/bind4")
int filter_bind4(struct bpf_sock_addr *ctx)
{
    return filter_bind(ctx);
}

// Filter for IPv6 bind operations
SEC("cgroup/bind6")
int filter_bind6(struct bpf_sock_addr *ctx)
{
    return filter_bind(ctx);
}

char LICENSE[] SEC("license") = "GPL";
