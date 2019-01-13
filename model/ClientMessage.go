package model

type ClientMessage struct {
    Type string
    Text string
    Dest string
    File string
    Request string
    Keywords []string
    Budget uint64
    Identity string
    Asset string
    Amount uint64
    Sign bool
    Encrypt bool
}

func (cm *ClientMessage) String() string {
    return "CLIENT MESSAGE " + cm.Text
}
