package model

import (
    "encoding/binary"
    "crypto/sha256"
    "encoding/hex"
	"strconv"
)

//VERIFY NAMES AFTER RIC PUSH

type VotationAnswerWrapper struct{
	Answer 		[]byte
	Question	string
	Origin		string
}

func (*VotationAnswerWrapper vaw) Hash() string{
	sha_256 := sha256.New()
	sha_256.Write(vaw.Answer)
	sha_256.Write([]byte(vaw.Question))
	sha_256.Write([]byte(vaw.Origin))
	
	return hex.EncodeToString(sha_256.Sum(nil))
}

func (*VotationAnswerWrapper vaw) Copy() VotationAnswerWrapper {
	new_answer := make([]byte, len(vaw.Answer))
	copy(new_answer,vaw.Answer)
	
	new_vaw := VotationAnswerWrapper{
		Answer: new_answer,
		Question: vaw.Question,
		Origin:	vaw.Origin,
	}
	
	return new_vaw
}

type VotationAnswer struct{
	Answer		bool
	Replier		string
}

func (*VotationAnswer va) Hash() string{
	sha_256 := sha256.New()
	ans := byte{}
	sha_256.Write([]byte(strconv.FormatBool(va.Answer)))
	sha_256.Write([]byte(va.Replier))
	
	return hex.EncodeToString(sha_256.Sum(nil))
}

func (*VotationAnswer va) Copy() VotationAnswer{
	new_va := VotationAnswer{
		Answer: va.Answer,
		Replier: va.Replier,
	}
	
	return new_va
}

