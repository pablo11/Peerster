package model

type ClientMessage struct {
    Text string
    Dest string
}

func (cm *ClientMessage) String() string {
    return "CLIENT MESSAGE " + cm.Text
}
