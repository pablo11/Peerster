package model

type ClientMessage struct {
    Type string
    Text string
    Dest string
    File string
    Request string
}

func (cm *ClientMessage) String() string {
    return "CLIENT MESSAGE " + cm.Text
}
