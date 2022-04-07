control L3forwarding_ing(inout headers_t hdr,
                         inout local_metadata_t local_metadata,
                         inout standard_metadata_t standard_metadata){

    action route(macAddr_t myMac, port_t port_num) {
        hdr.ethernet.srcAddr = hdr.ethernet.dstAddr;
        hdr.ethernet.dstAddr = myMac;
        hdr.ipv4.ttl = hdr.ipv4.ttl - 1;
        standard_metadata.egress_spec = port_num;
        local_metadata.forwarding_meta.fwd_done = true;
    }

    table l3forwarding_table {
        key = {
            hdr.ipv4.srcAddr: exact;
            hdr.ipv4.dstAddr: exact;
        }
        actions = {
            route;
            @defaultonly NoAction;
        }
        @name("l3f_exact_table")
        const default_action = NoAction;
    }

    apply {
        if (hdr.ipv4.isValid()) {
            l3forwarding_table.apply();
            /*
            Maybe this is better at some point.
            if(ipv4fwd_table.apply().hit){
                fwd_done=true;
            }
            */
        }
    }

}

