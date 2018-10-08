package model

type SimpleMessage struct {
    OriginalName string
    RelayPeerAddr string
    Contents string
}

func (sm *SimpleMessage) String(mode string) string {
    switch mode {
    case "client":
        return "CLIENT MESSAGE " + sm.Contents

    case "peer":
        return "SIMPLE MESSAGE origin " + sm.OriginalName + " from " + sm.RelayPeerAddr + " contents " + sm.Contents
    }

    return ""
}
