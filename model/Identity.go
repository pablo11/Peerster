package model

import (
	"crypto/rsa"
    "fmt"
    "encoding/hex"
	"crypto/sha256"
    "crypto/x509"
	"encoding/binary"
)

type Identity struct {
	Name      string
	PublicKey []byte
}

func (i *Identity) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(i.Name)))
	h.Write([]byte(i.Name))
	h.Write(i.PublicKey)
	copy(out[:], h.Sum(nil))
	return
}

func (i *Identity) HashStr() string {
    fixOut := i.Hash()
    out := fixOut[:]
	return hex.EncodeToString(out)
}

func (i *Identity) String() string {
	return "ID=" + i.Name + " (" + PublicKeyString(i.PublicKeyObj()) + ")"
}

func (i *Identity) PublicKeyObj() *rsa.PublicKey {
    publicKey, err := x509.ParsePKCS1PublicKey(i.PublicKey)

    if err != nil {
        fmt.Println(err)
    }

	return publicKey
}

func (i *Identity) Copy() Identity {
	newIdentity := Identity{
		Name:      i.Name,
		PublicKey: make([]byte, len(i.PublicKey)),
	}
    copy(newIdentity.PublicKey, i.PublicKey)

	return newIdentity
}

func (i *Identity) SetPublicKey(publicKey *rsa.PublicKey) {
     i.PublicKey = x509.MarshalPKCS1PublicKey(publicKey)
}
