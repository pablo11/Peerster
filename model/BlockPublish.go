package model

import (
    "encoding/binary"
    "crypto/sha256"
    "math/rand"
    "bytes"
)

type BlockPublish struct {
    Block Block
    HopLimit uint32
}

type Block struct {
    PrevHash [32]byte
    Nonce [32]byte
    Transactions []TxPublish
}

func (b *Block) Hash() (out [32]byte) {
    h := sha256.New()
    h.Write(b.PrevHash[:])
    h.Write(b.Nonce[:])
    binary.Write(h, binary.LittleEndian, uint32(len(b.Transactions)))
    for _, t := range b.Transactions {
        th := t.Hash()
        h.Write(th[:])
    }
    copy(out[:], h.Sum(nil))
    return
}

func (b *Block) Mine() [32]byte {
    var nonce [32]byte
    for {
        rand.Read(nonce[:])
        b.Nonce = nonce
        if b.IsValid() {
            return b.Hash()
        }
    }
}

func (b *Block) IsValid() bool {
    blockHash := b.Hash()
    return bytes.Equal(blockHash[0:2], []byte{0, 0})
}
