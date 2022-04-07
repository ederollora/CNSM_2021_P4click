#include "modulecode/l2forwarding.code.p4"
#include "modulecode/l3forwarding.code.p4"
#include "modulecode/aclhandler.code.p4"


control IngressImpl (inout headers_t hdr,
                     inout local_metadata_t local_metadata,
                     inout standard_metadata_t standard_metadata) {

    L2forwarding_ing() l2forwarding_ing;
    L3forwarding_ing() l3forwarding_ing;
    Aclhandler_ing() aclhandler_ing;
    

    apply {

        if (hdr.cpu_out.isValid()) {
            standard_metadata.egress_spec = hdr.cpu_out.egress_port;
            hdr.cpu_out.setInvalid();
            exit;
        }

        l2forwarding_ing.apply(hdr, local_metadata, standard_metadata);
        l3forwarding_ing.apply(hdr, local_metadata, standard_metadata);
        aclhandler_ing.apply(hdr, local_metadata, standard_metadata);
    }
}

