package gossip

import (
    "github.com/pablo11/Peerster/model"
)

func (g *Gossiper) HandlePktRumor(gp *model.GossipPacket, fromAddrStr string) {
    isRouteRumor := gp.Rumor.Text == ""
    if !isRouteRumor {
        g.printGossipPacket("received", fromAddrStr, gp)
    }

    g.updateRoutingTable(gp.Rumor, fromAddrStr)

    // If the message is the next one expected, store it
    if gp.Rumor.ID == g.getVectorClock(gp.Rumor.Origin) {
        g.incrementVectorClock(gp.Rumor.Origin)
        g.storeMessage(gp.Rumor, !isRouteRumor)
        g.sendRumorMessage(gp.Rumor, true, fromAddrStr)
    }

    // Send status message to the peer the rumor message was received from
    g.sendStatusMessage(fromAddrStr)
}
