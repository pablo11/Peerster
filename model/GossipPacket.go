package model

type GossipPacket struct {
    Simple *SimpleMessage
}

func (gp *GossipPacket) String(mode string) string {
    return gp.Simple.String(mode)
}
