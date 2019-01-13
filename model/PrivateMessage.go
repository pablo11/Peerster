package model

import (
    "strconv"
    "crypto/sha256"
    "encoding/binary"
)

type PrivateMessage struct {
    Origin string
    ID uint32
    Text string
    Destination string
    HopLimit uint32
    IsEncrypted bool
    Signature *Signature
}

func NewPrivateMessage(origin, text, dest string) *PrivateMessage {
    return &PrivateMessage{
        Origin: origin,
        ID: 0,
        Text: text,
        Destination: dest,
        HopLimit: 10,
        IsEncrypted: false,
    }
}

func (pm *PrivateMessage) String() string {
    hopLimitStr := strconv.FormatUint(uint64(pm.HopLimit), 10)
    return "PRIVATE origin " + pm.Origin + " hop-limit " + hopLimitStr + " contents " + pm.Text
}

func (pm *PrivateMessage) IntegrityHash() (out [32]byte) {
    sha_256 := sha256.New()
    sha_256.Write([]byte(pm.Origin))
    binary.Write(sha_256, binary.LittleEndian, pm.Origin)
    sha_256.Write([]byte(pm.Text))
    sha_256.Write([]byte(pm.Destination))
    copy(out[:], sha_256.Sum(nil))
    return
}
