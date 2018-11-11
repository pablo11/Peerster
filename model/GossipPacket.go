package model

type GossipPacket struct {
    Simple *SimpleMessage
    Rumor *RumorMessage
    Status *StatusPacket
    Private *PrivateMessage
    DataRequest *DataRequest
    DataReply *DataReply
}

func (gp *GossipPacket) String(mode, relayAddr string) string {
    switch {
        case gp.Simple != nil:
            return gp.Simple.String(mode)

        case gp.Rumor != nil:
            return gp.Rumor.String(mode, relayAddr)

        case gp.Status != nil:
            return gp.Status.String(relayAddr)

        case gp.Private != nil:
            return gp.Private.String()

        default:
            return ""
    }
}
