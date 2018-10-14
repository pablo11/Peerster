package model

type GossipPacket struct {
    Simple *SimpleMessage
    Rumor *RumorMessage
    Status *StatusPacket
}

func (gp *GossipPacket) String(mode string) string {
    if gp.Simple == nil {
        return ""
    }
    return gp.Simple.String(mode)
}
