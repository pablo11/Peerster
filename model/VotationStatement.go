package model

import (
    "crypto/sha256"
    "encoding/hex"
)

type VotationStatement struct{
	Question	string
	Origin		string
	AssetName	string
}

func (vs *VotationStatement) Hash() []byte{
	sha_256 := sha256.New()
	sha_256.Write([]byte(vs.Question))
	sha_256.Write([]byte(vs.Origin))
	sha_256.Write([]byte(vs.AssetName))
	return sha_256.Sum(nil)
}

func (vs *VotationStatement) Copy() VotationStatement {
	new_vs := VotationStatement{
		Question: vs.Question,
		Origin:	vs.Origin,
		AssetName: vs.AssetName,
	}
	
	return new_vs
} 

func (vs *VotationStatement) String() string {
    return "VOTATION_STATEMENT= FROM " + vs.Origin +" QUESTION "+vs.Question +" FOR ASSET "+vs.AssetName
}

func (vs *VotationStatement) GetId() string{
	return hex.EncodeToString(vs.Hash())
}

func GetVotationId(question string, assetName string, origin string) string{
	new_vs := &VotationStatement{
		Question: question,
		Origin:	origin,
		AssetName: assetName,
	}
	return hex.EncodeToString(new_vs.Hash())
}