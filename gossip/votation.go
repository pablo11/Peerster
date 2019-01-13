package gossip

import (
    "encoding/hex"
	"fmt"
	"math/rand"
	"github.com/pablo11/Peerster/model"
	//"github.com/pablo11/Peerster/util/debug"
)

func (g *Gossiper) LaunchVotation(question string, assetName string){
	//Create and put TxVotationStatement in pending Blocks
	//Send symmetric key to all peers

	//debug.Debug("Launching votating")
	vs := model.VotationStatement{
		Question: question,
		Origin:	g.Name,
		AssetName: assetName,
	}

    data := vs.Hash()
	sign := g.Sign(data[:])

	tx := model.Transaction{
		VotationStatement:	&vs,
		Signature: sign,
	}

	//debug.Debug("Checking votating correctness")
	isValid, errorMsg := g.Blockchain.isValidTx(&tx)
	if !isValid {
        fmt.Println("Discarding Tx: " + errorMsg)
        return
    }


    txCopy := tx.Copy()

	g.Blockchain.SendTxPublish(&txCopy)

}

func (g *Gossiper) AnswerVotation(question_subject string, assetName string, origin string, answer bool){
	//Get question corresponding to votation_id
	//debug.Debug("Answering vote question")

	votation_id := model.GetVotationId(question_subject,assetName,origin)

	g.Blockchain.VoteStatementMutex.Lock()
	question, questionExist := g.Blockchain.VoteStatement[votation_id]
	g.Blockchain.VoteStatementMutex.Unlock()

	if !questionExist{
		fmt.Println("❌ The question you'are trying to answer does not exist")
		return
	}

	va := model.VotationAnswer{
		Answer: answer,
	}

	g.QuestionKeyMutex.Lock()
	key, ok := g.QuestionKey[votation_id] //Get key received in private message
	g.QuestionKeyMutex.Unlock()

	if !ok {
		fmt.Println("Fail to retreive the key to answer to this question")
		return
	}

	key_byte, err := hex.DecodeString(key)
	if err != nil{
		fmt.Println("Cannot decode key")
		return
	}

	//Encrypt va
	va_enc, err := va.Encrypt(key_byte)

	if err != nil {
		fmt.Println("Error during symmetric encryption")
		return
	}

	vaw := model.VotationAnswerWrapper{
		Answer: va_enc,
		Question: question.Question,
		Origin: question.Origin,
		AssetName: question.AssetName,
		Replier: g.Name,
	}

    data := vaw.Hash()
	sign := g.Sign(data[:])

	tx := model.Transaction{
		VotationAnswerWrapper:	&vaw,
		Signature: sign,
	}

	//Send SendFileTx
	//debug.Debug("Checking correctness of vote answer")
	isValid, errorMsg := g.Blockchain.isValidTx(&tx)
	if !isValid {
        fmt.Println("Discarding Tx: " + errorMsg)
        return
    }
	g.Blockchain.SendTxPublish(&tx)

	//move from pending to completed? => This is done in GUI
}


func (g *Gossiper) sendKeyToAllPeers(questionId string, assetName string){

	key := make([]byte, 32)
	rand.Read(key)

	key_str := hex.EncodeToString(key)
	g.QuestionKeyMutex.Lock()
	g.QuestionKey[questionId] = key_str
	g.QuestionKeyMutex.Unlock()
	
	var peers []string
	//Send to all shareholders
	g.Blockchain.AssetsMutex.Lock()
	for p,_ := range g.Blockchain.Assets[assetName]{
		peers = append(peers, p) //Assume that peer with asset 0 have been removed
	}
	g.Blockchain.AssetsMutex.Unlock()

	for _,p := range peers{
		if p != g.Name {
			pm := g.NewEncryptedPrivateMessage(g.Name, createPMWithKey(key_str,questionId), p)
            g.SignPrivateMessage(pm)
			g.SendPrivateMessage(pm)
		}
	}
	//debug.Debug("Sending symmetric to all peers -> OK")
}
