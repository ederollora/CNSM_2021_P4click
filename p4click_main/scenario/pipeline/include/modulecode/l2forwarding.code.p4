control L2forwarding_ing(inout headers_t hdr,
                         inout local_metadata_t local_metadata,
                         inout standard_metadata_t standard_metadata){

    action drop() {
        mark_to_drop(standard_metadata);
    }

    action set_egress_port(port_t port_num) {
        standard_metadata.egress_spec = port_num;
        local_metadata.forwarding_meta.fwd_done = true;
    }

    table l2forwarding_table {
        key = {
            hdr.ethernet.dstAddr: exact;
        }
        actions = {
            set_egress_port;
            @defaultonly NoAction;
        }
        @name("l2f_exact_table")
        const default_action = NoAction;
    }

    apply {
        if (hdr.ethernet.isValid()) {
            l2forwarding_table.apply();
        }
    }

}
