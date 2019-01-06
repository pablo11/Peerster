package model

import (
    "encoding/binary"
    "crypto/sha256"
    "encoding/hex"
)

type TxPublish struct {
    File File
    HopLimit uint32
}

func (t *TxPublish) Hash() (out [32]byte) {
    h := sha256.New()
    binary.Write(h, binary.LittleEndian, uint32(len(t.File.Name)))
    h.Write([]byte(t.File.Name))
    h.Write(t.File.MetafileHash)
    copy(out[:], h.Sum(nil))
    return
}

func (t *TxPublish) HashStr() string {
    hash := t.Hash()
    return hex.EncodeToString(hash[:])
}
