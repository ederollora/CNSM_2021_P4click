#!/usr/bin/python

#  Copyright 2019-present Open Networking Foundation
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

import argparse

from mininet.cli import CLI
from mininet.log import setLogLevel, info
from mininet.net import Mininet
from mininet.node import Host, Node
from mininet.topo import Topo
from mininet.link import Link, Intf
from stratum import StratumBmv2Switch

CPU_PORT = 255

the_switch =  None

class IPv4Host(Host):
    """Host that can be configured with an IPv4 gateway (default route).
    """

    def config(self, mac=None, ip=None, defaultRoute=None, lo='up', gw=None,
               **_params):
        super(IPv4Host, self).config(mac, ip, defaultRoute, lo, **_params)
        self.cmd('ip -4 addr flush dev %s' % self.defaultIntf())
        self.cmd('ip -6 addr flush dev %s' % self.defaultIntf())
        self.cmd('ip -4 link set up %s' % self.defaultIntf())
        self.cmd('ip -4 addr add %s dev %s' % (ip, self.defaultIntf()))
        if gw:
            self.cmd('ip -4 route add default via %s' % gw)
        # Disable offload
        for attr in ["rx", "tx", "sg"]:
            cmd = "/sbin/ethtool --offload %s %s off" % (
                self.defaultIntf(), attr)
            self.cmd(cmd)

        def updateIP():
            return ip.split('/')[0]

        self.defaultIntf().updateIP = updateIP


class TaggedIPv4Host(Host):
    """VLAN-tagged host that can be configured with an IPv4 gateway
    (default route).
    """
    vlanIntf = None

    def config(self, mac=None, ip=None, defaultRoute=None, lo='up', gw=None,
               vlan=None, **_params):
        super(TaggedIPv4Host, self).config(mac, ip, defaultRoute, lo, **_params)
        self.vlanIntf = "%s.%s" % (self.defaultIntf(), vlan)
        # Replace default interface with a tagged one
        self.cmd('ip -4 addr flush dev %s' % self.defaultIntf())
        self.cmd('ip -6 addr flush dev %s' % self.defaultIntf())
        self.cmd('ip -4 link add link %s name %s type vlan id %s' % (
            self.defaultIntf(), self.vlanIntf, vlan))
        self.cmd('ip -4 link set up %s' % self.vlanIntf)
        self.cmd('ip -4 addr add %s dev %s' % (ip, self.vlanIntf))
        if gw:
            self.cmd('ip -4 route add default via %s' % gw)

        self.defaultIntf().name = self.vlanIntf
        self.nameToIntf[self.vlanIntf] = self.defaultIntf()

        # Disable offload
        for attr in ["rx", "tx", "sg"]:
            cmd = "/sbin/ethtool --offload %s %s off" % (
                self.defaultIntf(), attr)
            self.cmd(cmd)

        def updateIP():
            return ip.split('/')[0]

        self.defaultIntf().updateIP = updateIP

    def terminate(self):
        self.cmd('ip -4 link remove link %s' % self.vlanIntf)
        super(TaggedIPv4Host, self).terminate()

class NullIntf( Intf ):
    "A dummy interface with a blank name that doesn't do any configuration"
    def __init__( self, name, **params ):
        self.name = ''
        self.node ="alaki"

class NullLink( Link ):
    "A dummy link that doesn't touch either interface"
    def makeIntfPair( cls, intf1, intf2, *args, **kwargs ):
        pass
    def delete( self ):
        pass

class TutorialTopo(Topo):
    def addIntf( self, switch, intfName ):
        "Add intf intfName to switch"
        self.addLink( switch, switch, cls=NullLink,
                      intfName1=intfName, cls2=NullIntf )

    def __init__(self, *args, **kwargs):
        Topo.__init__(self, *args, **kwargs)

        # Leaves
        # gRPC port 50001
        s1 = self.addSwitch('s1', cls=StratumBmv2Switch, cpuport=CPU_PORT)
        s2 = self.addSwitch('s2', cls=StratumBmv2Switch, cpuport=CPU_PORT)
        s3 = self.addSwitch('s3', cls=StratumBmv2Switch, cpuport=CPU_PORT)
        s4 = self.addSwitch('s4', cls=StratumBmv2Switch, cpuport=CPU_PORT)

        self.addLink(s1, s2)
        self.addLink(s1, s3)
        self.addLink(s1, s4)
        self.addLink(s2, s4)


        # IPv4 hosts attached to leaf 1
        h1 = self.addHost('h1', cls=IPv4Host, mac="11:11:11:00:00:1A",
                           ip='192.168.1.1/24', gw='192.168.1.254')
        h2 = self.addHost('h2', cls=IPv4Host, mac="11:11:11:00:00:1B",
                           ip='192.168.1.2/24', gw='192.168.1.254')
        h3 = self.addHost('h3', cls=IPv4Host, mac="22:22:22:00:00:2A",
                           ip='192.168.2.1/24', gw='192.168.2.254')
        h4 = self.addHost('h4', cls=IPv4Host, mac="22:22.22:00:00:2B",
                           ip='192.168.2.2/24', gw='192.168.2.254')
        h5 = self.addHost('h5', cls=IPv4Host, mac="33:33:33:00:00:3A",
                           ip='192.168.3.1/24', gw='192.168.3.254')
        h6 = self.addHost('h6', cls=IPv4Host, mac="33:33:33:00:00:3B",
                           ip='192.168.3.2/24', gw='192.168.3.254')
        h7 = self.addHost('h7', cls=IPv4Host, mac="44:44.44:00:00:4A",
                           ip='192.168.4.1/24', gw='192.168.4.254')
        h8 = self.addHost('h8', cls=IPv4Host, mac="44:44:44:00:00:4B",
                           ip='192.168.4.2/24', gw='192.168.4.254')

        self.addLink(h1, s1)
        self.addLink(h2, s1)
        self.addLink(h3, s2)
        self.addLink(h4, s2)
        self.addLink(h5, s3)
        self.addLink(h6, s3)
        self.addLink(h7, s4)
        self.addLink(h8, s4)

        #self.addIntf(s1,'out_intf')

def main():
    net = Mininet(topo=TutorialTopo(), controller=None)

    h1 = net.get('h1')
    h2 = net.get('h2')
    h3 = net.get('h3')
    h4 = net.get('h4')
    h5 = net.get('h5')
    h6 = net.get('h6')
    h7 = net.get('h7')
    h8 = net.get('h8')

    h1.cmd("arp -i h1-eth0 -s 192.168.1.254 00:00:00:11:11:11")
    h2.cmd("arp -i h2-eth0 -s 192.168.1.254 00:00:00:11:11:11")
    h3.cmd("arp -i h3-eth0 -s 192.168.2.254 00:00:00:11:11:11")
    h4.cmd("arp -i h4-eth0 -s 192.168.2.254 00:00:00:11:11:11")
    h5.cmd("arp -i h5-eth0 -s 192.168.3.254 00:00:00:11:11:11")
    h6.cmd("arp -i h6-eth0 -s 192.168.3.254 00:00:00:11:11:11")
    h7.cmd("arp -i h7-eth0 -s 192.168.4.254 00:00:00:11:11:11")
    h8.cmd("arp -i h8-eth0 -s 192.168.4.254 00:00:00:11:11:11")

    #root = Node( 'root', inNamespace=False )
    #intf = net.addLink( root, net['s1']).intf1
    #root.setMAC("11:22:33:00:00:01", intf)

    net.start()
    CLI(net)
    net.stop()
    print '#' * 80
    print 'ATTENTION: Mininet was stopped! Perhaps accidentally?'
    print 'No worries, it will restart automatically in a few seconds...'
    print 'To access again the Mininet CLI, use `make mn-cli`'
    print 'To detach from the CLI (without stopping), press Ctrl-D'
    print 'To permanently quit Mininet, use `make stop`'
    print '#' * 80


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description='Mininet topology script to add hosts and a single switch')
    args = parser.parse_args()
    setLogLevel('info')

    main()
