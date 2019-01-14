package gossip

import (
	"fmt"
	"regexp"
    "github.com/pablo11/Peerster/model"
	"github.com/pablo11/Peerster/util/debug"
)

func (g *Gossiper) HandlePktPrivate(gp *model.GossipPacket, fromAddrStr string) {
	if gp.Private.Destination == g.Name {

		if gp.Private.Signature != nil {
			if !g.Blockchain.VerifyPrivateMessage(gp.Private) {
				fmt.Printf("Private message forged -> Refused")
				return
			}
		}

		// If the package is encrypted, decrypt it
		if gp.Private.IsEncrypted {
			debug.Debug("Receiving a message encrypted")

			g.DecryptPrivateMessage(gp.Private)
		} else {
			debug.Debug("Receiving a message in plain text")
		}

		// If the private message is for this node, display it
		g.printGossipPacket("", fromAddrStr, gp)

		if checkPMWithKey(gp.Private.Text) {
			key, question_id := getKeyFromPM(gp.Private.Text)

			g.Blockchain.VoteStatementMutex.Lock()
			question, questionExist := g.Blockchain.VoteStatement[question_id]
			g.Blockchain.VoteStatementMutex.Unlock()

			if questionExist && question.Origin == gp.Private.Origin {

				g.QuestionKeyMutex.Lock()
				g.QuestionKey[question_id] = key
				g.QuestionKeyMutex.Unlock()
			}
		}

	} else {
		// Forward the message and decrease the HopLimit
		pm := gp.Private
		fmt.Println("Forwarding private msg dest " + pm.Destination)
		if pm.HopLimit > 1 {
			pm.HopLimit -= 1
			g.SendPrivateMessage(pm)
		}
	}
}

func checkPMWithKey(msg string) bool {
	re := regexp.MustCompile("VOTATION KEY:.{64} QUESTION ID:.{64}")
	return re.MatchString(msg)
}

func getKeyFromPM(msg string) (string, string) {
	s1 := "VOTATION KEY:"
	key := msg[len(s1) : len(s1) + 64]
	questionId := msg[len(msg) - 64 : len(msg)]
	return key, questionId
}

func createPMWithKey(key string, questionId string) string {
	return "VOTATION KEY:" + key + " QUESTION ID:" + questionId
}
