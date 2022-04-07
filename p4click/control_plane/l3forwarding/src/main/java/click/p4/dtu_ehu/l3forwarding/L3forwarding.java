/*
 * Copyright 2019-present Open Networking Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package click.p4.dtu_ehu.l3forwarding;

import org.onlab.packet.MacAddress;
import org.onlab.packet.Ethernet;
import org.onlab.packet.IpAddress;
import org.onlab.packet.IPv4;
import org.onlab.packet.VlanId;
import org.onosproject.core.ApplicationId;
import org.onosproject.mastership.MastershipService;
import org.onosproject.net.ConnectPoint;
import org.onosproject.net.DeviceId;
import org.onosproject.net.Host;
import org.onosproject.net.PortNumber;
import org.onosproject.net.HostId;
import org.onosproject.net.Path;
import org.onosproject.net.device.DeviceService;
import org.onosproject.net.flow.DefaultTrafficSelector;
import org.onosproject.net.flow.FlowRule;
import org.onosproject.net.flow.FlowRuleService;
import org.onosproject.net.flow.TrafficSelector;
import org.onosproject.net.flow.criteria.PiCriterion;
import org.onosproject.net.host.HostService;
import org.onosproject.net.intf.Interface;
import org.onosproject.net.intf.InterfaceService;
import org.onosproject.net.packet.InboundPacket;
import org.onosproject.net.packet.PacketContext;
import org.onosproject.net.packet.PacketPriority;
import org.onosproject.net.packet.PacketProcessor;
import org.onosproject.net.packet.PacketService;
import org.onosproject.net.pi.model.PiActionId;
import org.onosproject.net.pi.model.PiActionParamId;
import org.onosproject.net.pi.model.PiMatchFieldId;
import org.onosproject.net.pi.runtime.PiAction;
import org.onosproject.net.pi.runtime.PiActionParam;
import org.onosproject.net.topology.TopologyService;
import org.osgi.service.component.annotations.Activate;
import org.osgi.service.component.annotations.Component;
import org.osgi.service.component.annotations.Deactivate;
import org.osgi.service.component.annotations.Reference;
import org.osgi.service.component.annotations.ReferenceCardinality;
import click.p4.dtu_ehu.l3forwarding.common.Utils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Set;
import java.util.stream.Collectors;

import static click.p4.dtu_ehu.l3forwarding.AppConstants.INITIAL_SETUP_DELAY;

/**
 * Simple L3forwarding based on Reactiveforwarding
 */
@Component(
        immediate = true,
        enabled = true
)
public class L3forwarding {

    private final Logger log = LoggerFactory.getLogger(getClass());


    private ApplicationId appId;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    protected TopologyService topologyService;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    private HostService hostService;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    private DeviceService deviceService;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    protected PacketService packetService;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    private InterfaceService interfaceService;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    private FlowRuleService flowRuleService;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    private MastershipService mastershipService;

    @Reference(cardinality = ReferenceCardinality.MANDATORY)
    private MainComponent mainComponent;


    private ReactivePacketProcessor processor = new ReactivePacketProcessor();

    @Activate
    protected void activate() {
        appId = mainComponent.getAppId();

        packetService.addProcessor(processor, PacketProcessor.director(2));
        requestIntercepts();

        //mainComponent.scheduleTask(this::setUpAllDevices, INITIAL_SETUP_DELAY);

        log.info("Started L3 Forwarding app");
    }

    @Deactivate
    protected void deactivate() {

        flowRuleService.removeFlowRulesById(appId);
        packetService.removeProcessor(processor);
        withdrawIntercepts();

        processor = null;

        log.info("Stopped L3 Forwarding app");
    }

    private void requestIntercepts() {
        TrafficSelector.Builder selector = DefaultTrafficSelector.builder();
        selector.matchEthType(Ethernet.TYPE_IPV4);
        packetService.requestPackets(selector.build(), PacketPriority.REACTIVE, appId);

    }

    private void withdrawIntercepts() {
        TrafficSelector.Builder selector = DefaultTrafficSelector.builder();
        selector.matchEthType(Ethernet.TYPE_IPV4);
        packetService.cancelPackets(selector.build(), PacketPriority.REACTIVE, appId);
    }

    private class ReactivePacketProcessor implements PacketProcessor {

        @Override
        public void process(PacketContext context) {

            if (context.isHandled()) {
                return;
            }

            InboundPacket pkt = context.inPacket();
            Ethernet ethPkt = pkt.parsed();

            if (ethPkt == null) {
                return;
            }

            if(isControlPacket(ethPkt))
                return;

            if (ethPkt.getEtherType() != Ethernet.TYPE_IPV4)
                return;

            IPv4 ipv4Packet = (IPv4) ethPkt.getPayload();

            //Set<Host> hosts = hostService.getHostsByMac(ethPkt.getDestinationMAC());
            IpAddress srcIp = IpAddress.valueOf(ipv4Packet.getSourceAddress());
            IpAddress dstIp = IpAddress.valueOf(ipv4Packet.getDestinationAddress());
            Set<Host> hosts = hostService.getHostsByIp(dstIp);


            if(hosts.isEmpty()){
                log.info("No host with destination IP [{}] was found.",
                        IpAddress.valueOf(ipv4Packet.getDestinationAddress()).getIp4Address().toString());
                return;
            }

            //Assume first host with an IP is OK
            Host dst = hosts.iterator().next();

            if (dst == null) {
                flood(context);
                return;
            }

            // If on edge just forward
            // This case should not happen, if L2 app is running
            if (pkt.receivedFrom().deviceId().equals(dst.location().deviceId())) {
                if (!context.inPacket().receivedFrom().port().equals(dst.location().port())) {
                    installRule(context, dst.location().port(), dst.mac());
                }
                return;
            }

            // Get paths
            Set<Path> paths =
                    topologyService.getPaths(topologyService.currentTopology(),
                            pkt.receivedFrom().deviceId(),
                            dst.location().deviceId());
            if (paths.isEmpty()) {
                log.warn("No paths found going from here {} for {} -> {}",
                        pkt.receivedFrom(),
                        srcIp.getIp4Address().getIp4Address().toString(),
                        dstIp.getIp4Address().getIp4Address().toString());
                // If there are no paths, flood and bail.
                flood(context);
                return;
            }


            // Otherwise, pick a path that does not lead back to where we
            // came from; if no such path, flood and bail.
            Path path = pickForwardPathIfPossible(paths, pkt.receivedFrom().port());
            if (path == null) {
                log.warn("Don't know where to go from here {} for {} -> {}",
                        pkt.receivedFrom(),
                        srcIp.getIp4Address().getIp4Address().toString(),
                        dstIp.getIp4Address().getIp4Address().toString());
                flood(context);
                return;
            }

            // Otherwise forward and be done with it.
            if(!pkt.receivedFrom().deviceId().equals(dst.location().deviceId())){
                installRule(context, path.src().port(), MacAddress.valueOf("00:00:00:11:11:11"));
            }else{
                installRule(context, path.src().port(), dst.mac());
            }
        }

    }

    // Indicates whether this is a control packet, e.g. LLDP, BDDP
    private boolean isControlPacket(Ethernet eth) {
        short type = eth.getEtherType();
        return type == Ethernet.TYPE_LLDP || type == Ethernet.TYPE_BSN;
    }

    // Selects a path from the given set that does not lead back to the
    // specified port if possible.
    private Path pickForwardPathIfPossible(Set<Path> paths, PortNumber notToPort) {
        for (Path path : paths) {
            if (!path.src().port().equals(notToPort)) {
                return path;
            }
        }
        return null;
    }

    // Floods the specified packet if permissible.
    private void flood(PacketContext context) {
        if (topologyService.isBroadcastPoint(topologyService.currentTopology(),
                context.inPacket().receivedFrom())) {
            packetOut(context, PortNumber.FLOOD);
        } else {
            context.block();
        }
    }

    // Sends a packet out the specified port.
    private void packetOut(PacketContext context, PortNumber portNumber) {
        context.treatmentBuilder().setOutput(portNumber);
        context.send();
    }

    // Install a rule forwarding the packet to the specified port.
    private void installRule(PacketContext context, PortNumber portNumber, MacAddress myMac) {
        //
        // We don't support (yet) buffer IDs in the Flow Service so
        // packet out first.
        //
        InboundPacket pkt = context.inPacket();
        Ethernet inPkt = pkt.parsed();

        DeviceId deviceId = pkt.receivedFrom().deviceId();

        if (inPkt.getEtherType() != Ethernet.TYPE_IPV4) {
            log.info("Trying to install a rule for a packet that is not IPv4, it is ethertype: {}...", inPkt.getEtherType());
            return;
        }

        IPv4 ipv4Packet = (IPv4) inPkt.getPayload();
        IpAddress srcIp = IpAddress.valueOf(ipv4Packet.getSourceAddress());
        IpAddress dstIp = IpAddress.valueOf(ipv4Packet.getDestinationAddress());

        final String tableId = "IngressImpl.l3forwarding_ing.l3forwarding_table";

        //Criterion
        final PiCriterion L3FwdCriterion = PiCriterion.builder()
                .matchExact(PiMatchFieldId.of("hdr.ipv4.srcAddr"),
                        srcIp.getIp4Address().toOctets())
                .matchExact(PiMatchFieldId.of("hdr.ipv4.dstAddr"),
                        dstIp.getIp4Address().toOctets())
                .build();

        //Action
        final PiAction l3UnicastAction = PiAction.builder()
                .withId(PiActionId.of("IngressImpl.l3forwarding_ing.route"))
                .withParameter(new PiActionParam(
                        PiActionParamId.of("myMac"),
                        myMac.toBytes()))
                .withParameter(new PiActionParam(
                        PiActionParamId.of("port_num"),
                        portNumber.toLong()))
                .build();

        // Build Flowerule takes PiCriterion and PiAction and buils Selector and Treatment.
        final FlowRule rule = Utils.buildFlowRule(
                deviceId,
                appId,
                tableId,
                L3FwdCriterion,
                l3UnicastAction);

        flowRuleService.applyFlowRules(rule);

        log.info("Added L3 unicast rule on switch {}...", deviceId);

        packetOut(context, portNumber);
    }

    /**
     * Returns a set of ports for the given device that are used to connect
     * hosts to the fabric.
     *
     * @param deviceId device ID
     * @return set of host facing ports
     */
    private Set<PortNumber> getHostFacingPorts(DeviceId deviceId) {
        // Get all interfaces configured via netcfg for the given device ID and
        // return the corresponding device port number. Interface configuration
        // in the netcfg.json looks like this:
        // "device:leaf1/3": {
        //   "interfaces": [
        //     {
        //       "name": "leaf1-3",
        //       "ips": ["2001:1:1::ff/64"]
        //     }
        //   ]
        // }
        return interfaceService.getInterfaces().stream()
                .map(Interface::connectPoint)
                .filter(cp -> cp.deviceId().equals(deviceId))
                .map(ConnectPoint::port)
                .collect(Collectors.toSet());
    }

}
