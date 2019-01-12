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
		AssetName: assetName,
	}
	
	sign := g.SignVotingStatement(&vs)
	
	tx := model.Transaction{
		VotationStatement:	&vs,
		Signature: sign,
	}
	
	isValid, errorMsg := g.Blockchain.isValidTx(&tx)
	if !isValid {
        fmt.Println("Discarding Tx: " + errorMsg)
        return
    }
	g.Blockchain.SendTxPublish(&tx)
	
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
	}
	
	g.QuestionKeyMutex.Lock()
	key, ok := g.QuestionKey[votation_id] //Get key received in private message
	g.QuestionKeyMutex.Unlock()
	
	if !ok {
		log.Fatal("Fail to retreive the key to answer to this question")
		return
	}
	
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
		AssetName: question.AssetName,
		Replier: g.Name, 
	}
	
	sign := g.SignVotationAnswerWrapper(&vaw)
	
	tx := model.Transaction{
		VotationAnswerWrapper:	&vaw,
		Signature: sign,
	}
	
	//Send SendFileTx
	isValid, errorMsg := g.Blockchain.isValidTx(&tx)
	if !isValid {
        fmt.Println("Discarding Tx: " + errorMsg)
        return
    }
	g.Blockchain.SendTxPublish(&tx)
	
	//move from pending to completed? => This is done in GUI
}

func (g *Gossiper) sendKeyToAllPeers(peers []string , key string, questionId string){	

	for _,p := range peers{
		pm := model.NewPrivateMessage(g.Name, createPMWithKey(key,questionId), p)
		
		//ENCRYPT PRIVATE !!
		g.SendPrivateMessage(pm)
	}
	
}