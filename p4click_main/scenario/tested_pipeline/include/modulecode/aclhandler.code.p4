control Aclhandler_ing(inout headers_t hdr,
                         inout local_metadata_t local_metadata,
                         inout standard_metadata_t standard_metadata){

    action send_to_cpu() {
        standard_metadata.egress_spec = CPU_PORT;
    }

    action clone_to_cpu() {
        clone3(CloneType.I2E, CPU_CLONE_SESSION_ID, { standard_metadata.ingress_port });
    }

    table acl_table {
        key = {
            standard_metadata.ingress_port:          ternary;
            hdr.ethernet.etherType:                 ternary;
            hdr.ethernet.dstAddr:                   ternary;
            hdr.ethernet.srcAddr:                   ternary;
        }
        actions = {
            send_to_cpu;
            clone_to_cpu;
            @defaultonly NoAction;
        }
        @name("acl_exact_table")
        const default_action = NoAction;
    }

    apply {
        if (hdr.ethernet.isValid() && local_metadata.forwarding_meta.fwd_done == false) {
            acl_table.apply();
        }
    }

}
