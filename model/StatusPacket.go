package model

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
        strPeersStatus += " peer " + sp.Want[i].Identifier + " nextID " + string(sp.Want[i].NextID)
    }

    return "STATUS from " + relayAddr + strPeersStatus
}
