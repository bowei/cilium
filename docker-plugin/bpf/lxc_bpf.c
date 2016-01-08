#include <iproute2/bpf_api.h>
#include <linux/ipv6.h>
#include <linux/if_ether.h>
#include <sys/socket.h>
#include <stdint.h>
#include "bpf_helpers.h"

#define TX_XMIT	0
#define TX_FRWD	1

#define ETH_HLEN 14

/* compiler workaround */
#define _htonl __builtin_bswap32

union v6addr {
        struct {
                __u32 p1;
                __u32 p2;
                __u32 p3;
                __u32 p4;
        };
        __u8 addr[16];
};

struct lxc_info {
	__u64 mac;
	int ifindex;
};

struct bpf_map_def SEC("maps") lxc_map = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u16),
	.value_size = sizeof(struct lxc_info),
	.max_entries = 1024,
};

static inline void set_dst_mac(struct __sk_buff *skb, char *mac)
{
        bpf_skb_store_bytes(skb, 0, mac, 6, 1);
}

static inline int do_redirect6(struct __sk_buff *skb, int nh_off)
{
	struct lxc_info *dst_lxc;
	__u16 lxc_id;
	__u8 hoplimit;
	union v6addr dst, dst_new;
        int *ifindex;
        char fmt[] = "skb %p len %d dst %x %x %x %x\n";

	/* FIXME: Validate source MAC and source IP */

	/* FIXME: Validate destination node ID and perform encap */

        dst.p1 = ntohl(load_word(skb, nh_off + offsetof(struct ipv6hdr, daddr) + sizeof(__u32) * 0));
        dst.p2 = ntohl(load_word(skb, nh_off + offsetof(struct ipv6hdr, daddr) + sizeof(__u32) * 1));
        dst.p3 = ntohl(load_word(skb, nh_off + offsetof(struct ipv6hdr, daddr) + sizeof(__u32) * 2));
        dst.p4 = ntohl(load_word(skb, nh_off + offsetof(struct ipv6hdr, daddr) + sizeof(__u32) * 3));

	bpf_trace_printk(fmt, sizeof(fmt), skb, skb->len, dst.p1, dst.p2, dst.p3, dst.p4);

	hoplimit = load_byte(skb, nh_off + offsetof(struct ipv6hdr, hop_limit));
	if (hoplimit <= 1) {
		/* FIXME: Handle */
		char fmt[] = "Hoplimit reached 0\n";
		bpf_trace_printk(fmt, sizeof(fmt));
		return -1;
	} else {
		__u8 new_hl;

		new_hl = hoplimit - 1;
                skb_store_bytes(skb, nh_off + offsetof(struct ipv6hdr, hop_limit),
                                &new_hl, sizeof(new_hl), 0);
		char fmt[] = "Decremented hoplimit\n";
		bpf_trace_printk(fmt, sizeof(fmt));
        }

	lxc_id = dst.p4 & 0xFFFF;

	dst_lxc = bpf_map_lookup_elem(&lxc_map, &lxc_id);
	if (dst_lxc) {
		set_dst_mac(skb, (char *) dst_lxc->mac);
		char fmt[] = "Found destination container locally\n";
		bpf_trace_printk(fmt, sizeof(fmt));
		bpf_clone_redirect(skb, dst_lxc->ifindex, 1);
	}

	return -1;
}

SEC("from-container")
int handle_ingress(struct __sk_buff *skb)
{
	int ret = 0, nh_off = BPF_LL_OFF + ETH_HLEN;

	if (likely(skb->protocol == __constant_htons(ETH_P_IPV6)))
		ret = do_redirect6(skb, nh_off);

	return ret;
}
BPF_LICENSE("GPL");