control DeparserImpl (packet_out packet,
                      in headers_t hdr) {

    apply {
        packet.emit(hdr.cpu_in);
        packet.emit(hdr.ethernet);
        packet.emit(hdr.ipv4);
    }
}
