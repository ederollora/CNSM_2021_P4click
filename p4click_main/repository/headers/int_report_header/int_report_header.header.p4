header int_report_header_t {
    bit<4> ver;
    bit<4> len;
    bit<3> nProt;
    bit<6> repMdBits;
    bit<6> rsvd;
    bit<1> d;
    bit<1> q;
    bit<1> f;
    bit<6> hw_id;
    bit<32> switch_id;
    bit<32> seq_no;
    bit<32> ingress_tstamp;
}
