header ethernet_t {
    macAddr_t dstAddr;
    macAddr_t srcAddr;
    bit<16>   etherType;
}

header ipv4_t {
    bit<4>    version;
    bit<4>    ihl;
    bit<8>    diffServ;
    bit<16>   totalLen;
    bit<16>   identification;
    bit<3>    flags;
    bit<13>   fragOffset;
    bit<8>    ttl;
    bit<8>    protocol;
    bit<16>   hdrChecksum;
    ip4Addr_t srcAddr;
    ip4Addr_t dstAddr;
}

@controller_header("packet_out")
header cpu_out_header_t {
    port_t egress_port;
    bit<7> _pad;
}

@controller_header("packet_in")
header cpu_in_header_t {
    port_t ingress_port;
    bit<7> _pad;
}

struct forwarding_meta_t {
    bool fwd_done;
}

struct local_metadata_t {
    forwarding_meta_t  forwarding_meta;
}

struct headers_t {
    ethernet_t         ethernet;
    ipv4_t             ipv4;
    cpu_out_header_t   cpu_out;
    cpu_in_header_t    cpu_in;
}

