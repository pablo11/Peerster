package gossip

import (
    "encoding/hex"
	"fmt"
	"math/rand"
	"github.com/pablo11/Peerster/model"
)

/*
Function to start a new vote 
question: the subject of the new vote
assetName: the asset on which question is asked
*/
func (g *Gossiper) LaunchVotation(question string, assetName string){
	//Create, validate and put TxVotationStatement in pending Blocks
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

	isValid, errorMsg := g.Blockchain.isValidTx(&tx)
	if !isValid {
        fmt.Println("Discarding Tx: " + errorMsg)
        return
    }


    txCopy := tx.Copy()

	g.Blockchain.SendTxPublish(&txCopy)

}

/*
Function to answer a vote
question_subject: the question that we are trying to answer
assetName: the name of the asset on which question was stated
origin: the name of the originator node of the question
answer: our answer to the question
*/
func (g *Gossiper) AnswerVotation(question_subject string, assetName string, origin string, answer bool){
	//Create, encrypt, validate and put TxAnswerVotation in pending Blocks


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

	isValid, errorMsg := g.Blockchain.isValidTx(&tx)
	if !isValid {
        fmt.Println("Discarding Tx: " + errorMsg)
        return
    }
	g.Blockchain.SendTxPublish(&tx)

}

/*
Function to send private key to answer a question to all shareholders
questionId: we send a symmetric key to answer the question identified by questionId
assetName: the name of the asset on which the question is
*/
func (g *Gossiper) sendKeyToAllPeers(questionId string, assetName string){

	key := make([]byte, 32)
	rand.Read(key)

	key_str := hex.EncodeToString(key)
	g.QuestionKeyMutex.Lock()
	g.QuestionKey[questionId] = key_str
	g.QuestionKeyMutex.Unlock()
	
	var peers []string
	//Retreive all shareholders
	g.Blockchain.AssetsMutex.Lock()
	for p,_ := range g.Blockchain.Assets[assetName]{
		//Assume that peer with asset 0 have been removed
		peers = append(peers, p) 
	}
	g.Blockchain.AssetsMutex.Unlock()

	for _,p := range peers{
		if p != g.Name {
			//Encrypt, sign and send
			pm := g.NewEncryptedPrivateMessage(g.Name, createPMWithKey(key_str,questionId), p)
            g.SignPrivateMessage(pm)
			g.SendPrivateMessage(pm)
		}
	}
	
}
