package gossip

import (
    "encoding/binary"
    "crypto/sha256"
    "encoding/hex"
)

func (*Gossiper g) launchVotation(question string){
	//Create and put TxVotationStatement in pending Blocks
	//Send symmetric key to all peers 
		//What would be the message kind?
	
	tx_vs := model.VotationStatement{
		Question: question,
		Origin:	g.Name,
	}
	
	//Put in pending blocks
	
}

func (*Gossiper g) answerVotation(votation_id string, answer bool){
	//Get question corresponding to votation_id
	
	va := model.VotationAnswer{
		Answer: answer,
		Replier: g.Name, 
	}
	
	//Encrypt va
	//Sign va
	
	vaw := model.VotationAnswerWrapped{
		Answer: ,
		Question: ,
		Origin: ,
		Signature: ,
	}
	
	//put in pending blocks
	
	//move from pending to completed?
}