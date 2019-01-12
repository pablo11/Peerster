package model

import (
    "crypto/sha256"
    "encoding/hex"
	"strconv"
	"crypto/rand"
	"crypto/aes"
    "crypto/cipher"
    "github.com/dedis/protobuf"
	"io"
	"errors"
)

//======================VOTATION ANSWER WRAPPED=========================

type VotationAnswerWrapper struct{
	Answer 		[]byte
	Question	string
	Origin		string
	AssetName	string
	Replier		string
}

func (vaw *VotationAnswerWrapper) Hash() []byte{
	sha_256 := sha256.New()
	sha_256.Write(vaw.Answer)
	sha_256.Write([]byte(vaw.Question))
	sha_256.Write([]byte(vaw.Origin))
	sha_256.Write([]byte(vaw.AssetName))
	sha_256.Write([]byte(vaw.Replier))
	
	return sha_256.Sum(nil)
}

func (vaw *VotationAnswerWrapper) Copy() VotationAnswerWrapper {
	new_answer := make([]byte, len(vaw.Answer))
	copy(new_answer,vaw.Answer)
	
	new_vaw := VotationAnswerWrapper{
		Answer: new_answer,
		Question: vaw.Question,
		Origin:	vaw.Origin,
		AssetName: vaw.AssetName,
		Replier: vaw.Replier,
	}
	
	return new_vaw
}

func (vaw *VotationAnswerWrapper) String() string {
    return "VOTATION_ANSWER_WRAPPED= FROM " + vaw.Origin +" QUESTION "+vaw.Question +" ASSET "+vaw.AssetName+" REPLIED BY "+vaw.Replier
}

func (vaw *VotationAnswerWrapper) GetVotationId() string{
	new_vs := &VotationStatement{
		Question: vaw.Question,
		Origin:	vaw.Origin,
		AssetName: vaw.AssetName,
	}
	return hex.EncodeToString(new_vs.Hash())
}

//=========================VOTATION ANSWER================================
type VotationAnswer struct{
	Answer		bool
}

func (va *VotationAnswer) Hash() string{
	sha_256 := sha256.New()
	sha_256.Write([]byte(strconv.FormatBool(va.Answer)))
	
	return hex.EncodeToString(sha_256.Sum(nil))
}

func (va *VotationAnswer) Copy() VotationAnswer{
	new_va := VotationAnswer{
		Answer: va.Answer,
	}
	
	return new_va
}

func (va *VotationAnswer) String() string {
    return "VOTATION_ANSWER= "+strconv.FormatBool(va.Answer)
}


func (va *VotationAnswer) Encrypt(key []byte) ([]byte, error) {

	va_encoded, err := protobuf.Encode(va) // Do we need to copy here?
	

    c, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(c)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    return gcm.Seal(nonce, nonce, va_encoded, nil), nil
}

func decrypt(ciphertext []byte, key []byte) (VotationAnswer, error) {
	var va_decoded VotationAnswer
	
    c, err := aes.NewCipher(key)
    if err != nil {
        return va_decoded, err
    }

    gcm, err := cipher.NewGCM(c)
    if err != nil {
        return va_decoded, err
    }

	//THIS COULD PROBABLY FAIL BE CAREFUL !
    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return va_decoded, errors.New("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    
	byte_decoded, err := gcm.Open(nil, nonce, ciphertext, nil)
	
	if err != nil {
		return va_decoded, err
	}
	
	va_decoded = VotationAnswer{}
    err = protobuf.Decode(byte_decoded, &va_decoded)
	
	return va_decoded,err
}
