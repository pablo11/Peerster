package gossip

import (
    "fmt"
    "github.com/pablo11/Peerster/model"
)

func (g *Gossiper) HandlePktPrivate(gp *model.GossipPacket, fromAddrStr string) {
    if gp.Private.Destination == g.Name {
        // If the private message is for this node, display it
        g.printGossipPacket("", fromAddrStr, gp)
    } else {
        // Forward the message and decrease the HopLimit
        pm := gp.Private
        fmt.Println("Forwarding private msg dest " + pm.Destination)
        if pm.HopLimit > 1 {
            pm.HopLimit -= 1
            g.SendPrivateMessage(pm)
        }
    }
}
