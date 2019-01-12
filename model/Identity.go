package model

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
    "crypto/x509"
)

type Identity struct {
	Name      string
	PublicKey rsa.PublicKey
}

func (i *Identity) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(i.Name)))
	h.Write([]byte(i.Name))
	h.Write(i.bytePublicKey())
	copy(out[:], h.Sum(nil))
	return
}

func (i *Identity) String() string {
	return "ID=" + i.Name + " (" + PublicKeyString(&i.PublicKey) + ")"
}

func (i *Identity) bytePublicKey() []byte {
	bytePublicKey := x509.MarshalPKCS1PublicKey(&i.PublicKey)
	return bytePublicKey
}

func (i *Identity) Copy() Identity {
	newIdentity := Identity{
		Name:      i.Name,
		PublicKey: i.PublicKey,
	}
	return newIdentity
}
