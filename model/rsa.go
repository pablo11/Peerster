package model

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

type Signature struct {
	Name      string
	BitString []byte
}

func NewPrivateKey() *rsa.PrivateKey {
	rng := rand.Reader
	privateKey, err := rsa.GenerateKey(rng, 2048)
	if err != nil {
		fmt.Printf("Bad private key: %v\n", err)
		return nil
	}

	return privateKey
}

func (s *Signature) Copy() Signature {
	newSignature := Signature{
		Name:      s.Name,
		BitString: make([]byte, len(s.BitString)),
	}

	copy(newSignature.BitString[:], s.BitString[:])
	return newSignature
}

func (s *Signature) PrintSignature() {
	fmt.Printf("🔏 Signature: Name=%v Hash(Bitstring)=%v\n", s.Name, hex.EncodeToString(s.BitString))
}

func PrivateKeyString(privateKey *rsa.PrivateKey) string {
	privateKeyByte := x509.MarshalPKCS1PrivateKey(privateKey)
	sha256PrivateKeyByte := sha256.Sum256(privateKeyByte)
	return hex.EncodeToString(sha256PrivateKeyByte[:])
}

func PublicKeyString(publicKey *rsa.PublicKey) string {
	publicKeyByte := x509.MarshalPKCS1PublicKey(publicKey)
	sha256PublicKeyByte := sha256.Sum256(publicKeyByte)
	return hex.EncodeToString(sha256PublicKeyByte[:])
}
