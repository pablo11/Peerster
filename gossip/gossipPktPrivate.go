package gossip

import (
    "fmt"
    "github.com/pablo11/Peerster/model"
	"github.com/pablo11/Peerster/util/debug"
	"regexp"
)

func (g *Gossiper) HandlePktPrivate(gp *model.GossipPacket, fromAddrStr string) {
    if gp.Private.Destination == g.Name {
        // If the private message is for this node, display it
        g.printGossipPacket("", fromAddrStr, gp)
		
		if checkPMWithKey(gp.Private.Text) {

			key, question_id := getKeyFromPM(gp.Private.Text)
			g.QuestionKeyMutex.Lock()
			g.QuestionKey[question_id] = key
			g.QuestionKeyMutex.Unlock()
			debug.Debug("Received a symmetric key for question "+question_id)
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

func checkPMWithKey (msg string) bool{
	re := regexp.MustCompile("VOTATION KEY:.{32} QUESTION ID:.{32}")
	if re.MatchString(msg) {
		return true
	}
	return false
}

func getKeyFromPM (msg string) (string,string){
	s1 := "VOTATION KEY:"
	key := msg[len(s1):len(s1)+32]
	questionId := msg[len(msg)-32:len(msg)]
	return key,questionId
}

func createPMWithKey(key string, questionId string) string{
	return "VOTATION KEY:"+key+" QUESTION ID:"+questionId
}