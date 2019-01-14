package model

import (
	"crypto/sha256"
	"encoding/hex"
)

/*
Structure include in a transaction used to start vote
Question: The subject of the question
Origin: The originator of the question
AssetName: Name of the asset on which voting in launch
*/
type VotationStatement struct {
	Question  string
	Origin    string
	AssetName string
}

func (vs *VotationStatement) Hash() (out [32]byte) {
	sha_256 := sha256.New()
	sha_256.Write([]byte(vs.Question))
	sha_256.Write([]byte(vs.Origin))
	sha_256.Write([]byte(vs.AssetName))
	copy(out[:], sha_256.Sum(nil))
	return
}

func (vs *VotationStatement) Copy() VotationStatement {
	return VotationStatement{
		Question:  vs.Question,
		Origin:    vs.Origin,
		AssetName: vs.AssetName,
	}
}

func (vs *VotationStatement) String() string {
	return "VOTATION_STATEMENT= FROM " + vs.Origin + " QUESTION " + vs.Question + " FOR ASSET " + vs.AssetName
}

func (vs *VotationStatement) GetId() string {
	hash := vs.Hash()
	return hex.EncodeToString(hash[:])
}

func GetVotationId(question string, assetName string, origin string) string {
	new_vs := &VotationStatement{
		Question:  question,
		Origin:    origin,
		AssetName: assetName,
	}
	hash := new_vs.Hash()
	return hex.EncodeToString(hash[:])
}
