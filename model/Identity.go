package model

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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
	hash := i.Hash()
	return hex.EncodeToString(hash[:])
}

func (i *Identity) String() string {
	return "ID=" + i.Name + " (" + PublicKeyString(i.PublicKeyObj()) + ")"
}

func (i *Identity) PublicKeyObj() *rsa.PublicKey {
	publicKey, err := x509.ParsePKCS1PublicKey(i.PublicKey)
	if err != nil {
		fmt.Println(err)
		return nil
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
