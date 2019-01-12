package model

import (
    "fmt"
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha256"
    "crypto/x509"
    "encoding/hex"
)

type Signature struct {
    Name string
    Signature []byte
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
        Name: s.Name,
        Signature: make([]byte, len(s.Signature)),
    }

    copy(newSignature.Signature[:], s.Signature[:])

    return newSignature
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
