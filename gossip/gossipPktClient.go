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
			var pm *model.PrivateMessage
			text := cm.Text

			if cm.Encrypt {
				pm = g.NewEncryptedPrivateMessage(g.Name, text, cm.Dest)
			} else {
				pm = model.NewPrivateMessage(g.Name, text, cm.Dest)
			}

			if pm != nil {
				go g.SendPrivateMessage(pm)
			}
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

	case "identity":
		go g.Blockchain.SendIdentityTx(cm.Identity)

	case "shareTx":
		go g.Blockchain.SendShareTx(cm.Asset, cm.Dest, cm.Amount)

	case "vote":
		go g.LaunchVotation(cm.Text, cm.Asset)

	case "voteAnswer":
		go g.AnswerVotation(cm.Text, cm.Asset, cm.Origin, cm.Answer)

	default:
		fmt.Println("WARNING: Unoknown client message type")
	}
}
