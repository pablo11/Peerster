package model

import (
    "fmt"
    "encoding/binary"
    "crypto/sha256"
    "math/rand"
    //"bytes"
    "encoding/hex"
    "strings"
    //"github.com/pablo11/Peerster/util/debug"
)

const MINING_DIFFICULTY int = 5 // Each unit of mining difficulty corresponds to 4 bits

type BlockPublish struct {
    Block Block
    HopLimit uint32
}

type Block struct {
    PrevHash [32]byte
    Nonce [32]byte
    Transactions []TxPublish
}

func (b *Block) HashStr() string {
    hash := b.Hash()
    return hex.EncodeToString(hash[:])
}

func (b *Block) PrevHashStr() string {
    return hex.EncodeToString(b.PrevHash[:])
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

func (b *Block) Mine() {
    var nonce [32]byte
    for {
        rand.Read(nonce[:])
        b.Nonce = nonce
        if b.IsValid() {
            hash := b.Hash()
            fmt.Println("FOUND-BLOCK " + hex.EncodeToString(hash[:]))
            return
        }
    }
}

func (b *Block) IsValid() bool {
    blockHash := b.Hash()
    blockHashStr := hex.EncodeToString(blockHash[:])
    return blockHashStr[0:MINING_DIFFICULTY] == strings.Repeat("0", MINING_DIFFICULTY)
    //return bytes.Equal(blockHash[0:MINING_DIFFICULTY], make([]byte, MINING_DIFFICULTY))
}

func (b *Block) String() string {
    filenames := make([]string, len(b.Transactions))
    for i, trx := range b.Transactions {
        filenames[i] = trx.File.Name
    }
    blockHash := b.Hash()
    //debug.Debug("🍎 PRINTING BLOCKCHAIN BLOCK HASH " + hex.EncodeToString(blockHash[:]))
    return hex.EncodeToString(blockHash[:]) + ":" + b.PrevHashStr() + ":" + strings.Join(filenames, ",")
}
