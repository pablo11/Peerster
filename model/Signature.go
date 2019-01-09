package model

import (
    
)

type Signature struct {
    Origin string
    Signature []byte
}

func (s *Signature) IsValid(publicKey []byte) bool {

    // TODO: implement

    return false
}

func Sign(origin string, privateKey, bytes []byte) Signature {
    s := Signature{
        Origin: origin,
        Signature: make([]byte, 0),
    }

    return s
}
