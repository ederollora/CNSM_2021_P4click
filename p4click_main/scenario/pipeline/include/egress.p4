

control EgressImpl (inout headers_t hdr,
                    inout local_metadata_t local_metadata,
                    inout standard_metadata_t standard_metadata) {

    

    apply {

        if (standard_metadata.egress_port == CPU_PORT) {
            hdr.cpu_in.setValid();
            hdr.cpu_in.ingress_port = standard_metadata.ingress_port;
            exit;
        }


    }
}

