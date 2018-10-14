package model

type StatusPacket struct {
    Want []PeerStatus
}

// Vector clock for peer "Identifier"
type PeerStatus struct {
    Identifier string
    NextID uint32
}
