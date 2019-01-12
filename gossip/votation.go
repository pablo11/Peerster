package gossip

import (
    "encoding/hex"
	"log"
	"fmt"
	"math/rand"
	"github.com/pablo11/Peerster/model"
)

func (g *Gossiper) launchVotation(question string, assetName string){
	//Create and put TxVotationStatement in pending Blocks
	//Send symmetric key to all peers 
		//What would be the message kind?
	
	vs := model.VotationStatement{
		Question: question,
		Origin:	g.Name,
	}
	
	sign := g.SignVotingStatement(&vs)
	
	tx := model.Transaction{
		VotationStatement:	&vs,
		Signature: sign,
	}
	
	g.Blockchain.addTxToPool(&tx) // DOES THIS SEND TXS TO ALL??
	
	key := make([]byte, 32)
	rand.Read(key)
	
	key_str := hex.EncodeToString(key)
	g.QuestionKeyMutex.Lock()
	g.QuestionKey[vs.GetId()] = key_str
	g.QuestionKeyMutex.Unlock()
	
	var peers []string
	//Send to all shareholders
	g.Blockchain.assetsMutex.Lock()
	for p,_ := range g.Blockchain.assets[assetName]{
		peers = append(peers, p) //Assume that peer with asset 0 have been removed
	}
	g.Blockchain.assetsMutex.Unlock()
	
	g.sendKeyToAllPeers(peers,key_str,vs.GetId())
}

func (g *Gossiper) answerVotation(votation_id string, answer bool){
	//Get question corresponding to votation_id
	
	g.Blockchain.voteStatementMutex.Lock()
	question := g.Blockchain.voteStatement[votation_id]
	g.Blockchain.voteStatementMutex.Unlock()
	
	
	va := model.VotationAnswer{
		Answer: answer,
		Replier: g.Name, 
	}
	
	g.QuestionKeyMutex.Lock()
	key := g.QuestionKey[votation_id] //Get key received in private message
	g.QuestionKeyMutex.Unlock()
	
	key_byte, err := hex.DecodeString(key)
	if err != nil{
		log.Fatal("Cannot decode key")
		return
	}
	
	//Encrypt va
	va_enc, err := va.Encrypt(key_byte)
	
	if err != nil {
		fmt.Println("Error during symmetric encryption")
		log.Fatal(err)
	}
	
	vaw := model.VotationAnswerWrapper{
		Answer: va_enc,
		Question: question.Question,
		Origin: question.Origin,
	}
	
	sign := g.SignVotationAnswerWrapper(&vaw)
	
	tx := model.Transaction{
		VotationAnswerWrapper:	&vaw,
		Signature: sign,
	}
	
	g.Blockchain.addTxToPool(&tx) // DOES THIS SEND TXS TO ALL??
	
	//move from pending to completed?
}

func (g *Gossiper) sendKeyToAllPeers(peers []string , key string, questionId string){	

	for _,p := range peers{
		//HAS TO BE CHANGED for that pm := NewPrivateMessage(g.Name, createPMWithKey(key,questionId), p)
		pm := &model.PrivateMessage{
			Origin: g.Name,
			ID: 0,
			Text: createPMWithKey(key,questionId),
			Destination: p,
			HopLimit: 10,
		}
		
		//ENCRYPT PRIVATE !!
		g.SendPrivateMessage(pm)
	}
	
}