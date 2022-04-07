control ComputeChecksumImpl (inout headers_t hdr,
                             inout local_metadata_t local_metadata) {

    apply {
        update_checksum(
            hdr.ipv4.isValid(),
                { hdr.ipv4.version,
              hdr.ipv4.ihl,
                  hdr.ipv4.diffServ,
                  hdr.ipv4.totalLen,
                  hdr.ipv4.identification,
                  hdr.ipv4.flags,
                  hdr.ipv4.fragOffset,
                  hdr.ipv4.ttl,
                  hdr.ipv4.protocol,
                  hdr.ipv4.srcAddr,
                  hdr.ipv4.dstAddr },
                hdr.ipv4.hdrChecksum,
                HashAlgorithm.csum16);
    }
}

