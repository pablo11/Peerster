package model

import (
    "encoding/binary"
    "crypto/sha256"
    "encoding/hex"
)

type VotationStatement struct{
	Question	string
	Origin		string
}

func (*VotationStatement vs) Hash() string{
	sha_256 := sha256.New()
	sha_256.Write([]byte(vs.Question))
	sha_256.Write([]byte(vs.Origin))
	return hex.EncodeToString(sha_256.Sum(nil))
}

func (*VotationStatement vs) Copy() VotationStatement {
	new_vs := VotationStatement{
		Question: vs.Question,
		Origin:	vs.Origin,
	}
	
	return new_vs
} 