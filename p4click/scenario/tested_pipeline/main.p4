#include <core.p4>
#include <v1model.p4>


#include "include/define.p4"
#include "include/typedef.p4"
#include "include/headers.p4"
#include "include/parser.p4"
#include "include/verify_checksum.p4"
#include "include/ingress.p4"
#include "include/egress.p4"
#include "include/compute_checksum.p4"
#include "include/deparser.p4"


V1Switch(
    ParserImpl(),
    VerifyChecksumImpl(),
    IngressImpl(),
    EgressImpl(),
    ComputeChecksumImpl(),
    DeparserImpl()
) main;
