package gossip

import (
    "fmt"
    "github.com/pablo11/Peerster/model"
)

func (g *Gossiper) HandlePktStatus(gp *model.GossipPacket, fromAddrStr string) {
    if (!DEBUG) {
        g.printGossipPacket("", fromAddrStr, gp)
    }

    g.compareVectorClocks(gp.Status, fromAddrStr)
}

func (g *Gossiper) compareVectorClocks(sp *model.StatusPacket, fromAddr string) {
    // Prepare g.status to compare vactors clocks
    tmpStatus := make(map[string]bool)
    g.statusMutex.Lock()
    for key, _ := range g.status {
        tmpStatus[key] = false
    }
    g.statusMutex.Unlock()

    // Compare the two vector clocks
    for i := 0; i < len(sp.Want); i++ {
        otherStatusPeer := sp.Want[i]

        g.statusMutex.Lock()
        statusPeer, exists := g.status[otherStatusPeer.Identifier]
        g.statusMutex.Unlock()
        if exists {
            tmpStatus[otherStatusPeer.Identifier] = true
            if otherStatusPeer.NextID > statusPeer.NextID {
                // The other peer has something more, so send StatusPacket
                g.sendStatusMessage(fromAddr)

                // Don't flip the coin and stop timer
                g.getChannelForPeer(fromAddr) <- false
                return
            } else if otherStatusPeer.NextID < statusPeer.NextID && otherStatusPeer.NextID > 0 {
                // The gossiper has something more, so send rumor of this thing
                //g.messagesMutex.Lock()
                rm := g.messages[otherStatusPeer.Identifier][otherStatusPeer.NextID - 1]
                //g.messagesMutex.Unlock()
                g.sendRumorMessage(rm, false, fromAddr)

                // Don't flip the coin and stop timer
                g.getChannelForPeer(fromAddr) <- false
                return
            }
        } else {
            // The other peer has something more, so send status
            g.sendStatusMessage(fromAddr)

            // Don't flip the coin and stop timer
            g.getChannelForPeer(fromAddr) <- false
            return
        }
    }

    g.statusMutex.Lock()
    defer g.statusMutex.Unlock()
    if len(sp.Want) == len(g.status) {
        // The two vectors are the same -> we are in sync with the peer
        if (!DEBUG) {
            fmt.Println("IN SYNC WITH " + fromAddr)
            fmt.Println()
        }

        // Flip the coin and stop timer
        g.getChannelForPeer(fromAddr) <- true
        return
    } else {
        // The peer vector cannot be longer than the gossiper vector clock, otherwise we don't get here
        // Find the first message from tmpStatus to send
        g.messagesMutex.Lock()
        defer g.messagesMutex.Unlock()
        for key, isVisited := range tmpStatus {
            if !isVisited && len(g.messages[key]) > 0 {
                rm := g.messages[key][0]
                g.sendRumorMessage(rm, false, fromAddr)

                // Don't flip the coin and stop timer
                g.getChannelForPeer(fromAddr) <- false
                return
            }
        }
    }
}
