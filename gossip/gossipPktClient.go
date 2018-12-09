package gossip

import (
    "fmt"
    "github.com/pablo11/Peerster/model"
)

func (g *Gossiper) HandlePktClient(cm *model.ClientMessage) {
    switch cm.Type {
        case "msg":
            fmt.Println(cm.String())
            fmt.Println()

            if cm.Dest == "" {
                go g.SendPublicMessage(cm.Text, true)
            } else {
                pm := model.NewPrivateMessage(g.Name, cm.Text, cm.Dest)
                go g.SendPrivateMessage(pm)
            }

        case "indexFile":
            go g.FileSharing.IndexFile(cm.File)

        case "downloadFile":
            go g.FileSharing.RequestFile(cm.File, cm.Dest, cm.Request)

        case "searchFile":
            if cm.Budget == 0 {
                go g.StartSearchRequest(2, cm.Keywords, true)
            } else {
                go g.StartSearchRequest(cm.Budget, cm.Keywords, false)
            }

        default:
            fmt.Println("WARNING: Unoknown client message type")
    }
}
