package model

type RumorMessage struct {
    Origin string
    ID uint32
    Text string
}

func (rm *RumorMessage) String(mode, relayAddr string) string {
    switch mode {
    case "received":
        return "RUMOR origin " + rm.Origin + " from " + relayAddr + " ID " + string(rm.ID) + " contents " + rm.Text

    case "mongering":
        return "MONGERING with " + relayAddr

    default:
        return ""
    }
}
