package model

import (
    "fmt"
    "crypto/rsa"
    "crypto/rand"
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
