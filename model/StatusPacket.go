package model

import (
    "strconv"
)

type StatusPacket struct {
    Want []PeerStatus
}

// Vector clock for peer "Identifier"
type PeerStatus struct {
    Identifier string
    NextID uint32
}

func (sp *StatusPacket) String(relayAddr string) string {
    strPeersStatus := ""
    for i := 0; i < len(sp.Want); i++ {
        nextIdStr := strconv.FormatUint(uint64(sp.Want[i].NextID), 10)
        strPeersStatus += " peer " + sp.Want[i].Identifier + " nextID " + nextIdStr
    }

    return "STATUS from " + relayAddr + strPeersStatus
}
