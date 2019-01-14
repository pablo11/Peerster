package gossip

import (
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/collections"
)

func (g *Gossiper) HandlePktSimple(gp *model.GossipPacket) {
    g.printGossipPacket("peer", "", gp)

    // Change the relay peer field to this node address
    receivedFrom := gp.Simple.RelayPeerAddr
    gp.Simple.RelayPeerAddr = g.address.String()

    // Broadcast the message to every peer except the one the message was received from
    go g.sendGossipPacket(gp, collections.Filter(g.peers, func(p string) bool {
        return p != receivedFrom
    }))
}
